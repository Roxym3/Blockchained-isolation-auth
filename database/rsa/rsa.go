package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log"
	"os"
	"time"
)

type UserProfile struct {
	PublicKey    string       `json:"public_key"`
	PersonalInfo PersonalInfo `json:"personal_info"`
	CreatedTime  time.Time    `json:"created_time"`
}

type PersonalInfo struct {
	Name        string `json:"name"`
	IDCardNum   string `json:"id_card_num"`
	CreditScore int    `json:"credit_score"`
}

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成密钥失败:%v", err)
	}

	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:  "RSA Private Key",
		Bytes: privDER,
	}
	privPEM, _ := os.Create("privKey.pem")
	defer privPEM.Close()
	pem.Encode(privPEM, &privBlock)

	pubilcKey := &privateKey.PublicKey
	pubDER := x509.MarshalPKCS1PublicKey(pubilcKey)

	pubBlock := pem.Block{
		Type:  "RSA Public Key",
		Bytes: pubDER,
	}
	pubPEM, _ := os.Create("pubKey.pem")
	defer pubPEM.Close()
	pem.Encode(pubPEM, &pubBlock)

	pubPEMString := string(pem.EncodeToMemory(&pubBlock))
	profile := UserProfile{
		PublicKey: pubPEMString,
		PersonalInfo: PersonalInfo{
			Name:        "张三",
			IDCardNum:   "370102199001011234",
			CreditScore: 750,
		},
		CreatedTime: time.Now(),
	}

	profileJSON, _ := json.MarshalIndent(profile, "", "")
	err = os.WriteFile("data.json", profileJSON, 0644)
	if err != nil {
		log.Fatalf("输出JSON失败:%v", err)
	}
}
