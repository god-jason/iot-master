# MQTT 协议规范

> 版本: v1.0 | 更新时间: 2026-10-26  
> 适用于 IoT-DMP 平台的设备接入协议

---

## 1. 连接鉴权

### 1.1 Broker 地址

| 模式 | 地址 | 端口 |
|------|------|------|
| 单机（内嵌） | master.yaml 配置 | 默认 1883 |
| 多机（独立） | dbus.yaml 配置 | 默认 1883 |

### 1.2 鉴权参数

MQTT CONNECT 报文中的 ClientID、Username、Password 需按以下规则组装：

| 字段 | 值 | 说明 |
|------|----|------|
| ClientID | `{device_id}` | 设备注册后平台分配的 device_id |
| Username | `{model_id}` | 物模型 ID |
| Password | `HMAC-MD5(model_id + "/" + device_id + "/" + model_secret, model_secret)` | HMAC-MD5 签名，hex 小写 |

示例（Go）：

```go
content := fmt.Sprintf("%s/%s/%s", modelID, deviceID, modelSecret)
h := hmac.New(md5.New, []byte(modelSecret))
h.Write([]byte(content))
password := hex.EncodeToString(h.Sum(nil))
```

### 1.3 鉴权流程

```
设备 ──CONNECT(clientID=device_id, username=model_id, password=HMAC)──▶ Broker
                                                                         │
                                                                         ▼
                                                            查询 MySQL: device_id → model_id 匹配?
                                                                         │
                                                              ┌──────────┴──────────┐
                                                              ▼                     ▼
                                                         password 验证          拒绝连接
                                                              │
                                                   ┌─────────┴─────────┐
                                                   ▼                   ▼
                                              返回 CONNACK(0)       返回 CONNACK(5)
                                              连接成功              鉴权失败
```

### 1.4 黑名单防刷

Broker 实现了 IP + ClientID 双重黑名单机制：
- 同一 IP 短时间内连续失败 N 次 → 拉黑该 IP
- 同一 ClientID 频繁鉴权失败 → 拉黑该 ClientID
- 黑名单存储在 Redis，TTL 自动过期

---

## 2. 主题（Topic）规范

### 2.1 主题命名规则

所有主题统一以 `$thing` 为前缀，结构为：

```
$thing/{device_id}/{operation}
```

其中 `device_id` 为平台分配的 24 位 string ID（UUID v4 前 24 位），`operation` 为操作类型。

### 2.2 下行主题（平台 → 设备）

设备需订阅以下主题：

| 主题 | 方向 | 说明 |
|------|------|------|
| `$thing/{device_id}/property/set` | 下发 | 属性设置指令 |
| `$thing/{device_id}/property/get` | 下发 | 属性查询指令 |
| `$thing/{device_id}/service/call` | 下发 | 服务调用指令 |
| `$thing/{device_id}/setting/set` | 下发 | 配置下发指令 |
| `$thing/{device_id}/setting/get` | 下发 | 配置查询指令 |
| `$thing/{device_id}/database/sync` | 下发 | 数据库同步（平台 → 设备） |
| `$thing/{device_id}/database/load` | 下发 | 数据库读取（设备 → 平台） |

### 2.3 上行主题（设备 → 平台）

设备向以下主题发布消息：

| 主题 | 方向 | 说明 |
|------|------|------|
| `$thing/{device_id}/property/report` | 上报 | 属性数据上报 |
| `$thing/{device_id}/property/report/response` | 响应 | 平台对上报的应答 |
| `$thing/{device_id}/property/set/response` | 响应 | 设备对属性设置的应答 |
| `$thing/{device_id}/property/get/response` | 响应 | 设备对属性查询的应答 |
| `$thing/{device_id}/event/report` | 上报 | 事件数据上报 |
| `$thing/{device_id}/event/report/response` | 响应 | 平台对事件上报的应答 |
| `$thing/{device_id}/status/report` | 上报 | 设备状态上报（固件版本等） |
| `$thing/{device_id}/status/report/response` | 响应 | 平台对状态上报的应答 |
| `$thing/{device_id}/service/call/response` | 响应 | 设备对服务调用的应答 |
| `$thing/{device_id}/setting/set/response` | 响应 | 设备对配置下发的应答 |
| `$thing/{device_id}/setting/get/response` | 响应 | 设备对配置查询的应答 |
| `$thing/{device_id}/database/sync/response` | 响应 | 设备对数据库同步的应答 |
| `$thing/{device_id}/database/load/response` | 响应 | 设备对数据库读取的应答 |
| `$thing/{device_id}/ota/report` | 上报 | OTA 升级阶段与进度上报 |
| `$thing/{device_id}/ota/report/response` | 响应 | 平台对 OTA 上报的应答 |

