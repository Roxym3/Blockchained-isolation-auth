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
	"os"
	"os/signal"
	"syscall"
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

	mongoClient, err := database.InitMongoDB(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer database.Disconnect(mongoClient)

	ticketColl := mongoClient.Database(cfg.Database.DBName).Collection(cfg.Database.Collection)
	log.Println("successfully connect to mongodb")

	taskPool := pool.NewPool(100)
	eventListener, err := listener.NewFabricListener(ticketColl, taskPool, &cfg.Fabric)
	if err != nil {
		return fmt.Errorf("create fabric listener: %w", err)
	}

	m := fabric.NewManger(cfg.Fabric.Peers, &cfg.Fabric, eventListener)
	go m.Start(ctx)

	// Start API Server
	apiServer := api.NewServer(ticketColl)
	go func() {
		if err := apiServer.Start("8081"); err != nil {
			log.Printf("API Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("cleaning up")

	taskPool.Wait()
	eventListener.Wm.End()

	log.Printf("exited")
	return nil
}
