# supers

`supers` 是一个用 Go 编写的轻量级进程管理守护进程（`superd`）及配套命令行客户端（`supers`）。

## 特性

- 进程管理：启动 / 停止 / 重启 / 状态查询  
- 自动重启：异常退出后自动重启  
- 日志收集与切割：按大小 / 天数滚动  
- 动态配置：`reload` 动态加载 `/etc/super/*.service`；`start` on-demand  
- 事件通知：内置 Webhook 支持  
- HTTP 控制接口 & CLI  
- 跨平台：Linux、macOS（未来可拓展 Windows）

---

## 快速开始

### 环境依赖

- Go 1.20+  
- systemd（用于管理 `superd`）  

### 编译

```bash
git clone https://github.com/litongjava/supers.git
cd supers

# 构建守护进程
go build -o bin/superd cmd/superd/main.go

# 构建客户端
go build -o bin/supers cmd/supers/main.go
````

### 安装

```bash
sudo mv bin/superd /usr/local/bin/
sudo mv bin/supers /usr/local/bin/
sudo chmod +x /usr/local/bin/superd /usr/local/bin/supers
```

---

## 配置

1. 创建服务描述目录：

   ```bash
   sudo mkdir -p /etc/super
   ```

2. 在 `/etc/super` 下添加你的 `.service` 文件。例如：

   ```ini
   # /etc/super/docker-io-proxy.service
   [Unit]
   Description=docker-io-proxy Java Web Service
   After=network.target

   [Service]
   Type=simple
   User=root
   WorkingDirectory=/data/apps/docker-io-proxy
   ExecStart=/usr/java/jdk1.8.0_211/bin/java -jar target/docker-io-proxy-1.0.0.jar --server.port=8004
   Restart=on-failure
   RestartSec=5s

   [Install]
   WantedBy=default.target
   ```

3. （可选）修改 `config/config.yml` 以调整 HTTP 端口、Webhook URL 等。

---

## 使用 systemd 管理 `superd`

1./data/apps/supers/config/config.yml
```yaml
app:
  port: 10405
  filePath: /data/upload
  password: 123456
```


2. 将以下内容保存为 `/etc/systemd/system/superd.service`：

```ini
[Unit]
Description=SuperD Process Management Daemon
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/data/apps/supers
ExecStart=/usr/local/bin/superd
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

2. 重新加载 systemd 并启用、启动服务：

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable superd
   sudo systemctl start superd
   ```

3. 查看 `superd` 运行状态：

   ```bash
   systemctl status superd
   ```

---

## 使用 supers 客户端

在系统上安装并启动 `superd` 后，你可以使用 `supers` 命令通过 Unix Socket 管理子进程：

```bash
# 列出所有服务及状态
supers list

# 查询某服务状态
supers status <service_name>

# 停止服务
supers stop <service_name>

# 启动（或 on-demand 启动）服务
supers start <service_name>

# 动态重载 /etc/super 下所有配置
supers reload
```

---

## 日志

每个子进程的日志文件位于：

```
/etc/super/logs/<service_name>/stdout.log
/etc/super/logs/<service_name>/stderr.log
```

---

## HTTP 控制接口

（如已启用）`superd` 默认监听 HTTP 在 `config/config.yml` 中配置的端口，提供增删改查、日志查看等 RESTful API，详见 [router 注册文档](./router.md)。

---

## 贡献

欢迎提 Issue & PR，共同完善更多功能 😊
