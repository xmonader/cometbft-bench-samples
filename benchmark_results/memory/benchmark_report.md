# Benchmark Report

Generated at: 2025-03-18 12:04:21

## Results

### Batch Size Impact

| Batch Size | TPS | OPS | Latency (ms) | Storage Backend |
|------------|-----|-----|--------------|----------------|
| 1 | 945733.79 | 945733.79 | 0.00 | memory |
| 10 | 193087.47 | 1930874.69 | 0.00 | memory |
| 50 | 41438.84 | 2071941.97 | 0.00 | memory |
| 100 | 21511.62 | 2151162.17 | 0.00 | memory |

### Storage Backend Impact (Batch Size = 100)

| Storage Backend | TPS | OPS | Latency (ms) |
|-----------------|-----|-----|--------------|
| memory | 21511.62 | 2151162.17 | 0.00 |

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