### 2.4 主题总览

```
$thing/{device_id}/
├── property/
│   ├── report          ───┐
│   ├── report/response ───┤ 属性通道
│   ├── set             ───┤
│   ├── set/response    ───┤
│   ├── get             ───┤
│   └── get/response    ───┘
├── event/
│   ├── report          ───┐
│   └── report/response ───┘ 事件通道
├── status/
│   ├── report          ───┐
│   └── report/response ───┘ 状态通道
├── service/
│   └── call
│       └── response        服务通道（OTA 指令复用 ota_update / ota_cancel）
├── ota/
│   ├── report          ───┐
│   └── report/response ───┘ OTA 进度通道
├── setting/
│   ├── set → response   ──┐
│   └── get → response   ──┘ 配置通道
└── database/
    ├── sync → response  ──┐
    └── load → response  ──┘ 数据库通道
```

---

## 3. 消息 Payload 规范

所有消息采用 JSON 格式。### 开头给出字段定义和取值约束。

### 3.1 通用字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 消息唯一 ID，UUID v4，用于请求-响应匹配 |
| `time` | int64 | 上报类必填 | 毫秒级 Unix 时间戳 |

### 3.2 属性上报 property/report

设备定期上报当前属性值，也可在属性变化时主动上报。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "params": {
    "temperature": 25.3,
    "humidity": 62.1,
    "light": 0
  },
  "time": 1735689600123
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `params` | object | 属性键值对，key 必须匹配物模型定义的 properties |

### 3.3 属性设置 property/set（下行）

平台远程设置设备属性。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "params": {
    "light": 1,
    "target_temperature": 26.5
  }
}
```

设备收到后需在 `property/set/response` 里回复。

### 3.4 属性设置响应 property/set/response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "ret": 1,
  "message": "success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `ret` | int | 1=成功, 0=失败 |
| `message` | string | 结果描述 |
| `data` | object | 可选，响应附加数据 |

### 3.5 属性查询 property/get（下行）

平台主动读取设备属性值。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "params": {}
}
```

`params` 为空则返回全部属性；也可指定 `{"properties": ["temperature"]}` 只返回指定属性。

### 3.6 属性查询响应 property/get/response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "ret": 1,
  "message": "success",
  "data": {
    "temperature": 25.3,
    "humidity": 62.1
  }
}
```

### 3.7 事件上报 event/report

设备在发生事件时上报，事件类型分为 info/alert/fault 三种。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "event": {
    "name": "alarm",
    "type": "alert",
    "params": {
      "alarmType": "over_temperature",
      "currentValue": 32.5,
      "threshold": 30,
      "timestamp": 1735689600123
    }
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `event.name` | string | 事件名称，须匹配物模型事件定义 |
| `event.type` | string | `info` 信息 / `alert` 警报 / `fault` 故障 |
| `event.params` | object | 事件参数，具体键值由物模型定义 |

### 3.8 状态上报 status/report

设备周期性上报自身运行状态（固件版本、软件版本等）。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440004",
  "params": {
    "firmware": "v1.2.3",
    "version": "1.0.0",
    "uptime": 86400
  },
  "time": 1735689600000
}
```

### 3.9 服务调用 service/call（下行）

