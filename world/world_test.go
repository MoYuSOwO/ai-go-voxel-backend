package world

import (
	"sync"
	"testing"
	"time"

	"github.com/aceld/zinx/ziface"
	"google.golang.org/protobuf/proto"
	pb "voxel-backend/proto"
)

// ============================= EntityManager 测试 =============================

// TestEntityManagerBasic 测试EntityManager基本功能
func TestEntityManagerBasic(t *testing.T) {
	em := NewEntityManager()

	// 创建测试实体
	entity := NewEntity(1, EntityTypePlayer, Vector3{X: 0, Y: 0, Z: 0})

	// 测试添加实体
	if !em.AddEntity(entity) {
		t.Error("AddEntity应该成功")
	}

	// 测试重复添加
	if em.AddEntity(entity) {
		t.Error("重复添加实体应该失败")
	}

	// 测试获取实体
	retrievedEntity, exists := em.GetEntity(1)
	if !exists {
		t.Error("GetEntity应该找到实体")
	}
	if retrievedEntity.GetID() != 1 {
		t.Error("获取的实体ID不正确")
	}

	// 测试获取不存在的实体
	_, exists = em.GetEntity(999)
	if exists {
		t.Error("不存在的实体应该返回false")
	}

	// 测试实体数量
	if em.EntityCount() != 1 {
		t.Errorf("实体数量期望为1，实际为%d", em.EntityCount())
	}

	// 测试按类型获取实体
	entities := em.GetEntitiesByType(EntityTypePlayer)
	if len(entities) != 1 {
		t.Errorf("按类型获取实体数量期望为1，实际为%d", len(entities))
	}

	// 测试移除实体
	if !em.RemoveEntity(1) {
		t.Error("RemoveEntity应该成功")
	}

	// 测试移除不存在的实体
	if em.RemoveEntity(999) {
		t.Error("移除不存在的实体应该失败")
	}

	// 测试移除后数量
	if em.EntityCount() != 0 {
		t.Errorf("移除后实体数量期望为0，实际为%d", em.EntityCount())
	}
}

// TestEntityManagerConcurrent 测试EntityManager并发安全性
func TestEntityManagerConcurrent(t *testing.T) {
	em := NewEntityManager()
	const numGoroutines = 100
	const numEntitiesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发添加实体
	for i := 0; i < numGoroutines; i++ {
		go func(startID int) {
			defer wg.Done()
			for j := 0; j < numEntitiesPerGoroutine; j++ {
				entityID := uint64(startID*numEntitiesPerGoroutine + j + 1)
				entity := NewEntity(entityID, EntityTypePlayer, Vector3{})
				em.AddEntity(entity)
			}
		}(i)
	}

	wg.Wait()

	// 验证实体数量
	expectedCount := numGoroutines * numEntitiesPerGoroutine
	if em.EntityCount() != expectedCount {
		t.Errorf("并发添加后实体数量期望为%d，实际为%d", expectedCount, em.EntityCount())
	}

	// 并发读取实体
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numEntitiesPerGoroutine; j++ {
				entityID := uint64(j + 1)
				_, exists := em.GetEntity(entityID)
				// 不检查是否存在，因为ID可能超出范围
				_ = exists
			}
		}()
	}

	wg.Wait()

	// 并发移除实体
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(startID int) {
			defer wg.Done()
			for j := 0; j < numEntitiesPerGoroutine; j++ {
				entityID := uint64(startID*numEntitiesPerGoroutine + j + 1)
				em.RemoveEntity(entityID)
			}
		}(i)
	}

	wg.Wait()

	// 验证所有实体已移除
	if em.EntityCount() != 0 {
		t.Errorf("并发移除后实体数量期望为0，实际为%d", em.EntityCount())
	}
}

// ============================= World 基本功能测试 =============================

// TestWorldBasic 测试World基本功能
func TestWorldBasic(t *testing.T) {
	world := NewWorld()

	// 测试初始状态
	if world.PlayerCount() != 0 {
		t.Errorf("新World玩家数量期望为0，实际为%d", world.PlayerCount())
	}

	// 测试ID生成
	playerID1 := world.generatePlayerID()
	playerID2 := world.generatePlayerID()
	if playerID1 == playerID2 {
		t.Error("生成的玩家ID应该不同")
	}
	if playerID1 != 1 {
		t.Errorf("第一个玩家ID期望为1，实际为%d", playerID1)
	}

	entityID1 := world.generateEntityID()
	entityID2 := world.generateEntityID()
	if entityID1 == entityID2 {
		t.Error("生成的实体ID应该不同")
	}
	if entityID1 != 1 {
		t.Errorf("第一个实体ID期望为1，实际为%d", entityID1)
	}
}

