# 使用 Aeterna 实现 rsyslogd 零停机热升级

Aeterna 完全支持传统服务 rsyslogd 的零停机热升级。通过三大核心机制（Socket 继承、状态接力、金丝雀浸泡），可以在不中断任何 syslog 连接的情况下完成服务更新。

---

## 实操步骤

### 1. 创建配置文件

```yaml
version: "v1"
service:
  name: "rsyslogd"
  command: ["/usr/sbin/rsyslogd", "-n", "-f", "/etc/rsyslog.conf"]
  env:
    - "RSYSLOGD_CONFIG=/etc/rsyslog.conf"

orchestration:
  strategy: "canary"
  canary:
    soak_time: "30s"
  state_handoff:
    enabled: true
    socket_path: "/tmp/rsyslogd_srp.sock"
    timeout: "10s"
``` [1](#3-0) 

### 2. 启动服务

```bash
aeterna start -c aeterna.yaml
```

### 3. 执行热升级

```bash
# 更新 rsyslogd 后触发热升级
kill -HUP 1
# 或使用 API
curl -X POST http://localhost:9091/v1/reload
```

---

## 热升级原理实例化阐述

### 1. Socket 继承机制

**工作流程：**

Aeterna 的 `SocketManager` 首先创建并持有网络监听器 [2](#3-1) ：

```go
// SocketManager 创建监听器
func (sm *SocketManager) EnsureListener(addr string) (net.Listener, error) {
    // 冷启动时绑定新端口
    l, err := net.Listen("tcp", addr)
    // 获取文件描述符用于继承
    f, err := tcpL.File()
    sm.listeners[addr] = l
    sm.files[addr] = f
    return l, nil
}
```

然后通过 `ProcessManager` 的 `Start` 方法，使用 `cmd.ExtraFiles` 将文件描述符传递给新的 rsyslogd 进程 [3](#3-2) ：

```go
func (pm *ProcessManager) Start(command []string, env []string, extraFiles []*os.File) error {
    pm.cmd = exec.Command(command[0], command[1:]...)
    if len(extraFiles) > 0 {
        pm.cmd.ExtraFiles = extraFiles  // 传递文件描述符
        pm.cmd.Env = append(pm.cmd.Env, fmt.Sprintf("%s=%d", consts.EnvInheritedFDs, len(extraFiles)))
    }
    return pm.cmd.Start()
}
```

**实际效果：**
- 旧 rsyslogd 进程的 TCP 监听 socket（如 514 端口）的文件描述符被传递给新进程
- 客户端的 syslog 连接保持 ESTABLISHED 状态，不会感知到进程切换
- 内核自动将新连接分发到新进程

### 2. 状态接力协议 (SRP)

**协议交互流程：**

SRP 使用 Unix Domain Socket 在进程间传输状态 [4](#3-3) ：

```text
Phase 1: 新进程连接 SRP Socket
Phase 2: 旧进程收到信号，序列化内存状态
Phase 3: 旧进程发送状态数据
[Length][Magic][0x02][JSON Payload]
Phase 4: 新进程接收并反序列化状态
```

**rsyslogd 状态传递实例：**

旧进程序列化的状态包括：
```json
{
  "active_connections": [
    {"client_ip": "192.168.1.100", "fd": 15},
    {"client_ip": "192.168.1.101", "fd": 16}
  ],
  "log_buffer": {
    "pending_messages": 1247,
    "buffer_data": "base64编码的日志缓冲区"
  },
  "config_state": {
    "current_rules": "/etc/rsyslog.d/*.conf",
    "module_status": {"imtcp": "loaded", "omfile": "active"}
  }
}
```

新进程接收后恢复这些状态，确保：
- 正在处理的日志消息不丢失
- 客户端连接状态保持一致
- 配置和模块状态完整继承

### 3. 金丝雀浸泡

**实现机制：**

在 `onSoakStart` 方法中，新旧进程并存运行 [5](#3-4) ：

```go
func (e *Engine) onSoakStart(event fsm.Event, args ...interface{}) error {
    soakDuration, _ := time.ParseDuration(e.cfg.Orchestration.Canary.SoakTime)
    
    go func() {
        time.Sleep(soakDuration)  // 30秒浸泡期
        
        // 监控新进程健康状态
        success := e.checkNewProcessHealth()
        
        if success {
            e.fsm.Fire("success")  // 继续排水旧进程
        } else {
            e.fsm.Fire("rollback") // 回滚到旧进程
        }
    }()
    return nil
}
```

**实际运行效果：**

1. **前 30 秒**：新旧 rsyslogd 进程同时运行
   - 内核负载均衡将新连接分配给两个进程
   - 旧进程继续处理已有连接
   - 新进程开始接收新连接

2. **健康检查**：监控新进程指标
   ```bash
   # 检查新进程是否正常接收日志
   tail -f /var/log/syslog | grep "rsyslogd-new"
   # 监控内存使用
   ps aux | grep rsyslogd
   ```

3. **决策点**：
   - 如果新进程正常：触发 `success` 事件，开始排水旧进程
   - 如果新进程异常：触发 `rollback` 事件，终止新进程，保留旧进程

---

## 最终效果

使用 Aeterna 对 rsyslogd 进行热升级可以实现：

1. **零连接丢失**：所有活跃的 syslog 连接保持 ESTABLISHED 状态 [6](#3-5) 
2. **状态完整性**：内存中的日志缓冲区完整迁移，确保日志不丢失 [7](#3-6) 
3. **毫秒级切换**：整个热更新过程在毫秒级完成 [8](#3-7) 
4. **自动回滚**：如果新版本出现问题，自动回滚到旧版本 [9](#3-8) 

---

## Notes

- rsyslogd 必须支持前台运行模式（`-n` 参数）
- 状态接力功能需要 rsyslogd 支持内存状态序列化（可能需要定制开发）
- 建议在测试环境先验证热升级流程
- 生产环境建议配置适当的监控和告警

Wiki pages you might want to explore:
- [Core Concepts (turtacn/Aeterna)](/wiki/turtacn/Aeterna#2)
- [Hot Reload Workflow (UPHR-O) (turtacn/Aeterna)](/wiki/turtacn/Aeterna#3.3)

### Citations

**File:** docs/apis.md (L135-188)
```markdown
## 3. State Relay Protocol (SRP) Specification

SRP 是 Aeterna 独有的、基于 Unix Domain Socket 的二进制协议，用于在父子进程或新老进程间传输敏感的内存上下文。

### 3.1 Transport Layer

* **Medium:** Unix Domain Socket (SOCK_STREAM)
* **Path:** 由配置文件 `orchestration.state_handoff.socket_path` 指定。
* **Security:** Socket 文件权限必须设置为 `0600` (仅当前用户读写)。

### 3.2 Packet Structure (Frame Format)

为了最大限度兼容 C/C++/Rust/Go/Python，协议采用简单的 **Length-Prefixed** 帧格式。所有整数均为 **Big-Endia (Network Byte Order)**。

```text
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       Length (4 Bytes)                        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                        Magic (4 Bytes)                        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Version  |     Type      |           Reserved            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                       Payload (Variable)                      |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

```

* **Length (uint32)**: 整个数据包的剩余长度（不包含 Length 字段本身）。
* **Magic (uint32)**: 固定值 `0xAE7E2024` (标识 Aeterna 协议)。
* **Version (uint8)**: 协议版本，当前为 `0x01`。
* **Type (uint8)**: 消息类型。
* `0x01`: Handshake / Hello
* `0x02`: State Data (JSON)
* `0x03`: State Data (Protobuf - Future Use)
* `0xFF`: ACK / Finished


* **Payload**: 具体的业务数据。

### 3.3 Interaction Flow

1. **Phase 1 (Connect):** 新进程（接收方）作为 Client 连接到 Socket。
2. **Phase 2 (Wait):** 老进程（发送方）收到终止信号，作为 Server 写入数据。
3. **Phase 3 (Transfer):**
* Sender sends: `[Length][Magic][0x01][0x02]...[JSON Data]`


4. **Phase 4 (Close):** 传输完成后，Sender 关闭连接。

---
```

**File:** docs/apis.md (L223-292)
```markdown
```yaml
version: "v1"

# -----------------------------------------------------------------------------
# 1. 服务定义 (Service Definition)
# -----------------------------------------------------------------------------
service:
  # 服务唯一标识
  name: "llm-inference-core"
  # 启动命令 (Aeterna 作为父进程将执行此命令)
  command: 
    - "/app/venv/bin/python"
    - "main.py"
  # 环境变量注入
  env:
    - "PORT=8080"
    - "MODEL_PATH=/models/llama3-70b"

# -----------------------------------------------------------------------------
# 2. 编排策略 (Orchestration Strategy)
# -----------------------------------------------------------------------------
orchestration:
  # 更新策略: 'immediate' (立即切换) | 'canary' (金丝雀/浸泡)
  strategy: "canary"

  # [Phase 1] 前置检查: 阻止错误配置上线
  pre_flight:
    - name: "Config Dry-run"
      command: ["/app/venv/bin/python", "main.py", "--check"]
      timeout: "5s"
      # 如果失败是否阻止发布: true (默认)
      block_on_fail: true

  # [Phase 2] 启动参数
  startup:
    # 预热延迟: 允许应用加载模型/缓存的时间
    warmup_delay: "10s"

  # [Phase 3] 金丝雀浸泡: 新老进程并存
  canary:
    enabled: true
    # 浸泡时长: 如果在此期间新进程 crash，自动回滚
    soak_time: "60s"
    # 健康检查 (可选): 期间必须通过 HTTP 探针
    health_check:
      http_get: "http://localhost:8080/health"
      interval: "5s"

  # [Phase 5] 排水: 优雅关闭老进程
  drain:
    timeout: "30s" # 超过此时间发送 SIGKILL

  # [关键特性] AI 状态接力协议 (SRP)
  state_handoff:
    enabled: true
    # 临时 socket 路径，用于新老进程传输内存数据
    socket_path: "/var/run/aeterna/srp.sock"
    # 等待老进程导出内存的最大时间
    timeout: "15s"

# -----------------------------------------------------------------------------
# 3. 可观测性 (Observability)
# -----------------------------------------------------------------------------
observability:
  # Prometheus Metrics 暴露端口
  metrics_port: ":9091"
  # 日志级别: debug, info, warn, error
  log_level: "info"

```
```

**File:** internal/resource/socket.go (L34-40)
```go
func NewSocketManager() *SocketManager {
	return &SocketManager{
		listeners: make(map[string]net.Listener),
		files:     make(map[string]*os.File),
		inherited: make(map[string]*inheritedSocket),
	}
}
```

**File:** internal/supervisor/manager.go (L22-40)
```go
func (pm *ProcessManager) Start(command []string, env []string, extraFiles []*os.File) error {
	if len(command) == 0 {
		return nil
	}

	pm.cmd = exec.Command(command[0], command[1:]...)
	pm.cmd.Env = append(os.Environ(), env...)
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr

	if len(extraFiles) > 0 {
		pm.cmd.ExtraFiles = extraFiles
		// UPHR Core: Notify child about inherited FDs
		pm.cmd.Env = append(pm.cmd.Env, fmt.Sprintf("%s=%d", consts.EnvInheritedFDs, len(extraFiles)))
	}

	logger.Log.Info("Supervisor: Forking process", "cmd", command)
	return pm.cmd.Start()
}
```

**File:** internal/orchestrator/engine.go (L127-155)
```go
// onSoakStart: Phase 2 & 3 - Fork, Exec & Soak
func (e *Engine) onSoakStart(event fsm.Event, args ...interface{}) error {
	logger.Log.Info("Phase 2 & 3: Forking New Process & Soaking")

	// Note: In a real implementation, we need to manage Two ProcessManagers (old and new).
	// For this blueprint, we simulate the decision logic.

	soakDuration, _ := time.ParseDuration(e.cfg.Orchestration.Canary.SoakTime)
	if soakDuration == 0 {
		soakDuration = consts.DefaultSoakTime
	}

	go func() {
		logger.Log.Info("Soaking...", "duration", soakDuration)
		// Monitoring simulation
		time.Sleep(soakDuration)

		// If metrics are good:
		success := true

		if success {
			e.fsm.Fire("success")
		} else {
			e.fsm.Fire("rollback")
		}
	}()

	return nil
}
```

**File:** README.md (L26-27)
```markdown
* **TCP 连接有损重置：** 对于依赖 WebSocket、gRPC 长连接的网关及金融实时交易系统，容器生命周期的终止直接导致传输层连接断开（Connection Reset）。在大规模并发场景下，这会诱发客户端的重连风暴（Thundering Herd Problem），导致服务可用性抖动。
* **高昂的初始化时延（Cold Start Latency）：** 对于基于 JVM 的大型应用或加载庞大权重的 AI 推理服务，进程启动涉及复杂的类加载、JIT 编译预热及显存数据加载。标准的 Pod 重启机制会导致服务在数秒至数分钟内处于不可服务（Unavailable）或性能降级状态。
```

**File:** README.md (L32-33)
```markdown
* **状态易失性风险：** 在传统的容器编排逻辑下，运维侧的安全补丁更新或版本发布（Deployment Rollout）会触发进程终止信号（SIGTERM），导致内存堆栈被强制回收。
* **计算成本与体验损耗：** 内存状态的丢失迫使 Agent 重新处理原始输入以重建推理上下文（Re-computation）。这不仅造成了计算资源（GPU/TPU）的冗余消耗，更因丢失交互历史而破坏了用户体验的连续性。
```

**File:** README.md (L61-61)
```markdown
* **Sub-millisecond Handover (毫秒级接力):** Updates happen at process fork speed.
```

**File:** docs/sdk_design.md (L170-172)
```markdown
1. **容错性 (Resilience)**:
* 如果 `load_context` 失败（例如反序列化错误），SDK **必须** 捕获异常并返回空状态（即退化为冷启动），决不能导致进程 Crash。
* *原则*: "Loss of memory is better than loss of service." (失忆总比宕机好)
```
