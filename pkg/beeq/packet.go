package beeq

import (
	"bytes"
	"encoding/binary"
)

// MQTT 协议命令类型常量
const (
	CMD_CONNECT     = 1  // CONNECT: 客户端连接请求
	CMD_CONNACK     = 2  // CONNACK: 连接确认
	CMD_PUBLISH     = 3  // PUBLISH: 发布消息
	CMD_PUBACK      = 4  // PUBACK: QoS 1消息确认
	CMD_PUBREC      = 5  // PUBREC: QoS 2消息接收确认
	CMD_PUBREL      = 6  // PUBREL: QoS 2消息释放
	CMD_PUBCOMP     = 7  // PUBCOMP: QoS 2消息完成
	CMD_SUBSCRIBE   = 8  // SUBSCRIBE: 订阅请求
	CMD_SUBACK      = 9  // SUBACK: 订阅确认
	CMD_UNSUBSCRIBE = 10 // UNSUBSCRIBE: 取消订阅请求
	CMD_UNSUBACK    = 11 // UNSUBACK: 取消订阅确认
	CMD_PINGREQ     = 12 // PINGREQ: 心跳请求
	CMD_PINGRESP    = 13 // PINGRESP: 心跳响应
	CMD_DISCONNECT  = 14 // DISCONNECT: 断开连接
)

// Packet MQTT报文结构
// 包含固定头部、剩余长度和负载
type Packet struct {
	FixedHeader  byte   // 固定头部（命令类型 + 标志位）
	RemainingLen int    // 剩余长度
	Payload      []byte // 报文负载
}

// ReadPacket 从缓冲区读取一个MQTT报文
// buf: 包含报文数据的缓冲区
// 返回值: 解析后的报文和已消费的字节数，报文为空表示数据不完整
func ReadPacket(buf []byte) (*Packet, int) {
	if len(buf) < 2 {
		return nil, 0
	}

	fh := buf[0]
	offset := 1
	multiplier := 1
	remainingLen := 0

	// MQTT剩余长度采用可变长度编码，最多4字节
	for i := 0; i < 4; i++ {
		if offset >= len(buf) {
			return nil, 0
		}
		byteValue := buf[offset]
		offset++
		remainingLen += int(byteValue&0x7F) * multiplier
		if byteValue&0x80 == 0 {
			break
		}
		multiplier *= 128
	}

	totalLen := offset + remainingLen
	if len(buf) < totalLen {
		return nil, 0
	}

	return &Packet{
		FixedHeader:  fh,
		RemainingLen: remainingLen,
		Payload:      buf[offset:totalLen],
	}, totalLen
}

// Type 获取报文命令类型
// 返回值: 命令类型常量
func (p *Packet) Type() byte {
	return p.FixedHeader >> 4
}

// Dup 获取Dup标志位
// 返回值: true表示重复发送，false表示首次发送
func (p *Packet) Dup() bool {
	return (p.FixedHeader & 0x08) != 0
}

// QoS 获取QoS等级
// 返回值: 0表示至多一次，1表示至少一次，2表示恰好一次
func (p *Packet) QoS() byte {
	return (p.FixedHeader >> 1) & 0x03
}

// Retain 获取Retain标志位
// 返回值: true表示保留消息，false表示不保留
func (p *Packet) Retain() bool {
	return (p.FixedHeader & 0x01) != 0
}

// Build 构建完整的MQTT报文字节流
// 返回值: 完整的报文字节数组
func (p *Packet) Build() []byte {
	length := len(p.Payload)
	buf := make([]byte, 1+4+length)
	buf[0] = p.FixedHeader

	offset := 1
	if length == 0 {
		buf[offset] = 0x00
		offset++
	} else {
		for length > 0 {
			byteValue := byte(length % 128)
			length /= 128
			if length > 0 {
				byteValue |= 0x80
			}
			buf[offset] = byteValue
			offset++
		}
	}

	copy(buf[offset:], p.Payload)
	return buf[:offset+len(p.Payload)]
}