// TestWorldPlayerManagement 测试World玩家管理
func TestWorldPlayerManagement(t *testing.T) {
	world := NewWorld()

	// 创建测试玩家
	entityID := world.generateEntityID()
	playerID := world.generatePlayerID()
	player := NewPlayer(entityID, playerID, 1001, nil, Vector3{X: 0, Y: 0, Z: 0})

	// 测试添加玩家
	if !world.AddPlayer(player) {
		t.Error("AddPlayer应该成功")
	}

	// 测试重复添加
	if world.AddPlayer(player) {
		t.Error("重复添加玩家应该失败")
	}

	// 测试玩家数量
	if world.PlayerCount() != 1 {
		t.Errorf("添加玩家后数量期望为1，实际为%d", world.PlayerCount())
	}

	// 测试获取玩家
	retrievedPlayer, exists := world.GetPlayer(playerID)
	if !exists {
		t.Error("GetPlayer应该找到玩家")
	}
	if retrievedPlayer.Connection.PlayerID != playerID {
		t.Error("获取的玩家ID不正确")
	}

	// 测试通过连接ID获取玩家
	connPlayer, exists := world.GetPlayerByConnID(1001)
	if !exists {
		t.Error("GetPlayerByConnID应该找到玩家")
	}
	if connPlayer.Connection.PlayerID != playerID {
		t.Error("通过连接ID获取的玩家ID不正确")
	}

	// 测试移除玩家
	if !world.RemovePlayer(playerID) {
		t.Error("RemovePlayer应该成功")
	}

	// 测试移除不存在的玩家
	if world.RemovePlayer(999) {
		t.Error("移除不存在的玩家应该失败")
	}

	// 测试移除后数量
	if world.PlayerCount() != 0 {
		t.Errorf("移除玩家后数量期望为0，实际为%d", world.PlayerCount())
	}
}

// TestWorldConcurrent 测试World并发安全性
func TestWorldConcurrent(t *testing.T) {
	world := NewWorld()
	const numGoroutines = 50
	const numPlayersPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发添加玩家
	for i := 0; i < numGoroutines; i++ {
		go func(startID int) {
			defer wg.Done()
			for j := 0; j < numPlayersPerGoroutine; j++ {
				playerID := uint32(startID*numPlayersPerGoroutine + j + 1)
				entityID := world.generateEntityID()
				connID := uint64(playerID) + 1000
				player := NewPlayer(entityID, playerID, connID, nil, Vector3{})
				world.AddPlayer(player)
			}
		}(i)
	}

	wg.Wait()

	// 验证玩家数量
	expectedCount := numGoroutines * numPlayersPerGoroutine
	if world.PlayerCount() != expectedCount {
		t.Errorf("并发添加后玩家数量期望为%d，实际为%d", expectedCount, world.PlayerCount())
	}

	// 并发读取玩家
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numPlayersPerGoroutine; j++ {
				playerID := uint32(j + 1)
				_, exists := world.GetPlayer(playerID)
				_ = exists
			}
		}()
	}

	wg.Wait()

	// 并发移除玩家
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(startID int) {
			defer wg.Done()
			for j := 0; j < numPlayersPerGoroutine; j++ {
				playerID := uint32(startID*numPlayersPerGoroutine + j + 1)
				world.RemovePlayer(playerID)
			}
		}(i)
	}

	wg.Wait()

	// 验证所有玩家已移除
	if world.PlayerCount() != 0 {
		t.Errorf("并发移除后玩家数量期望为0，实际为%d", world.PlayerCount())
	}
}

// ============================= Mock 连接实现 =============================

// mockConnection 模拟Zinx连接，用于测试
type mockConnection struct {
	ziface.IConnection
	sentMessages []struct {
		msgID uint32
		data  []byte
	}
}

func (m *mockConnection) SendMsg(msgID uint32, data []byte) error {
	m.sentMessages = append(m.sentMessages, struct {
		msgID uint32
		data  []byte
	}{msgID, data})
	return nil
}

func (m *mockConnection) GetLastMessage() (uint32, []byte) {
	if len(m.sentMessages) == 0 {
		return 0, nil
	}
	last := m.sentMessages[len(m.sentMessages)-1]
	return last.msgID, last.data
}

// ============================= 登录流程测试 =============================

