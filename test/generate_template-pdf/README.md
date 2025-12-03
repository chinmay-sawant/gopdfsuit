# k6 Load Testing for GoPdfSuit

This directory contains k6 load testing scripts for the GoPdfSuit PDF generation API.

## 📋 Prerequisites

- Ubuntu WSL2 (or any Linux distribution)
- k6 load testing tool
- GoPdfSuit server running on `http://localhost:8080`

## � Test Data

All tests load the payload from an external JSON file:
- **Source**: `sampledata/amazon/amazon_receipt.json`
- The JSON is loaded once at initialization time using k6's `open()` function
- This ensures consistent test data across all test files and easy payload updates

## �🚀 Quick Installation

### Install k6 on Ubuntu WSL2

Run the installation script:

```bash
chmod +x install_k6.sh
./install_k6.sh
```

Or install manually:

```bash
# Add k6 GPG key
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69

# Add k6 repository
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list

# Update and install
sudo apt-get update
sudo apt-get install k6
```

Verify installation:

```bash
k6 version
```

## 📁 Test Files

| File | Description | Duration | VUs |
|------|-------------|----------|-----|
| `smoke_test.js` | Quick verification test | 10s | 1 |
| `load_test.js` | Complete load test with multiple scenarios | ~13min | 1-50 |
| `spike_test.js` | Sudden traffic spike simulation | ~90s | 5-100 |
| `soak_test.js` | Extended duration stability test | 30min | 10 |

## 🏃 Running Tests

### Before Running Tests

Make sure the GoPdfSuit server is running:

```bash
# Start the server (from project root)
make run
# or
go run cmd/gopdfsuit/main.go
```

### Smoke Test (Quick Verification)

```bash
k6 run smoke_test.js
```

### Load Test (Full Suite)

```bash
k6 run load_test.js
```

### Spike Test

```bash
k6 run spike_test.js
```

### Soak Test (30 minutes)

```bash
k6 run soak_test.js
```

### Custom Configuration

Override virtual users and duration:

```bash
# Run with 20 VUs for 1 minute
k6 run --vus 20 --duration 1m smoke_test.js
```

Override base URL:

```bash
# Test against different server
k6 run -e BASE_URL=http://your-server:8080 load_test.js
```

## 📊 Test Scenarios

### Load Test Scenarios

The `load_test.js` contains three scenarios that run sequentially:

1. **Smoke (0-30s)**: 1 VU - Basic functionality verification
2. **Load (35s-6m)**: Ramps up to 10 VUs - Normal expected load
3. **Stress (6m-13m)**: Ramps up to 50 VUs - Beyond normal capacity

### Spike Test

Simulates sudden traffic bursts:
- Warm up: 5 VUs
- Spike: Sudden increase to 100 VUs
- Recovery: Back to normal load

### Soak Test

Sustained load over 30 minutes to identify:
- Memory leaks
- Resource exhaustion
- Performance degradation over time

## 📈 Metrics & Thresholds

### Key Metrics

| Metric | Description |
|--------|-------------|
| `http_req_duration` | Time to complete HTTP request |
| `http_req_failed` | Rate of failed requests |
| `pdf_generation_time` | Custom metric for PDF generation |
| `errors` | Custom error rate |

### Default Thresholds

| Test | p95 Response Time | Error Rate |
|------|-------------------|------------|
| Smoke | < 3s | < 5% |
| Load | < 5s | < 10% |
| Spike | < 10s | < 30% |
| Soak | < 5s | < 5% |

## 📝 Output & Reports

### Console Output

By default, k6 outputs results to the console with a summary.

### JSON Output

```bash
k6 run --out json=results.json load_test.js
```

### HTML Report (with k6-reporter)

```bash
# Install k6-reporter extension first
k6 run --out json=results.json load_test.js
# Then convert to HTML using k6-reporter
```

### InfluxDB + Grafana

```bash
k6 run --out influxdb=http://localhost:8086/k6 load_test.js
```

## 🔧 Customization

### Modify Payload

Edit the JSON file at `sampledata/amazon/amazon_receipt.json` to test with different data. All test files will automatically use the updated payload.

### Using a Different JSON File

To use a different JSON file, update the `open()` path in each test file:

```javascript
// Change this line in the test files
const amazonReceiptPayload = JSON.parse(open('../../sampledata/amazon/amazon_receipt.json'));

// To use a different file
const customPayload = JSON.parse(open('../sampledata/your/custom_file.json'));
```

### Add New Endpoints

Create a new test file or add scenarios to existing tests:

```javascript
import http from 'k6/http';
import { check } from 'k6';

export default function () {
    // Test your endpoint
    const response = http.get('http://localhost:8080/your-endpoint');
    check(response, {
        'status is 200': (r) => r.status === 200,
    });
}
```

## 🐛 Troubleshooting

### Connection Refused

Make sure the server is running:

```bash
curl http://localhost:8080/gopdfsuit
```

### High Error Rate

- Check server logs for errors
- Verify the payload format is correct
- Ensure server has sufficient resources

### Slow Response Times

- Monitor server CPU/memory
- Check for database bottlenecks
- Consider increasing server resources

## 📚 Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 GitHub](https://github.com/grafana/k6)
- [k6 Examples](https://k6.io/docs/examples/)

## 📄 License

This test suite is part of the GoPdfSuit project.
