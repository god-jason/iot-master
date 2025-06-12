# 物联大师

[![Go](https://github.com/god-jason/iot-master/actions/workflows/go.yml/badge.svg)](https://github.com/god-jason/iot-master/actions/workflows/go.yml)

物联大师是开源免费的物联网数据中台，区别于传统的物联网平台，物联大师采用Go语言进行编程实现，有以下诸多优点：

- 单一程序文件，免安装
- 最低内存只需要12MB
- 插件机制，支持功能自由扩展
- 能够支持多种操作系统和处理器架构
- 支持使用Lua脚本扩展协议
- 支持使用JS脚本实现边缘计算
- 支持智能家居应用场景，定时和联动控制
- 支持WebRTC点对点视频传输


# MQTT主题规范

物联大师的插件主要通过MQTT总线相互通讯，

## 一、连接消息

MQTT主题：

```javascript
`link/${linker_name}/${link_id}/#`
```

| 消息   | 主题        | 内容   | 说明                        |
|------|-----------|------|---------------------------|
| 打开连接 | .../open  | json | {...options, remote:"ip"} |
| 关闭连接 | .../close | none |                           |
| 接收数据 | .../up    | bin  | 原始二进制                     |
| 发送数据 | .../down  | bin  | 原始二进制                     |

### 备注：

1. linker_name：连接器名称，比如：serial、can、tcp-client、tcp-server、udp-client、udp-server、udp-group、gnet-server
2. link_id：连接ID

## 二、协议插件消息

MQTT主题：

```javascript
`protocol/${protocol_name}/${linker_name}/${link_id}/#`
```

| 消息   | 主题         | 内容   | 说明                                          |
|------|------------|------|---------------------------------------------|
| 打开连接 | .../open   | json | {devices:[{product_id,device_id,slave...}]} |
| 关闭连接 | .../close  | none |                                             |
| 接收数据 | .../up     | bin  | 将原始数据定向发送到协议插件                              |
| 数据轮询 | .../poll   | json | {msg_id, }                                  |
| 数据同步 | .../sync   | json | {msg_id, device_id}                         |
| 读取数据 | .../read   | json | {msg_id, device_id, points:[]}              |
| 写入数据 | .../write  | json | {msg_id, device_id, values:{k:v}}           |
| 执行操作 | .../action | json | {msg_id, device_id, action, parameters}     |

### 备注：

1. protocol_name：协议名称，推荐与插件名称一致，比如：modbus、dlt645、iec104……
2. linker_name：连接器名称，目的是方便做数据下发，不用二次转换
3. 修改设备，实际操作是先删除，再添加
4. 发送数据使用 link/${linker_name}/${link_id}/down
5. 插件需要订阅产品消息，同步产品配置
6. msg_id：消息ID

## 三、产品消息

MQTT主题：

```javascript
`product/${product_id}/#`
```

| 消息   | 主题                        | 内容   | 说明 |
|------|---------------------------|------|----|
| 产品配置 | .../config/${config_name} | json |    |
| 产品模型 | .../model                 | json |    |

### 备注：

1. config_name：配置名称，一般是协议名称，比如：modbus、dlt645、iec104……

## 四、设备消息

MQTT主题：

```javascript
`device/${device_id}/#`
```

| 消息     | 主题                  | 内容   | 说明                                       |
|--------|---------------------|------|------------------------------------------|
| 上传属性   | .../values          | json | {k:v}                                    |
| 读取属性   | .../sync            | json | {msg_id}                                 |
| 读取属性响应 | .../sync/response   | json | {msg_id, values:{k:v}}                   |
| 读取属性   | .../read            | json | {msg_id, points:[k,]}                    |
| 读取属性响应 | .../read/response   | json | {msg_id, values:{k:v}}                   |
| 修改属性   | .../write           | json | {msg_id, values:{k:v}}                   |
| 修改属性响应 | .../write/response  | json | {msg_id, results:{k:bool}}               |
| 上报事件   | .../event           | json | {title, message, level,}                 |
| 执行操作   | .../action          | json | {msg_id, action:"reboot", parameters:{}} |
| 执行操作响应 | .../action/response | json | {msg_id, action:"reboot", result: {}}    |

### 备注：

1. device_id：设备ID
2. msg_id：消息ID

## 五、项目消息

MQTT主题：

```javascript
`project/${project_id}/${device_id}/#`
```

| 消息   | 主题         | 内容   | 说明                       |
|------|------------|------|--------------------------|
| 上传属性 | .../values | json | {k:v}                    |
| 上报事件 | .../event  | json | {title, message, level,} |

### 备注：

1. project_id：项目ID
2.

## 六、空间消息

MQTT主题：

```javascript
`space/${space_id}/${device_id}/#`
```

| 消息   | 主题         | 内容   | 说明                       |
|------|------------|------|--------------------------|
| 上传属性 | .../values | json | {k:v}                    |
| 上报事件 | .../event  | json | {title, message, level,} |

### 备注：

1. space_id：空间ID

## 七、异常消息

| 消息   | 主题                       | 内容   | 说明                             |
|------|--------------------------|------|--------------------------------|
| 推送消息 | push/${device_id}/values | json | {user_id, device_id, event:{}} |

# 协议库支持统计

- [x] Modbus RTU
- [x] Modbus TCP
- [ ] Modbus ASCII（使用比较少，暂不做支持）
- [ ] PLC
-
    - [x] Siemens 西门子 s7 fetchwrite mpi ppi
-
    - [x] Mitsubishi 三菱 melsec
-
    - [x] Omron 欧姆龙 fins hostlink
-
    - [ ] AB df1
-
    - [ ] Delta 台达 dvp
-
    - [ ] Keyence 基恩士 melsec
-
    - [ ] Panasonic 松下 melsec newtocol
-
    - [ ] Fuji 富士 spb
-
    - [ ] Fatek 永宏
- [ ] BACnet智能建筑协议
- [ ] KNX智能建筑协议
- [x] DL/T645-1997、2007 多功能电表通讯规约
- [ ] DL/T698.45-2017 国网电力通讯规约
- [x] CJ/T188-2004、2018 户用计量仪表数据传输技术条件
- [ ] IEC 101/103/104 电力系统远程控制和监视的通信协议
- [ ] IEC 61850 电力系统自动化领域全球通用协议
- [ ] SL/T427-2021 水资源监测数据传输规约
- [ ] SL/T651-2014 水文监测数据通信规约
- [ ] SL/T812.1-2021 水利监测数据传输规约
- [ ] SZY206-2016 水资源监测数据传输规约

# 联系方式

南京真格智能科技有限公司 [链接](https://zgwit.com)

- 邮箱：[jason@zgwit.com](mailto:jason@zgwit.com)
- 手机：[15161515197](tel:15161515197)(微信同号)

# 开源协议

[GNU GPLv3](https://github.com/god-jason/iot-master/blob/main/LICENSE)

`补充：产品仅限个人免费使用，商业需求请联系我们`
