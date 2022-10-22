all: build/ppa-cli.arm64.macos build/ppa-cli.x86_64.macos build/ppa-cli.x86_64.windows.exe

build:
	mkdir -p build

build/ppa-cli.arm64.macos: build
	GOOS=darwin GOARCH=arm64 go build -o $@ ./cmd/ppa-cli

build/ppa-cli.x86_64.macos: build
	GOOS=darwin GOARCH=amd64 go build -o $@ ./cmd/ppa-cli

build/ppa-control.app: build
	rm -rf build/ppa-control.app
	cd cmd/ui-test \
       && MACOSX_DEPLOYMENT_TARGET=10.11 \
          GOARCH=amd64 \
          CGO_CFLAGS=-mmacosx-version-min=10.11 \
          CGO_LDFLAGS=-mmacosx-version-min=10.11 \
          CGO_ENABLED=1 fyne package -os darwin --name ppa-control \
       && mv ppa-control.app ../../build

build/ppa-cli.x86_64.windows.exe: build
	GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/ppa-cli

build/ppa-control.exe: build
	cd cmd/ui-test \
       && CGO_ENABLED=1 fyne package -os windows --name ppa-control \
       && mv ppa-control.exe ../../build
