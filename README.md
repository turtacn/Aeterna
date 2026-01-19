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
</div>

---

## 📖 Mission / 核心使命

To democratize **Zero-Downtime In-Place Evolution** for every backend service and AI Agent. Aeterna acts as a universal PID 1 supervisor that manages socket inheritance, state handoff, and canary soaking, ensuring your services evolve without ever dropping a connection or losing context.

致力于让 **“零中断原地进化”** 成为所有后端服务和 AI 智能体的标配。Aeterna 作为一个通用的 PID 1 守护进程，管理 Socket 继承、内存状态接力（State Handoff）以及金丝雀浸泡，确保您的服务在迭代升级时，连接不断、记忆不丢。

## 🚀 Why Aeterna? / 核心价值

In the era of **Agentic AI** and **Real-time Services**, standard Kubernetes Rolling Updates are disruptive:
在 **Agentic AI（智能体）** 和 **实时服务** 时代，传统的 K8s 滚动更新存在严重缺陷：

1.  **Connection Severing (连接中断):** Killing a Pod disconnects all active users. (销毁 Pod 意味着断开所有在线用户)
2.  **Context Amnesia (上下文遗忘):** AI Agents lose their in-memory thought chains and cache. (AI Agent 会丢失内存中的思维链和缓存)
3.  **Cold Starts (冷启动):** New processes take time to warm up. (新进程预热耗时漫长)

**Aeterna** solves this by orchestrating the update **inside the container**. It treats the process as ephemeral but the connections and state as persistent assets.

## ✨ Key Features / 功能特性

* **⚡ Sub-millisecond Handover (毫秒级接力):** Updates happen at process fork speed.
* **🔌 Socket Inheritance (Socket 继承):** Seamlessly passes TCP/UDP/Unix listeners to the new version.
* **🧠 State Relay Protocol (SRP - 状态接力协议):** Uniquely designed for AI Agents to transfer in-memory context (Context Windows, RAG Cache) to the new process via IPC before exiting.
* **🛡️ Orchestrated Safety (编排式安全):** Built-in **Pre-flight Checks**, **Canary Soaking**, and **Auto-Rollback**.
* **🌐 Polyglot Support (多语言支持):** Works with Go, Python (AI-First), Java, Rust, etc.

## 🛠️ Getting Started / 快速开始

### Installation / 安装

```bash
go install [github.com/turtacn/Aeterna/cmd/aeterna@latest](https://github.com/turtacn/Aeterna/cmd/aeterna@latest)

```

### Basic Usage / 基本使用

Aeterna runs as the entrypoint of your container.
Aeterna 作为容器的入口点运行。

**1. Configuration (`aeterna.yaml`):**

```yaml
service:
  name: "agent-core"
  command: ["python", "agent.py"]

orchestration:
  mode: "hot-relay"
  soak_time: "30s"      # Soak time for canary observation (浸泡观察时间)
  state_handoff: true   # Enable memory context transfer (开启状态接力)

```

**2. Run / 启动:**

```bash
aeterna start -c aeterna.yaml

```

**3. Trigger Update / 触发升级:**

Replace your binary or script, then run:
替换二进制文件或脚本后运行：

```bash
aeterna reload

```

## 🤝 Contributing / 贡献指南

We welcome contributors! Please see [CONTRIBUTING.md](https://www.google.com/search?q=CONTRIBUTING.md) for details.

## 📄 License / 许可证

Distributed under the Apache 2.0 License. See `LICENSE` for more information.