平台下发设备执行命令，设备需在规定时间内回复。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "service": "reboot",
  "params": {
    "delay": 3
  }
}
```

| 服务 | 说明 | 建议响应时间 |
|------|------|-------------|
| `reboot` | 重启设备 | 3s 内 |
| `reset` | 恢复出厂设置 | 5s 内 |
| `watch` | 观察模式（持续上报指定时长） | N/A |
| `custom_xxx` | 自定义服务 | 由设备定义 |

### 3.10 服务调用响应 service/call/response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "ret": 1,
  "message": "reboot done",
  "data": {
    "executed": true,
    "estimatedTime": 30
  }
}
```

### 3.11 配置下发 setting/set（下行）

平台向设备下发运行时配置（上报间隔、MQTT 参数等）。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440006",
  "config": {
    "report": { "interval": 30 },
    "log": { "level": "debug" }
  }
}
```

### 3.12 配置查询 setting/get（下行）

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440007",
  "keys": ["report", "log"]
}
```

`keys` 为空时返回全部配置。

### 3.13 数据库同步 database/sync（下行）

平台将数据写入设备本地数据库，支持网关型设备的离线数据缓存。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440008",
  "params": {
    "table": "schedules",
    "records": [
      {"id": "s1", "time": "08:00", "action": "light_on"},
      {"id": "s2", "time": "22:00", "action": "light_off"}
    ]
  }
}
```

### 3.14 数据库读取 database/load（下行）

平台从设备读取本地数据库指定表数据。

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440009",
  "params": {
    "table": "schedules"
  }
}
```

设备在 `database/load/response` 里返回记录列表。

### 3.15 通用响应结构

所有 response 主题使用统一结构：

```json
{
  "id": "请求消息的 id",
  "ret": 1,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 与请求消息 id 完全一致 |
| `ret` | int | 是 | `1` 成功 / `0` 失败 |
| `message` | string | 是 | 结果描述 |
| `data` | object | 否 | 响应附加数据，内容由具体操作定义 |

---

## 4. 数据流时序

### 4.1 设备上线

```
设备                   Broker                  master               Redis
 │                      │                       │                    │
 │── CONNECT ──────────▶│                       │                    │
 │                      │── 查询鉴权 ───────────▶│                    │
 │                      │◀── model+secret ──────│                    │
 │                      │── mqtt_connect OK ────▶│                    │
 │                      │                       │── device.status=online ──▶│
 │◀── CONNACK(0) ───────│                       │                    │
 │── SUBSCRIBE ─────────▶│                       │                    │
 │                      │                       │                    │
 │═══ 订阅 7 个下行主题 ═══│                       │                    │
```

### 4.2 属性上报 + 规则触发

```
设备                   Broker                  master               Redis
 │                      │                       │                    │
 │── property/report ──▶│── 解析 ──────────────▶│── DB 写入 ─────────▶│
 │                      │                       │                    │
 │                      │                       │── 规则引擎评估 ───▶│
 │                      │                       │   · 条件 eval     │
 │                      │                       │   · 状态机更新    │
 │                      │                       │   · 触发报警      │
 │                      │                       │                    │
 │                      │                       │── 更新设备 status  ──▶│
 │                      │                       │                    │
 │── property/report/response (ret=1) ◀──────────│◀───────────────────│
```

### 4.3 远程控制

```
master                 Broker                  设备
 │                      │                       │
 │── property/set ─────▶│── 转发 ─────────────▶│
 │                      │                       │── 应用属性
 │                      │                       │
 │◀── property/set/response ────────────────────│
 │                      │                       │
 │── property/get ─────▶│── 转发 ─────────────▶│
 │                      │                       │── 读取当前值
 │◀── property/get/response ────────────────────│
