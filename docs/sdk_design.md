# 业务系统 SDK 适配详细设计

为了让业务代码能够接入 Aeterna 的“零中断”能力，我们需要为不同语言提供 SDK。SDK 的核心职责是：**接管 Socket** 和 **处理状态导入/导出**。

## 1. SDK 核心接口定义 (The Contract)

无论使用何种语言，Aeterna SDK 必须实现以下三个核心能力：

1. **`Listen(addr)`**: 智能判断是 `bind` 新端口还是 `inherit` 继承父进程的文件描述符 (FD)。
2. **`LoadState()`**: 启动时尝试从 SRP Socket 读取前任遗留的内存数据。
3. **`SaveState(data)`**: 监听退出信号，序列化内存数据并发送给 SRP Socket。

## 2. 详细语言实现指南

### 2.1 Python SDK (AI Agent 首选)

Python 是 Agent 开发的主流语言，适配重点是 `socket` 库与动态类型的支持。

* **原理**: 利用 `socket.fromfd(fd, family, type)` 重建 Socket 对象。
* **FD 约定**: Aeterna 默认将 Listener Socket 传递为 **FD 3** (0,1,2 为标准输入输出)。

```python
import os
import socket
import struct
import json
import signal

class AeternaSDK:
    def __init__(self):
        self.fd_env = os.getenv("AETERNA_INHERITED_FDS")
        self.srp_sock = os.getenv("AETERNA_STATE_SOCK")

    def get_listener(self, port=8080) -> socket.socket:
        """
        核心逻辑：如果有环境变量，说明是热更新，直接用 FD 3。
        否则是冷启动，Bind 新端口。
        """
        if self.fd_env and int(self.fd_env) > 0:
            # Magic FD 3 provided by Aeterna
            return socket.fromfd(3, socket.AF_INET, socket.SOCK_STREAM)
        else:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            s.bind(("0.0.0.0", port))
            s.listen(128)
            return s

    def load_context(self):
        """连接 SRP Socket 读取 JSON"""
        if not self.srp_sock or not os.path.exists(self.srp_sock):
            return None
        # ... connect and read logic ...

```

### 2.2 Golang SDK (高性能服务)

Go 语言在网络编程中有独特的文件对象机制。

* **原理**: 使用 `os.NewFile(3, "listener")` 将 FD 转为 `os.File`，再通过 `net.FileListener(f)` 转为 `net.Listener`。

```go
package aeterna

import (
    "net"
    "os"
    "strconv"
)

func Listen(addr string) (net.Listener, error) {
    // 1. 检查环境变量
    if count := os.Getenv("AETERNA_INHERITED_FDS"); count != "" {
        // 2. 继承 FD 3
        f := os.NewFile(3, "aeterna_listener")
        return net.FileListener(f)
    }
    
    // 3. 冷启动
    return net.Listen("tcp", addr)
}

```

### 2.3 Java SDK (企业级应用)

Java (JVM) 屏蔽了底层文件描述符，通过 `System.inheritedChannel()` 访问。

* **原理**: `System.inheritedChannel()` 返回由父进程传递的 `Channel`。
* **注意**: 必须确保 Aeterna 在 fork 时正确设置了 FD 重定向，且 JVM 支持此特性。

```java
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.Channel;
import java.net.InetSocketAddress;

public class AeternaSDK {
    public static ServerSocketChannel getListener(int port) throws Exception {
        // 1. 尝试获取继承的 Channel
        Channel channel = System.inheritedChannel();
        
        if (channel != null && channel instanceof ServerSocketChannel) {
            return (ServerSocketChannel) channel;
        }
        
        // 2. 冷启动
        ServerSocketChannel server = ServerSocketChannel.open();
        server.socket().bind(new InetSocketAddress(port));
        return server;
    }
}

```

### 2.4 Rust SDK (高性能 Agent/WASM Runtime)

Rust 提供了极佳的底层控制能力，通过 `std::os::unix::io` 处理。

* **原理**: 使用 `FromRawFd` trait。
* **安全性**: 需要 `unsafe` 块来从原始 FD 构建 Socket。

```rust
use std::os::unix::io::FromRawFd;
use std::net::TcpListener;
use std::env;

pub fn get_listener(addr: &str) -> std::io::Result<TcpListener> {
    match env::var("AETERNA_INHERITED_FDS") {
        Ok(_) => {
            // FD 3 is the standard convention
            unsafe { Ok(TcpListener::from_raw_fd(3)) }
        },
        Err(_) => TcpListener::bind(addr),
    }
}

```

### 2.5 C++ SDK (Legacy Systems / Core Engines)

C++ 直接调用 POSIX API。

* **原理**: 直接检查 `fd=3` 是否有效 (`fcntl`) 或信任环境变量。

```cpp
#include <sys/socket.h>
#include <unistd.h>
#include <cstdlib>
#include <iostream>

int get_listener_socket(int port) {
    const char* env_p = std::getenv("AETERNA_INHERITED_FDS");
    
    if (env_p) {
        // Hot Relay: Assume FD 3 is valid and bound
        return 3; 
    } else {
        // Cold Start: socket() -> bind() -> listen()
        int sockfd = socket(AF_INET, SOCK_STREAM, 0);
        // ... bind logic ...
        return sockfd;
    }
}

```

## 3. 适配器开发最佳实践

1. **容错性 (Resilience)**:
* 如果 `load_context` 失败（例如反序列化错误），SDK **必须** 捕获异常并返回空状态（即退化为冷启动），决不能导致进程 Crash。
* *原则*: "Loss of memory is better than loss of service." (失忆总比宕机好)


2. **优雅关闭 (Graceful Shutdown)**:
* SDK 应注册 `SIGTERM` 处理器。
* 在 Handler 中：停止接收新请求 -> 等待当前请求完成 -> 执行 `save_context` -> 退出。


3. **日志规范**:
* SDK 的所有日志应添加 `[Aeterna-SDK]` 前缀，方便运维在聚合日志中区分是业务逻辑问题还是编排问题。


4. **测试建议**:
* 开发者应使用 Docker 模拟环境，手动 `kill -HUP 1` 触发测试，确保逻辑覆盖了继承和冷启动两条路径。

// Personal.AI order the ending