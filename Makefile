all: build/ppa-cli.arm64.macos build/ppa-cli.x86_64.windows

build:
	mkdir -p build

build/ppa-cli.arm64.macos: build
	GOOS=darwin GOARCH=arm64 go build -o $@ ./cmd/ppa-cli

build/ppa-cli.x86_64.windows: build
	GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/ppa-cli