// ConnectPacket CONNECT报文结构
type ConnectPacket struct {
	ProtocolName    string // 协议名称，固定为"MQTT"
	ProtocolVersion byte   // 协议版本，3.1.1为4
	UsernameFlag    bool   // 是否包含用户名
	PasswordFlag    bool   // 是否包含密码
	WillRetain      bool   // Will消息是否保留
	WillQoS         byte   // Will消息的QoS等级
	WillFlag        bool   // 是否设置Will消息
	CleanSession    bool   // 是否清除会话
	KeepAlive       uint16 // 保持连接时间(秒)
	ClientID        string // 客户端ID，即设备ID
	Username        string // 用户名
	Password        string // 密码
	WillTopic       string // Will消息主题
	WillMessage     []byte // Will消息内容
}

// ParseConnect 解析CONNECT报文负载
// payload: CONNECT报文的负载数据
// 返回值: 解析后的ConnectPacket，nil表示解析失败
func ParseConnect(payload []byte) *ConnectPacket {
	p := &ConnectPacket{}
	offset := 0

	if offset+2 > len(payload) {
		return nil
	}
	protoLen := binary.BigEndian.Uint16(payload[offset:])
	offset += 2
	if offset+int(protoLen) > len(payload) {
		return nil
	}
	p.ProtocolName = string(payload[offset : offset+int(protoLen)])
	offset += int(protoLen)

	if offset >= len(payload) {
		return nil
	}
	p.ProtocolVersion = payload[offset]
	offset++

	if offset >= len(payload) {
		return nil
	}
	flags := payload[offset]
	offset++
	p.UsernameFlag = (flags & 0x80) != 0
	p.PasswordFlag = (flags & 0x40) != 0
	p.WillRetain = (flags & 0x20) != 0
	p.WillQoS = (flags >> 3) & 0x03
	p.WillFlag = (flags & 0x04) != 0
	p.CleanSession = (flags & 0x02) != 0

	if offset+2 > len(payload) {
		return nil
	}
	p.KeepAlive = binary.BigEndian.Uint16(payload[offset:])
	offset += 2

	if offset+2 > len(payload) {
		return nil
	}
	clientIDLen := binary.BigEndian.Uint16(payload[offset:])
	offset += 2
	if offset+int(clientIDLen) > len(payload) {
		return nil
	}
	p.ClientID = string(payload[offset : offset+int(clientIDLen)])
	offset += int(clientIDLen)

	if p.WillFlag {
		if offset+2 > len(payload) {
			return nil
		}
		willTopicLen := binary.BigEndian.Uint16(payload[offset:])
		offset += 2
		if offset+int(willTopicLen) > len(payload) {
			return nil
		}
		p.WillTopic = string(payload[offset : offset+int(willTopicLen)])
		offset += int(willTopicLen)

		if offset+2 > len(payload) {
			return nil
		}
		willMessageLen := binary.BigEndian.Uint16(payload[offset:])
		offset += 2
		if offset+int(willMessageLen) > len(payload) {
			return nil
		}
		p.WillMessage = payload[offset : offset+int(willMessageLen)]
		offset += int(willMessageLen)
	}

	if p.UsernameFlag {
		if offset+2 > len(payload) {
			return nil
		}
		usernameLen := binary.BigEndian.Uint16(payload[offset:])
		offset += 2
		if offset+int(usernameLen) > len(payload) {
			return nil
		}
		p.Username = string(payload[offset : offset+int(usernameLen)])
		offset += int(usernameLen)
	}

	if p.PasswordFlag {
		if offset+2 > len(payload) {
			return nil
		}
		passwordLen := binary.BigEndian.Uint16(payload[offset:])
		offset += 2
		if offset+int(passwordLen) > len(payload) {
			return nil
		}
		p.Password = string(payload[offset : offset+int(passwordLen)])
		offset += int(passwordLen)
	}

	return p
}

