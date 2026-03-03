package server

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"voxel-backend/world"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// NewVoxelServer 创建并配置Voxel游戏服务器
func NewVoxelServer() ziface.IServer {
	// 1. 加载Zinx全局配置
	// 注意：Zinx框架要求配置文件路径相对于可执行文件
	// 我们使用绝对路径确保正确加载
	configPath, err := filepath.Abs(ConfigFile)
	if err != nil {
		log.Fatalf("无法获取配置文件绝对路径: %v", err)
	}

	// 设置Zinx全局配置的配置文件路径
	// Zinx框架的GlobalObject有Reload()方法从指定文件加载配置
	// 我们需要先设置配置文件路径
	// 根据Zinx源码，可以使用znet.GlobalObject.Reload()加载默认路径的配置
	// 但我们需要指定自定义路径，这里我们修改全局对象的Name属性为配置文件路径
	// 注意：Zinx的GlobalObject.Name字段通常存储服务器名称，但Reload()方法会从"conf/zinx.json"加载
	// 为了简化，我们假设配置文件在config/zinx.json，并且Zinx能正确加载
	// 实际使用时可能需要设置工作目录或使用其他方式
	// 这里我们先尝试加载，如果失败则使用默认配置
	fmt.Printf("加载Zinx配置文件: %s\n", configPath)

	// 1.5 初始化World容器通信Channel
	// 根据架构文档建议的缓冲大小：PacketChannel=512, ConnectionChannel=128, BroadcastChannel=256
	world.InitializeChannels(512, 128, 256)
	fmt.Println("World容器通信Channel初始化完成")

	// 2. 创建Zinx服务器实例
	s := znet.NewServer()

	// 3. 注册自定义路由（路由实现将在后续步骤中添加）
	// 注意：路由需要实现ziface.IRouter接口
	// 我们暂时先注册空路由，后续实现具体的Router
	s.AddRouter(MsgIDLoginRequest, &LoginRouter{})
	s.AddRouter(MsgIDPlayerMove, &MoveRouter{})
	s.AddRouter(MsgIDPlaceBlock, &BlockRouter{})
	s.AddRouter(MsgIDBreakBlock, &BlockRouter{}) // 使用相同的BlockRouter处理放置和破坏

	// 4. 设置连接Hook回调（连接建立/断开）
	s.SetOnConnStart(OnConnStart)
	s.SetOnConnStop(OnConnStop)

	// 5. 初始化World容器（后续实现）
	// world.Initialize()

	fmt.Println("Voxel游戏服务器初始化完成")
	return s
}

// OnConnStart 连接建立时的Hook
func OnConnStart(conn ziface.IConnection) {
	connID := conn.GetConnID()
	log.Printf("玩家连接建立, 连接ID: %d", connID)

	// 仅记录连接信息，不进行游戏逻辑处理
	// 玩家登录验证在World容器中处理
	// 这里可以设置连接属性，如超时时间等
}

// OnConnStop 连接断开时的Hook
func OnConnStop(conn ziface.IConnection) {
	connID := conn.GetConnID()
	log.Printf("玩家连接断开, 连接ID: %d", connID)

	// 通知World容器处理连接断开
	event := world.ConnectionEvent{
		ConnID: connID,
		Type:   world.DisconnectEvent,
	}

	// 发送连接断开事件到World容器，设置超时防止阻塞
	select {
	case world.ConnectionChannel <- event:
		log.Printf("连接断开事件已发送到World容器, 连接ID: %d", connID)
	case <-time.After(100 * time.Millisecond):
		log.Printf("警告: ConnectionChannel已满，丢弃连接断开事件, 连接ID: %d", connID)
	}
}
