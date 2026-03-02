package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"voxel-backend/server"
	"voxel-backend/world"
)

func main() {
	fmt.Println("=== Voxel游戏后端服务器启动 ===")
	fmt.Println("版本: MVP 里程碑1 (Zinx + Protobuf 集成)")
	fmt.Println("======================================")

	// 启动World容器的模拟goroutine
	// 注意：实际World容器将在后续里程碑实现
	// 这里仅演示Channel通信机制
	go simulateWorldContainer()

	// 创建并启动Zinx服务器
	s := server.NewVoxelServer()

	// 设置信号处理，实现优雅关闭
	setupSignalHandler(s)

	fmt.Println("服务器正在启动，监听端口 8888...")
	fmt.Println("使用 Ctrl+C 停止服务器")

	// 启动服务器（阻塞调用）
	s.Serve()
}

// simulateWorldContainer 模拟World容器，接收并处理事件
// 在真实实现中，World容器将运行Tick循环并处理游戏逻辑
func simulateWorldContainer() {
	log.Println("World容器模拟器启动")

	// 从Channel读取事件并打印日志
	for {
		select {
		case packetEvent := <-world.GetPacketChannel():
			log.Printf("World容器收到数据包事件: 连接ID=%d, 包ID=%d, 数据=%v",
				packetEvent.ConnID, packetEvent.PacketID, packetEvent.Data)

		case connEvent := <-world.GetConnectionChannel():
			eventType := "未知"
			switch connEvent.Type {
			case world.ConnectEvent:
				eventType = "连接"
			case world.DisconnectEvent:
				eventType = "断开连接"
			}
			log.Printf("World容器收到连接事件: 连接ID=%d, 类型=%s",
				connEvent.ConnID, eventType)

		case broadcastEvent := <-world.GetBroadcastChannel():
			log.Printf("World容器收到广播事件: 目标玩家数=%d, 包ID=%d",
				len(broadcastEvent.TargetPlayerIDs), broadcastEvent.PacketID)
		}
	}
}

// setupSignalHandler 设置信号处理，实现优雅关闭
func setupSignalHandler(s interface {
	Stop()
}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-c
		log.Printf("收到信号 %v，开始优雅关闭...", sig)

		// 停止服务器
		s.Stop()

		log.Println("服务器已停止")
		os.Exit(0)
	}()
}
