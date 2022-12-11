all: build/ppa-cli.arm64.macos build/ppa-cli.x86_64.macos build/ppa-cli.x86_64.windows.exe

build:
	mkdir -p build

build/ppa-cli.arm64.macos: build
	go generate ./...
	GOOS=darwin GOARCH=arm64 go build -o $@ ./cmd/ppa-cli

build/ppa-cli.x86_64.macos: build
	go generate ./...
	GOOS=darwin GOARCH=amd64 go build -o $@ ./cmd/ppa-cli

build/ppa-control.app: build
	rm -rf build/ppa-control.app
	go generate ./...
	cd cmd/ui-test \
       && MACOSX_DEPLOYMENT_TARGET=10.11 \
          GOARCH=amd64 \
          CGO_CFLAGS=-mmacosx-version-min=10.11 \
          CGO_LDFLAGS=-mmacosx-version-min=10.11 \
          CGO_ENABLED=1 fyne package -os darwin --name ppa-control \
       && mv ppa-control.app ../../build

build/ppa-cli.x86_64.windows.exe: build
	go generate ./...
	GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/ppa-cli

build/ppa-control.exe: build
	go generate ./...
	cd cmd/ui-test \
       && CGO_ENABLED=1 fyne package -os windows --name ppa-control \
       && mv ppa-control.exe ../../build

lint:
	#docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v --skip-dirs cmd/pcap
	golangci-lint run -v

test:
	go test ./...

goreleaser:
	goreleaser release --snapshot --rm-dist

tag-release:
	git tag ${VERSION}
