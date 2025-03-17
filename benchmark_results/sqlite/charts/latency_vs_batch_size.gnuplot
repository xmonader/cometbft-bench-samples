set terminal png size 800,600
set output "benchmark_results/sqlite/charts/latency_vs_batch_size.png"
set title "Latency vs. Batch Size"
set xlabel "Batch Size"
set ylabel "Latency (ms)"
set grid
set key top left
set logscale x
set datafile separator ","
# Skip header line
plot "benchmark_results/sqlite/benchmark_results.csv" using 1:4 every ::1 with linespoints title "Latency"
