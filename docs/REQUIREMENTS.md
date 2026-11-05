# IoT-DMP 需求规格说明书

> 版本: v2.0 | 更新时间: 2026-10-26

---

## 1. 项目概述

### 1.1 项目定位

物联网设备管理平台（IoT Device Management Platform），为设备厂商、系统集成商和终端用户提供设备全生命周期管理、数据采集、远程控制、规则自动化、智能场景等一站式 IoT 能力。

### 1.2 技术栈

| 层次 | 技术 | 说明 |
|------|------|------|
| 主后端 | Go 1.25 + Gin | REST API + WebSocket |
| 数据存储 | MySQL 8.0 (xorm) | 关系型数据库，设备/模型/规则等业务数据（xorm v1.4.3 ORM，自动建表） |
| 缓存 | Redis 6.0 | 会话缓存、规则引擎状态、分布式锁 |
| 时序数据 | InfluxDB 2.0 | 设备属性/事件历史曲线 |
| 边缘网关 | Go + gnet | 自研 MQTT Broker，高性能网络框架 |
| 管理后台 | Vue 3 + Element Plus + Pinia + Vite | admin 前端 |
| 手机端 | UniApp (Vue 3) | 跨端 H5/微信小程序 |
| 设备模拟器 | Go + paho.mqtt.golang | 集成测试 build tag |
| 外部集成 | 微信小程序 SDK、微信支付 V3、DeepSeek AI | AI 语音助手/支付 |
| 部署 | Docker | Dockerfile.master / Dockerfile |

### 1.3 架构

```
┌─────────────────────────────────────────────────────────┐
│                  IoT-DMP 系统架构                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │ admin 管理后台 │  │ app 手机端    │  │ 开放 API 调用方 │  │
│  └──────┬───────┘  └──────┬───────┘  └───────┬───────┘  │
│         │ HTTP              │ HTTP             │ HTTP    │
│         ▼                   ▼                  ▼        │
│  ┌─────────────────────────────────────────────────┐    │
│  │  master 主后端（Gin + MySQL(xorm) + Redis/Memory）│    │
│  │  /api/* 业务路由                                  │    │
│  │  /open/* 开放接口                                  │    │
│  └────────────────────┬────────────────────────────┘    │
│                       │                                 │
│            ┌──────────┴──────────┐                       │
│            ▼                     ▼                       │
│  ┌─────────────────────┐  ┌─────────────────────────┐   │
│  │ gateway 内嵌/独立    │  │ MySQL / Redis / InfluxDB  │   │
│  │ pkg/broker MQTT      │  │                         │   │
│  │ engine 规则引擎      │  │                         │   │
│  │ handler 消息处理     │  │                         │   │
│  └──────────┬──────────┘  └─────────────────────────┘   │
│             │ MQTT                                        │
│             ▼                                             │
│  ┌─────────────────────┐                                 │
│  │ test/ 设备模拟器      │  ← go test -tags integration │
│  └─────────────────────┘                                 │
└─────────────────────────────────────────────────────────┘
```

---

## 2. 目录结构

```
iot-dmp/
├── main.go                       # 主后端入口
├── cmd/gateway/                  # 边缘网关独立程序入口
├── test/                         # 集成测试（-tags integration）
│   └── device_test.go            # 设备模拟器
│
├── internal/                     # 项目私有代码
│   ├── api/                      # HTTP handler + 路由注册
│   ├── config/                   # 配置加载
│   ├── gateway/                  # 网关引擎 + handler（MQTT 消息处理）
│   ├── types/                    # 业务数据结构
│   └── weixin/                   # 微信对接
│
├── pkg/                          # 公共可复用库（第三方可用）
│   ├── broker/                   # 自研 MQTT Broker（gnet）
│   ├── web/                      # Gin 引擎封装、中间件
│   ├── db/                       # MySQL(xorm) 封装层（Engine + CRUD + ID 生成）
│   ├── memstore/                 # 缓存抽象层（Redis 或 内存，自动降级）
│   ├── curd/                     # 泛型 CRUD 框架
│   ├── history/                  # InfluxDB 时序数据
│   ├── lib/                      # 工具库（Cache/Map/RingBuffer）
│   ├── logger/                   # 日志封装
│   └── pricing/                  # 计价计费
│
├── admin/                        # 管理后台前端（Vue3）
├── app/                          # 手机端（UniApp）
├── go.mod / go.sum
├── master.yaml / dbus.yaml / device.yaml
├── Dockerfile / Dockerfile.master
├── build.sh / build_linux.sh
└── docs/                         # 文档
```

---

## 3. 核心模块需求

### 3.1 用户与认证

| 功能 | 需求 |
|------|------|
| 用户注册/登录 | 用户名密码注册 + JWT 登录态 |
| 管理员 | `admin=true` 用户拥有系统设置入口 |
| 个人中心 | 查看/修改昵称、邮箱、手机、头像 |
| 子账号 | 主账号可创建/管理子账号 |
| 登录锁定 | `lock=true` 禁止密码修改 |
| 微信登录 | app 端支持微信 code2session |
| JWT | 3 套密钥（api / node / open），分角色签发 |

### 3.2 设备管理

| 功能 | 需求 |
|------|------|
| 设备注册 | `/open/register` SN + 物模型 → deviceID |
| 设备列表 | 分页 + 关键字 + 分组筛选 |
| 设备详情 | 基础信息、实时状态、历史曲线、事件、日志 |
| 设备控制 | 通过 MQTT 下发属性设置/读取/服务调用 |
| 分组管理 | 树状结构，设备归属多个分组 |
| 异常报警 | 规则引擎触发的告警事件列表 |
| 设备日志 | 设备操作/运维日志搜索 |
| 子设备 | 网关型设备下挂载子设备 |
| 共享/绑定 | 设备 owner 与 family/user 的绑定关系 |

