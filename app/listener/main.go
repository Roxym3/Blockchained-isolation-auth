package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID           = "PrivacyMSP"
	channelName     = "cross-domain-channel"
	chaincodeName   = "authcc"
	cryptoPath      = "/home/roxy3/blockchained_isolation_auth/fabric/test-network/organizations/peerOrganizations/privacy.com"
	certPath        = cryptoPath + "/users/Admin@privacy.com/msp/signcerts/Admin@privacy.com-cert.pem"
	keyDir          = cryptoPath + "/users/Admin@privacy.com/msp/keystore"
	tlsRootCertPath = cryptoPath + "/peers/peer0.privacy.com/tls/ca.crt"

	mongoURI = "mongodb://privacy_admin:PrivacyPassword@localhost:27017"
	dbName   = "privacy_db"
	colName  = "cross_domain_tickets"
)

type PeerNode struct {
	Endpoint string
	HostName string
}

var privacyPeers = []PeerNode{
	{Endpoint: "localhost:9051", HostName: "peer0.privacy.com"},
	{Endpoint: "localhost:9061", HostName: "peer1.privacy.com"},
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("unable to connect to mongodb:%v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	ticketCollection := mongoClient.Database(dbName).Collection(colName)
	log.Println("connect to mongodb in privacy domain successfully!")

	for {
		for _, node := range privacyPeers {
			log.Printf("trying to connect to [%s](%s)...", node.HostName, node.Endpoint)

			gw, err := connectToFabric(node)
			if err != nil {
				log.Printf("failed to connect to %s,%v,switching for next", node.HostName, err)
			}

			log.Printf("connected to %s! Begin to listen for event", node.HostName)

			listenToEvents(gw, ticketCollection)
			gw.Close()
			log.Printf("lost connection to %s", node.HostName)
		}

		log.Println("Peers in Privacy domain are unavailable! Retry after 5s")
		time.Sleep(5 * time.Second)
	}
}

func listenToEvents(gw *client.Gateway, collection *mongo.Collection) {
	network := gw.GetNetwork(channelName)
	ctxEvents, cancelEvents := context.WithCancel(context.Background())
	defer cancelEvents()

	events, err := network.ChaincodeEvents(ctxEvents, chaincodeName)
	if err != nil {
		log.Printf("failed to subscribe stream events:%v", err)
		return
	}

	for event := range events {
		if event.EventName == "TicketIssuedEvent" {
			var ticketData map[string]interface{}
			if err := json.Unmarshal(event.Payload, &ticketData); err != nil {
				continue
			}

			if target, ok := ticketData["target_domain"].(string); ok && target == mspID {
				ticketData["_id"] = ticketData["ticket_id"]
				ticketData["received_at"] = time.Now().Format(time.RFC3339)

				_, err := collection.UpdateOne(context.Background(),
					bson.M{"_id": ticketData["_id"]},
					bson.M{"$set": ticketData},
					options.Update().SetUpsert(true),
				)

				if err != nil {
					log.Printf("failed to write data:%v", err)
				} else {
					log.Printf("successfully catch and save ticket[%s]", ticketData["ticket_id"])
				}
			}
		}
	}
}

func connectToFabric(node PeerNode) (*client.Gateway, error) {
	cert, err := loadCertificate(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate:%v", err)
	}

	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate object:%v", err)
	}

	sign, err := newSigner()
	if err != nil {
		return nil, fmt.Errorf("failed to create signer:%v", err)
	}

	tlsCert, err := loadCertificate(tlsRootCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tls certificate:%v", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCert)

	transportCredentials := credentials.NewClientTLSFromCert(certPool, node.HostName)
	grpcConn, err := grpc.NewClient(node.Endpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection:%v", err)
	}

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("failed to initalize grpc connection:%v", err)
	}
	return gw, nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate:%v", err)
	}

	return identity.CertificateFromPEM(certificatePEM)
}

func newSigner() (identity.Sign, error) {
	files, err := os.ReadDir(keyDir)
	if err != nil {
		return nil, err
	}

	privateKeyPEM, err := os.ReadFile(filepath.Join(keyDir, files[0].Name()))
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySign(privateKey)
}
