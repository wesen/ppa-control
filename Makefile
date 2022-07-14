all: ppa-cli.arm64.macos

ppa-cli.arm64.macos:
	mkdir -p build
	go build -o build/ppa-cli.arm64.macos ./cmd/ppa-cli