// BuildConnack 构建CONNACK报文
// sessionPresent: 是否存在会话
// returnCode: 返回码，0表示成功，其他表示失败原因
// 返回值: CONNACK报文
func BuildConnack(sessionPresent bool, returnCode byte) *Packet {
	payload := make([]byte, 2)
	if sessionPresent {
		payload[0] = 0x01
	}
	payload[1] = returnCode
	return &Packet{
		FixedHeader: CMD_CONNACK << 4,
		Payload:     payload,
	}
}

// PublishPacket PUBLISH报文结构
type PublishPacket struct {
	TopicName string // 消息主题
	PacketID  uint16 // 消息包ID（QoS 0时为0）
	Payload   []byte // 消息负载
	Dup       bool   // 是否重复发送
	QoS       byte   // QoS等级
	Retain    bool   // 是否保留消息
}

// ParsePublish 解析PUBLISH报文负载
// payload: PUBLISH报文的负载数据
// qos: QoS等级
// 返回值: 解析后的PublishPacket，nil表示解析失败
func ParsePublish(payload []byte, qos byte) *PublishPacket {
	p := &PublishPacket{}
	offset := 0

	if offset+2 > len(payload) {
		return nil
	}
	topicLen := binary.BigEndian.Uint16(payload[offset:])
	offset += 2
	if offset+int(topicLen) > len(payload) {
		return nil
	}
	p.TopicName = string(payload[offset : offset+int(topicLen)])
	offset += int(topicLen)

	if qos > 0 {
		if offset+2 > len(payload) {
			return nil
		}
		p.PacketID = binary.BigEndian.Uint16(payload[offset:])
		offset += 2
	}

	p.Payload = payload[offset:]
	return p
}

// BuildPublish 构建PUBLISH报文
// p: PublishPacket结构
// 返回值: PUBLISH报文
func BuildPublish(p *PublishPacket) *Packet {
	buf := &bytes.Buffer{}
	topicLen := make([]byte, 2)
	binary.BigEndian.PutUint16(topicLen, uint16(len(p.TopicName)))
	buf.Write(topicLen)
	buf.WriteString(p.TopicName)

	if p.QoS > 0 {
		packetID := make([]byte, 2)
		binary.BigEndian.PutUint16(packetID, p.PacketID)
		buf.Write(packetID)
	}

	buf.Write(p.Payload)

	fh := byte(CMD_PUBLISH<<4) | boolToByte(p.Dup)<<3 | p.QoS<<1 | boolToByte(p.Retain)
	return &Packet{
		FixedHeader: fh,
		Payload:     buf.Bytes(),
	}
}

// BuildPuback 构建PUBACK报文
// packetID: 消息包ID
// 返回值: PUBACK报文
func BuildPuback(packetID uint16) *Packet {
	payload := make([]byte, 2)
	binary.BigEndian.PutUint16(payload, packetID)
	return &Packet{
		FixedHeader: CMD_PUBACK << 4,
		Payload:     payload,
	}
}

// BuildPubrec 构建PUBREC报文
// packetID: 消息包ID
// 返回值: PUBREC报文
func BuildPubrec(packetID uint16) *Packet {
	payload := make([]byte, 2)
	binary.BigEndian.PutUint16(payload, packetID)
	return &Packet{
		FixedHeader: CMD_PUBREC << 4,
		Payload:     payload,
	}
}

// BuildPubrel 构建PUBREL报文
// packetID: 消息包ID
// 返回值: PUBREL报文
func BuildPubrel(packetID uint16) *Packet {
	payload := make([]byte, 2)
	binary.BigEndian.PutUint16(payload, packetID)
	return &Packet{
		FixedHeader: CMD_PUBREL<<4 | 0x02,
		Payload:     payload,
	}
}

