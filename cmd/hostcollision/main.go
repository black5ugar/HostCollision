package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"hostCollision/internal/app"
	"hostCollision/internal/banner"
	"hostCollision/internal/config"
)

func main() {
	banner.Print()
	cfg, err := config.FromFlags()
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, cfg); err != nil {
		log.Printf("scan failed: %v", err)
		os.Exit(1)
	}
}
