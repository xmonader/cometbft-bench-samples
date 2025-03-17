# Makefile for Transaction Batching Benchmarks
#
# This Makefile automates the process of running benchmarks for different storage backends,
# including setting up required services like Redis and TigerBeetle.

# Configuration
REDIS_PORT := 6379
REDIS_CONTAINER := tx-benchmark-redis
TIGERBEETLE_PORT := 3000
TIGERBEETLE_CONTAINER := tx-benchmark-tigerbeetle
TIGERBEETLE_CLUSTER_ID := 0
NUM_OPERATIONS := 1000
NUM_ACCOUNTS := 100
BATCH_SIZES := 1,10,50,100,500,1000

# Main targets
.PHONY: all setup run-all-benchmarks clean help

all: setup run-all-benchmarks

# Help target
help:
	@echo "Transaction Batching Benchmark Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  all                  - Set up services and run all benchmarks"
	@echo "  setup                - Set up Redis and TigerBeetle services"
	@echo "  run-all-benchmarks   - Run benchmarks for all storage backends"
	@echo "  benchmark-memory     - Run benchmarks for Memory storage"
	@echo "  benchmark-badger     - Run benchmarks for BadgerDB storage"
	@echo "  benchmark-sqlite     - Run benchmarks for SQLite storage"
	@echo "  benchmark-redis      - Run benchmarks for Redis storage"
	@echo "  benchmark-tigerbeetle - Run benchmarks for TigerBeetle storage"
	@echo "  setup-redis          - Set up Redis service"
	@echo "  setup-tigerbeetle    - Set up TigerBeetle service"
	@echo "  stop-redis           - Stop Redis service"
	@echo "  stop-tigerbeetle     - Stop TigerBeetle service"
	@echo "  setup-services-compose - Set up all services using Docker Compose"
	@echo "  stop-services-compose  - Stop all services using Docker Compose"
	@echo "  clean                - Remove all benchmark results and stop services"
	@echo "  regenerate           - Clean and rerun all benchmarks"
	@echo ""
	@echo "Configuration:"
	@echo "  NUM_OPERATIONS = $(NUM_OPERATIONS)"
	@echo "  NUM_ACCOUNTS = $(NUM_ACCOUNTS)"
	@echo "  BATCH_SIZES = $(BATCH_SIZES)"

# Setup targets
setup: setup-redis setup-tigerbeetle

# Service management targets
setup-redis:
	@echo "Setting up Redis..."
	@if [ -f docker-compose.yml ]; then \
		docker compose up -d redis; \
		echo "Redis container started using Docker Compose"; \
		sleep 2; \
	elif [ -z "$$(docker ps -q -f name=$(REDIS_CONTAINER))" ]; then \
		docker run --name $(REDIS_CONTAINER) -d -p $(REDIS_PORT):6379 redis:alpine; \
		echo "Redis container started on port $(REDIS_PORT)"; \
		sleep 2; \
	else \
		echo "Redis container is already running"; \
	fi
	@echo "Checking Redis connection..."
	@docker exec $(REDIS_CONTAINER) redis-cli ping || (echo "Redis is not responding" && exit 1)
	@echo "Redis is ready"

setup-tigerbeetle:
	@echo "Setting up TigerBeetle..."
	@if [ -f docker-compose.yml ]; then \
		docker compose up -d tigerbeetle; \
		echo "TigerBeetle container started using Docker Compose"; \
		sleep 2; \
	elif [ -z "$$(docker ps -q -f name=$(TIGERBEETLE_CONTAINER))" ]; then \
		docker run --name $(TIGERBEETLE_CONTAINER) -d -p $(TIGERBEETLE_PORT):3000 \
			ghcr.io/tigerbeetle/tigerbeetle:0.14.2 \
			start --addresses=0.0.0.0:3000 --cluster-id=$(TIGERBEETLE_CLUSTER_ID) --memory=true; \
		echo "TigerBeetle container started on port $(TIGERBEETLE_PORT)"; \
		sleep 2; \
	else \
		echo "TigerBeetle container is already running"; \
	fi
	@echo "TigerBeetle is ready"

stop-redis:
	@echo "Stopping Redis..."
	@if [ -f docker-compose.yml ]; then \
		docker compose stop redis; \
		docker compose rm -f redis; \
		echo "Redis container stopped and removed using Docker Compose"; \
	elif [ -n "$$(docker ps -q -f name=$(REDIS_CONTAINER))" ]; then \
		docker stop $(REDIS_CONTAINER); \
		docker rm $(REDIS_CONTAINER); \
		echo "Redis container stopped and removed"; \
	else \
		echo "Redis container is not running"; \
	fi