// BuildPubcomp 构建PUBCOMP报文
// packetID: 消息包ID
// 返回值: PUBCOMP报文
func BuildPubcomp(packetID uint16) *Packet {
	payload := make([]byte, 2)
	binary.BigEndian.PutUint16(payload, packetID)
	return &Packet{
		FixedHeader: CMD_PUBCOMP << 4,
		Payload:     payload,
	}
}

// SubscribePacket SUBSCRIBE报文结构
type SubscribePacket struct {
	PacketID uint16 // 消息包ID
	Topics   []struct {
		Name string // 订阅主题
		QoS  byte   // QoS等级
	}
}

// ParseSubscribe 解析SUBSCRIBE报文负载
// payload: SUBSCRIBE报文的负载数据
// 返回值: 解析后的SubscribePacket，nil表示解析失败
func ParseSubscribe(payload []byte) *SubscribePacket {
	p := &SubscribePacket{}
	offset := 0

	if offset+2 > len(payload) {
		return nil
	}
	p.PacketID = binary.BigEndian.Uint16(payload[offset:])
	offset += 2

	for offset < len(payload) {
		if offset+2 > len(payload) {
			return nil
		}
		topicLen := binary.BigEndian.Uint16(payload[offset:])
		offset += 2
		if offset+int(topicLen) > len(payload) {
			return nil
		}
		topic := string(payload[offset : offset+int(topicLen)])
		offset += int(topicLen)

		if offset >= len(payload) {
			return nil
		}
		qos := payload[offset]
		offset++

		p.Topics = append(p.Topics, struct {
			Name string
			QoS  byte
		}{topic, qos})
	}

	return p
}

// BuildSuback 构建SUBACK报文
// packetID: 消息包ID
// qosList: 每个订阅主题的QoS等级列表
// 返回值: SUBACK报文
func BuildSuback(packetID uint16, qosList []byte) *Packet {
	buf := &bytes.Buffer{}
	packetIDBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(packetIDBytes, packetID)
	buf.Write(packetIDBytes)
	buf.Write(qosList)
	return &Packet{
		FixedHeader: CMD_SUBACK << 4,
		Payload:     buf.Bytes(),
	}
}

// UnsubscribePacket UNSUBSCRIBE报文结构
type UnsubscribePacket struct {
	PacketID uint16   // 消息包ID
	Topics   []string // 要取消订阅的主题列表
}

// ParseUnsubscribe 解析UNSUBSCRIBE报文负载
// payload: UNSUBSCRIBE报文的负载数据
// 返回值: 解析后的UnsubscribePacket，nil表示解析失败
func ParseUnsubscribe(payload []byte) *UnsubscribePacket {
	p := &UnsubscribePacket{}
	offset := 0

	if offset+2 > len(payload) {
		return nil
	}
	p.PacketID = binary.BigEndian.Uint16(payload[offset:])
	offset += 2

	for offset < len(payload) {
		if offset+2 > len(payload) {
			return nil
		}
		topicLen := binary.BigEndian.Uint16(payload[offset:])
		offset += 2
		if offset+int(topicLen) > len(payload) {
			return nil
		}
		topic := string(payload[offset : offset+int(topicLen)])
		offset += int(topicLen)
		p.Topics = append(p.Topics, topic)
	}

	return p
}

// BuildUnsuback 构建UNSUBACK报文
// packetID: 消息包ID
// 返回值: UNSUBACK报文
func BuildUnsuback(packetID uint16) *Packet {
	payload := make([]byte, 2)
	binary.BigEndian.PutUint16(payload, packetID)
	return &Packet{
		FixedHeader: CMD_UNSUBACK << 4,
		Payload:     payload,
	}
}

// BuildPingresp 构建PINGRESP报文
// 返回值: PINGRESP报文
func BuildPingresp() *Packet {
	return &Packet{
		FixedHeader: CMD_PINGRESP << 4,
		Payload:     nil,
	}
}

// boolToByte 将布尔值转换为字节
// b: 布尔值
// 返回值: 1表示true，0表示false
func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
