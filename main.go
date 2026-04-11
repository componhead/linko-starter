package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"boot.dev/linko/internal/store"
)

var Logger log.Logger

func main() {
	Logger = *log.New(os.Stderr, "DEBUG: ", log.LstdFlags)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	httpPort := flag.Int("port", 8899, "port to listen on")
	dataDir := flag.String("data", "./data", "directory to store data")
	flag.Parse()
	// use a logger
	Logger.Printf("The Lisan al-Gaib arrived")
	status := run(ctx, cancel, *httpPort, *dataDir)
	cancel()
	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc, httpPort int, dataDir string) int {
	st, err := store.New(dataDir)
	if err != nil {
		Logger.Print("failed to create store: %v", err)
		return 1
	}
	s := newServer(*st, httpPort, cancel)
	var serverErr error
	go func() {
		serverErr = s.start()
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.shutdown(shutdownCtx); err != nil {
		Logger.Printf("failed to shutdown server: %v", err)
		return 1
	}
	if serverErr != nil {
		Logger.Print("server error: %v", serverErr)
		return 1
	}
	return 0
}
