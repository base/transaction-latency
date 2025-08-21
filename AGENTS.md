# AGENTS Guidelines for This Repository

This repository contains a Go-based tool for testing transaction latency on Base blockchain with Flashblocks support. When working on the project interactively with an agent (e.g. the Codex CLI) please follow the guidelines below for safe and efficient development.

## 1. Use Docker for Consistent Testing

* **Always use Docker** for running tests to ensure consistent environment.
* **Test with small transaction counts first** (`NUMBER_OF_TRANSACTIONS=1-5`) during development.
* **Use testnet endpoints** for development and testing before mainnet.
* **Monitor gas costs** when testing on mainnet.

## 2. Keep Dependencies in Sync

If you update dependencies:

1. Update using Go modules: `go get -u <package>`.
2. Run `go mod tidy` to clean up dependencies.
3. Test changes locally before committing.
4. Verify compatibility with Go 1.24+ as specified in the project.

## 3. Environment Configuration

Create a `.env` file for testing (never commit):

```env
PRIVATE_KEY=your_test_private_key
TO_ADDRESS=0xc2F695613de0885dA3bdd18E8c317B9fAf7d4eba
POLLING_INTERVAL_MS=50
BASE_NODE_ENDPOINT_1=https://your-flashblocks-endpoint
BASE_NODE_ENDPOINT_2=https://your-standard-endpoint
REGION=test
NUMBER_OF_TRANSACTIONS=5
SEND_TXN_SYNC=true
RUN_ENDPOINT2_TESTING=true
```

## 4. Code Quality Checks

Before completing any task, run these quality checks:

| Command                  | Purpose                                  |
| ------------------------ | ---------------------------------------- |
| `go fmt ./...`           | Format code to Go standards             |
| `go vet ./...`           | Run static analysis                     |
| `go test ./...`          | Run unit tests                          |
| `golangci-lint run`      | Run comprehensive linting (if installed)|

## 5. Testing Workflow

Test changes progressively:

1. **Build verification**: Ensure code compiles
   ```bash
   go build -o test-latency main.go
   ```

2. **Docker build**: Test container build
   ```bash
   docker build -t transaction-latency .
   ```

3. **Small test run**: Test with minimal transactions
   ```bash
   docker run -v $(pwd)/data:/data --env-file .env --rm -it transaction-latency
   ```

4. **Results verification**: Check output in `data/` directory

## 6. Development Best Practices

* Use proper error handling for all RPC calls.
* Log important events for debugging.
* Validate environment variables before use.
* Handle network timeouts gracefully.
* Clean up resources properly (close connections).

## 7. Docker Development

When modifying the Dockerfile:

* Keep the image minimal (use multi-stage builds if needed).
* Pin dependency versions for reproducibility.
* Use non-root user for running the application.
* Mount data directory for persistent results.

## 8. Testing Modes

### Synchronous Mode (`SEND_TXN_SYNC=true`)
* Uses `eth_sendRawTransactionSync` method.
* Requires Flashblocks-enabled endpoint.
* Provides instant confirmation receipts.
* Best for testing Flashblocks performance.

### Asynchronous Mode (`SEND_TXN_SYNC=false`)
* Uses standard `eth_sendTransaction` method.
* Works with any Base endpoint.
* Polls for receipts using `POLLING_INTERVAL_MS`.
* Good for compatibility testing.

## 9. Data Analysis

Results are saved to `data/` directory:
* Review CSV files for latency measurements.
* Compare endpoint1 vs endpoint2 performance.
* Check for anomalies in inclusion delays.
* Validate block numbers for consistency.

## 10. Useful Commands Recap

| Command                                        | Purpose                           |
| ---------------------------------------------- | --------------------------------- |
| `go build -o test-latency main.go`            | Build the binary                  |
| `docker build -t transaction-latency .`        | Build Docker image                |
| `docker run -v $(pwd)/data:/data --env-file .env --rm -it transaction-latency` | Run test |
| `go fmt ./...`                                 | Format Go code                    |
| `go mod tidy`                                  | Clean up dependencies             |

## 11. Safety Reminders

* **Never commit private keys** or sensitive data.
* **Use testnet first** for all development testing.
* **Monitor gas usage** to avoid unexpected costs.
* **Start with few transactions** to validate logic.
* **Check endpoint compatibility** before testing.
* **Keep test data** for analysis and debugging.

## 12. Common Issues

* "Error loading .env file" in Docker is expected and harmless (env vars loaded via --env-file).
* Ensure endpoints support required RPC methods for chosen mode.
* Flashblocks endpoints required for synchronous mode.
* Network latency affects actual confirmation times.

---

Following these practices ensures safe testing, prevents unnecessary gas costs, and maintains code quality. Always test with small transaction counts and testnet endpoints during development before scaling up to production testing.