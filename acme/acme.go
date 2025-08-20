package acme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/acme"
)

// ChallengeType 定义 ACME 验证类型
type ChallengeType uint8

const (
	ChallengeHTTP01 ChallengeType = iota // HTTP-01 验证
	ChallengeDNS01                       // DNS-01 验证
)

// ChallengeTask 用于存储每个验证任务的信息
type ChallengeTask struct {
	Type      ChallengeType   // 验证类型
	Token     string          // 验证令牌
	KeyAuth   string          // 验证密钥
	Domain    string          // 域名
	Status    bool            // 验证状态
	challenge *acme.Challenge // ACME 挑战信息
	authURI   string          // 授权 URI
}

// Auth 用于存储 ACME 认证信息
type Auth struct {
	Domain  []string        // 申请域名列表
	Email   string          // 申请邮箱
	AcmeURL string          // ACME 服务器地址 默认用 Let's Encrypt
	Ctx     context.Context // 上下文用来控制请求超时
	order   *acme.Order     // 订单信息
	client  *acme.Client    // ACME 客户端
}

// 申请证书初始准备
// 返回验证信息并进行处理
func (auth *Auth) AuthorizeOrder() ([]ChallengeTask, error) {
	// 设置 ACME 服务器地址
	if auth.AcmeURL == "" {
		auth.AcmeURL = acme.LetsEncryptURL
	}
	// 设置上下文
	if auth.Ctx == nil {
		auth.Ctx = context.Background()
	}

	// 创建 ACME 客户端
	accountKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	auth.client = &acme.Client{
		Key:          accountKey,
		DirectoryURL: auth.AcmeURL,
	}

	// 注册账户
	account := &acme.Account{Contact: []string{"mailto:" + auth.Email}}
	_, err = auth.client.Register(auth.Ctx, account, acme.AcceptTOS)
	if err != nil && err != acme.ErrAccountAlreadyExists {
		return nil, err
	}

	// 下订单

	auth.order, err = auth.client.AuthorizeOrder(auth.Ctx, acme.DomainIDs(auth.Domain...))
	if err != nil {
		return nil, err
	}

	tasks := make([]ChallengeTask, 0)
	for _, authzURL := range auth.order.AuthzURLs {
		authz, _ := auth.client.GetAuthorization(auth.Ctx, authzURL)
		for _, ch := range authz.Challenges {
			switch ch.Type {
			case "http-01":
				TokenResponse, err := auth.client.HTTP01ChallengeResponse(ch.Token)
				if err != nil {
					return nil, err
				}
				tasks = append(tasks, ChallengeTask{
					Type:      ChallengeHTTP01,
					Token:     ch.Token,
					KeyAuth:   TokenResponse,
					Domain:    authz.Identifier.Value,
					challenge: ch,
					authURI:   authz.URI,
				})
			case "dns-01":
				keyAuth, err := auth.client.DNS01ChallengeRecord(ch.Token)
				if err != nil {
					return nil, err
				}

				tasks = append(tasks, ChallengeTask{
					Type:      ChallengeDNS01,
					Token:     ch.Token,
					KeyAuth:   keyAuth,
					Domain:    authz.Identifier.Value,
					challenge: ch,
					authURI:   authz.URI,
				})
			}
		}
	}
	return tasks, nil
}

type Certificate struct {
	Domain []string // 域名
	Cert   []byte   // 证书内容
	Key    []byte   // 私钥内容
}

// 验证生成证书
// http解析或者dns解析完成后则开始生成证书
func (auth *Auth) CreateOrderCert(tasks []ChallengeTask) (Certificate, error) {
	domain := []string{}
	for _, task := range tasks {
		if task.Status {
			// 接受挑战
			if _, err := auth.client.Accept(auth.Ctx, task.challenge); err != nil {
				return Certificate{}, err
			}
			// 等待验证
			if _, err := auth.client.WaitAuthorization(auth.Ctx, task.authURI); err != nil {
				return Certificate{}, err
			}
			domain = append(domain, task.Domain)
		}
	}

	// 生成域名私钥
	domainKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	// 生成 CSR
	csrDER, _ := x509.CreateCertificateRequest(
		rand.Reader,
		&x509.CertificateRequest{DNSNames: domain},
		domainKey,
	)
	// 获取证书
	certDER, _, err := auth.client.CreateOrderCert(auth.Ctx, auth.order.FinalizeURL, csrDER, true)
	if err != nil {
		return Certificate{}, err
	}

	// 保存完整证书链
	var certPEM []byte
	for _, der := range certDER {
		certPEM = append(certPEM, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})...)
	}
	// 保存私钥 (PKCS#8 格式更通用)
	privateKeyByte, err := x509.MarshalPKCS8PrivateKey(domainKey)
	if err != nil {
		return Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyByte})

	cert := Certificate{
		Domain: domain,
		Cert:   certPEM,
		Key:    keyPEM, // 私钥将在下面保存
	}

	return cert, nil
}
