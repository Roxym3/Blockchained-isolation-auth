package fabric

import (
	"crypto/x509"
	"fmt"
	"listener/pkg/config"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func ConnectToFabric(peerNode config.PeerNode, nodeConfig *config.Fabric) (*client.Gateway, error) {
	cert, err := LoadFabricCert(nodeConfig.CertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load peer certificate:%v", err)
	}

	id, err := identity.NewX509Identity(nodeConfig.MspID, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate object:%v", err)
	}

	signer, err := newSigner(nodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer:%v", err)
	}

	tlsCert, err := LoadFabricCert(nodeConfig.TLSRootCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tls certificate:%v", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCert)

	transportCredentials := credentials.NewClientTLSFromCert(certPool, peerNode.HostName)
	grpcConn, err := grpc.NewClient(peerNode.Endpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection:%v", err)
	}

	gw, err := client.Connect(id,
		client.WithSign(signer),
		client.WithClientConnection(grpcConn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithSubmitTimeout(1*time.Minute))
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("failed to initalize grpc connection: %v", err)
	}

	return gw, nil
}

func newSigner(fabricConfig *config.Fabric) (identity.Sign, error) {
	files, err := os.ReadDir(fabricConfig.KeyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read key directory:%v", err)
	}

	keyPath := filepath.Join(fabricConfig.KeyDir, files[0].Name())

	privateKey, err := LoadFabricKey(keyPath)
	return identity.NewPrivateKeySign(privateKey)
}
