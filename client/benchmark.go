package client

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// BenchmarkConfig represents the configuration for a benchmark
type BenchmarkConfig struct {
	NumTransactions  int   // Number of transactions to generate
	BatchSizes       []int // Batch sizes to test
	NumUsers         int   // Number of users to simulate
	MaxAmount        int   // Maximum amount per operation
	RandomRecipients bool  // Whether to use random recipients
}

// DefaultBenchmarkConfig returns a default benchmark configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		NumTransactions:  1000,
		BatchSizes:       []int{1, 10, 50, 100, 500},
		NumUsers:         10,
		MaxAmount:        100,
		RandomRecipients: true,
	}
}

// BenchmarkResult represents the result of a benchmark
type BenchmarkResult struct {
	BatchSize             int           // Batch size
	TotalOperations       int           // Total number of operations
	TotalTransactions     int           // Total number of transactions
	TotalTime             time.Duration // Total time taken
	OperationsPerSecond   float64       // Operations per second
	TransactionsPerSecond float64       // Transactions per second
}

// String returns a string representation of the benchmark result
func (r *BenchmarkResult) String() string {
	return fmt.Sprintf(
		"Batch Size: %d, Operations: %d, Transactions: %d, Time: %v, Ops/s: %.2f, Tx/s: %.2f",
		r.BatchSize, r.TotalOperations, r.TotalTransactions, r.TotalTime,
		r.OperationsPerSecond, r.TransactionsPerSecond,
	)
}

// Benchmark runs a benchmark with the given configuration
func Benchmark(config *BenchmarkConfig, client *Client, processFunc func(*types.Transaction) error) ([]*BenchmarkResult, error) {
	// Initialize random number generator
	rand.Seed(time.Now().UnixNano())

	// Create results slice
	results := make([]*BenchmarkResult, len(config.BatchSizes))

	// Run benchmark for each batch size
	for i, batchSize := range config.BatchSizes {
		// Calculate number of transactions needed
		numTransactions := config.NumTransactions / batchSize
		if numTransactions == 0 {
			numTransactions = 1
		}

		// Start timer
		startTime := time.Now()

		// Generate and process transactions
		totalOperations := 0
		for j := 0; j < numTransactions; j++ {
			// Generate operations for this batch
			operations := make([]types.Operation, 0, batchSize)
			for k := 0; k < batchSize; k++ {
				// Generate random recipient and amount
				var recipient int
				if config.RandomRecipients {
					recipient = rand.Intn(config.NumUsers) + 1
					// Ensure recipient is not the sender
					if recipient == client.GetUserID() {
						recipient = (recipient % config.NumUsers) + 1
					}
				} else {
					// Use a fixed recipient (user 2) for simplicity
					recipient = 2
				}
				amount := rand.Intn(config.MaxAmount) + 1

				// Create operation
				operation := client.CreateTransferOperation(recipient, amount)
				operations = append(operations, operation)
			}

			// Create transaction
			tx, err := client.CreateTransaction(operations)
			if err != nil {
				return nil, fmt.Errorf("failed to create transaction: %w", err)
			}

			// Process transaction
			if err := processFunc(tx); err != nil {
				return nil, fmt.Errorf("failed to process transaction: %w", err)
			}

			totalOperations += len(operations)
		}

		// Calculate elapsed time
		elapsedTime := time.Since(startTime)

		// Calculate operations per second
		opsPerSecond := float64(totalOperations) / elapsedTime.Seconds()
		txPerSecond := float64(numTransactions) / elapsedTime.Seconds()

		// Create result
		results[i] = &BenchmarkResult{
			BatchSize:             batchSize,
			TotalOperations:       totalOperations,
			TotalTransactions:     numTransactions,
			TotalTime:             elapsedTime,
			OperationsPerSecond:   opsPerSecond,
			TransactionsPerSecond: txPerSecond,
		}
	}

	return results, nil
}