// TestWorldLogin 测试World登录流程
func TestWorldLogin(t *testing.T) {
	world := NewWorld()

	// 创建模拟连接
	mockConn := &mockConnection{}

	// 创建登录请求事件
	loginReq := &pb.LoginRequest{
		PlayerName: "TestPlayer",
	}

	// 注意：这里需要将Protobuf消息序列化，因为PacketEvent的Data字段是interface{}
	// 但在实际代码中，Router已经解析为具体类型，所以这里直接使用指针
	packetEvent := PacketEvent{
		ConnID:   1001,
		PacketID: uint32(pb.PacketID_LOGIN_REQUEST),
		Data:     loginReq,
		Conn:     mockConn,
	}

	// 处理登录请求
	world.handleLoginRequest(packetEvent)

	// 验证玩家已添加到世界
	if world.PlayerCount() != 1 {
		t.Errorf("登录后玩家数量期望为1，实际为%d", world.PlayerCount())
	}

	// 验证登录响应
	msgID, data := mockConn.GetLastMessage()
	if msgID != uint32(pb.PacketID_LOGIN_RESPONSE) {
		t.Errorf("响应消息ID期望为LOGIN_RESPONSE(%d)，实际为%d", pb.PacketID_LOGIN_RESPONSE, msgID)
	}

	// 解析响应
	loginResp := &pb.LoginResponse{}
	err := proto.Unmarshal(data, loginResp)
	if err != nil {
		t.Fatalf("解析登录响应失败: %v", err)
	}

	if loginResp.Status != 0 {
		t.Errorf("登录响应状态期望为0，实际为%d", loginResp.Status)
	}

	// 验证玩家ID
	playerID := loginResp.PlayerId
	player, exists := world.GetPlayer(playerID)
	if !exists {
		t.Error("登录后应该能找到玩家")
	}
	if player.Connection.ConnID != 1001 {
		t.Errorf("玩家连接ID期望为1001，实际为%d", player.Connection.ConnID)
	}
}

// TestWorldDisconnect 测试World断线处理
func TestWorldDisconnect(t *testing.T) {
	world := NewWorld()

	// 添加一个测试玩家
	entityID := world.generateEntityID()
	playerID := world.generatePlayerID()
	player := NewPlayer(entityID, playerID, 1001, nil, Vector3{})
	world.AddPlayer(player)

	// 验证玩家存在
	if world.PlayerCount() != 1 {
		t.Fatalf("添加玩家后数量期望为1，实际为%d", world.PlayerCount())
	}

	// 创建断线事件
	connEvent := ConnectionEvent{
		ConnID: 1001,
		Type:   DisconnectEvent,
	}

	// 处理断线事件
	world.handleConnection(connEvent)

	// 验证玩家已移除
	if world.PlayerCount() != 0 {
		t.Errorf("断线后玩家数量期望为0，实际为%d", world.PlayerCount())
	}
}

// TestWorldTick 测试World的tick循环
func TestWorldTick(t *testing.T) {
	world := NewWorld()

	// 启动World的Run循环（在独立goroutine中）
	stopCh := make(chan struct{})
	go func() {
		// 简化版Run循环，只执行一次tick
		world.tick()
		close(stopCh)
	}()

	// 等待tick完成
	select {
	case <-stopCh:
		// tick执行成功
	case <-time.After(100 * time.Millisecond):
		t.Error("tick执行超时")
	}
}

// ============================= 集成测试 =============================

// TestWorldLoginDisconnectIntegration 测试登录和断线集成流程
func TestWorldLoginDisconnectIntegration(t *testing.T) {
	world := NewWorld()

	// 模拟登录
	mockConn := &mockConnection{}
	loginReq := &pb.LoginRequest{PlayerName: "TestPlayer"}
	packetEvent := PacketEvent{
		ConnID:   2001,
		PacketID: uint32(pb.PacketID_LOGIN_REQUEST),
		Data:     loginReq,
		Conn:     mockConn,
	}
	world.handleLoginRequest(packetEvent)

	// 验证登录成功
	if world.PlayerCount() != 1 {
		t.Fatalf("登录后玩家数量期望为1，实际为%d", world.PlayerCount())
	}

	// 获取玩家ID
	_, data := mockConn.GetLastMessage()
	loginResp := &pb.LoginResponse{}
	proto.Unmarshal(data, loginResp)
	playerID := loginResp.PlayerId

	// 模拟断线
	connEvent := ConnectionEvent{
		ConnID: 2001,
		Type:   DisconnectEvent,
	}
	world.handleConnection(connEvent)

	// 验证玩家已移除
	if world.PlayerCount() != 0 {
		t.Errorf("断线后玩家数量期望为0，实际为%d", world.PlayerCount())
	}

	// 验证玩家无法再获取
	_, exists := world.GetPlayer(playerID)
	if exists {
		t.Error("断线后玩家不应该存在")
	}
}

