all: build/ppa-cli.arm64.macos build/ppa-cli.x86_64.windows.exe

build:
	mkdir -p build

build/ppa-cli.arm64.macos: build
	GOOS=darwin GOARCH=arm64 go build -o $@ ./cmd/ppa-cli

build/ppa-control.app: build
	cd cmd/ui-test \
       && CGO_ENABLED=1 fyne package -os darwin --name ppa-control \
       && mv ppa-control.app ../../build

build/ppa-cli.x86_64.windows.exe: build
	GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/ppa-cli

build/ppa-control.exe: build
	cd cmd/ui-test \
       && CGO_ENABLED=1 fyne package -os windows --name ppa-control \
       && mv ppa-control.exe ../../build
