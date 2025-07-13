https://chatgpt.com/c/687411ee-6524-8012-aa40-dc863c9eb9bf

---

Below is a **clean‑slate event‑driven design** that keeps all of the domain logic you already wrote (ping, preset recall, master‑volume, discovery, simulation, etc.) but **re‑packages every interaction as Watermill messages**.
Nothing in the design depends on a specific broker—during development you can use Watermill’s in‑process **gochannel** implementation and later swap to Kafka, NATS, Google Cloud Pub/Sub … without touching the services themselves.

---

## 1. Messages first – a very small contract

```go
// ─── COMMANDS ────────────────────────────────────────────────────────────
type PingCmd struct { DeviceID string }              // ""  == broadcast
type PresetRecallCmd struct { DeviceID string; Index int }
type MasterVolumeCmd struct { DeviceID string; Volume float32 }

// ─── EVENTS ──────────────────────────────────────────────────────────────
type DeviceDiscovered struct { DeviceID, Addr, IFace string }
type DeviceLost       struct { DeviceID, Addr, IFace string }

type UdpPacketIn  struct { DeviceID string; Raw []byte } // parsed later
type UdpPacketOut struct { DeviceID string; Raw []byte } // debugging tap

// optional high‑level semantic events produced by decoding UdpPacketIn
type PongReceived    struct { DeviceID string }
type PresetRecalled  struct { DeviceID string; Index int }
```

All structs are plain Go, **encoded with Watermill’s JSON marshaler**.
Topics (channel names) follow a single rule:

```
commands.device.<id>        – command for one device
commands.broadcast           – command for everyone
events.discovery             – discovered / lost devices
events.udp.in.<id>           – raw UDP ➜ bus
events.udp.out.<id>          – raw bus ➜ UDP (debug only)
events.device.<id>           – decoded high‑level events (optional)
```

---

## 2. High‑level component map

```
                           +----------------------+
                           |      CLI / API       |
                           |  (publishes commands |
                           |   subscribes events) |
                           +----------+-----------+
                                      |
                                  commands.*
                                      |
             +------------------------+---------------------+
             |                                              |
+------------v-----------+                      +-----------v-----------+
|  Device‑Command Router |                      |  Device‑Event Router  |
|  (Watermill Router)    |                      |  (Watermill Router)   |
+------------+-----------+                      +-----------+-----------+
             |                                              |
┌────────────v─────────────┐                   ┌────────────v───────────┐
|  Device‑Gateway <id=N>   | …  one per device |  Device‑State Project. | (optional)
|  ↔ UDP socket            |                   |  in‑memory / Redis / DB|
└────────────┬─────────────┘                   └────────────┬───────────┘
             |                                              |
             | discovery publishes DeviceDiscovered/Lost    |
             |                                              |
        +----v----+                                         |
        |Discover |                                         |
        +---------+                                         |
```

*Every box is an independent process/goroutine with its own Watermill **Router***.

---

## 3. Core building blocks

### 3.1 Discovery service

* **Inputs**: none
* **Outputs**: `events.discovery`
* **Behaviour**: re‑use your existing interface scan and peer‑timeout logic, but **publish `DeviceDiscovered` / `DeviceLost`** instead of calling `MultiClient.Add/Cancel`.

### 3.2 Device‑Gateway (one instance per network device)

* **Inputs**

  * `commands.device.<id>` – unicasts
  * `commands.broadcast`    – fan‑out
* **Outputs**

  * `events.udp.out.<id>` (optional tap)
  * `events.udp.in.<id>`   – every packet read from the UDP socket
* **Lifecycle**

  * Spun up by a **factory** that subscribes to `events.discovery` and starts/stops gateways dynamically.

*Replacing `SingleDevice`:*

```go
router.AddHandler(
    "udp‑send",
    "commands.device."+id,
    pubSub,                      // input
    "", "",                      // no output
    pubSub,
    func(msg *message.Message) error {
        var cmd PingCmd
        _ = json.Unmarshal(msg.Payload, &cmd)

        // translate cmd -> bytes (reuse existing encoder)
        udpConn.WriteTo(...)          // no retry; Watermill nack==requeue
        // Optionally publish tap:
        pubSub.Publish("events.udp.out."+id, message.NewMsg(uuid.New().String(), rawBytes))

        msg.Ack()
        return nil
    })
```

### 3.3 Device‑Command Router

A tiny Watermill router that **forks** commands:

```
commands.broadcast  ─┐
                     ├─▶ fan‑out to every   → commands.device.<id>
commands.device.<id> ┘
```

