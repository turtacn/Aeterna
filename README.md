# Aeterna

<div align="center">
  <!--
  <img src="docs/images/logo.png" alt="Aeterna Logo" width="200" height="200">
  <br /> 
  -->
  <h1>Aeterna</h1>
  <p><strong>The Eternal Process Orchestrator for Agentic AI & High-Availability Services</strong></p>
  <p><strong>面向 Agentic AI 与高可用服务的"永恒"进程编排引擎</strong></p>
  
  
  
  [![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/turtacn/Aeterna/actions)
  [![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
  [![Go Report Card](https://goreportcard.com/badge/github.com/turtacn/Aeterna)](https://goreportcard.com/report/github.com/turtacn/Aeterna)
  
  [Quick Start](#quick-start) • [Architecture](docs/architecture.md) • [API Docs](docs/apis.md) • [SDK Design](docs/sdk_design.md) 
</div>

---

## Background / 背景

云原生环境下的有状态服务连续性挑战与 Aeterna 架构解析

**1. 基础设施层的固有矛盾：不可变性与服务连续性**
在基于 Kubernetes 的现代云原生编排体系中，“不可变基础设施（Immutable Infrastructure）”范式确立了以容器为最小部署单元的标准。然而，该范式通常采用“销毁-重建”的滚动更新（Rolling Update）策略，这在处理长连接与高初始化成本应用时存在显著的架构缺陷：

* **TCP 连接有损重置：** 对于依赖 WebSocket、gRPC 长连接的网关及金融实时交易系统，容器生命周期的终止直接导致传输层连接断开（Connection Reset）。在大规模并发场景下，这会诱发客户端的重连风暴（Thundering Herd Problem），导致服务可用性抖动。
* **高昂的初始化时延（Cold Start Latency）：** 对于基于 JVM 的大型应用或加载庞大权重的 AI 推理服务，进程启动涉及复杂的类加载、JIT 编译预热及显存数据加载。标准的 Pod 重启机制会导致服务在数秒至数分钟内处于不可服务（Unavailable）或性能降级状态。

**2. 领域挑战：Agentic AI 的易失性内存状态管理**
随着人工智能从无状态推理向 Agentic AI（智能体）演进，应用架构的本质发生了从 Stateless 向 Stateful 的转变。现代 Agent 依赖于驻留在内存中的高维状态，包括思维链（Chain of Thought, CoT）中间结果、会话上下文窗口（Context Window）及短期记忆索引。

* **状态易失性风险：** 在传统的容器编排逻辑下，运维侧的安全补丁更新或版本发布（Deployment Rollout）会触发进程终止信号（SIGTERM），导致内存堆栈被强制回收。
* **计算成本与体验损耗：** 内存状态的丢失迫使 Agent 重新处理原始输入以重建推理上下文（Re-computation）。这不仅造成了计算资源（GPU/TPU）的冗余消耗，更因丢失交互历史而破坏了用户体验的连续性。

**3. 架构解决方案：基于 PID 1 的进程热接力（In-Place Hot Relay）**
针对上述问题，Aeterna 提出了一种基于容器内进程编排的解决方案，旨在实现计算逻辑与运行时资产（网络连接、内存状态）的解耦。作为容器内的 PID 1 初始化进程，Aeterna 引入了以下核心机制：

* **文件描述符传递（File Descriptor Passing）：** 利用 Unix Domain Socket 的 `SCM_RIGHTS` 辅助消息功能，在父子进程间原子性地传递监听 Socket 的文件描述符。此机制确保在进程二进制文件更新期间，TCP 连接保持 ESTABLISHED 状态，实现对客户端透明的热升级。
* **状态接力协议（State Relay Protocol, SRP）：** 定义了一套标准化的进程间通信（IPC）协议，用于易失性状态的序列化与迁移。在旧进程终止前，通过共享内存（Shared Memory）或管道，将关键业务状态（如 AI 上下文向量、Session 缓存）传输至新启动的进程，实现应用层状态的无损继承（State Handover）。


## Mission / 主要作用

To democratize **Zero-Downtime In-Place Evolution** for every backend service and AI Agent. Aeterna acts as a universal PID 1 supervisor that manages socket inheritance, state handoff, and canary soaking, ensuring your services evolve without ever dropping a connection or losing context.

致力于让 **“零中断原地进化”** 成为所有后端服务和 AI 智能体的标配。Aeterna（`Eternal Uptime`） 作为一个通用的 PID 1 守护进程，管理 Socket 继承、内存状态接力（State Handoff）以及金丝雀浸泡，确保您的服务在迭代升级时，连接不断、记忆不丢。

## Why Aeterna? / 核心价值

In the era of **Agentic AI** and **Real-time Services**, standard Kubernetes Rolling Updates are disruptive:
在 **Agentic AI（智能体）** 和 **实时服务** 时代，传统的 K8s 滚动更新存在严重缺陷：

1.  **Connection Severing (连接中断):** Killing a Pod disconnects all active users. (销毁 Pod 意味着断开所有在线用户)
2.  **Context Amnesia (上下文遗忘):** AI Agents lose their in-memory thought chains and cache. (AI Agent 会丢失内存中的思维链和缓存)
3.  **Cold Starts (冷启动):** New processes take time to warm up. (新进程预热耗时漫长)

**Aeterna** solves this by orchestrating the update **inside the container**. It treats the process as ephemeral but the connections and state as persistent assets.

## Key Features / 功能特性

* **Sub-millisecond Handover (毫秒级接力):** Updates happen at process fork speed.
* **Socket Inheritance (Socket 继承):** Seamlessly passes TCP/UDP/Unix listeners to the new version.
* **State Relay Protocol (SRP - 状态接力协议):** Uniquely designed for AI Agents to transfer in-memory context (Context Windows, RAG Cache) to the new process via IPC before exiting.
* **️Orchestrated Safety (编排式安全):** Built-in **Pre-flight Checks**, **Canary Soaking**, and **Auto-Rollback**.
* **Polyglot Support (多语言支持):** Works with Go, Python (AI-First), Java, Rust, etc.

## Getting Started / 快速开始

### Installation / 安装

```bash
go install github.com/turtacn/Aeterna/cmd/aeterna@latest
```

### Basic Usage / 基本使用

Aeterna runs as the entrypoint of your container.
Aeterna 作为容器的入口点运行。

**1. Configuration (`aeterna.yaml`):**

```yaml
version: "1.0"
service:
  name: "agent-core"
  command: ["python", "agent.py"]
  env:
    - "PYTHONUNBUFFERED=1"

orchestration:
  strategy: "hot-relay"
  canary:
    enabled: true
    soak_time: "30s"    # Soak time for canary observation (浸泡观察时间)
  state_handoff:
    enabled: true       # Enable memory context transfer (开启状态接力)
    socket_path: "/tmp/aeterna_srp.sock"

observability:
  metrics_port: ":9090"
  log_level: "info"
```

**2. Run / 启动:**

```bash
aeterna start -c aeterna.yaml
```

**3. Trigger Update / 触发升级:**

To trigger a hot reload, send a `SIGHUP` signal to the Aeterna process:
要触发热重载，请向 Aeterna 进程发送 `SIGHUP` 信号：

```bash
kill -HUP $(pgrep aeterna)
```

## Architecture / 架构

Aeterna operates as a PID 1 supervisor within a container, managing the lifecycle of the underlying business process.

1.  **Phase 1: Pre-flight Checks**: Runs user-defined hooks to ensure the environment is ready for an update.
2.  **Phase 2: Startup**: Launches the new version of the process.
3.  **Phase 2.5: SRP Handover**: Facilitates state transfer via the State Relay Protocol.
4.  **Phase 3: Soaking**: Observes the new process for a specified duration (Canary phase).
5.  **Phase 5: Drain**: Gracefully shuts down the old process once the new one is confirmed stable.

## Development / 开发

### Prerequisites
- Go 1.21+
- Python 3.9+ (for SDK and AI Agent examples)

### Running Tests
```bash
go test ./...
```

### Generating Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## SDK Integration / SDK 集成

### Python (AI Agents)
```python
from aeterna import AeternaClient

client = AeternaClient()

# Inherit the listening socket
listener = client.get_listener_socket()

# Load memory context from the previous process
context = client.load_context()

# ... Your Agent Logic ...
```


## License / 许可证

Distributed under the Apache 2.0 License. See `LICENSE` for more information.