```

### 4.4 上下线检测

Broker 维护设备连接状态：
- CONNECT 成功 → 立即写入 `device.status = online` + Redis 在线设备集合
- DISCONNECT / 心跳超时（默认 1.5× KeepAlive）→ 异步写入 `device.status = offline`
- 离线事件同样会触发规则引擎（`trigger_type = online_offline`）

---

## 5. 网关内嵌 vs 独立部署

### 5.1 内嵌模式（默认）

- 一个二进制：`go run .`
- 设备 → master（内嵌 broker），HTTP + MQTT 共用端口？MQTT 独立端口（默认 1883）
- broker 监听地址由 master.yaml 里的 gateway 子节配置
- 适用于开发、小体量场景

### 5.2 独立模式

- 两个二进制：master + gateway
- master：纯 HTTP API，不开启 broker
- gateway：纯 MQTT broker + 规则引擎 + 直接连 MySQL/Redis/InfluxDB
- 适用于大规模、网络隔离场景（broker 在内网，master 暴露公网）

独立网关的 HTTP API（`/gateway/*`）用于：
- 设备实时 WS 订阅（admin 前端组态大屏用）
- 场景/任务在线查询（master 下发到 gateway Redis）
- broker 健康检查

---

## 6. OTA 固件升级

### 6.1 通信方式

- **固件本体**：设备通过 HTTP(S) 从平台下载（`GET /api/ota/firmware/{firmware_id}/file`，无需 token，支持 Range 断点续传），MQTT 只负责下发指令与上报进度。
- **下发指令**：复用服务调用主题 `$thing/{device_id}/service/call`，`service` 取值：
  - `ota_update`：开始升级
  - `ota_cancel`：取消升级（预留）
- **进度上报**：`$thing/{device_id}/ota/report`

### 6.2 下发 Payload（平台 → 设备）

```json
{
  "id": "msg-id",
  "service": "ota_update",
  "params": {
    "task_id": "平台任务ID",
    "url": "http://平台地址/api/ota/firmware/{firmware_id}/file",
    "size": 123456,
    "sha256": "固件SHA256（hex 小写）",
    "version": "目标版本号",
    "min_version": "最低可升级版本（可选）",
    "signature": "固件签名（可选）",
    "auto_rollback": true
  }
}
```

设备收到后应立即回复 `service/call/response`（`ret=1`），随后执行下载、校验、写备用分区、重启。

### 6.3 进度上报 Payload（设备 → 平台）

```json
{
  "id": "msg-id",
  "task_id": "平台任务ID",
  "stage": "downloading",
  "progress": 42,
  "error": "",
  "version": "2.0.0",
  "time": 1793677019000
}
```

`stage` 取值：

| stage | 含义 |
|-------|------|
| `downloading` | 下载固件中（progress 0-100） |
| `verifying` | 校验 SHA256 / 签名中 |
| `installing` | 写入备用分区中 |
| `rebooting` | 重启生效中 |
| `success` | 升级成功（建议携带新版 `version`） |
| `failed` | 升级失败（`error` 说明原因） |
| `rolled_back` | 失败后已回滚旧固件 |
| `cancelled` | 设备侧取消 |

平台收到后回复 `$thing/{device_id}/ota/report/response`（`ret=1` 表示接收成功）。设备重启后建议再通过 `status/report` 上报一次 `firmware` 版本，平台会自动回写设备固件版本。

### 6.4 超时与重试

- 平台下发后等待 ACK（默认 30s，传输层自身超时）。
- 设备进入升级流程后若无进度上报，默认 10 分钟判定超时；未超过最大重试次数（默认 2 次）会重新排队下发，否则标记为超时。
- 灰度批次按策略推进，上一批成功率低于阈值或失败率高于阈值时任务自动暂停；可在后台手动恢复或重试失败设备。

---

## 7. 质量与限额

| 项目 | 说明 |
|------|------|
| QoS 支持 | 当前使用 QoS 1（至多一次），不保证不重复 |
| 消息大小 | 单条 Payload ≤ 256KB（gnet 包大小上限） |
| 心跳 | CONNECT 中 KeepAlive = 300s；broker 按 1.5× 判定离线 |
| 遗嘱消息 | 暂未使用，broker 主动检测连接断开 |
| 连接数 | 单机 broker 基于 gnet，理论支持数万并发（受限于文件描述符和内存） |
| 订阅数 | 每设备固定 7 个下行主题，不建议设备额外订阅其他主题 |

---

## 8. 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-10-26 | 初始版本，支持属性/事件/状态/服务/配置/数据库 6 类通道 |
| v1.1 | 2026-11-03 | 新增 OTA 固件升级通道（HTTP 下载 + `ota/report` 进度上报） |
