package otp

// Go示例：生成16字节随机密钥（Base32编码）
import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

const (
	DefaultTimeStep     = 30 // 默认时间步长(秒)
	DefaultDigits       = 6  // 默认OTP位数
	DefaultSecretLength = 16 // 默认密钥长度(字节)
)

// GenerateSecret 生成随机Base32编码的TOTP密钥
func GenerateSecret() (string, error) {
	key := make([]byte, DefaultSecretLength)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate random key: %v", err)
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(key)
	secret = secret[:DefaultSecretLength]
	return secret, nil
}

// GenerateOTP 生成基于时间的一次性密码
func GenerateOTPWithTime(secret string, t time.Time) (string, error) {
	// 解码Base32密钥
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid secret: %v", err)
	}

	// 计算时间计数器
	counter := t.Unix() / DefaultTimeStep

	// 计算HMAC-SHA1
	h := hmac.New(sha1.New, key)
	err = binary.Write(h, binary.BigEndian, counter)
	if err != nil {
		return "", fmt.Errorf("failed to write counter: %v", err)
	}
	hash := h.Sum(nil)

	// 动态截断
	offset := hash[len(hash)-1] & 0x0F
	truncated := hash[offset : offset+4]

	// 转换为数字
	code := binary.BigEndian.Uint32(truncated) & 0x7FFFFFFF
	code %= uint32(math.Pow10(DefaultDigits))

	// 格式化为6位数字
	return fmt.Sprintf("%06d", code), nil
}

// GenerateOTPURI 生成TOTP URI，可用于生成QR码
func GenerateOTPURI(method, issuer, accountName, secret string) string {
	return fmt.Sprintf(
		"otpauth://%s/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		method,      // 协议类型
		issuer,      // 机构
		accountName, // 账户名称
		secret,
		issuer, DefaultDigits, DefaultTimeStep,
	)
}
