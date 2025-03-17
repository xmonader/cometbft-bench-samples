# Transaction Batching Benchmark System

This document explains how to use the benchmark system with the provided Makefile to easily run benchmarks for different storage backends.

## Prerequisites

Before running the benchmarks, make sure you have the following installed:

1. **Go** (1.16 or later)
2. **Docker** (for Redis and TigerBeetle services)
3. **Gnuplot** (for generating charts)

You can check if you have all prerequisites installed by running:

```bash
make check-prerequisites
```

## Quick Start

To run all benchmarks with the default configuration:

```bash
make
```

This will:
1. Set up Redis and TigerBeetle services using Docker
2. Run benchmarks for all storage backends
3. Generate reports and charts

Alternatively, you can use Docker Compose to set up the services:

```bash
# Set up services using Docker Compose
make setup-services-compose

# Run all benchmarks
make run-all-benchmarks
```

## Available Commands

### Main Commands

- `make all` - Set up services and run all benchmarks
- `make setup` - Set up Redis and TigerBeetle services
- `make run-all-benchmarks` - Run benchmarks for all storage backends
- `make clean` - Remove all benchmark results and stop services
- `make regenerate` - Clean and rerun all benchmarks
- `make help` - Show available commands and configuration

### Individual Benchmark Commands

Run benchmarks for specific storage backends:

- `make benchmark-memory` - Run benchmarks for Memory storage
- `make benchmark-badger` - Run benchmarks for BadgerDB storage
- `make benchmark-sqlite` - Run benchmarks for SQLite storage
- `make benchmark-redis` - Run benchmarks for Redis storage
- `make benchmark-tigerbeetle` - Run benchmarks for TigerBeetle storage

### Service Management Commands

- `make setup-redis` - Set up Redis service
- `make setup-tigerbeetle` - Set up TigerBeetle service
- `make stop-redis` - Stop Redis service
- `make stop-tigerbeetle` - Stop TigerBeetle service
- `make setup-services-compose` - Set up all services using Docker Compose
- `make stop-services-compose` - Stop all services using Docker Compose

## Configuration

The Makefile uses the following default configuration:

- **NUM_OPERATIONS**: 1000 (number of operations to perform)
- **NUM_ACCOUNTS**: 1000 (number of accounts to create)
- **BATCH_SIZES**: 1,10,50,100,500,1000 (batch sizes to benchmark)
- **REDIS_PORT**: 6379 (port for Redis service)
- **TIGERBEETLE_PORT**: 3000 (port for TigerBeetle service)

You can modify these values in the Makefile if needed.

## Benchmark Results

After running the benchmarks, results will be available in the `benchmark_results` directory:

```
benchmark_results/
├── memory/
│   ├── benchmark_report.md
│   ├── benchmark_results.csv
│   └── benchmark_charts/
├── badger/
│   ├── benchmark_report.md
│   ├── benchmark_results.csv
│   └── benchmark_charts/
├── sqlite/
│   ├── benchmark_report.md
│   ├── benchmark_results.csv
│   └── benchmark_charts/
├── redis/
│   ├── benchmark_report.md
│   ├── benchmark_results.csv
│   └── benchmark_charts/
└── tigerbeetle/
    ├── benchmark_report.md
    ├── benchmark_results.csv
    └── benchmark_charts/
```

Each directory contains:
- A CSV file with raw benchmark data
- A Markdown report with analysis
- Charts showing performance metrics

## Simple Explanations

The benchmark system also includes simple explanations for each storage backend:

- [SQLite Explanation](benchmark_results/sqlite/simple_explanation.md)
- [Redis Explanation](benchmark_results/redis/simple_explanation.md)
- [TigerBeetle Explanation](benchmark_results/tigerbeetle/simple_explanation.md)
- [Storage Comparison for Beginners](benchmark_results/storage_comparison_for_dummies.md)
- [Visual Comparison](benchmark_results/visual_comparison.md)
- [Final Summary](benchmark_results/final_summary.md)

## Troubleshooting

### Docker Issues

If you encounter issues with Docker containers:

1. Check if containers are running: `docker ps`
2. Check container logs: `docker logs tx-benchmark-redis` or `docker logs tx-benchmark-tigerbeetle`
3. Try stopping and restarting the containers: `make stop-redis && make setup-redis`

### Docker Compose Issues

If you're using Docker Compose and encounter issues:

1. Check if containers are running: `docker-compose ps`
2. Check container logs: `docker-compose logs redis` or `docker-compose logs tigerbeetle`
3. Try stopping and restarting the services: `make stop-services-compose && make setup-services-compose`

### Benchmark Issues

If benchmarks fail:

1. Check that all prerequisites are installed: `make check-prerequisites`
2. Ensure Redis and TigerBeetle services are running: `make setup`
3. Try running individual benchmarks to isolate the issue: `make benchmark-memory`

## Advanced Usage

### Running with Different Parameters

To run with different parameters, you can modify the Makefile variables:

```bash
# Edit the Makefile to change NUM_OPERATIONS, NUM_ACCOUNTS, or BATCH_SIZES
# Then run the benchmarks
make regenerate
```

### Adding New Storage Backends

To add a new storage backend:

1. Implement the Storage interface in the `storage` package
2. Register the storage backend in the `storage` package
3. Add a new benchmark target in the Makefile
4. Add the backend to the `run-all-benchmarks` target
