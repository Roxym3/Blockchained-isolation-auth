package utils

import (
	"crypto/x509"
	"fmt"
	"listener/fabric"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"gopkg.in/yaml.v3"
)

func LoadConfig(configFile string) (*fabric.FabricConfig, error) {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config:%v", err)
	}

	var wrapper fabric.ConfigWrapper
	if err := yaml.Unmarshal(configData, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to resolve config data:%v", err)
	}
	return &wrapper.FConfig, nil
}

func LoadKey(keyPath string) (*x509.Certificate, error) {
	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key:%v", err)
	}

	privateKey, err := identity.CertificateFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to convert key:%v", err)
	}

	return privateKey, nil
}

func LoadCert(filename string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read Certificate:%v", err)
	}

	return identity.CertificateFromPEM(certPEM)
}
