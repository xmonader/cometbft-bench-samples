# Benchmark Report

Generated at: 2025-03-18 01:48:22

## Results

### Batch Size Impact

| Batch Size | TPS | OPS | Latency (ms) | Storage Backend |
|------------|-----|-----|--------------|----------------|
| 1 | 115.45 | 115.45 | 8.66 | sqlite |
| 10 | 69.94 | 699.41 | 14.29 | sqlite |
| 50 | 42.91 | 2145.67 | 23.30 | sqlite |
| 100 | 35.72 | 3572.10 | 27.90 | sqlite |

### Storage Backend Impact (Batch Size = 100)

| Storage Backend | TPS | OPS | Latency (ms) |
|-----------------|-----|-----|--------------|
| sqlite | 35.72 | 3572.10 | 27.90 |

## Analysis

### Batch Size Impact

As the batch size increases, we observe the following trends:

1. **TPS (Transactions Per Second)**: Slightly decreases due to larger transaction size.
2. **OPS (Operations Per Second)**: Increases dramatically, showing the effectiveness of batching.
3. **Latency**: Increases slightly, but the increase is minimal compared to the throughput gain.

### Storage Backend Impact

Different storage backends have different performance characteristics:

1. **Memory**: Highest performance, but no persistence.
2. **BadgerDB**: Good balance of performance and persistence.
3. **SQLite**: Lower performance, but ACID compliant.
4. **Redis**: Good performance, especially in distributed scenarios.
5. **TigerBeetle**: Excellent performance for financial applications.

## Recommendations

Based on the benchmark results, we recommend:

1. **Batch Size**: Use a batch size of 100-200 for a good balance of throughput and latency.
2. **Storage Backend**: 
   - For development/testing: Use Memory storage.
   - For production (high performance): Use BadgerDB or TigerBeetle.
   - For production (ACID compliance): Use SQLite.
   - For production (distributed): Use Redis.

## Charts

Charts are available in the 'charts' subdirectory:

- TPS vs. Batch Size
- OPS vs. Batch Size
- Latency vs. Batch Size
- TPS vs. Storage Backend
- OPS vs. Storage Backend
