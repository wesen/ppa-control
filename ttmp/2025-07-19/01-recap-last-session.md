Here’s what we did last week for the **PPA protocol reverse engineering** work, focused around PCAP sniffing and analysis, and which documents are relevant to that workflow.

---

## ✅ Summary of Last Week’s Work

### 🧪 PCAP Capture & Analysis Setup

We built a **complete capture and analysis toolchain** inside `ppa-control/ttmp/2025-07-13/pcap/`:

* 📥 **Recording scripts**:

  * `quick-capture.sh`: Fast capture (default 30s)
  * `record-ppa-session.sh`: Full session capture
  * `record-specific-operations.sh`: Guided capture per operation (e.g. mute, volume, preset)

* 🔍 **Analysis tools**:

  * `analyze-ppa-captures.sh`: Generates message type, status, sequence, and payload analysis
  * `test-protocol-implementation.sh`: Replays known commands via `ppa-cli` and captures responses
  * `protocol-fuzzer.sh`: Sends edge-case/invalid commands to capture error responses

All of these are documented in:

* 📄 `ppa-control/ttmp/2025-07-13/pcap/README.md`

---

### 🐞 Bug Investigation: SLL2 Format

You discovered that the pcap tool was **not analyzing packets properly**.

Cause:

* We used `tcpdump -i any`, which produces **LINUX\_SLL2** (Linux cooked v2) format
* `gopacket` didn't decode these, leading to `DecodeFailure` and silent drops

Fixes:

* Switched to default interface (e.g. `wlp170s0`) via:

  ```bash
  DEFAULT_INTERFACE=$(ip route | grep default | awk '{print $5}')
  ```
* Removed debug `fmt.Printf("Opening...")` calls that broke `--output-format jsonl` parsing

Full write-up:

* 📄 `ppa-control/ttmp/2025-07-13/02-explanation-why-pcap-sll2-was-not-working.md`

---

### 🔎 Protocol Discovery (Based on Captures)

From analyzing several `.pcap` files using our tools:

* **Message Types Found**:

  * `0` = Ping
  * `2` = DeviceData
  * `4` = PresetRecall
  * `3`, `6`, `9`, `10` = Undocumented but present (especially 10 = streaming)

* **Status Codes**:

  * `0x0101`, `0x0102`, `0x0001`, `0x0002`, plus undocumented `0x0000`

* **Streaming Behavior**:

  * Type 10 packets seen at \~9Hz
  * 144-byte payloads → suspected real-time metering

Analysis summaries live in:

* 📁 `ppa-control/ttmp/2025-07-13/pcap-analysis/`

  * `summary-report.md` — high-level overview
  * `connection-analysis.md`, `mute-operations-analysis.md`, etc. — per operation
  * `official-doc-analysis.md` — comparing the Simon Hoffmann doc to actual traffic
  * `oracle-protocol-analysis.md` — detailed inferred v3 protocol structure

---

### 📜 Document References Used

These two were key:

* 🧾 `ppa-control/doc/PPA commands_SimonHoffmann.pdf`

  * Only covers v2 (Ping, DeviceData, PresetRecall, LiveCmd)
  * Lacks streaming/message fragmentation
* 📄 `ppa-control/doc/PPA commands_SimonHoffmann.txt`

  * Text dump used for faster cross-referencing with PCAP captures

---

### 🔮 Oracle Inference: PPA v3 Protocol

The Oracle helped identify:

* **Type 10**: `LiveStatusStream` — channel metering, float32s
* **Type 6**: `BulkParameterBlock` — device config dumps (fragmented)
* **Type 3**: `PresetDirectory` — sends preset list
* **Type 9**: `TransactionControl` — commit/rollback wrapper
* **Deprecation**: Type 1 (LiveCmd) seems unused in new clients

📄 `oracle-protocol-analysis.md` has the full breakdown.

---

## 📌 Suggested Next Actions

Here’s what you could record next to extend the reverse engineering effort:

| Scenario                | Script to Use                   | Goal                                    |
| ----------------------- | ------------------------------- | --------------------------------------- |
| Full startup            | `record-ppa-session.sh`         | Baseline: all types, handshake, idle    |
| Live metering (Type 10) | `quick-capture.sh` (60s)        | Capture metering stream, float decoding |
| Preset recall           | `record-specific-operations.sh` | Confirm Type 4 behavior                 |
| Force LiveCmd (Type 1)  | Use CLI or old client           | See if it’s still supported             |
| Config save/load        | `record-specific-operations.sh` | Trigger Type 6 / 9 activity             |

Let me know when you’ve recorded more, and I’ll analyze them!

