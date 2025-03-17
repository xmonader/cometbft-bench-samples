set terminal png size 800,600
set output "benchmark_charts/ops_vs_storage.png"
set title "Operations Per Second (OPS) vs. Storage Backend"
set xlabel "Storage Backend"
set ylabel "OPS"
set grid
set key top right
set style data histogram
set style histogram cluster gap 1
set style fill solid border -1
set boxwidth 0.9
set xtic rotate by -45 scale 0
set datafile separator ","
# Skip header line
plot "benchmark_results/sqlite/benchmark_results.csv" using 3:xtic(5) every ::1 title "OPS"
