package server

import "voxel-backend/proto"

// 消息ID映射：将proto.PacketID转换为Zinx使用的uint32 MsgID
// 注意：Zinx的MsgID是uint32，而proto.PacketID是int32，需要转换
const (
	// 使用proto.PacketID的值直接作为MsgID，因为枚举值从0开始，符合要求
	// 但Zinx可能期望非零MsgID？我们保留原值
	// 这些常量用于代码可读性
	MsgIDLoginRequest     = uint32(proto.PacketID_LOGIN_REQUEST)
	MsgIDLoginResponse    = uint32(proto.PacketID_LOGIN_RESPONSE)
	MsgIDPlayerMove       = uint32(proto.PacketID_PLAYER_MOVE)
	MsgIDPlaceBlock       = uint32(proto.PacketID_PLACE_BLOCK)
	MsgIDBreakBlock       = uint32(proto.PacketID_BREAK_BLOCK)
	MsgIDChunkData        = uint32(proto.PacketID_CHUNK_DATA)
	MsgIDPlayerListUpdate = uint32(proto.PacketID_PLAYER_LIST_UPDATE)
	MsgIDHeartbeat        = uint32(proto.PacketID_HEARTBEAT)
)

// 配置文件名
const ConfigFile = "config/zinx.json"
