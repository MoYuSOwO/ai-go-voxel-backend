package world

import (
	"fmt"
	"sync"
	"time"

	"github.com/aceld/zinx/ziface"
	"google.golang.org/protobuf/proto"
	pb "voxel-backend/proto"
)

// ============================= 世界容器定义 =============================

// World 世界容器
// 这是整个游戏服务器的核心容器，协调所有子系统的工作
// 采用NASA级安全军规：极度安全的并发设计，零魔数，显式优于隐式
type World struct {
	// players 在线玩家映射表
	// key: 玩家ID (uint32), value: 玩家实体指针 (*Player)
	// 注意：玩家ID是游戏内逻辑ID，不同于Zinx的连接ID
	players map[uint32]*Player

	// playersMu 保护players映射的读写锁
	// 使用独立的锁而不是嵌入在结构体中，因为players需要频繁的并发访问
	// 读锁（RLock）允许多个goroutine同时读取玩家列表，写锁（Lock）独占修改
	playersMu sync.RWMutex

	// entityManager 实体管理器
	// 负责管理游戏中所有实体（玩家、物品、怪物等）的生命周期
	// EntityManager内部有自己的锁，这里不需要额外加锁
	entityManager *EntityManager

	// tickInterval 游戏刻间隔
	// 游戏刻（Tick）是游戏逻辑更新的基本时间单位
	// 设置为50ms，即20TPS（Ticks Per Second），这是体素游戏的常见频率
	// 使用time.Duration类型，便于与time包的其他函数配合使用
	tickInterval time.Duration

	// stopCh 停止信号通道
	// 当需要停止World容器时，向此通道发送信号
	// 使用空结构体struct{}作为通道类型，因为它不占用内存
	stopCh chan struct{}

	// nextPlayerID 下一个玩家ID
	// 用于分配全局唯一的玩家ID，采用简单的自增方式
	// 必须使用原子操作或加锁保证并发安全，这里选择加锁因为操作不频繁
	nextPlayerID uint32
	nextIDMu     sync.Mutex

	// nextEntityID 下一个实体ID
	// 用于分配全局唯一的实体ID，实体ID使用uint64，范围更大
	// 同样需要加锁保证并发安全
	nextEntityID uint64
	nextEntityMu sync.Mutex
}

// ============================= 世界容器创建与初始化 =============================

// NewWorld 创建新的世界容器
// 这是World结构体的工厂函数，确保正确初始化所有字段
func NewWorld() *World {
	// 创建World实例
	w := &World{
		players: make(map[uint32]*Player),
		// playersMu 使用零值初始化即可
		entityManager: NewEntityManager(),
		tickInterval:  50 * time.Millisecond, // 50ms = 20TPS
		stopCh:        make(chan struct{}),
		nextPlayerID:  1, // 从1开始，0保留为无效ID
		nextEntityID:  1, // 从1开始，0保留为无效ID
		// nextIDMu 和 nextEntityMu 使用零值初始化
	}

	return w
}

// ============================= 游戏刻循环 =============================

// Run 启动世界容器的游戏刻循环
// 这是World的主循环，在一个独立的goroutine中运行
// 采用经典的for-select模式监听多个通道
func (w *World) Run() {
	// 创建定时器，每tickInterval触发一次
	ticker := time.NewTicker(w.tickInterval)
	// 确保函数退出时停止定时器，防止goroutine泄漏
	defer ticker.Stop()

	// 主循环
	for {
		select {
		case <-ticker.C:
			// 游戏刻触发，执行游戏逻辑更新
			w.tick()

		case packetEvent := <-GetPacketChannel():
			// 收到来自网络层的数据包事件
			// 调用handlePacket方法处理数据包
			w.handlePacket(packetEvent)

		case connEvent := <-GetConnectionChannel():
			// 收到连接事件（连接建立或断开）
			// 调用handleConnection方法处理连接事件
			w.handleConnection(connEvent)

		case <-w.stopCh:
			// 收到停止信号，退出循环
			// 注意：这里不关闭任何通道，由调用者负责清理
			return
		}
	}
}

