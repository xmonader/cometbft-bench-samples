# Benchmark Instructions

This document provides detailed instructions on how to run benchmarks for the transaction batching system and generate performance reports.

## Overview

The benchmarking system allows you to measure the performance of different storage backends and batch sizes. It generates reports and charts to help you analyze the results.

## Prerequisites

Before running the benchmarks, make sure you have the following installed:

1. Go 1.16 or later
2. Gnuplot (for generating charts)

To install Gnuplot:

```bash
# Ubuntu/Debian
sudo apt-get install gnuplot

# macOS
brew install gnuplot

# Windows
# Download and install from http://www.gnuplot.info/download.html
```

## Running Benchmarks

### Basic Usage

To run benchmarks with default settings:

```bash
go run cmd/benchmark/main.go
```

This will run benchmarks for all storage backends with various batch sizes and save the results to the `benchmark_results` directory.

### Customizing Benchmarks

You can customize the benchmarks using command-line flags:

```bash
go run cmd/benchmark/main.go \
  --storage-backends=memory,badger \
  --batch-sizes=1,10,50,100,500,1000 \
  --num-accounts=1000 \
  --num-operations=10000 \
  --output-dir=my_benchmark_results
```

Available flags:

- `--storage-backends`: Comma-separated list of storage backends to benchmark (default: "memory,badger,sqlite,redis,tigerbeetle")
- `--batch-sizes`: Comma-separated list of batch sizes to benchmark (default: "1,10,50,100,500,1000")
- `--num-accounts`: Number of accounts to create (default: 1000)
- `--num-operations`: Number of operations to perform (default: 10000)
- `--output-dir`: Directory to store benchmark results (default: "benchmark_results")
- `--generate-charts`: Whether to generate charts from the results (default: true)

### Running Benchmarks for Specific Storage Backends

#### In-Memory Storage

```bash
# Create a directory for benchmark results
mkdir -p benchmark_results/memory

# Run benchmarks for in-memory storage with various batch sizes
go run cmd/benchmark/main.go \
  --storage-backends=memory \
  --batch-sizes=1,10,50,100,500,1000 \
  --num-accounts=1000 \
  --num-operations=10000 \
  --output-dir=benchmark_results/memory
```

#### BadgerDB Storage

```bash
# Create a directory for benchmark results
mkdir -p benchmark_results/badger

# Run benchmarks for BadgerDB storage with various batch sizes
go run cmd/benchmark/main.go \
  --storage-backends=badger \
  --batch-sizes=1,10,50,100,500,1000 \
  --num-accounts=1000 \
  --num-operations=10000 \
  --output-dir=benchmark_results/badger
```

## Benchmark Results

After running the benchmarks, the following files will be generated in the output directory:

1. `benchmark_results.csv`: Raw benchmark data in CSV format
2. `benchmark_report.md`: Markdown report with analysis of the benchmark results

If you enabled chart generation, a `benchmark_charts` directory will also be created with the following charts:

1. `tps_vs_batch_size.png`: Transactions per second vs. batch size
2. `ops_vs_batch_size.png`: Operations per second vs. batch size
3. `latency_vs_batch_size.png`: Latency vs. batch size
4. `tps_vs_storage.png`: Transactions per second vs. storage backend
5. `ops_vs_storage.png`: Operations per second vs. storage backend

## Generating Charts Manually

If you want to generate charts from existing benchmark results:

```bash
# Make the script executable
chmod +x cmd/benchmark/generate_charts.sh

# Generate charts from CSV file
./cmd/benchmark/generate_charts.sh benchmark_results/memory/benchmark_results.csv
```

This will create charts in the `benchmark_results/memory/benchmark_charts/` directory.

## Interpreting the Results

The benchmark results include the following metrics:

- **TPS (Transactions Per Second)**: The number of batched transactions processed per second
- **OPS (Operations Per Second)**: The number of individual operations processed per second
- **Latency**: The average time to process a single batched transaction

### Batch Size Impact

As the batch size increases, you should observe:

1. **TPS**: Slightly decreases due to larger transaction size
2. **OPS**: Increases dramatically, showing the effectiveness of batching
3. **Latency**: Increases slightly, but the increase is minimal compared to the throughput gain

### Storage Backend Impact

Different storage backends have different performance characteristics:

1. **Memory**: Highest performance, but no persistence
2. **BadgerDB**: Good balance of performance and persistence
3. **SQLite**: Lower performance, but ACID compliant
4. **Redis**: Good performance, especially in distributed scenarios
5. **TigerBeetle**: Excellent performance for financial applications

## Example Benchmark Report

Here's an example of what the benchmark report might look like:

```markdown
# Benchmark Report

Generated at: 2025-03-18 00:00:00

## Results

### Batch Size Impact

| Batch Size | TPS | OPS | Latency (ms) |
|------------|-----|-----|--------------|
| 1 | 9,500.00 | 9,500.00 | 10.00 |
| 10 | 9,400.00 | 94,000.00 | 11.00 |
| 50 | 9,200.00 | 460,000.00 | 12.00 |
| 100 | 9,000.00 | 900,000.00 | 13.00 |
| 500 | 8,500.00 | 4,250,000.00 | 15.00 |
| 1000 | 8,000.00 | 8,000,000.00 | 18.00 |

### Storage Backend Impact (Batch Size = 100)

| Storage Backend | TPS | OPS | Latency (ms) |
|-----------------|-----|-----|--------------|
| memory | 9,000.00 | 900,000.00 | 13.00 |
| badger | 8,500.00 | 850,000.00 | 15.00 |
| sqlite | 7,000.00 | 700,000.00 | 18.00 |
| redis | 8,000.00 | 800,000.00 | 16.00 |
| tigerbeetle | 8,800.00 | 880,000.00 | 14.00 |
```

## Conclusion

By running these benchmarks, you can determine the optimal batch size and storage backend for your specific use case. The results will help you make informed decisions about how to configure your system for maximum performance.
