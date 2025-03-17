set terminal png size 800,600
set output "benchmark_results/memory/charts/tps_vs_storage.png"
set title "Transactions Per Second (TPS) vs. Storage Backend"
set xlabel "Storage Backend"
set ylabel "TPS"
set grid
set key top right
set style data histogram
set style histogram cluster gap 1
set style fill solid border -1
set boxwidth 0.9
set xtic rotate by -45 scale 0
set datafile separator ","
# Skip header line
plot "benchmark_results/memory/benchmark_results.csv" using 2:xtic(5) every ::1 title "TPS"
