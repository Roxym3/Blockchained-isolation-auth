package listener

import (
	"context"
	"encoding/json"
	"fmt"
	config "listener/pkg/config"
	pool "listener/pkg/pool"
	"log"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventListener interface {
	Listen(ctx context.Context, gw *client.Gateway)
}

type FabricListener struct {
	Collection *mongo.Collection
	Pool       *pool.TaskPool
	Config     *config.Fabric
	Wm         *Watermark
}

func NewFabricListener(col *mongo.Collection, pool *pool.TaskPool, cfg *config.Fabric) (*FabricListener, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	startOffs, err := Load(ctx, col)
	if err != nil {
		return nil, fmt.Errorf("Error: %w\n", err)
	}

	log.Printf("recover checkpoint: %d\n", startOffs)
	wm := NewWatermark(startOffs, col)
	return &FabricListener{
		Collection: col,
		Pool:       pool,
		Config:     cfg,
		Wm:         wm,
	}, nil
}

func (l *FabricListener) Listen(ctx context.Context, gw *client.Gateway) {
	network := gw.GetNetwork(l.Config.ChannelName)
	ctxEvents, cancelEvents := context.WithCancel(ctx)
	defer cancelEvents()

	l.Wm.mu.Lock()
	startBlock := l.Wm.offset
	l.Wm.mu.Unlock()

	log.Printf("start to listen, current height: %d", startBlock)

	events, err := network.ChaincodeEvents(ctxEvents, l.Config.ChaincodeName, client.WithStartBlock(startBlock))
	if err != nil {
		log.Printf("failed to subscribe stream events:%v", err)
		return
	}

	for event := range events {
		eventName := event.EventName
		payload := event.Payload
		blockNum := event.BlockNumber

		l.Pool.AddTask(pool.NewTask(func() error {
			err := l.handleEvents(ctx, payload, eventName)
			if err == nil {
				l.Wm.Mark(blockNum + 1)
			}
			return err
		}))
	}
}

func (l *FabricListener) handleEvents(ctx context.Context, payload []byte, eventName string) error {
	if eventName != "TicketIssuedEvent" && eventName != "TicketRevokedEvent" {
		return nil
	}

	var ticketData map[string]interface{}
	if err := json.Unmarshal(payload, &ticketData); err != nil {
		log.Printf("failed to resolve ticket data %s: %v", eventName, err)
		return nil
	}

	ticketID, ok := ticketData["ticket_id"].(string)
	if !ok || ticketID == "" {
		log.Printf("Ignored event %s: missing or empty ticket_id\n", eventName)
		return nil
	}

	if eventName == "TicketIssuedEvent" {
		target, ok := ticketData["target_domain"].(string)
		if !ok || target != l.Config.MspID {
			return nil
		}
	}

	ticketData["_id"] = ticketData["ticket_id"]
	ticketData["sys_updated_at"] = time.Now().Format(time.RFC3339)
	ticketData["sys_last_event"] = eventName

	return l.upsertWithRetry(ctx, ticketID, ticketData)
}

func (l *FabricListener) upsertWithRetry(ctx context.Context, ticketID string, ticketData map[string]any) error {
	retryTime := 0
	for {
		if err := l.execUpsert(ticketID, ticketData); err == nil {
			log.Printf("catch and write ticket data successfully[%s]", ticketData["ticket_id"])
			return nil
		}

		retryTime++
		log.Printf("failed to write into mongodb:%s,retried %d times", ticketID, retryTime)

		sleepTime := time.Duration(retryTime) * time.Second
		if sleepTime > 10*time.Second {
			sleepTime = 18 * time.Second
		}
		select {
		case <-ctx.Done():
			log.Printf("stop retrying...")
			return nil
		case <-time.After(sleepTime):
		}
	}
}

func (l *FabricListener) execUpsert(ticketID string, ticketData map[string]any) error {
	dbCtx, dbCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer dbCancel()
	_, err := l.Collection.UpdateOne(
		dbCtx,
		bson.M{"_id": ticketData["_id"]},
		bson.M{"$set": ticketData},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("Error upsert: %w", err)
	}
	return nil
}
