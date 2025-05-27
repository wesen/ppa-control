# PPA Control Agent Guidelines

## Commands
- **Build**: `make all` (cross-platform builds) or `go build ./cmd/ppa-cli`
- **Test**: `make test` or `go test ./...`
- **Test single package**: `go test ./path/to/package`
- **Lint**: `make lint` or `golangci-lint run -v`
- **Generate code**: `go generate ./...`
- **Release build**: `make goreleaser`

## Project Structure
- `cmd/`: Main applications (ppa-cli, pcap, ui-test, ppa-web)
- `lib/`: Reusable library code (client, protocol, utils)
- `test/`: Test utilities and fixtures
- Skip `cmd/pcap` directory in linting (configured in .golangci.yml)

## Go Guidelines

- Uses Go 1.23+ with modules
- When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
- When building web applications, use htmx, bootstrap and the templ templating language.
- Always use a context argument when appropriate.
- Use cobra for command-line applications.
- Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
- Use github.com/pkg/errors for wrapping errors.
- When starting goroutines, use errgroup.
- Don't create new go.mod in the subdirectories, instead rely on the top level one
- Create apps in self-contained folders, usually in cmd/apps or in cmd/experiments
- Use Fyne for GUI applications
- Use github.com/rs/zerolog/log for structured logging
- Follow standard naming (CamelCase for exported, camelCase for unexported)
- Import style: Standard library first, then third-party, then local (`ppa-control/lib/...`)
- Define enums with `go:generate stringer`, use typed constants
- Testing: Table-driven tests with struct slices, descriptive test names
- Return errors explicitly, use structured concurrency
- Pre-commit hooks use lefthook (configured in lefthook.yml)

## Web Guidelines

Use bun, react and rtk-query. Use typescript.
Use bootstrap for styling.

## Debugging Guidelines

If me or you the LLM agent seem to go down too deep in a debugging/fixing rabbit hole in our conversations, remind me to take a breath and think about the bigger picture instead of hacking away. Say: "I think I'm stuck, let's TOUCH GRASS".  IMPORTANT: Don't try to fix errors by yourself more than twice in a row. Then STOP. Don't do anything else.

## General Guidelines

Run the format_file tool at the end of each response.