// Stop 停止世界容器
// 向stopCh发送停止信号，使Run循环退出
func (w *World) Stop() {
	// 使用select防止stopCh已关闭导致panic
	select {
	case w.stopCh <- struct{}{}:
		// 成功发送停止信号
	case <-time.After(100 * time.Millisecond):
		// 发送超时，可能World已经停止
		// 这里记录日志但不需要处理，因为超时是安全的
	}
}

// tick 执行游戏刻逻辑
// 目前仅打印日志，为后续的实体移动和区块更新预留位置
// 在Go语言中，tick方法会在World的主goroutine中执行，因此不需要加锁
// 但后续如果要在tick中访问共享数据（如players），仍然需要加锁
func (w *World) tick() {
	// 目前仅打印日志，验证tick循环正常运行
	// 后续在这里添加实体移动、区块更新等游戏逻辑
	fmt.Println("Tick run")

	// 注意：这里打印日志会影响性能，正式版本中应移除或改为调试级别日志
	// 但为了开发阶段验证tick循环是否正常工作，保留此日志
}

// ============================= ID分配方法 =============================

// generatePlayerID 生成新的玩家ID
// 线程安全：使用互斥锁保护自增操作
func (w *World) generatePlayerID() uint32 {
	w.nextIDMu.Lock()
	defer w.nextIDMu.Unlock()

	// 获取当前ID并递增
	id := w.nextPlayerID
	w.nextPlayerID++

	// 注意：这里没有检查ID溢出，因为uint32最大值是42亿，足够使用
	// 如果服务器运行时间极长导致ID溢出，需要处理循环或扩展为uint64
	return id
}

// generateEntityID 生成新的实体ID
// 线程安全：使用互斥锁保护自增操作
func (w *World) generateEntityID() uint64 {
	w.nextEntityMu.Lock()
	defer w.nextEntityMu.Unlock()

	// 获取当前ID并递增
	id := w.nextEntityID
	w.nextEntityID++

	return id
}

// ============================= 玩家管理方法 =============================

// AddPlayer 添加玩家到世界
// 线程安全：使用写锁保护players映射的插入操作
func (w *World) AddPlayer(player *Player) bool {
	if player == nil {
		// 防御性编程：不允许添加nil玩家
		return false
	}

	playerID := player.Connection.PlayerID

	w.playersMu.Lock()
	defer w.playersMu.Unlock()

	// 检查玩家是否已存在
	if _, exists := w.players[playerID]; exists {
		// 玩家已存在，不重复添加
		return false
	}

	// 添加玩家到映射表
	w.players[playerID] = player

	// 同时将玩家实体添加到实体管理器
	// 注意：Player结构体嵌入了Entity，因此可以直接作为IEntity使用
	w.entityManager.AddEntity(player.Entity)

	return true
}

// RemovePlayer 从世界移除玩家
// 线程安全：使用写锁保护players映射的删除操作
func (w *World) RemovePlayer(playerID uint32) bool {
	w.playersMu.Lock()
	defer w.playersMu.Unlock()

	// 检查玩家是否存在
	player, exists := w.players[playerID]
	if !exists {
		return false
	}

	// 从players映射中删除
	delete(w.players, playerID)

	// 同时从实体管理器中移除实体
	// 注意：实体ID是uint64，需要从Player的Entity中获取
	w.entityManager.RemoveEntity(player.GetID())

	return true
}

// GetPlayer 获取玩家
// 线程安全：使用读锁保护players映射的查找操作
func (w *World) GetPlayer(playerID uint32) (*Player, bool) {
	w.playersMu.RLock()
	defer w.playersMu.RUnlock()

	player, exists := w.players[playerID]
	return player, exists
}