// TestWorldRaceConditions 测试竞态条件（使用-race标志）
func TestWorldRaceConditions(t *testing.T) {
	world := NewWorld()
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(3)

	// goroutine 1: 添加玩家
	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			entityID := world.generateEntityID()
			playerID := world.generatePlayerID()
			player := NewPlayer(entityID, playerID, uint64(i+1), nil, Vector3{})
			world.AddPlayer(player)
		}
	}()

	// goroutine 2: 读取玩家
	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			world.PlayerCount()
			// 随机读取一些玩家ID
			playerID := uint32(i%100) + 1
			world.GetPlayer(playerID)
		}
	}()

	// goroutine 3: 移除玩家
	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			playerID := uint32(i%100) + 1
			world.RemovePlayer(playerID)
		}
	}()

	wg.Wait()
	// 如果没有发生数据竞争，测试通过
}

// TestWorldSendPacketToPlayer 测试SendPacketToPlayer方法
func TestWorldSendPacketToPlayer(t *testing.T) {
	world := NewWorld()

	// 创建模拟连接
	mockConn := &mockConnection{}

	// 添加一个测试玩家
	entityID := world.generateEntityID()
	playerID := world.generatePlayerID()
	player := NewPlayer(entityID, playerID, 1001, mockConn, Vector3{})
	world.AddPlayer(player)

	// 测试成功发送
	heartbeat := &pb.Heartbeat{}
	err := world.SendPacketToPlayer(playerID, uint32(pb.PacketID_HEARTBEAT), heartbeat)
	if err != nil {
		t.Errorf("SendPacketToPlayer应该成功: %v", err)
	}

	// 验证消息已发送
	msgID, _ := mockConn.GetLastMessage()
	if msgID != uint32(pb.PacketID_HEARTBEAT) {
		t.Errorf("消息ID期望为HEARTBEAT(%d)，实际为%d", pb.PacketID_HEARTBEAT, msgID)
	}
	// Heartbeat消息是空的，所以数据可能为空，这是正常的
	// 我们只验证消息ID正确即可

	// 测试玩家不存在的情况
	err = world.SendPacketToPlayer(999, uint32(pb.PacketID_HEARTBEAT), heartbeat)
	if err == nil {
		t.Error("发送给不存在的玩家应该返回错误")
	}

	// 测试玩家连接为nil的情况
	entityID2 := world.generateEntityID()
	playerID2 := world.generatePlayerID()
	player2 := NewPlayer(entityID2, playerID2, 1002, nil, Vector3{})
	world.AddPlayer(player2)
	err = world.SendPacketToPlayer(playerID2, uint32(pb.PacketID_HEARTBEAT), heartbeat)
	if err == nil {
		t.Error("玩家连接为nil时应该返回错误")
	}

	// 测试Protobuf序列化错误（传入nil消息）
	err = world.SendPacketToPlayer(playerID, uint32(pb.PacketID_HEARTBEAT), nil)
	if err == nil {
		t.Error("nil消息应该返回序列化错误")
	}
}

// TestWorldHandlePacketErrorPaths 测试handlePacket的错误路径
func TestWorldHandlePacketErrorPaths(t *testing.T) {
	world := NewWorld()

	// 测试未知数据包类型
	mockConn := &mockConnection{}
	packetEvent := PacketEvent{
		ConnID:   1001,
		PacketID: 9999, // 未知数据包ID
		Data:     nil,
		Conn:     mockConn,
	}

	// 应该处理未知数据包而不panic
	world.handlePacket(packetEvent)

	// 测试LoginRequest类型断言失败
	packetEvent2 := PacketEvent{
		ConnID:   1002,
		PacketID: uint32(pb.PacketID_LOGIN_REQUEST),
		Data:     "invalid data", // 错误的类型
		Conn:     mockConn,
	}

	// 应该处理类型断言失败而不panic
	world.handlePacket(packetEvent2)
}

