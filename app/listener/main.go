package main

import (
	"context"
	"fmt"
	"listener/pkg/api"
	"listener/pkg/config"
	"listener/pkg/database"
	"listener/pkg/fabric"
	"listener/pkg/listener"
	"listener/pkg/pool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

const MaxRetries = 16

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL :%v", err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load("./config.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	mongoClient, ticketColl, err := setupDB(cfg)
	if err != nil {
		return err
	}
	defer database.Disconnect(mongoClient)

	taskPool := pool.NewPool(100)

	eventListener, err := startFabric(ctx, ticketColl, taskPool, cfg)
	if err != nil {
		return err
	}

	apiServer := startServer(cfg, ticketColl)
	<-ctx.Done()
	shutdown(apiServer, taskPool, eventListener)
	log.Printf("exited")
	return nil
}

func setupDB(cfg *config.Config) (*mongo.Client, *mongo.Collection, error) {
	mongoClient, err := database.InitMongoDB(&cfg.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("database error: %w", err)
	}

	ticketColl := mongoClient.Database(cfg.Database.DBName).Collection(cfg.Database.Collection)
	log.Println("successfully connect to mongodb")

	return mongoClient, ticketColl, nil
}

func startFabric(ctx context.Context, ticketColl *mongo.Collection, taskPool *pool.TaskPool, cfg *config.Config) (*listener.FabricListener, error) {
	eventListener, err := listener.NewFabricListener(ticketColl, taskPool, &cfg.Fabric)
	if err != nil {
		return nil, fmt.Errorf("create fabric listener: %w", err)
	}

	m := fabric.NewManger(cfg.Fabric.Peers, &cfg.Fabric, eventListener)
	go m.Start(ctx)
	return eventListener, nil
}

func startServer(cfg *config.Config, ticketColl *mongo.Collection) *api.Server {
	apiServer := api.NewServer(ticketColl, cfg.Fabric.MspID)
	go func() {
		if err := apiServer.Start("8081"); err != nil && err != http.ErrServerClosed {
			log.Printf("API Server error: %v", err)
		}
	}()

	return apiServer
}

func shutdown(apiServer *api.Server, taskPool *pool.TaskPool, eventListener *listener.FabricListener) error {
	log.Println("cleaning up")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	taskPool.Wait()
	eventListener.Wm.End()
	return nil
}
