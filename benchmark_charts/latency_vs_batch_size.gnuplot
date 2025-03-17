set terminal png size 800,600
set output "benchmark_charts/latency_vs_batch_size.png"
set title "Latency vs. Batch Size"
set xlabel "Batch Size"
set ylabel "Latency (ms)"
set grid
set key top left
set logscale x
plot "benchmark_results/tigerbeetle/benchmark_results.csv" using 1:4 with linespoints title "Latency"