// GetPlayerByConnID 根据连接ID获取玩家
// 线程安全：使用读锁遍历players映射
// 注意：这个方法效率较低（O(n)），但连接ID查找不频繁，可以接受
func (w *World) GetPlayerByConnID(connID uint64) (*Player, bool) {
	w.playersMu.RLock()
	defer w.playersMu.RUnlock()

	// 遍历所有玩家，查找匹配的连接ID
	for _, player := range w.players {
		if player.Connection.ConnID == connID {
			return player, true
		}
	}

	return nil, false
}

// PlayerCount 返回在线玩家数量
// 线程安全：使用读锁保护映射的长度获取
func (w *World) PlayerCount() int {
	w.playersMu.RLock()
	defer w.playersMu.RUnlock()

	return len(w.players)
}

// ============================= 辅助方法 =============================

// GetEntityManager 获取实体管理器
// 注意：返回的是指针，调用者可以直接使用EntityManager的方法
// EntityManager内部有自己的锁，因此这里不需要加锁
func (w *World) GetEntityManager() *EntityManager {
	return w.entityManager
}

// SendPacketToPlayer 向指定玩家发送数据包
// 这是网络层发送消息的辅助方法，封装了Protobuf序列化和Zinx发送逻辑
func (w *World) SendPacketToPlayer(playerID uint32, packetID uint32, message proto.Message) error {
	// 首先获取玩家
	player, exists := w.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found: %d", playerID)
	}

	// 检查玩家连接组件是否存在
	if player.Connection == nil {
		return fmt.Errorf("player connection component is nil: %d", playerID)
	}

	// 获取Zinx连接引用
	conn := player.Connection.Conn
	if conn == nil {
		// 连接可能为nil（例如在测试环境中）
		return fmt.Errorf("player connection is nil: %d", playerID)
	}

	// 防御性编程：检查消息是否为nil
	// 在Go语言中，interface{}可以包含nil值，但proto.Marshal期望非nil消息
	if message == nil {
		return fmt.Errorf("protobuf message is nil")
	}

	// 将Protobuf消息序列化为字节
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf message: %v", err)
	}

	// 通过Zinx连接发送消息
	// 根据Zinx框架的IConnection接口，SendMsg方法接受消息ID和数据负载
	err = conn.SendMsg(packetID, data)
	if err != nil {
		return fmt.Errorf("failed to send message via Zinx: %v", err)
	}

	return nil
}

// BroadcastToPlayers 向多个玩家广播消息
func (w *World) BroadcastToPlayers(playerIDs []uint32, packetID uint32, message proto.Message) {
	// 暂时留空，后续实现
}

// ============================= 事件处理方法 =============================

// handlePacket 处理数据包事件
// 这是网络层与游戏逻辑层之间的桥梁，根据数据包类型分发到具体的处理器
// 注意：这个方法在World的主goroutine中执行，因此不需要加锁
func (w *World) handlePacket(event PacketEvent) {
	// 根据数据包类型分发处理
	switch event.PacketID {
	case uint32(pb.PacketID_LOGIN_REQUEST):
		// 登录请求数据包
		w.handleLoginRequest(event)
	// 后续在这里添加其他数据包的处理逻辑
	default:
		// 未知的数据包类型，记录日志但不断开连接
		fmt.Printf("Unknown packet ID: %d\n", event.PacketID)
	}
}