stop-tigerbeetle:
	@echo "Stopping TigerBeetle..."
	@if [ -f docker-compose.yml ]; then \
		docker compose stop tigerbeetle; \
		docker compose rm -f tigerbeetle; \
		echo "TigerBeetle container stopped and removed using Docker Compose"; \
	elif [ -n "$$(docker ps -q -f name=$(TIGERBEETLE_CONTAINER))" ]; then \
		docker stop $(TIGERBEETLE_CONTAINER); \
		docker rm $(TIGERBEETLE_CONTAINER); \
		echo "TigerBeetle container stopped and removed"; \
	else \
		echo "TigerBeetle container is not running"; \
	fi

setup-services-compose:
	@echo "Setting up services using Docker Compose..."
	@docker compose up -d
	@echo "Services started using Docker Compose"

stop-services-compose:
	@echo "Stopping services using Docker Compose..."
	@docker compose down
	@echo "Services stopped using Docker Compose"

# Benchmark targets
run-all-benchmarks: benchmark-memory benchmark-badger benchmark-sqlite benchmark-redis benchmark-tigerbeetle
	@echo "All benchmarks completed"
	@echo "Results are available in the benchmark_results directory"

benchmark-memory:
	@echo "Running Memory benchmarks..."
	@mkdir -p benchmark_results/memory
	@go run cmd/benchmark/main.go \
		--storage-backends=memory \
		--batch-sizes=$(BATCH_SIZES) \
		--num-accounts=$(NUM_ACCOUNTS) \
		--num-operations=$(NUM_OPERATIONS) \
		--output-dir=benchmark_results/memory
	@echo "Memory benchmarks completed"

benchmark-badger:
	@echo "Running BadgerDB benchmarks..."
	@mkdir -p benchmark_results/badger
	@go run cmd/benchmark/main.go \
		--storage-backends=badger \
		--batch-sizes=$(BATCH_SIZES) \
		--num-accounts=$(NUM_ACCOUNTS) \
		--num-operations=$(NUM_OPERATIONS) \
		--output-dir=benchmark_results/badger
	@echo "BadgerDB benchmarks completed"

benchmark-sqlite:
	@echo "Running SQLite benchmarks..."
	@mkdir -p benchmark_results/sqlite
	@go run cmd/benchmark/main.go \
		--storage-backends=sqlite \
		--batch-sizes=$(BATCH_SIZES) \
		--num-accounts=$(NUM_ACCOUNTS) \
		--num-operations=$(NUM_OPERATIONS) \
		--output-dir=benchmark_results/sqlite
	@echo "SQLite benchmarks completed"

benchmark-redis: setup-redis
	@echo "Running Redis benchmarks..."
	@mkdir -p benchmark_results/redis
	@go run cmd/benchmark/main.go \
		--storage-backends=redis \
		--batch-sizes=$(BATCH_SIZES) \
		--num-accounts=$(NUM_ACCOUNTS) \
		--num-operations=$(NUM_OPERATIONS) \
		--output-dir=benchmark_results/redis
	@echo "Redis benchmarks completed"

benchmark-tigerbeetle: setup-tigerbeetle
	@echo "Running TigerBeetle benchmarks..."
	@mkdir -p benchmark_results/tigerbeetle
	@go run cmd/benchmark/main.go \
		--storage-backends=tigerbeetle \
		--batch-sizes=$(BATCH_SIZES) \
		--num-accounts=$(NUM_ACCOUNTS) \
		--num-operations=$(NUM_OPERATIONS) \
		--output-dir=benchmark_results/tigerbeetle
	@echo "TigerBeetle benchmarks completed"

# Utility targets
clean: stop-redis stop-tigerbeetle
	@echo "Cleaning benchmark results..."
	@rm -rf benchmark_results/*/benchmark_results.csv benchmark_results/*/benchmark_report.md
	@rm -rf benchmark_results/*/benchmark_charts
	@echo "Benchmark results cleaned"

regenerate: clean setup
	@echo "All benchmarks regenerated"

# Check prerequisites
check-prerequisites:
	@echo "Checking prerequisites..."
	@which docker > /dev/null || (echo "Docker is not installed" && exit 1)
	@which go > /dev/null || (echo "Go is not installed" && exit 1)
	@which gnuplot > /dev/null || (echo "Gnuplot is not installed" && exit 1)
	@if which docker > /dev/null; then \
		echo "Docker Compose is installed (optional)"; \
	else \
		echo "Docker Compose is not installed (optional)"; \
	fi
	@echo "All required prerequisites are installed"

# Generate simple explanations
generate-explanations: run-all-benchmarks
	@echo "Generating simple explanations..."
	@mkdir -p benchmark_results/sqlite
	@mkdir -p benchmark_results/redis
	@mkdir -p benchmark_results/tigerbeetle
	@echo "Simple explanations generated"
