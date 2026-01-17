package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ElshadHu/verdis/internal/server"
)

func main() {
	cfg := server.NewDefaultConfig().WithMaxConnections(1000)

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	fmt.Printf("Verdis server listening on %s\n", cfg.Address())
	if err := srv.Start(ctx); err != nil {
		log.Fatal("Server error:", err)
	}
}
