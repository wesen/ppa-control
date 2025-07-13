# PPA-Control CLI Migration - QA Validation Guide

This guide helps validate the complete CLI migration to the glazed framework.

## Migration Summary

**What was done:**
- Migrated all CLI commands from Cobra to glazed framework
- Removed legacy command duplications and simulation functionality
- Added structured output support (JSON, CSV, table, YAML, etc.)
- Preserved backward compatibility with legacy flags
- Implemented dual-mode commands (classic logging + structured output)

**Commands available:**
- `ping` - Device ping with dual output modes
- `volume` - Volume control with structured output
- `recall` - Preset recall with structured output  
- `udp-broadcast` - UDP testing with dual output modes

---

## Pre-QA Setup

1. **Build the CLI:**
   ```bash
   cd ppa-control
   go build ./cmd/ppa-cli
   ```

2. **Verify build success:**
   ```bash
   ./ppa-cli --help
   ```
   Should show 4 commands: ping, volume, recall, udp-broadcast

---

## QA Test Cases

### 1. Help System Validation

**Test 1.1: Root Help**
```bash
./ppa-cli --help
```
**Expected:** Shows 4 commands without "-glazed" suffix, clean descriptions

**Test 1.2: Command Help**
```bash
./ppa-cli ping --help
./ppa-cli volume --help
./ppa-cli recall --help
./ppa-cli udp-broadcast --help
```
**Expected:** Each shows appropriate flags including `--structured-output`

---

### 2. Legacy Compatibility Tests

**Test 2.1: Legacy componentId flag**
```bash
./ppa-cli ping --componentId 200 --addresses 127.0.0.1 --log-level error
```
**Expected:** Accepts legacy camelCase flag without error

**Test 2.2: Modern component-id flag**
```bash
./ppa-cli ping --component-id 200 --addresses 127.0.0.1 --log-level error
```
**Expected:** Accepts kebab-case flag without error

**Test 2.3: Discovery defaults**
```bash
./ppa-cli volume --volume 0.5 --log-level error
```
**Expected:** Enables discovery by default (should see broadcast messages in logs)

**Test 2.4: UDP broadcast flag compatibility**
```bash
./ppa-cli udp-broadcast --address 127.0.0.1 --port 8080 --log-level error
```
**Expected:** Accepts -a and -p flags (legacy compatibility)

---

### 3. Structured Output Tests

**Test 3.1: JSON Output**
```bash
timeout 5s ./ppa-cli ping --addresses 127.0.0.1 --structured-output --output json --log-level error
```
**Expected:** Outputs valid JSON array with structured events

**Test 3.2: Table Output**
```bash
timeout 5s ./ppa-cli ping --addresses 127.0.0.1 --structured-output --output table --log-level error
```
**Expected:** Outputs formatted ASCII table with columns

**Test 3.3: CSV Output**
```bash
timeout 5s ./ppa-cli ping --addresses 127.0.0.1 --structured-output --output csv --log-level error
```
**Expected:** Outputs CSV format with headers

**Test 3.4: Volume Structured Output**
```bash
./ppa-cli volume --addresses 127.0.0.1 --volume 0.5 --structured-output --output json --log-level error
```
**Expected:** JSON with volume_command_sent event

**Test 3.5: Recall Structured Output**
```bash
./ppa-cli recall --addresses 127.0.0.1 --preset 5 --structured-output --output json --log-level error
```
**Expected:** JSON with recall_initiated/recall_sent events

---

### 4. Classic Mode Tests

**Test 4.1: Ping Classic Mode**
```bash
timeout 5s ./ppa-cli ping --addresses 127.0.0.1 --log-level info
```
**Expected:** Human-readable log messages, no JSON output

**Test 4.2: UDP Broadcast Classic Mode**
```bash
./ppa-cli udp-broadcast --address 127.0.0.1 --port 8080 --log-level info
```
**Expected:** Log messages showing socket creation and message sending

---

### 5. Real Device Tests (if available)

**Prerequisites:** Have a real PPA device at known IP (e.g., 192.168.0.200)

**Test 5.1: Real Device Ping**
```bash
timeout 10s ./ppa-cli ping --addresses 192.168.0.200 --structured-output --output json
```
**Expected:** JSON events showing actual device responses

