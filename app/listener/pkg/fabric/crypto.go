package fabric

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

func LoadFabricKey(keyPath string) (crypto.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key:%v", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to convert key:%v", err)
	}

	return privateKey, nil
}

func LoadFabricCert(filename string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read Certificate:%v", err)
	}

	return identity.CertificateFromPEM(certPEM)
}
