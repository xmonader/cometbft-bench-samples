package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	abciserver "github.com/cometbft/cometbft/abci/server"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/xmonader/test_batched_tx_tendermint/app"
)

// Configuration represents the application configuration
type Configuration struct {
	StateFile  string `json:"state_file"`
	PrivateKey string `json:"private_key"`
}

var (
	configFile = flag.String("config", "config.json", "Path to the configuration file")
	abciAddr   = flag.String("abci", "tcp://0.0.0.0:26658", "ABCI server address")
)

func main() {
	// Parse command-line flags
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create logger
	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))

	// Create application
	application := app.NewApplication(config.StateFile, logger)

	// Create ABCI server
	srv, err := abciserver.NewServer(*abciAddr, "socket", application)
	if err != nil {
		log.Fatalf("Failed to create ABCI server: %v", err)
	}

	// Set logger
	srv.SetLogger(logger)

	// Start the server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start ABCI server: %v", err)
	}
	defer srv.Stop()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Exit gracefully
	fmt.Println("Shutting down...")
}

// loadConfig loads the configuration from a file
func loadConfig(path string) (*Configuration, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer file.Close()

	// Decode the JSON
	var config Configuration
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	return &config, nil
}
