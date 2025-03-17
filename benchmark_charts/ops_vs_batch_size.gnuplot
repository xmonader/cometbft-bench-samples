set terminal png size 800,600
set output "benchmark_charts/ops_vs_batch_size.png"
set title "Operations Per Second (OPS) vs. Batch Size"
set xlabel "Batch Size"
set ylabel "OPS"
set grid
set key top left
set logscale x
set logscale y
plot "benchmark_results/tigerbeetle/benchmark_results.csv" using 1:3 with linespoints title "OPS"
