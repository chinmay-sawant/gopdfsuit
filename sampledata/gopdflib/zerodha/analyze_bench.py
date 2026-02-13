
import sys
import numpy as np

def parse_runs(filename):
    throughputs = []
    latencies = []
    try:
        with open(filename, 'r') as f:
            for line in f:
                if "Throughput:" in line:
                    val = float(line.split(":")[1].strip().split()[0])
                    throughputs.append(val)
                if "Avg Latency:" in line:
                    val = float(line.split(":")[1].strip().split()[0])
                    latencies.append(val)
    except FileNotFoundError:
        return [], []
    return throughputs, latencies

t24, l24 = parse_runs("1.24.txt")
t26, l26 = parse_runs("1.26.txt")

def print_stats(name, data, unit):
    if not data:
        print(f"No data for {name}")
        return
    print(f"--- {name} ({unit}) ---")
    print(f"Runs: {len(data)}")
    print(f"Mean: {np.mean(data):.2f}")
    print(f"Median: {np.median(data):.2f}")
    print(f"Min: {np.min(data):.2f}")
    print(f"Max: {np.max(data):.2f}")
    print(f"StdDev: {np.std(data):.2f}")
    print()

print("### Benchmark Analysis")
print_stats("Go 1.24 Throughput", t24, "ops/sec")
print_stats("Go 1.26 Throughput", t26, "ops/sec")

print_stats("Go 1.24 Avg Latency", l24, "ms")
print_stats("Go 1.26 Avg Latency", l26, "ms")
