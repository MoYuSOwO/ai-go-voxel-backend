package server

import (
	"testing"
)

// TestConstants 测试消息ID常量是否正确映射
func TestConstants(t *testing.T) {
	// 测试消息ID常量不为零值（除了可能的LOGIN_REQUEST为0）
	// 这里我们主要测试常量是否正确定义
	if MsgIDLoginRequest != 0 {
		t.Errorf("MsgIDLoginRequest期望为0，实际为%d", MsgIDLoginRequest)
	}
	if MsgIDLoginResponse != 1 {
		t.Errorf("MsgIDLoginResponse期望为1，实际为%d", MsgIDLoginResponse)
	}
	if MsgIDPlayerMove != 2 {
		t.Errorf("MsgIDPlayerMove期望为2，实际为%d", MsgIDPlayerMove)
	}
	if MsgIDPlaceBlock != 3 {
		t.Errorf("MsgIDPlaceBlock期望为3，实际为%d", MsgIDPlaceBlock)
	}
	if MsgIDBreakBlock != 4 {
		t.Errorf("MsgIDBreakBlock期望为4，实际为%d", MsgIDBreakBlock)
	}
	if MsgIDChunkData != 5 {
		t.Errorf("MsgIDChunkData期望为5，实际为%d", MsgIDChunkData)
	}
	if MsgIDPlayerListUpdate != 6 {
		t.Errorf("MsgIDPlayerListUpdate期望为6，实际为%d", MsgIDPlayerListUpdate)
	}
	if MsgIDHeartbeat != 7 {
		t.Errorf("MsgIDHeartbeat期望为7，实际为%d", MsgIDHeartbeat)
	}
}

// TestConfigFile 测试配置文件路径
func TestConfigFile(t *testing.T) {
	if ConfigFile != "config/zinx.json" {
		t.Errorf("ConfigFile期望为'config/zinx.json'，实际为'%s'", ConfigFile)
	}
}

// TestNewVoxelServer 测试服务器创建不panic
// 注意：这是一个简单的烟雾测试，仅验证函数调用不崩溃
func TestNewVoxelServer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewVoxelServer panicked: %v", r)
		}
	}()

	// 由于Zinx框架需要配置文件，我们无法在没有配置文件的情况下完全测试
	// 这里只测试函数调用本身，实际配置文件可能不存在
	// 在完整测试环境中应该提供测试配置文件
	_ = NewVoxelServer()
}
