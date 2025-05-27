# Project Overview and Recent Progress

## 1. Purpose and Scope

This project is a control and monitoring tool for PPA (DSP speaker system) devices, featuring both a command-line packet capture tool (`pcap`) and a web interface for device discovery, control, and monitoring. The web interface uses modern Go web development practices (Gorilla Mux, zerolog, htmx, templ) and supports real-time updates via Server-Sent Events (SSE).

## 2. What We Accomplished Last Time

### Major Features Added/Changed
- **Device Discovery UI**: The web interface now has a discovery section that allows users to start/stop device discovery and see a live-updating list of discovered devices. Users can connect to a device directly from this list.
- **Web Server Logging**: Introduced structured logging with zerolog, including request IDs, timing, response sizes, and panic recovery with stack traces.
- **Improved Locking Patterns**: Refactored server state management to avoid deadlocks and improve concurrency safety, especially around discovery and notification logic.
- **SSE Improvements**: SSE endpoints now use proper event formatting, heartbeats, buffered channels, and robust error handling. CORS headers are set for cross-origin support.
- **Router Refactor**: Migrated to Gorilla Mux for route handling and moved route definitions to a dedicated `router` package.
- **Types Refactor**: Moved shared types (AppState, DeviceInfo, PacketInfo, ServerInterface) to a new `cmd/ppa-web/types` package to break import cycles and clarify dependencies.
- **Packet Filtering Enhancements**: The `pcap` tool now supports `--exclude-packets` and allows filtering by numeric message type as well as named types.

### Key Files Touched
- `cmd/ppa-web/handler/handler.go`: HTTP handlers, SSE logic
- `cmd/ppa-web/server/server.go`: Server state, discovery, logging, notification
- `cmd/ppa-web/types/types.go`: Shared types/interfaces
- `cmd/ppa-web/router/router.go`: Route definitions (new)
- `cmd/ppa-web/templates/`: Templ files for UI
- `cmd/pcap/packet-handler.go`, `cmd/pcap/main.go`: Packet filtering logic
- `lib/protocol/ppa-protocol.go`: Added `ToMap()` for header serialization

## 3. Key Findings and Technical Insights
- **Concurrency**: Locking must be carefully managed when updating state and notifying listeners to avoid deadlocks and race conditions.
- **SSE**: Buffered channels and periodic heartbeats are essential for robust SSE connections, especially with htmx clients.
- **Type Organization**: Moving shared types to a dedicated package (`types`) helps break import cycles and clarifies the contract between server and handler.
- **Logging**: Structured, contextual logging (with request IDs and stack traces) is invaluable for debugging and observability.
- **Extensibility**: The new packet filtering logic in `pcap` is flexible and supports both named and numeric message types.

## 4. Next Steps
- [ ] **Testing**: Add more tests for the new discovery and SSE logic, especially around concurrency and error handling.
- [ ] **UI Polish**: Improve the discovery UI (e.g., show device details, last seen time, better error messages).
- [ ] **Documentation**: Expand README and in-code comments to help new contributors understand the architecture.
- [ ] **API/CLI Consistency**: Ensure that packet filtering options are consistent and well-documented across CLI and web.
- [ ] **Security**: Review CORS and input validation for the web interface.
- [ ] **Performance**: Profile the server under load (many devices, many clients) and optimize as needed.
- [ ] **Refactor**: Consider further modularization (e.g., move discovery logic to its own package).

## 5. Key Resources
- **Entry Points**:
  - `cmd/ppa-web/main.go` (web server startup)
  - `cmd/pcap/main.go` (packet capture CLI)
- **Types and State**: `cmd/ppa-web/types/types.go`
- **Templates**: `cmd/ppa-web/templates/`
- **Routing**: `cmd/ppa-web/router/router.go`
- **Discovery Logic**: `cmd/ppa-web/server/server.go`
- **Packet Handling**: `cmd/pcap/packet-handler.go`
- **Protocol Definitions**: `lib/protocol/ppa-protocol.go`
- **Logging**: `zerolog` (see `main.go` for setup)
- **UI**: Uses htmx, templ, Bootstrap

## 6. Saving Future Research
- All future research, findings, and next steps should be saved in `ttmp/YYYY-MM-DD/0X-XXX.md` as per project guidelines.

---

**For a new developer:**
- Start by reading this document and the files listed above.
- Run the web server (`go run cmd/ppa-web/main.go`) and experiment with device discovery.
- Try the `pcap` tool with various filtering options.
- Review the logging output and SSE events in the browser (network tab).
- When making changes, document your findings and next steps in the `ttmp/` directory as shown here. 