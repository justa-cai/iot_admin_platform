<div align="center">

# IoT Admin Platform

**物联网设备管理与监控系统**

基于 MQTT 的全栈 IoT 管理平台，包含嵌入式 MQTT Broker、设备管理、规则引擎、实时数据可视化与固件 OTA 升级

**[English](README_EN.md)** | 中文

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![SQLite](https://img.shields.io/badge/SQLite-WAL-003B57?style=flat&logo=sqlite)](https://www.sqlite.org/)
[![MQTT](https://img.shields.io/badge/MQTT-5.0-660066?style=flat&logo=mqtt)](https://mqtt.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## 界面预览

<table>
<tr>
<td width="50%"><img src="docs/screenshots/dashboard.png" alt="仪表盘"/></td>
<td width="50%"><img src="docs/screenshots/devices.png" alt="设备管理"/></td>
</tr>
<tr>
<td align="center"><b>仪表盘</b> - 实时统计卡片、消息吞吐量趋势、设备状态分布</td>
<td align="center"><b>设备管理</b> - 设备列表、搜索筛选、分组标签管理</td>
</tr>
<tr>
<td width="50%"><img src="docs/screenshots/rules.png" alt="规则引擎"/></td>
<td width="50%"><img src="docs/screenshots/messages.png" alt="消息历史"/></td>
</tr>
<tr>
<td align="center"><b>规则引擎</b> - Topic 通配符匹配、条件评估、告警触发</td>
<td align="center"><b>消息历史</b> - 消息查询、Topic 搜索、时间范围筛选</td>
</tr>
<tr>
<td width="50%"><img src="docs/screenshots/telemetry.png" alt="数据查询"/></td>
<td width="50%"><img src="docs/screenshots/firmware.png" alt="固件管理"/></td>
</tr>
<tr>
<td align="center"><b>数据查询</b> - 遥测数据趋势图表、时间范围快速选择</td>
<td align="center"><b>固件管理</b> - 固件上传、SHA256 校验、OTA 升级任务</td>
</tr>
</table>

---

## 系统架构

```
┌──────────────┐     ┌───────────────────────────────────────────┐
│              │     │            IoT Admin Server                │
│  IoT 设备    │     │                                           │
│  (MQTT 5.0)  ├────►│  ┌─────────────┐   ┌──────────────────┐  │
│              │     │  │ Mochi MQTT  │   │    规则引擎       │  │
│  温湿度传感器 │ TCP │  │   Broker    ├──►│  Topic 通配匹配   │  │
│  压力传感器   ├────►│  │  :1883      │   │  条件评估         │  │
│  GPS 定位器  │ WS  │  │  :8083      │   │  告警/转发        │  │
│  烟雾报警器  ├────►│  └──────┬──────┘   └──────────────────┘  │
│              │     │         │                                  │
└──────────────┘     │  ┌──────▼──────┐   ┌──────────────────┐  │
                     │  │ OnPublish   │   │  WebSocket Hub   │  │
┌──────────────┐     │  │   Hook      ├──►│   实时推送       │  │
│              │     │  └──────┬──────┘   └────────┬─────────┘  │
│   浏览器     │     │         │                    │            │
│  (Vue 3 SPA) │◄────│  ┌──────▼──────┐            │            │
│              │ WS  │  │   SQLite    │            │            │
│  仪表盘      ├────►│  │  (WAL 模式) │            │            │
│  控制台      │ API │  └─────────────┘            │            │
│              ├────►│                              │            │
└──────────────┘     │  ┌───────────────┐  ┌───────▼─────────┐  │
                     │  │  Gin REST API │  │   静态文件       │  │
                     │  │  :8080        │  │  (Vue SPA)      │  │
                     │  └───────────────┘  └─────────────────┘  │
                     └───────────────────────────────────────────┘
```

---

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| **后端** | Go 1.23 + Gin | 高性能 HTTP 框架 |
| **MQTT Broker** | [Mochi MQTT v2](https://github.com/mochi-mqtt/server) | 嵌入式 Broker，TCP + WebSocket 双协议 |
| **数据库** | SQLite (WAL) | 零配置、单文件部署 |
| **前端** | Vue 3 + TypeScript + Element Plus | SPA 组件自动导入 |
| **图表** | ECharts | 实时数据可视化 |
| **状态管理** | Pinia | Vue 3 状态管理 |
| **认证** | JWT + bcrypt | 基于角色的访问控制 |
| **部署** | 单二进制 | Go 嵌入 Vue SPA 静态文件 |

---

## 功能特性

### 设备管理
- 设备 CRUD，自动生成凭证（Key / Secret）
- 分层分组树形管理
- 灵活的标签系统，支持自定义颜色
- MQTT Hook 实现在线 / 离线状态追踪
- 丰富元数据支持（位置、型号、固件版本）

### MQTT Broker
- 嵌入式 Mochi MQTT v2，无需外部依赖
- TCP (`:1883`) + WebSocket (`:8083`) 双协议
- 设备 Key / Secret 认证
- 自动上下线状态管理
- OnPublish Hook 消息拦截

### 规则引擎
- Topic 通配符匹配（如 `telemetry/+/data`）
- 条件评估：`gt`、`gte`、`lt`、`lte`、`eq`、`neq`、`contains`
- 动作类型：告警通知、MQTT 消息转发、HTTP Webhook
- 可配置冷却时间，避免告警风暴
- 规则执行日志记录

### 实时仪表盘
- 4 个 KPI 统计卡片（设备总数、在线设备、今日消息、活跃规则）
- 消息吞吐量趋势图（6h / 12h / 24h / 3d）
- 设备状态分布饼图
- 最近告警列表

### 遥测与消息
- WebSocket 实时消息控制台
- 消息历史查询，支持 Topic 搜索与时间筛选
- 遥测数据趋势图表
- 快捷时间范围按钮（1h / 6h / 24h / 7d / 30d）

### 固件 OTA
- 固件上传，SHA256 校验
- 批量 OTA 升级任务创建
- 升级进度与状态追踪
- 固件下载接口

### 系统管理
- JWT 认证，支持 Token 刷新
- RBAC 权限：管理员 / 操作员 / 观察者
- 用户管理（增删改查、密码重置）
- CORS 跨域支持

---

## 快速开始

### 环境要求

- Go 1.23+（需启用 CGO，用于 SQLite）
- Node.js 18+
- npm

### 1. 克隆与安装

```bash
git clone https://github.com/justa-cai/iot_admin_platform.git
cd iot_admin_platform

# 安装前端依赖
cd frontend && npm install && cd ..
```

### 2. 启动开发服务

```bash
# 终端 1 - 后端（API :8080 + MQTT :1883 + WS :8083）
make backend-dev

# 终端 2 - 前端（Vite 开发服务器 :5173）
make frontend-dev
```

打开 http://localhost:5173，使用 `admin` / `admin123` 登录

### 3. 生产构建

```bash
make build
./iot-admin-server
```

单二进制文件同时提供 API + MQTT + 前端服务。

### 4. 填充演示数据

```bash
python3 loadtest/seed_demo.py
```

自动创建 20 个设备、5 个分组、7 个标签、6 条规则和 4 个用户。

---

## 项目结构

```
iot_admin/
├── backend/
│   ├── cmd/server/main.go           # 入口文件
│   ├── internal/
│   │   ├── api/                     # Gin 处理器 + 中间件
│   │   │   ├── handler/             # Auth、Device、Rule、Dashboard 等
│   │   │   ├── middleware/          # JWT 认证、CORS
│   │   │   └── router.go           # API 路由定义
│   │   ├── config/                  # Viper 配置管理
│   │   ├── model/                   # 数据模型
│   │   ├── mqtt/                    # Mochi Broker + Hooks
│   │   ├── rule/                    # 规则引擎（匹配 + 动作）
│   │   ├── store/sqlite/            # SQLite 数据访问层
│   │   └── ws/                      # WebSocket Hub
│   ├── config.yaml
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/                     # Axios API 客户端
│   │   ├── components/              # StatCard、LineChart、GaugeChart
│   │   ├── composables/             # useWebSocket、useECharts
│   │   ├── layouts/                 # AdminLayout、AuthLayout
│   │   ├── router/                  # Vue Router 路由配置
│   │   ├── stores/                  # Pinia 状态管理
│   │   ├── styles/                  # 全局 SCSS 主题
│   │   ├── types/                   # TypeScript 类型定义
│   │   └── views/                   # 页面组件
│   └── vite.config.ts
├── clients/
│   ├── go/                          # Go MQTT 发布/订阅客户端
│   └── python/                      # Python MQTT 发布/订阅客户端
├── loadtest/
│   ├── main.go                      # Go 并发负载测试
│   ├── seed_demo.py                 # 演示数据填充脚本
│   └── demo_sim.py                  # 设备模拟器
├── docs/screenshots/                # UI 截图
├── Makefile
└── .gitignore
```

---

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| **认证** | | |
| POST | `/api/v1/auth/login` | 登录，返回 JWT |
| POST | `/api/v1/auth/register` | 注册（仅管理员） |
| GET | `/api/v1/auth/profile` | 当前用户信息 |
| **设备** | | |
| GET/POST | `/api/v1/devices` | 设备列表 / 创建设备 |
| GET/PUT/DELETE | `/api/v1/devices/:id` | 设备详情 / 更新 / 删除 |
| **分组与标签** | | |
| CRUD | `/api/v1/groups` | 分组管理 |
| CRUD | `/api/v1/tags` | 标签管理 |
| **消息** | | |
| POST | `/api/v1/messages/publish` | 发布 MQTT 消息 |
| GET | `/api/v1/messages/history` | 消息历史 |
| GET | `/api/v1/messages/topics` | Topic 树 |
| **规则** | | |
| CRUD | `/api/v1/rules` | 规则管理 |
| PUT | `/api/v1/rules/:id/enable` | 启用 / 禁用规则 |
| GET | `/api/v1/rules/:id/logs` | 规则执行日志 |
| **遥测** | | |
| GET | `/api/v1/telemetry` | 查询遥测数据 |
| GET | `/api/v1/telemetry/latest` | 各设备最新数据 |
| **仪表盘** | | |
| GET | `/api/v1/dashboard/stats` | 统计概览 |
| GET | `/api/v1/dashboard/throughput` | 吞吐量图表数据 |
| **固件** | | |
| POST | `/api/v1/firmware/upload` | 上传固件 |
| GET | `/api/v1/firmware/:id/download` | 下载固件 |
| POST | `/api/v1/ota` | 创建 OTA 升级任务 |
| **WebSocket** | | |
| WS | `/api/v1/ws` | 实时事件推送 |

---

## 负载测试

100 个并发设备，500ms 发送间隔：

```
=== Final Results ===
Duration:        30.9s
Total Published: 5,943
Total Errors:    0
Avg Publish Rate: 192 msg/s
Connected:       100 / 100
```

```bash
# 运行负载测试
make loadtest

# 自定义参数
NUM_DEVICES=200 MSG_INTERVAL_MS=200 DURATION_SECS=60 make loadtest
```

---

## IoT 客户端示例

### Go 发布端

```go
opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
opts.SetUsername("your-device-key").SetPassword("your-device-secret")
client := mqtt.NewClient(opts)
client.Connect()

payload, _ := json.Marshal(map[string]interface{}{
    "temperature": 25.5,
    "humidity":    60.0,
})
client.Publish("telemetry/device-001/data", 0, false, payload)
```

### Python 订阅端

```python
import paho.mqtt.client as mqtt

client = mqtt.Client()
client.username_pw_set("your-device-key", "your-device-secret")
client.connect("localhost", 1883)

client.subscribe("telemetry/+/data")
client.on_message = lambda c, u, msg: print(msg.payload.decode())
client.loop_forever()
```

---

## 配置说明

`backend/config.yaml`：

```yaml
server:
  port: 8080

sqlite:
  path: data/iot_admin.db

mqtt:
  tcp_port: 1883
  ws_port: 8083

jwt:
  secret: your-secret-key
  expire_hours: 24
```

---

## 许可证

[MIT](LICENSE)
