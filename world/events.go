package world

import (
	"github.com/aceld/zinx/ziface"
	"google.golang.org/protobuf/proto"
)

// ============================= 事件定义 =============================

// PacketEvent 数据包事件 - 从Zinx Router传递到World容器
// 这是网络层与游戏逻辑层之间的桥梁，遵循"搬运工"模式
type PacketEvent struct {
	ConnID   uint64             // Zinx连接ID（Zinx使用uint64作为连接ID）
	PacketID uint32             // 对应Protobuf PacketID
	Data     interface{}        // 解析后的Protobuf消息（如*proto.LoginRequest）
	Conn     ziface.IConnection // Zinx连接引用（用于回包）
}

// ConnectionEvent 连接事件
type ConnectionEvent struct {
	ConnID uint64
	Type   EventType // ConnectEvent / DisconnectEvent
}

// BroadcastEvent 广播事件 - 从World容器发送到网络层
type BroadcastEvent struct {
	TargetPlayerIDs []uint32      // 目标玩家ID列表（nil表示广播给所有人）
	PacketID        uint32        // 包类型
	Message         proto.Message // Protobuf消息
}

// EventType 事件类型枚举
type EventType uint8

const (
	ConnectEvent    EventType = 1
	DisconnectEvent EventType = 2
)

// ============================= 通道定义 =============================

// 全局Channel定义，用于网络层与游戏逻辑层之间的通信
// 注意：这些Channel是单向的，确保层与层之间的解耦
var (
	// PacketChannel 用于从Router传递数据包到World容器
	// 缓冲大小可配置，默认512，防止网络层阻塞
	PacketChannel chan PacketEvent

	// ConnectionChannel 用于传递连接事件
	ConnectionChannel chan ConnectionEvent

	// BroadcastChannel 用于从World容器发送广播到网络层
	BroadcastChannel chan BroadcastEvent
)

// InitializeChannels 初始化全局Channel
// 必须在服务器启动前调用，设置适当的缓冲大小
func InitializeChannels(packetBuffer, connBuffer, broadcastBuffer int) {
	PacketChannel = make(chan PacketEvent, packetBuffer)
	ConnectionChannel = make(chan ConnectionEvent, connBuffer)
	BroadcastChannel = make(chan BroadcastEvent, broadcastBuffer)
}

// GetPacketChannel 获取PacketChannel（只读视图）
func GetPacketChannel() <-chan PacketEvent {
	return PacketChannel
}

// GetConnectionChannel 获取ConnectionChannel（只读视图）
func GetConnectionChannel() <-chan ConnectionEvent {
	return ConnectionChannel
}

// GetBroadcastChannel 获取BroadcastChannel（只读视图）
func GetBroadcastChannel() <-chan BroadcastEvent {
	return BroadcastChannel
}
