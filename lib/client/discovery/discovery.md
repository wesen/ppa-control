# PPA Device Discovery System

The discovery system allows automatic detection of PPA devices on the network. It uses UDP broadcast messages to discover devices and maintains a list of active devices with timeouts.

## Architecture

The discovery system consists of several components working together:

1. **Interface Discovery**: Monitors network interfaces and manages their addition/removal
2. **Interface Manager**: Manages UDP clients for each interface
3. **Discovery Loop**: Handles peer timeouts and message processing

### Concurrency Setup

The system runs three main goroutines:

```
                                ┌─────────────────┐
                                │  Main Context   │
                                └────────┬────────┘
                                         │
                     ┌───────────────────┴───────────────────┐
                     │                                       │
            ┌────────┴────────┐                    ┌────────┴────────┐
            │ Interface       │                    │ Discovery       │
            │ Discoverer (GR1)│                    │ Loop (GR2)     │
            └────────┬────────┘                    └────────┬────────┘
                     │                                      │
                     │                              ┌───────┴───────┐
                     │                              │ Per-Interface │
                     └──────────────────────────────► UDP Clients   │
                      (Add/Remove Interfaces)       └───────────────┘
```

1. **Interface Discoverer (GR1)**:
   - Monitors network interfaces
   - Sends interface updates through channels (`addedInterfaceCh` and `removedInterfaceCh`)
   - Runs in `interfaceDiscoverer.Run()`

2. **Discovery Loop (GR2)**:
   - Manages peer timeouts
   - Processes received messages
   - Sends discovery status updates
   - Handles interface changes from GR1
   - Runs in the second goroutine in `Discover()`

3. **Per-Interface UDP Clients**:
   - Created/destroyed by Interface Manager
   - Handle UDP communication on each interface
   - Send and receive broadcast messages

## Usage Example

Basic usage with CommandContext (from ping.go):

```go
// Setup discovery if enabled
cmdCtx.SetupDiscovery()

// Start multiclient
cmdCtx.StartMultiClient()

// Handle discovery messages
cmdCtx.RunInGroup(func() error {
    for {
        select {
        case msg := <-cmdCtx.Channels.DiscoveryCh:
            if newClient, err := cmdCtx.HandleDiscoveryMessage(msg); err != nil {
                return err
            } else if newClient != nil {
                // Send ping immediately to newly discovered client
                newClient.SendPing()
            }
        }
    }
})
```

Direct usage with interface monitoring:

```go
// Create channels
discoveryCh := make(chan discovery.PeerInformation)
interfaceManager := discovery.NewInterfaceManager(5001, receiveCh)
interfaceDiscoverer := discovery.NewInterfaceDiscoverer(interfaceManager, []string{"eth0", "wlan0"})

// Monitor interface changes
go func() {
    for {
        select {
        case newIface := <-interfaceDiscoverer.addedInterfaceCh:
            log.Info().
                Str("iface", newIface).
                Msg("Network interface added")

        case removedIface := <-interfaceDiscoverer.removedInterfaceCh:
            log.Info().
                Str("iface", removedIface).
                Msg("Network interface removed")
        }
    }
}()

// Start discovery
ctx := context.Background()
go discovery.Discover(ctx, discoveryCh, []string{"eth0", "wlan0"}, 5001)

// Handle discovery events
for msg := range discoveryCh {
    switch msg.(type) {
    case discovery.PeerDiscovered:
        log.Info().
            Str("addr", msg.GetAddress()).
            Str("iface", msg.GetInterface()).
            Msg("New device discovered")
    case discovery.PeerLost:
        log.Info().
            Str("addr", msg.GetAddress()).
            Str("iface", msg.GetInterface()).
            Msg("Device lost")
    }
}
```

## Web UI Integration

The discovery system is well-suited for web UI integration through its message channels. The system provides three types of events:

1. **Device Events**:
   - `PeerDiscovered`: When a new device is found
   - `PeerLost`: When a device hasn't responded for 30 seconds

2. **Interface Events**:
   - Interface added: When a new network interface becomes available
   - Interface removed: When a network interface is removed or disabled

To integrate with a web UI:

1. Store the discovery and interface channels in your server state
2. Process messages in goroutines to update UI state
3. Use websockets or server-sent events to push updates to the browser

Example integration:

```go
type AppState struct {
    DiscoveredDevices map[string]DeviceInfo  // addr -> info
    ActiveInterfaces  map[string]bool        // iface -> active
    // ... other state fields
}

// Start goroutines to process events
go func() {
    // Process device discovery
    for msg := range discoveryCh {
        switch m := msg.(type) {
        case discovery.PeerDiscovered:
            server.SetState(func(state *AppState) {
                state.DiscoveredDevices[m.GetAddress()] = DeviceInfo{
                    Address:   m.GetAddress(),
                    Interface: m.GetInterface(),
                    LastSeen:  time.Now(),
                }
            })
        case discovery.PeerLost:
            server.SetState(func(state *AppState) {
                delete(state.DiscoveredDevices, m.GetAddress())
            })
        }
    }
}()

go func() {
    // Process interface changes
    for {
        select {
        case newIface := <-interfaceDiscoverer.addedInterfaceCh:
            server.SetState(func(state *AppState) {
                state.ActiveInterfaces[newIface] = true
            })
        case removedIface := <-interfaceDiscoverer.removedInterfaceCh:
            server.SetState(func(state *AppState) {
                delete(state.ActiveInterfaces, removedIface)
            })
        }
    }
}()
```

## Timeouts and Cleanup

- Devices are considered lost after 30 seconds of no response
- Interface clients are cleaned up when interfaces are removed
- All goroutines and resources are cleaned up when the context is cancelled

## Error Handling

The discovery system handles several types of errors:

1. Interface errors (e.g., interface disappears)
2. UDP socket errors
3. Context cancellation

All errors are propagated through the error group and will cause graceful shutdown of the discovery system. 