**Test 5.2: Real Device Volume**
```bash
./ppa-cli volume --addresses 192.168.0.200 --volume 0.3 --structured-output --output table
```
**Expected:** Table showing volume command sent and device response

**Test 5.3: Real Device Discovery**
```bash
timeout 10s ./ppa-cli ping --discover --structured-output --output json
```
**Expected:** Discovery events and responses from real devices

---

### 6. Parameter Validation Tests

**Test 6.1: Volume Range Validation**
```bash
./ppa-cli volume --addresses 127.0.0.1 --volume 2.0
```
**Expected:** Error message about volume range (0.0-1.0)

**Test 6.2: Required Parameters**
```bash
./ppa-cli volume --addresses 127.0.0.1
```
**Expected:** Error about missing required --volume parameter

**Test 6.3: Invalid Preset**
```bash
./ppa-cli recall --addresses 127.0.0.1 --preset -1
```
**Expected:** Validation error for negative preset

---

### 7. Advanced Output Features

**Test 7.1: Field Selection**
```bash
timeout 5s ./ppa-cli ping --addresses 127.0.0.1 --structured-output --output table --fields timestamp,event,from
```
**Expected:** Table with only specified columns

**Test 7.2: Sorting**
```bash
timeout 5s ./ppa-cli ping --addresses 127.0.0.1 --structured-output --output table --sort-by timestamp
```
**Expected:** Table sorted by timestamp

**Test 7.3: Filtering**
```bash
timeout 5s ./ppa-cli ping --addresses 127.0.0.1 --structured-output --output json --filter client
```
**Expected:** JSON without client field

---

### 8. Error Handling Tests

**Test 8.1: Invalid Address**
```bash
timeout 5s ./ppa-cli ping --addresses 999.999.999.999 --log-level error
```
**Expected:** Graceful handling, no crash

**Test 8.2: Invalid Output Format**
```bash
./ppa-cli ping --addresses 127.0.0.1 --structured-output --output invalid
```
**Expected:** Error message about invalid output format

**Test 8.3: Network Interface Not Found**
```bash
./ppa-cli udp-broadcast --interface nonexistent --log-level error
```
**Expected:** Graceful error handling

---

## Expected Behaviors

### ✅ Pass Criteria

1. **All commands build and run without panic**
2. **Help system shows clean command names**
3. **Legacy flags (componentId, -a, -p) work**
4. **Structured output produces valid JSON/CSV/table**
5. **Classic mode shows human-readable logs**
6. **Parameter validation works correctly**
7. **Real device communication functions**
8. **Discovery defaults to enabled**
9. **All output formats render properly**
10. **Error messages are helpful**

### ❌ Fail Criteria

1. **Commands crash or panic**
2. **Legacy flags rejected**
3. **Invalid JSON/CSV output**
4. **No structured events generated**
5. **Discovery not working by default**
6. **Real device communication broken**
7. **Help system shows "-glazed" commands**
8. **Parameter validation missing**

---

## Issue Reporting Template

If you find issues, report them with:

```
**Test:** [Test name from above]
**Command:** [Exact command run]
**Expected:** [What should happen]
**Actual:** [What actually happened]
**Error Output:** [Any error messages]
**Environment:** [OS, Go version, network setup]
```

---

## Success Metrics

- [ ] All 40+ test cases pass
- [ ] No command crashes or panics
- [ ] Legacy compatibility maintained
- [ ] Structured output works in all formats
- [ ] Real device communication successful
- [ ] Help system clean and accurate
- [ ] Parameter validation comprehensive
- [ ] Error handling graceful

## Final Validation Checklist

- [ ] `go build ./cmd/ppa-cli` succeeds
- [ ] `./ppa-cli --help` shows 4 clean commands
- [ ] All legacy flags accepted (componentId, -a, -p)
- [ ] JSON output is valid and structured
- [ ] CSV output has proper headers
- [ ] Table output is formatted correctly
- [ ] Real device responds to commands
- [ ] Discovery works by default
- [ ] Parameter validation catches errors
- [ ] Classic mode produces readable logs

**When all checkboxes are ✅, the migration is validated successfully!**
