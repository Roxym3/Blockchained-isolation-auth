package listener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Watermark struct {
	mu           sync.Mutex
	offset       uint64
	completed    map[uint64]bool
	dbCollection *mongo.Collection

	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewWatermark(startOffs uint64, col *mongo.Collection) *Watermark {
	wm := &Watermark{
		offset:       startOffs,
		completed:    make(map[uint64]bool),
		dbCollection: col,
		stopCh:       make(chan struct{}),
	}
	wm.wg.Add(1)
	go wm.flusher(3 * time.Second)
	return wm
}

func (wm *Watermark) Mark(block uint64) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.completed[block] = true

	for wm.completed[wm.offset] {
		delete(wm.completed, wm.offset)
		wm.offset++
	}
}

func (wm *Watermark) flusher(interval time.Duration) {
	defer wm.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastOffs uint64 = wm.offset

	for {
		select {
		case <-ticker.C:
			wm.sync(&lastOffs)
		case <-wm.stopCh:
			wm.sync(&lastOffs)
			log.Printf("flusher stopped")
			return
		}
	}
}

func (wm *Watermark) sync(lastOffs *uint64) error {
	wm.mu.Lock()
	currentOffs := wm.offset
	wm.mu.Unlock()

	if currentOffs == *lastOffs {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := wm.dbCollection.UpdateOne(
		ctx,
		bson.M{"_id": "sys_checkpoint"},
		bson.M{"$set": bson.M{"offset": currentOffs}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("update error at checkpoint %d: %w\n", currentOffs, err)
	} else {
		*lastOffs = currentOffs
	}

	return nil
}

func (wm *Watermark) End() {
	close(wm.stopCh)
	wm.wg.Wait()
}

func Load(ctx context.Context, col *mongo.Collection) (uint64, error) {
	var result struct {
		Offs uint64 `bson:"offset"`
	}

	err := col.FindOne(ctx, bson.M{"_id": "sys_checkpoint"}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		} else {
			return 0, fmt.Errorf("load error: %w", err)
		}
	}
	return result.Offs, nil
}
