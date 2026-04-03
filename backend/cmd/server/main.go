package main

import (
	"backend/internal/config"
	"backend/internal/infra/db"
	"backend/internal/infra/logger"
	"backend/internal/router"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	path = "./configs/config.yaml"
)

func main() {
	// init config
	cfg := config.LoadConfig(path)

	// init logger
	logger.InitLogger(cfg)

	// init db
	db.InitDB(cfg)
	// db.AutoMigrate()

	// init server
	r := router.SetupRouter(cfg.Server.Mode)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server run failed, err: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Println("Receive shutdown signal: ", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed, err: %v", err)
	}

	signal.Stop(quit)
	close(quit)

	log.Println("Server shutdown")
}
