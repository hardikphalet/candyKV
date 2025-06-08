package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hardikphalet/go-redis/internal/server"
)

// main is the entry point for the Redis server.
// It creates a new server instance and starts it.
// It also sets up signal handling for graceful shutdown.
func main() {
	srv := server.New(":6379")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting Redis server on port 6379...")
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	if err := srv.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
}
