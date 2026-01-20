# Aeterna: Process Hot-Swap & State Handover Tool

Aeterna is a lightweight process supervisor (PID 1) designed to handle **hot-swaps** for stateful applications without dropping TCP connections or losing in-memory context.

It is **not** a "magic protocol." It is a wrapper around standard Unix system calls (`fork`, `exec`, `SCM_RIGHTS`, `memfd_create`) to solve a specific problem: updating Python-based AI agents without forcing a cold boot or context re-computation.

## The Problem

Standard Kubernetes rolling updates (`SIGTERM` -> `SIGKILL`) are destructive for stateful agents.

1. **Connection Loss:** Clients are disconnected. Reconnecting 10k clients creates a thundering herd.
2. **Memory Loss:** An AI agent's "short-term memory" (Python list/dict, PyTorch tensors) is wiped. Re-computing context is CPU/GPU expensive.

Existing tools (Nginx, systemd) handle socket inheritance but do not provide a standard mechanism for **application-level state handover** before the old process exits. Aeterna fills that gap.

## Implementation Details

No magic. Just plumbing.

### 1. Socket Inheritance (Zero-Downtime)

We do not use a proxy. We use **File Descriptor Passing**.

* **Mechanism:** Aeterna holds the listener socket. When a reload is triggered, it forks the new process and passes the listener FD (typically FD 3) via environment variables.
* **Syscall:** The child process uses `python: socket.fromfd(3, ...)` instead of `bind()`.
* **Result:** The kernel queue for the socket remains intact. The TCP handshake is never reset.

### 2. State Handover (Zero-Amnesia)

*Formerly referred to as "SRP" (marketing fluff), this is just IPC.*

We support two modes of passing state from Old PID -> New PID:

* **Mode A: Pipe (Small Data)**
* Uses a Unix Domain Socket to stream JSON/Pickle data.
* Suitable for session IDs, simple conversation history (< 1MB).


* **Mode B: Shared Memory (Large Tensors)**
* **Mechanism:** Uses `memfd_create` (Linux only) or `shm_open` to allocate a RAM block.
* The file descriptor for this memory block is passed to the child process.
* **Result:** Zero-copy transfer. The new process `mmap`s the existing memory region. The old process unmaps and exits.



## Supported Targets

**Current Focus: Python (Linux/macOS)**

* We prioritize Python because AI Agents (PyTorch/LangChain) have the most expensive initialization costs.
* *Note:* Python's GIL and memory model make "true" thread migration impossible. We only migrate **serializable data** and **raw memory buffers**.

**Status of other languages:**

* **Go/Rust:** Technically supported via the generic FD passing, but you must implement the state serialization logic yourself.
* **Java:** **NOT SUPPORTED.** The JVM heap is too complex to serialize efficiently for this approach. Don't ask.

## Performance Benchmarks

Don't trust "millisecond" claims. Run the benchmarks yourself: `make test-benchmark`.

**Test Environment:** AMD Ryzen 9 5950X, Linux 6.5, Python 3.11.

| Scenario | State Size | Mechanism | Handover Latency (Time to Ready) |
| --- | --- | --- | --- |
| **Simple Chat Agent** | 50 KB (JSON) | Unix Socket Pipe | ~15 ms |
| **RAG Vector Cache** | 100 MB (Binary) | Shared Memory (`memfd`) | ~40 ms |
| **Large Context** | 1 GB (Tensors) | Shared Memory (`memfd`) | ~120 ms* |
| **Cold Boot (Baseline)** | N/A | Disk Load + Init | **4500 ms** |

**Note: 120ms includes the time to verify memory integrity, not just the map switching.*

## Usage

### 1. Build

```bash
make build

```

### 2. Configure (Simple YAML)

Don't over-engineer the config.

```yaml
# aeterna.yaml
service:
  command: ["python3", "my_agent.py"]
  env:
    - PORT=8080

handoff:
  # Use 'pipe' for small text, 'shm' for large tensors
  mechanism: pipe 
  timeout: 5s

```

### 3. Run

```bash
./bin/aeterna start -c aeterna.yaml

```

To trigger a hot swap:

```bash
pkill -HUP aeterna

```