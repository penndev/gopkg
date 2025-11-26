package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/penndev/gopkg/acme"
)

// go run acme/cmd/example.go
func main() {
	auth := &acme.Auth{
		Domain: []string{"example.com", "*.example.com"},
		Email:  "your@email.com",
	}
	tasks, err := auth.AuthorizeOrder()
	if err != nil {
		panic(err)
	}
	log.Println("验证方式DNS HTTP任意认证一种即可:", len(tasks))
	for i, task := range tasks {
		switch task.Type {
		case acme.ChallengeHTTP01:
			fmt.Printf("HTTP TXT: http://%s/.well-known/acme-challenge/%s\n", task.Domain, task.Token)
			fmt.Printf("Resp: %s\n", task.KeyAuth)
		case acme.ChallengeDNS01:
			fmt.Printf("DNS Record:(通配符 %t) _acme-challenge.%s IN TXT %s\n", task.Wildcard, task.Domain, task.KeyAuth)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("当前认证方式已完成？(y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if strings.EqualFold(input, "y") {
			tasks[i].Status = true
		}

	}

	cert, err := auth.CreateOrderCert(tasks)
	if err != nil {
		panic(err)
	}
	log.Println("证书生成成功")
	log.Printf("域名: %v\n", cert.Domain)
	log.Printf("证书内容:\n%s\n", string(cert.Cert))
	log.Printf("私钥内容:\n%s\n", string(cert.Key))
}
