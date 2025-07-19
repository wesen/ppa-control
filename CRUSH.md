# Commands
make lint          # Run golangci-lint with config from .golangci.yml
make test          # Run all Go tests
make build         # Create build directory structure
go test ./... -v   # Run single test with verbose output
go test ./cmd/pcap -run TestPacketHandler -v  # Run specific test

# Style Guidelines
- Use goimports for import sorting (std lib, external, internal)
- Error handling: use pkg/errors with stack traces
- Naming: CamelCase for exported, camelCase for unexported
- Interfaces: suffix with "er" (e.g., Discoverer, Handler)
- Constants: UPPER_CASE with context prefix
- Use zerolog for logging, avoid fmt.Printf in production
- Follow standard Go formatting (gofmt)
- Skip cmd/pcap directory from linting (see .golangci.yml)

# Structure
- cmd/ contains CLI binaries (ppa-cli, ppa-web, pcap)
- lib/ contains shared libraries and protocols
- mobile/ contains React Native app
- test/ contains integration tests