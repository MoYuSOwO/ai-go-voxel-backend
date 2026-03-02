package server

import (
	"log"
	"time"

	"google.golang.org/protobuf/proto"
	pb "voxel-backend/proto"
	"voxel-backend/world"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// ============================= 基础定义 =============================
// 注意：事件和Channel定义已移动到world包，此处直接使用world包中的定义

// ============================= 基础Router =============================

// BaseRouter 基础路由，提供公共功能
// 注意：Zinx框架的znet.BaseRouter已经提供了基础实现
// 但我们可能需要自定义一些公共逻辑，这里暂时直接使用znet.BaseRouter
// 具体Router可以嵌入znet.BaseRouter

// ============================= LoginRouter =============================

// LoginRouter 处理登录请求
type LoginRouter struct {
	znet.BaseRouter
}

// Handle 处理登录请求
// 核心搬运工逻辑：解析Protobuf → 封装为PacketEvent → 扔给World容器
func (r *LoginRouter) Handle(request ziface.IRequest) {
	// 1. 从Zinx请求中获取Protobuf消息
	msg := request.GetMessage()
	loginReq := &pb.LoginRequest{}

	// 使用proto.Unmarshal解析消息数据
	// 注意：msg.GetData()返回的是[]byte，即Protobuf编码的二进制数据
	if err := proto.Unmarshal(msg.GetData(), loginReq); err != nil {
		// 协议解析失败，根据安全军规必须关闭连接
		log.Printf("登录请求Protobuf解析失败: %v, 关闭连接", err)
		request.GetConnection().Stop()
		return
	}

	// 2. 封装为PacketEvent（仅搬运数据）
	event := world.PacketEvent{
		ConnID:   request.GetConnection().GetConnID(),
		PacketID: msg.GetMsgID(),
		Data:     loginReq,
		Conn:     request.GetConnection(), // 保留连接引用用于后续回包
	}

	// 3. 将事件通过Channel发送给World容器（绝对不在Router中写业务逻辑！）
	// 根据安全军规：Channel发送必须设置超时，防止World容器阻塞导致Router挂起
	select {
	case world.PacketChannel <- event:
		// 成功提交给游戏逻辑层
		log.Printf("登录请求已转发到World容器, 连接ID: %d, 玩家名: %s",
			event.ConnID, loginReq.PlayerName)
	case <-time.After(100 * time.Millisecond):
		// Channel满，记录日志但不断开连接
		log.Printf("警告: PacketChannel已满，丢弃登录请求, 连接ID: %d", event.ConnID)
	}
}

// ============================= MoveRouter =============================

// MoveRouter 处理玩家移动
type MoveRouter struct {
	znet.BaseRouter
}

// Handle 处理玩家移动请求
func (r *MoveRouter) Handle(request ziface.IRequest) {
	msg := request.GetMessage()
	moveReq := &pb.PlayerMove{}

	if err := proto.Unmarshal(msg.GetData(), moveReq); err != nil {
		log.Printf("玩家移动请求Protobuf解析失败: %v, 关闭连接", err)
		request.GetConnection().Stop()
		return
	}

	event := world.PacketEvent{
		ConnID:   request.GetConnection().GetConnID(),
		PacketID: msg.GetMsgID(),
		Data:     moveReq,
		Conn:     request.GetConnection(),
	}

	select {
	case world.PacketChannel <- event:
		// 成功提交
		log.Printf("移动请求已转发, 连接ID: %d, 玩家ID: %d",
			event.ConnID, moveReq.PlayerId)
	case <-time.After(100 * time.Millisecond):
		log.Printf("警告: PacketChannel已满，丢弃移动请求, 连接ID: %d", event.ConnID)
	}
}

// ============================= BlockRouter =============================

// BlockRouter 处理方块放置和破坏
type BlockRouter struct {
	znet.BaseRouter
}

// Handle 处理方块操作请求
// 注意：这个Router同时处理PLACE_BLOCK和BREAK_BLOCK两种消息
func (r *BlockRouter) Handle(request ziface.IRequest) {
	msg := request.GetMessage()
	packetID := msg.GetMsgID()

	// 根据消息ID决定解析哪种消息
	switch packetID {
	case MsgIDPlaceBlock:
		placeReq := &pb.PlaceBlock{}
		if err := proto.Unmarshal(msg.GetData(), placeReq); err != nil {
			log.Printf("放置方块请求Protobuf解析失败: %v, 关闭连接", err)
			request.GetConnection().Stop()
			return
		}

		event := world.PacketEvent{
			ConnID:   request.GetConnection().GetConnID(),
			PacketID: packetID,
			Data:     placeReq,
			Conn:     request.GetConnection(),
		}

		select {
		case world.PacketChannel <- event:
			log.Printf("放置方块请求已转发, 连接ID: %d, 玩家ID: %d",
				event.ConnID, placeReq.PlayerId)
		case <-time.After(100 * time.Millisecond):
			log.Printf("警告: PacketChannel已满，丢弃放置方块请求, 连接ID: %d", event.ConnID)
		}

	case MsgIDBreakBlock:
		breakReq := &pb.BreakBlock{}
		if err := proto.Unmarshal(msg.GetData(), breakReq); err != nil {
			log.Printf("破坏方块请求Protobuf解析失败: %v, 关闭连接", err)
			request.GetConnection().Stop()
			return
		}

		event := world.PacketEvent{
			ConnID:   request.GetConnection().GetConnID(),
			PacketID: packetID,
			Data:     breakReq,
			Conn:     request.GetConnection(),
		}

		select {
		case world.PacketChannel <- event:
			log.Printf("破坏方块请求已转发, 连接ID: %d, 玩家ID: %d",
				event.ConnID, breakReq.PlayerId)
		case <-time.After(100 * time.Millisecond):
			log.Printf("警告: PacketChannel已满，丢弃破坏方块请求, 连接ID: %d", event.ConnID)
		}

	default:
		log.Printf("未知方块操作消息ID: %d, 关闭连接", packetID)
		request.GetConnection().Stop()
	}
}
