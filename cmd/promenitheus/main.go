package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Avinash7390/Promenitheus/pkg/config"
	"github.com/Avinash7390/Promenitheus/pkg/metrics"
	"github.com/Avinash7390/Promenitheus/pkg/scraper"
	"github.com/Avinash7390/Promenitheus/pkg/storage"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	port := flag.Int("port", 9090, "Port to expose metrics on")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	registry := metrics.NewMetricRegistry()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scr := scraper.NewScraper(cfg, registry)
	scr.Start(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
	}()

	server := storage.NewServer(registry, *port)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