### 3.3 物模型

| 功能 | 需求 |
|------|------|
| 物模型 CRUD | 属性/事件/服务/配置/控件的可视化定义 |
| 模型设计器 | 画布式编辑器，支持拖拽 |
| 模型校验 | 网关引擎启动时按模型校验设备上报 |

### 3.4 规则引擎

| 功能 | 需求 |
|------|------|
| 触发类型 | 属性变化 / 事件上报 / 上下线 / 定时 |
| 条件求值 | expr-lang 表达式引擎 + 结构化点选配置 |
| 报警状态机 | 正常 → 待确认 → 告警中 → 恢复（Redis 存储运行时状态） |
| 动作执行 | 设备控制 / 通知 / 场景联动等 |
| 规则日志 | 每次执行的输入/条件匹配/动作结果 |
| 旁路缓冲 | 独立 goroutine 消费，不阻塞设备上报主链路 |

### 3.5 智能场景与定时任务

| 功能 | 需求 |
|------|------|
| 场景 CRUD | 在线模式（master 执行）+ 离线模式（gateway 执行） |
| 场景动作 | 按序执行，单设备不在线时标记失败继续后续动作 |
| 手动执行 | 支持即时触发验证 |
| 定时任务 | 5 段 cron，定时触发场景；与规则引擎共用 cron 匹配器 |
| 过滤加载 | 按 `mode` 字段过滤，避免 master 和 gateway 重复执行 |

### 3.6 家庭管理

| 功能 | 需求 |
|------|------|
| 家庭 CRUD | 用户可创建多个家庭，设置图标/名称 |
| 房间 CRUD | 家庭下划分房间，排序 |
| 成员管理 | 邀请/移除/设为管理员，family_id 归属 |
| 邀请码 | 生成 UUID 格式邀请码，5 分钟有效 |
| 设备归属 | 设备通过 binding 关联家庭 + 房间 |

### 3.7 组态大屏

| 功能 | 需求 |
|------|------|
| 画面 CRUD | 画布创建/编辑/预览 |
| 组件库 | 设备控件（图表/仪表/开关等） |
| 实时数据 | WebSocket 订阅设备属性 |

### 3.8 协议模板

| 功能 | 需求 |
|------|------|
| 完整 CRUD | 管理员可见，定义 Modbus/自定义协议寄存器映射 |

---

## 4. 数据模型概要

| 集合 | 说明 | 关键字段 |
|------|------|---------|
| user | 用户 | username, password(hash), nickname, admin, lock, balance(分), open_id, union_id, app_open_id |
| device | 设备 | model_id, user_id, family_id, online, status(map), firmware |
| model | 物模型 | name, type(direct/gateway), properties, events, services, configs, controls |
| group | 分组 | name, parent_id（树状） |
| family | 家庭 | name, icon, user_id(owner) |
| room | 房间 | name, family_id, sort |
| member | 家庭成员 | family_id, user_id, admin(bool) |
| binding | 设备绑定 | device_id, user_id / family_id, room_id |
| rule | 规则 | name, trigger_type, condition(expr), enabled, actions, state(status) |
| scene | 场景 | name, mode(online/offline), enabled, actions |
| task | 定时任务 | name, cron(5段), scene_id, enabled |
| protocol | 协议模板 | name, registers, description |
| scada | 组态画面 | name, data(canvas JSON), components |
| component | 组件库 | name, type, schema |
| device_event | 设备事件 | device_id, name, type(info/alert/fault), params, status(open/confirmed/resolved), handled |
| device_log | 设备日志 | device_id, type, message, user_name |

---

## 5. 部署方式

### 5.1 单机模式（开发/小体量）

master 内嵌 gateway broker，一个进程搞定：

```yaml
# master.yaml
server:
  host: 0.0.0.0
  port: 8080
mysql:
  dsn: root:password@tcp(127.0.0.1:3306)/iot_dmp?charset=utf8mb4&parseTime=True&loc=Local
redis:
  addr: 127.0.0.1:6379   # 留空则自动降级为内存存储
```

```bash
go run .                     # master（自动内嵌 MQTT broker）
cd admin && npm run dev      # 管理后台
```

### 5.2 多机模式（生产/大规模）

master 和 gateway 独立部署，gateway 直连 MySQL/Redis/InfluxDB：

```bash
# 机器 A：master（无 MQTT，纯 HTTP）
go run .

# 机器 B：gateway（MQTT broker + 规则引擎）
go run ./cmd/gateway     # 读取 dbus.yaml
```

gateway 的 `dbus.yaml` 需配置 MySQL/Redis 直连地址。

### 5.3 Docker

```bash
# 主后端
docker build -f Dockerfile.master -t dmp-master .
docker run -p 8080:8080 dmp-master

# 全部（master + gateway）
docker build -t iot-dmp .
docker run -p 8080:8080 -p 8081:8081 -p 1883:1883 iot-dmp
```

---

## 6. 测试

### 6.1 单元测试

```bash
go test ./...
go test ./internal/api/...
```

### 6.2 集成测试（设备模拟器）

需要后端和 gateway 已启动：

```bash
go test -tags integration -v ./test/ -timeout 5m -run TestDeviceSimulator
```

或编译为二进制独立运行：

```bash
go test -tags integration -c -o test-device.exe ./test/
./test-device.exe -test.v -test.timeout 5m
```

### 6.3 API 冒烟测试

```bash
# 登录获取 token
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'

# 设备列表
curl -X POST http://localhost:8080/api/device/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"page":1,"page_size":10}'
```
