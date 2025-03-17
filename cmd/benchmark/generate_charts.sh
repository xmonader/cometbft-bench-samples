#!/bin/bash

# This script generates charts from benchmark results using gnuplot.
# It takes a benchmark results file as input and generates charts for:
# 1. Transactions per second (TPS) vs. Batch Size
# 2. Operations per second (OPS) vs. Batch Size
# 3. Latency vs. Batch Size
# 4. TPS vs. Storage Backend
# 5. OPS vs. Storage Backend

set -e

# Check if gnuplot is installed
if ! command -v gnuplot &> /dev/null; then
    echo "Error: gnuplot is not installed. Please install it first."
    echo "On Ubuntu/Debian: sudo apt-get install gnuplot"
    echo "On macOS: brew install gnuplot"
    exit 1
fi

# Check if input file is provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <benchmark_results.csv>"
    exit 1
fi

INPUT_FILE="$1"
OUTPUT_DIR="benchmark_charts"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Generate TPS vs. Batch Size chart
cat > "$OUTPUT_DIR/tps_vs_batch_size.gnuplot" << EOL
set terminal png size 800,600
set output "$OUTPUT_DIR/tps_vs_batch_size.png"
set title "Transactions Per Second (TPS) vs. Batch Size"
set xlabel "Batch Size"
set ylabel "TPS"
set grid
set key top right
set logscale x
set datafile separator ","
# Skip header line
plot "$INPUT_FILE" using 1:2 every ::1 with linespoints title "TPS"
EOL

# Generate OPS vs. Batch Size chart
cat > "$OUTPUT_DIR/ops_vs_batch_size.gnuplot" << EOL
set terminal png size 800,600
set output "$OUTPUT_DIR/ops_vs_batch_size.png"
set title "Operations Per Second (OPS) vs. Batch Size"
set xlabel "Batch Size"
set ylabel "OPS"
set grid
set key top left
set logscale x
set logscale y
set datafile separator ","
# Skip header line
plot "$INPUT_FILE" using 1:3 every ::1 with linespoints title "OPS"
EOL

# Generate Latency vs. Batch Size chart
cat > "$OUTPUT_DIR/latency_vs_batch_size.gnuplot" << EOL
set terminal png size 800,600
set output "$OUTPUT_DIR/latency_vs_batch_size.png"
set title "Latency vs. Batch Size"
set xlabel "Batch Size"
set ylabel "Latency (ms)"
set grid
set key top left
set logscale x
set datafile separator ","
# Skip header line
plot "$INPUT_FILE" using 1:4 every ::1 with linespoints title "Latency"
EOL

# Generate TPS vs. Storage Backend chart
cat > "$OUTPUT_DIR/tps_vs_storage.gnuplot" << EOL
set terminal png size 800,600
set output "$OUTPUT_DIR/tps_vs_storage.png"
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
plot "$INPUT_FILE" using 2:xtic(5) every ::1 title "TPS"
EOL

# Generate OPS vs. Storage Backend chart
cat > "$OUTPUT_DIR/ops_vs_storage.gnuplot" << EOL
set terminal png size 800,600
set output "$OUTPUT_DIR/ops_vs_storage.png"
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
plot "$INPUT_FILE" using 3:xtic(5) every ::1 title "OPS"
EOL

# Run gnuplot on all files
for plot_file in "$OUTPUT_DIR"/*.gnuplot; do
    echo "Generating chart from $plot_file..."
    gnuplot "$plot_file"
done

echo "Charts generated successfully in $OUTPUT_DIR directory."
echo "The following charts were created:"
echo "- TPS vs. Batch Size: $OUTPUT_DIR/tps_vs_batch_size.png"
echo "- OPS vs. Batch Size: $OUTPUT_DIR/ops_vs_batch_size.png"
echo "- Latency vs. Batch Size: $OUTPUT_DIR/latency_vs_batch_size.png"
echo "- TPS vs. Storage Backend: $OUTPUT_DIR/tps_vs_storage.png"
echo "- OPS vs. Storage Backend: $OUTPUT_DIR/ops_vs_storage.png"

# Make the script executable
chmod +x "$0"
