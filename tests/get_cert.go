package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

func createHTTPClient() *http.Client {
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return d.DialContext(ctx, "tcp4", addr)
		},
		TLSHandshakeTimeout:   60 * time.Second,
		ResponseHeaderTimeout: 180 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   10,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   300 * time.Second,
	}
}

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func main() {
	domains := []string{"deeppluse.dpdns.org"}
	email := os.Getenv("CF_API_EMAIL")

	if email == "" {
		log.Fatalf("请设置环境变量 CF_API_EMAIL")
	}

	apiKey := os.Getenv("CF_API_KEY")
	if apiKey == "" {
		log.Fatalf("请设置环境变量 CF_API_KEY")
	}

	useStaging := true
	caURL := lego.LEDirectoryStaging

	if len(os.Args) > 1 && os.Args[1] == "production" {
		useStaging = false
		caURL = lego.LEDirectoryProduction
	}

	log.Printf("=== Let's Encrypt DNS验证证书获取工具 ===")
	log.Printf("域名: %v", domains)
	log.Printf("邮箱: %s", email)
	log.Printf("DNS: Cloudflare")
	if useStaging {
		log.Printf("环境: 测试环境 (Staging)")
	} else {
		log.Printf("环境: 生产环境 (Production)")
	}
	log.Printf("")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成私钥失败: %v", err)
	}

	user := &MyUser{
		Email: email,
		key:   privateKey,
	}

	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = caURL
	legoConfig.HTTPClient = createHTTPClient()

	log.Printf("创建LEGO客户端...")

	client, err := lego.NewClient(legoConfig)
	if err != nil {
		log.Fatalf("创建LEGO客户端失败: %v", err)
	}

	log.Printf("配置Cloudflare DNS验证...")

	dnsProvider, err := cloudflare.NewDNSProvider()
	if err != nil {
		log.Fatalf("创建DNS提供商失败: %v", err)
	}

	err = client.Challenge.SetDNS01Provider(dnsProvider)
	if err != nil {
		log.Fatalf("设置DNS提供商失败: %v", err)
	}

	log.Printf("注册Let's Encrypt账户...")
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatalf("注册用户失败: %v", err)
	}
	user.Registration = reg

	log.Printf("请求证书...")
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatalf("获取证书失败: %v", err)
	}

	log.Printf("证书获取成功!")

	certPath := "cert.pem"
	keyPath := "key.pem"

	err = os.WriteFile(certPath, certificates.Certificate, 0644)
	if err != nil {
		log.Fatalf("保存证书失败: %v", err)
	}

	err = os.WriteFile(keyPath, certificates.PrivateKey, 0600)
	if err != nil {
		log.Fatalf("保存私钥失败: %v", err)
	}

	log.Printf("=== 证书已保存 ===")
	log.Printf("证书文件: %s", certPath)
	log.Printf("私钥文件: %s", keyPath)

	if useStaging {
		log.Printf("")
		log.Printf("=== 测试环境证书获取成功 ===")
		log.Printf("要获取正式证书，请运行: go run get_cert.go production")
	}
}
