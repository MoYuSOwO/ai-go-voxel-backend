package world

import (
	"testing"
	"time"
)

// TestInitializeChannels 测试Channel初始化
func TestInitializeChannels(t *testing.T) {
	// 重置Channel（如果之前已初始化）
	PacketChannel = nil
	ConnectionChannel = nil
	BroadcastChannel = nil

	// 测试初始化
	InitializeChannels(10, 5, 8)

	if PacketChannel == nil {
		t.Error("PacketChannel未初始化")
	}
	if ConnectionChannel == nil {
		t.Error("ConnectionChannel未初始化")
	}
	if BroadcastChannel == nil {
		t.Error("BroadcastChannel未初始化")
	}

	// 测试缓冲大小
	// 注意：我们无法直接获取Channel的缓冲大小，但可以测试能否发送指定数量的消息而不阻塞
	for i := 0; i < 10; i++ {
		select {
		case PacketChannel <- PacketEvent{}:
			// 成功
		case <-time.After(10 * time.Millisecond):
			t.Errorf("PacketChannel缓冲大小可能小于10，在第%d次发送时阻塞", i)
		}
	}

	// 清理
	close(PacketChannel)
	close(ConnectionChannel)
	close(BroadcastChannel)
}

// TestEventTypes 测试事件类型枚举
func TestEventTypes(t *testing.T) {
	if ConnectEvent != 1 {
		t.Errorf("ConnectEvent期望为1，实际为%d", ConnectEvent)
	}
	if DisconnectEvent != 2 {
		t.Errorf("DisconnectEvent期望为2，实际为%d", DisconnectEvent)
	}
}

// TestGetChannelFunctions 测试Channel获取函数
func TestGetChannelFunctions(t *testing.T) {
	// 重新初始化
	InitializeChannels(5, 5, 5)
	defer func() {
		close(PacketChannel)
		close(ConnectionChannel)
		close(BroadcastChannel)
	}()

	// 测试获取只读Channel
	packetCh := GetPacketChannel()
	if packetCh == nil {
		t.Error("GetPacketChannel返回nil")
	}

	connCh := GetConnectionChannel()
	if connCh == nil {
		t.Error("GetConnectionChannel返回nil")
	}

	broadcastCh := GetBroadcastChannel()
	if broadcastCh == nil {
		t.Error("GetBroadcastChannel返回nil")
	}

	// 验证这些Channel是只读的（编译时检查，这里无法测试）
	// 但我们可以测试它们与原始Channel相同
	// 通过发送到原始Channel，从只读Channel接收来验证
	testEvent := PacketEvent{ConnID: 123}
	select {
	case PacketChannel <- testEvent:
		// 成功
	case <-time.After(10 * time.Millisecond):
		t.Error("无法发送到PacketChannel")
	}

	select {
	case received := <-packetCh:
		if received.ConnID != 123 {
			t.Errorf("从只读Channel接收到错误数据: %v", received)
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("无法从只读Channel接收")
	}
}

// TestPacketEvent 测试PacketEvent结构
func TestPacketEvent(t *testing.T) {
	event := PacketEvent{
		ConnID:   1001,
		PacketID: 1,
		Data:     "test data",
		Conn:     nil, // 实际使用中应为ziface.IConnection
	}

	if event.ConnID != 1001 {
		t.Errorf("PacketEvent.ConnID期望为1001，实际为%d", event.ConnID)
	}
	if event.PacketID != 1 {
		t.Errorf("PacketEvent.PacketID期望为1，实际为%d", event.PacketID)
	}
	if event.Data != "test data" {
		t.Errorf("PacketEvent.Data期望为'test data'，实际为'%v'", event.Data)
	}
}
