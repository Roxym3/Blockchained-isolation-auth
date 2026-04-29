package fabric

import (
	"context"
	"listener/pkg/config"
	"listener/pkg/listener"
	"log"
	"time"
)

type Manager struct {
	peers    []config.PeerNode
	cfg      *config.Fabric
	listener *listener.FabricListener
}

func NewManger(peers []config.PeerNode, cfg *config.Fabric, listener *listener.FabricListener) *Manager {
	return &Manager{
		peers:    peers,
		cfg:      cfg,
		listener: listener,
	}
}

func (m *Manager) Start(ctx context.Context) {
OuterLoop:
	for {
		for _, node := range m.peers {
			select {
			case <-ctx.Done():
				break
			default:
			}
			log.Printf("connecting to %s: %s\n", node.HostName, node.Endpoint)
			gw, err := ConnectToFabric(node, m.cfg)
			if err != nil {
				log.Printf("connect error %s: %v, switch for next", node.HostName, err)
				continue
			}
			log.Printf("successfully connect to %s: %s", node.HostName, node.Endpoint)
			m.listener.Listen(ctx, gw)
			gw.Close()
		}
		select {
		case <-ctx.Done():
			break OuterLoop
		case <-time.After(5 * time.Second):
		}
	}
}