Because Watermill routers can publish inside a handler, no custom MultiClient map is needed—the fan‑out handler simply iterates over the currently known device IDs kept in memory (updated via `events.discovery`).

### 3.4 Device‑Event Router

Subscribes to `events.udp.in.<id>`, runs your parser (`protocol.ParseHeader`) and publishes higher‑level events (`PongReceived`, `PresetRecalled` …) to `events.device.<id>` for anybody interested (CLI, UI, databases).

### 3.5 Optional projections

If you need current device state, spin up a read‑model service that consumes `events.device.*` and stores the last known status in Redis, a DB, or just in memory.

---

## 4. CLI / Glazed commands rewritten

```go
func (c *PingCommand) Run(ctx context.Context, layers *layers.ParsedLayers) error {
    pubSub := gochannel.NewGoChannel(gochannel.Config{Buffered: true})
    router  := message.NewRouter(routerConfig)

    // Fire initial discovery (optional)
    pubSub.Publish("commands.broadcast",
        message.NewMessage(uuid.New(), mustJSON(PingCmd{})))

    // Subscribe to responses
    router.AddNoPublisherHandler(
        "print‑pongs",
        "events.device.#",   // wildcard
        pubSub,
        func(m *message.Message) error {
            var e PongReceived
            _ = json.Unmarshal(m.Payload, &e)
            fmt.Printf("PONG from %s\n", e.DeviceID)
            return nil
        })

    return router.Run(ctx)
}
```

No `MultiClient`, no shared channels, no goroutine book‑keeping—the router plus Pub/Sub glue everything.

---

## 5. How the new architecture solves your pain points

| Current pain                                                 | Event‑driven replacement                                                                                                         |
| ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| `MultiClient` must track a map and fan‑out commands manually | Fan‑out is just *publishing once* to `commands.broadcast`. Watermill delivers the same message to every active device gateway.   |
| Complex shutdown with `sync.WaitGroup`                       | Each router runs under its own context; cancel the context and Watermill closes handlers gracefully.                             |
| Coupling between discovery and client creation               | Discovery simply emits events; a small factory subscribes and starts gateways.                                                   |
| Adding a simulation layer required extra hooks               | A simulator is just **another subscriber** to `commands.device.<id>` that writes back to `events.udp.in.<id>`. No special cases. |
| Hard to test                                                 | Swap the broker to `gochannel`, start routers in the same test, publish commands, assert events—no sockets necessary.            |

---

## 6. Migration plan in three incremental steps

1. **Introduce the message contracts** (section 1) and wrap your existing `MultiClient` with an adapter that publishes/consumes them. Everything else keeps running.
2. Replace `MultiClient` + `SingleDevice` with the new **Device‑Gateway**; at this point the adapter becomes a thin shim and can be deleted.
3. Swap discovery, CLI, simulation to publish/consume messages directly—now the entire codebase is Watermill‑native and fully decoupled.

Because each step leaves runnable binaries, you can refactor with confidence.

---

## 7. Code scaffolding

A minimal boot file common to every micro‑service:

```go
func main() {
    logger := watermill.NewStdLogger(false, false)
    router, _ := message.NewRouter(message.RouterConfig{}, logger)

    pubSub := gochannel.NewGoChannel(gochannel.Config{Buffered: true})

    // registerHandlers(router, pubSub)  ← your handlers here

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    if err := router.Run(ctx); err != nil {
        log.Fatal().Err(err).Msg("router stopped")
    }
}
```

From here each service just calls `router.AddHandler` / `AddNoPublisherHandler` as shown above.

---

### That’s the whole architecture:

* **Messages are the API**
* **Routers + Pub/Sub wires everything**
* No global maps, no custom wait‑groups, no direct function calls between packages.

Drop‑in, incremental, and perfectly aligned with Watermill’s philosophy.


---

### Yes – request/response is fully preserved

In the PPA protocol every outbound UDP frame already carries a **16‑bit `SequenceNumber`**.
We simply expose it (plus a Watermill‐level correlation ID) so that any component can match replies to the originating command without sharing memory or locks.

Below is a pragmatic template that uses **two complementary correlation keys**:

| Level                      | Field                     | Purpose                                                                                                            |
| -------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Watermill message metadata | `correlation_id` (UUID)   | Relates the *command message* published on `commands.*` with the *reply message* published on `events.device.<id>` |
| PPA wire protocol          | `SequenceNumber` (uint16) | Exactly what the amplifier sees on the wire, echoed verbatim in its response                                       |

