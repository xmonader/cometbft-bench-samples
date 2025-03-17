# Transaction Batching with Tendermint/CometBFT

This project demonstrates how to implement transaction batching with Tendermint/CometBFT to increase throughput. By batching multiple operations into a single transaction, we can significantly increase the number of operations processed per second.

## Overview

Tendermint/CometBFT is a Byzantine Fault Tolerant (BFT) consensus engine that can handle up to 10,000 transactions per second. However, since the meaning of a "transaction" is user-defined, we can batch multiple operations into a single transaction to increase the effective throughput.

This project implements a simple financial application that supports account creation and money transfers. It demonstrates how to batch multiple transfer operations into a single transaction, with each operation having its own signature, and then submit the batch to the Tendermint/CometBFT network.

## Features

- **Multiple Storage Backends**: Support for various storage backends including in-memory, BadgerDB, SQLite, Redis, and TigerBeetle.
- **Transaction Batching**: Batch multiple operations into a single transaction to increase throughput.
- **Per-Operation Signatures**: Each operation has its own signature, allowing operations from different users to be batched together.
- **Digital Signatures**: Sign operations with Ed25519 keys to ensure authenticity and integrity.
- **Benchmarking**: Tools to benchmark the performance of different batching strategies and storage backends.

## Architecture

The project is organized into several packages:

- **app**: Contains the ABCI application implementation.
- **client**: Provides a client library for interacting with the application.
- **cmd**: Contains command-line tools for running the application and benchmarks.
- **crypto**: Implements cryptographic operations such as signing and verification.
- **storage**: Provides various storage backends for persisting application state.
- **types**: Defines common data structures used throughout the application.

## Getting Started

### Prerequisites

- Go 1.16 or later
- Tendermint/CometBFT

### Installation

1. Clone the repository:

```bash
git clone https://github.com/xmonader/test_batched_tx_tendermint.git
cd test_batched_tx_tendermint
```

2. Build the application:

```bash
go build -o batched_tx_app
```

3. Run the application with Tendermint:

```bash
./batched_tx_app
```

In another terminal, run Tendermint:

```bash
tendermint init
tendermint node --proxy_app=tcp://localhost:26658
```

### Configuration

The application can be configured using a JSON configuration file. Here's an example:

```json
{
  "storage": {
    "backend": "memory",
    "config": {}
  },
  "batch_size": 100,
  "private_key": "path/to/private_key.pem"
}
```

Available storage backends:

- **memory**: In-memory storage (no persistence)
- **badger**: BadgerDB storage
- **sqlite**: SQLite storage
- **redis**: Redis storage
- **tigerbeetle**: TigerBeetle storage

Each storage backend has its own configuration options. See the documentation for details.

## Usage

### Creating an Account

```bash
curl -X POST http://localhost:26657/broadcast_tx_commit?tx=0x$(echo -n '{"type":"create_account","id":1,"initial_balance":1000}' | xxd -p)
```

### Transferring Money

```bash
curl -X POST http://localhost:26657/broadcast_tx_commit?tx=0x$(echo -n '{"type":"transfer","from":1,"to":2,"amount":100}' | xxd -p)
```

### Batching Transfers

```bash
curl -X POST http://localhost:26657/broadcast_tx_commit?tx=0x$(echo -n '{"type":"batch","operations":[{"type":"transfer","from":1,"to":2,"amount":10},{"type":"transfer","from":1,"to":3,"amount":20}]}' | xxd -p)
```

## Benchmarking

The project includes benchmarking tools to measure the performance of different batching strategies and storage backends.

```bash
go run cmd/benchmark/main.go --batch-sizes=1,10,50,100,500,1000 --num-accounts=1000 --num-operations=10000 --storage-backends=memory,badger
```

This will run benchmarks with various batch sizes, 1000 accounts, and 10,000 total operations using both the in-memory and BadgerDB storage backends.

For detailed instructions on running benchmarks and generating reports, see [Benchmark Instructions](docs/benchmark_instructions.md).

## Performance Analysis

See [Transaction Batching Analysis](docs/transaction_batching_analysis.md) for a detailed analysis of the performance of different batching strategies and storage backends.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