// TestWorldHandleLoginRequestErrorPaths 测试handleLoginRequest的错误路径
func TestWorldHandleLoginRequestErrorPaths(t *testing.T) {
	world := NewWorld()
	mockConn := &mockConnection{}

	// 测试空玩家名
	loginReq := &pb.LoginRequest{
		PlayerName: "", // 空名字
	}
	packetEvent := PacketEvent{
		ConnID:   1001,
		PacketID: uint32(pb.PacketID_LOGIN_REQUEST),
		Data:     loginReq,
		Conn:     mockConn,
	}

	world.handleLoginRequest(packetEvent)
	// 应该处理空名字而不添加玩家
	if world.PlayerCount() != 0 {
		t.Error("空玩家名不应该添加玩家")
	}

	// 测试玩家添加失败的情况（模拟重复玩家ID）
	// 这需要更复杂的模拟，暂时跳过
}

// TestWorldRunStop 测试World的Run和Stop方法
func TestWorldRunStop(t *testing.T) {
	world := NewWorld()

	// 初始化Channel（避免nil通道）
	InitializeChannels(10, 5, 5)

	// 启动World的Run循环（在独立goroutine中）
	stopCh := make(chan struct{})
	go func() {
		// 运行一小段时间后停止
		time.Sleep(10 * time.Millisecond)
		world.Stop()
		close(stopCh)
	}()

	// 启动World
	go world.Run()

	// 等待停止完成
	select {
	case <-stopCh:
		// 成功停止
	case <-time.After(100 * time.Millisecond):
		t.Error("World停止超时")
	}
}

// TestWorldGetEntityManager 测试GetEntityManager方法
func TestWorldGetEntityManager(t *testing.T) {
	world := NewWorld()

	em := world.GetEntityManager()
	if em == nil {
		t.Error("GetEntityManager不应该返回nil")
	}

	// 验证EntityManager可用
	entity := NewEntity(1, EntityTypePlayer, Vector3{})
	em.AddEntity(entity)

	if em.EntityCount() != 1 {
		t.Errorf("EntityManager实体数量期望为1，实际为%d", em.EntityCount())
	}
}

// TestEntityMethods 测试Entity的GetPosition/SetPosition等方法
func TestEntityMethods(t *testing.T) {
	entity := NewEntity(1, EntityTypePlayer, Vector3{X: 10, Y: 20, Z: 30})

	// 测试GetPosition
	pos := entity.GetPosition()
	if pos.X != 10 || pos.Y != 20 || pos.Z != 30 {
		t.Errorf("GetPosition返回值不正确: %v", pos)
	}

	// 测试SetPosition
	newPos := Vector3{X: 40, Y: 50, Z: 60}
	entity.SetPosition(newPos)
	pos = entity.GetPosition()
	if pos.X != 40 || pos.Y != 50 || pos.Z != 60 {
		t.Errorf("SetPosition后GetPosition返回值不正确: %v", pos)
	}

	// 测试GetRotation/SetRotation
	if entity.GetRotation() != 0 {
		t.Errorf("初始Rotation期望为0，实际为%f", entity.GetRotation())
	}

	entity.SetRotation(1.5)
	if entity.GetRotation() != 1.5 {
		t.Errorf("SetRotation后GetRotation返回值不正确: %f", entity.GetRotation())
	}
}

// TestEntityManagerGetAllEntities 测试GetAllEntities方法
func TestEntityManagerGetAllEntities(t *testing.T) {
	em := NewEntityManager()

	// 添加多个实体
	entity1 := NewEntity(1, EntityTypePlayer, Vector3{})
	entity2 := NewEntity(2, EntityTypeItem, Vector3{})
	entity3 := NewEntity(3, EntityTypeMonster, Vector3{})

	em.AddEntity(entity1)
	em.AddEntity(entity2)
	em.AddEntity(entity3)

	// 获取所有实体
	allEntities := em.GetAllEntities()
	if len(allEntities) != 3 {
		t.Errorf("GetAllEntities返回实体数量期望为3，实际为%d", len(allEntities))
	}

	// 验证所有实体都存在
	_, exists1 := allEntities[1]
	_, exists2 := allEntities[2]
	_, exists3 := allEntities[3]
	if !exists1 || !exists2 || !exists3 {
		t.Error("GetAllEntities返回的映射缺少某些实体")
	}
}

// TestEntityManagerClear 测试Clear方法
func TestEntityManagerClear(t *testing.T) {
	em := NewEntityManager()

	// 添加实体
	entity := NewEntity(1, EntityTypePlayer, Vector3{})
	em.AddEntity(entity)

	if em.EntityCount() != 1 {
		t.Fatalf("添加实体后数量期望为1，实际为%d", em.EntityCount())
	}

	// 清空
	em.Clear()

	if em.EntityCount() != 0 {
		t.Errorf("Clear后实体数量期望为0，实际为%d", em.EntityCount())
	}
}
