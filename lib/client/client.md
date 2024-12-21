# PPA Control Client Architecture

This document explains the client architecture of PPA Control, including how single devices, multi-client management, and discovery work together.

## Overview

The client architecture consists of three main components:
1. Single Device Client (`SingleDevice`)
2. Multi-Client Manager (`MultiClient`)
3. Discovery System

## Single Device Client

A single device client handles communication with one PPA device. It implements the `Client` interface:

```go
type Client interface {
    SendPing()
    SendPresetRecallByPresetIndex(index int)
    SendMasterVolume(volume float32)
    Run(ctx context.Context, receivedCh *chan ReceivedMessage) error
    Name() string
}
```

Each client maintains its own connection to a device and handles:
- Command sending (ping, preset recall, volume control)
- Message receiving
- Connection lifecycle management

## Multi-Client Manager

The `MultiClient` acts as an orchestrator for multiple single device clients. Key features:

```go
type MultiClient struct {
    clients    map[string]Client       // Maps address to client
    cancels    map[string]context.CancelFunc
    receivedCh chan ReceivedMessage    // Buffered channel (size 10)
    // ... other fields
}
```

### Client Management
- Dynamically adds/removes clients
- Thread-safe operations using mutex
- Broadcasts commands to all connected devices
- Aggregates received messages into a single channel

### Message Flow
1. Each single client writes to MultiClient's `receivedCh`
2. MultiClient forwards messages to the application's channel
3. When stopping:
   ```go
   func (mc *MultiClient) Run(ctx context.Context, receivedCh *chan ReceivedMessage) error {
       // ... context done case ...
       // Cancel all clients
       for _, cancel := range mc.cancels {
           cancel()
       }
       // Wait for all clients to finish
       mc.wg.Wait()
       // Channel cleanup handled by Go's GC
   }
   ```

## Discovery System

The discovery system dynamically finds PPA devices on the network:

1. Sends discovery messages on specified interfaces
2. Notifies about discovered/lost peers via channel:
   ```go
   type PeerInformation interface {
       GetAddress() string
       GetInterface() string
   }
   ```

3. Integration with MultiClient:
   ```go
   // When peer discovered
   multiClient.AddClient(ctx, msg.GetAddress(), msg.GetInterface(), componentId)
   
   // When peer lost
   multiClient.CancelClient(msg.GetAddress())
   ```

## Lifecycle Management

1. **Startup**:
   - Create MultiClient
   - Start discovery (if enabled)
   - Add static clients (from addresses)
   - Start message handling

2. **Runtime**:
   - Dynamic client addition/removal
   - Command broadcasting
   - Message aggregation

3. **Shutdown**:
   - Context cancellation propagates to all clients
   - Wait for client goroutines to finish
   - Clean channel shutdown

## Error Handling

- Client errors are logged but don't stop other clients
- Context cancellation provides clean shutdown
- Thread-safe operations prevent race conditions
- Buffered channels prevent message loss

## Best Practices

1. Always use context for lifecycle management
2. Handle discovery events promptly
3. Use proper error handling in client implementations
4. Maintain thread safety when managing clients
5. Clean up resources on shutdown 