---

## 1. Enriching the command before it leaves the bus

```go
func sendPing(router *message.Router, pubSub message.PubSub, id string) {
    seq := nextSeq()                      // uint16 atomic counter (per device)
    cmd := PingCmd{DeviceID: id, Seq: seq}

    msg := message.NewMessage(uuid.NewString(), mustJSON(cmd))
    msg.Metadata.Set("correlation_id", msg.UUID)   // self‑correlated
    msg.Metadata.Set("sequence", strconv.Itoa(int(seq)))

    pubSub.Publish("commands.device."+id, msg)     // fan‑out layer unchanged
}
```

Nothing else in the system invents a correlation ID; *exactly one* is generated at the origin and travels through every handler unchanged.

---

## 2. Gateway behaviour

```
                          Watermill                     UDP
commands.device.<id> ──▶ (1) encode & send  ───▶  device
                         (2) save   (cid,seq)◀──┐
                                                │
events.udp.in.<id> ◀── (3) parse reply ◀──  device
                         (4) match cid & publish PongReceived
```

```go
// (1) encode & (2) remember
router.AddHandler(
  "udp‑send", "commands.device."+id, pubSub, "",
  pubSub, func(msg *message.Message) error {

    seq, _ := strconv.Atoi(msg.Metadata.Get("sequence"))
    encodePing(uint16(seq), udpBuf)

    // cache by seq until timeout
    cache.Set(uint16(seq), msg.Metadata.Get("correlation_id"), cache.DefaultExpiration)

    udp.WriteTo(udpBuf.Bytes(), remote)
    msg.Ack()
    return nil
})

// (3) read & (4) correlate
router.AddHandler(
  "udp‑recv", "events.udp.in."+id, pubSub, "events.device."+id,
  pubSub, func(msg *message.Message) (*message.Message, error) {

    hdr, _ := protocol.ParseHeader(msg.Payload)
    cid, ok := cache.Get(hdr.SequenceNumber)   // returns correlation_id

    out := message.NewMessage(uuid.NewString(), mustJSON(PongReceived{
        DeviceID: id,
    }))
    if ok {
        out.Metadata.Set("correlation_id", cid.(string))
    }
    return out, nil
})
```

### Notes

* Gateway keeps only the **minimal (`seq -> correlation_id`) map** and automatically evicts entries after a configurable reply timeout.
  No goroutines block; the reply comes asynchronously on `events.device.<id>`.
* If you ever need to multiplex several wire requests before the same response arrives, store **slice of cids** per sequence.

---

## 3. Consuming the reply from a CLI / API

```go
cid := publishPing("broadcast")   // helper returns the cid it set

replySub, _ := pubSub.Subscribe(ctx, "events.device.#")
for {
    m, _ := replySub.Receive(ctx)
    if m.Metadata.Get("correlation_id") == cid {
        fmt.Println("Got my reply:", string(m.Payload))
        break
    }
}
```

Because Watermill subscription filters are topic‑based only, correlation happens client‑side by reading metadata.
If you prefer true request/response semantics (single reply channel), publish replies to:

```
events.reply.<correlation_id>
```

and then **subscribe to that exact topic**; Watermill will deliver only matching messages and you can cancel the subscription after the first one.

---

## 4. Handling missing responses

The gateway is the right place to detect timeouts:

```go
cache.OnEvicted(func(k string, v interface{}) {
    cid := v.(string)
    pubSub.Publish("events.device."+id, message.NewMessage(
        uuid.NewString(),
        mustJSON(ResponseTimedOut{DeviceID: id}),
        message.SetMetadata("correlation_id", cid),
    ))
})
```

Your upper‑level code now receives either `PongReceived` **or** `ResponseTimedOut` with the same `correlation_id`.

---

## 5. Zero‑copy simulators and tests

Simulators consume `commands.device.<id>` the same way as a gateway, but instead of writing UDP they *immediately* publish a crafted `events.udp.in.<id>` message reusing metadata from the command—so correlation works out of the box and no fixtures change when you switch between real hardware and simulated devices.

---

### Take‑away

* **Request/response is preserved**—the correlation ID lives in Watermill metadata, the sequence number remains on the wire.
* A **single tiny in‑memory map per gateway** is enough; nothing else in the system needs to track pending requests.
* Consumers can choose between filtering by metadata on a shared topic or subscribing to a dedicated reply topic for perfect 1‑to‑1 semantics.

This keeps the event‑driven architecture pure while giving you the full convenience of synchronous request/response when you need it.
