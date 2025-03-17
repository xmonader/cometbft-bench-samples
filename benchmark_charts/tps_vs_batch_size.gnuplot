set terminal png size 800,600
set output "benchmark_charts/tps_vs_batch_size.png"
set title "Transactions Per Second (TPS) vs. Batch Size"
set xlabel "Batch Size"
set ylabel "TPS"
set grid
set key top right
set logscale x
set datafile separator ","
# Skip header line
plot "benchmark_results/sqlite/benchmark_results.csv" using 1:2 every ::1 with linespoints title "TPS"