// handleLoginRequest 处理登录请求
// 这是玩家进入游戏世界的第一个步骤，完成身份验证和玩家实体创建
func (w *World) handleLoginRequest(event PacketEvent) {
	// 首先断言Data字段为*proto.LoginRequest类型
	// 在Go语言中，类型断言用于从interface{}中提取具体类型
	loginReq, ok := event.Data.(*pb.LoginRequest)
	if !ok {
		// 类型断言失败，说明数据包格式错误
		fmt.Printf("Invalid login request data type: %T\n", event.Data)
		// 可以考虑断开连接，但这里仅记录日志
		return
	}

	// 验证玩家名称合法性
	// 非空即可，实际游戏中可能需要更复杂的验证（如长度、字符限制等）
	playerName := loginReq.GetPlayerName()
	if playerName == "" {
		fmt.Printf("Empty player name from connID: %d\n", event.ConnID)
		// 可以发送错误响应，但这里简单返回
		return
	}

	// 生成全局唯一的玩家ID
	// 使用World内部的ID生成器，保证线程安全
	playerID := w.generatePlayerID()

	// 生成全局唯一的实体ID
	entityID := w.generateEntityID()

	// 创建玩家实体
	// 初始位置设置为世界原点 (0, 0, 0)，实际游戏中可能需要更合理的出生点
	initialPosition := Vector3{X: 0, Y: 0, Z: 0}
	player := NewPlayer(entityID, playerID, event.ConnID, event.Conn, initialPosition)

	// 将玩家添加到世界容器中
	success := w.AddPlayer(player)
	if !success {
		fmt.Printf("Failed to add player to world: playerID=%d\n", playerID)
		return
	}

	// 构造登录响应消息
	// 使用Protobuf生成的LoginResponse结构体
	loginResp := &pb.LoginResponse{
		Status:   0, // 0表示成功
		PlayerId: playerID,
	}

	// 发送登录响应给客户端
	// 注意：这里使用事件中的连接直接发送，因为玩家尚未正式加入世界
	// 也可以使用w.SendPacketToPlayer，但此时玩家尚未添加到映射表
	// 所以我们直接使用event.Conn发送
	data, err := proto.Marshal(loginResp)
	if err != nil {
		fmt.Printf("Failed to marshal login response: %v\n", err)
		return
	}

	// 发送响应包
	// 响应包的消息ID是LOGIN_RESPONSE
	err = event.Conn.SendMsg(uint32(pb.PacketID_LOGIN_RESPONSE), data)
	if err != nil {
		fmt.Printf("Failed to send login response: %v\n", err)
		// 发送失败，可能需要从世界中移除玩家
		w.RemovePlayer(playerID)
		return
	}

	// 登录成功，记录日志
	fmt.Printf("Player login successful: name=%s, playerID=%d, entityID=%d, connID=%d\n",
		playerName, playerID, entityID, event.ConnID)
}

// handleConnection 处理连接事件
func (w *World) handleConnection(event ConnectionEvent) {
	// 根据事件类型处理连接建立或断开
	switch event.Type {
	case ConnectEvent:
		// 连接建立事件
		// 目前不需要特殊处理，玩家登录通过LoginRequest数据包处理
		fmt.Printf("Connection established: connID=%d\n", event.ConnID)
	case DisconnectEvent:
		// 连接断开事件
		// 需要清理对应的玩家数据
		fmt.Printf("Connection disconnected: connID=%d\n", event.ConnID)
		// 调用处理玩家断线的方法
		w.handlePlayerDisconnect(event.ConnID)
	default:
		fmt.Printf("Unknown connection event type: %d\n", event.Type)
	}
}

// handlePlayerDisconnect 处理玩家断线
// 这是一个内部方法，由handleConnection调用
func (w *World) handlePlayerDisconnect(connID uint64) {
	// 根据连接ID查找玩家
	player, found := w.GetPlayerByConnID(connID)
	if !found {
		// 未找到玩家，可能是连接建立后尚未登录
		fmt.Printf("Player not found for connID: %d\n", connID)
		return
	}

	// 从世界中移除玩家
	playerID := player.Connection.PlayerID
	success := w.RemovePlayer(playerID)
	if success {
		fmt.Printf("Player removed after disconnect: playerID=%d, connID=%d\n", playerID, connID)
	} else {
		fmt.Printf("Failed to remove player after disconnect: playerID=%d, connID=%d\n", playerID, connID)
	}
}

// 确保ziface包被使用（避免编译器报"imported and not used"错误）
var _ ziface.IConnection = (ziface.IConnection)(nil)
