package world

import (
	"math"
	"sync"
)

// ============================= 实体类型定义 =============================

// EntityType 实体类型枚举
// 注意：使用uint8以节省内存，最多支持255种实体类型
type EntityType uint8

const (
	EntityTypePlayer  EntityType = iota + 1 // 玩家实体，值=1
	EntityTypeItem                          // 物品实体（未来扩展）
	EntityTypeMonster                       // 怪物实体（未来扩展）
	// 可以继续添加其他实体类型
)

// ComponentType 组件类型枚举
// 组件是实体功能的模块化单元，采用组合模式而非继承
type ComponentType uint8

const (
	HealthComponent     ComponentType = iota + 1 // 生命值组件
	ConnectionComponent                          // 网络连接组件
	InventoryComponent                           // 背包组件（未来扩展）
	MovementComponent                            // 移动组件（未来扩展）
	// 可以继续添加其他组件类型
)

// Vector3 三维坐标结构体
// 注意：这里使用float32以节省内存，与Protobuf的Vector3消息兼容
type Vector3 struct {
	X float32
	Y float32
	Z float32
}

// Entity 基础实体结构体
// 采用NASA级安全军规：显式优于隐式，零魔数，极度安全的并发
type Entity struct {
	ID       uint64     // 全局唯一ID（64位）
	Type     EntityType // 实体类型
	Position Vector3    // 三维坐标
	Rotation float32    // 朝向（弧度）

	// 组件映射（类型 → 组件实例）
	// 使用sync.RWMutex保护并发访问，因为多个goroutine可能同时读写组件
	mu         sync.RWMutex
	components map[ComponentType]interface{}
}

// NewEntity 创建新实体
// 注意：必须调用此函数创建实体，以确保组件映射正确初始化
func NewEntity(id uint64, entityType EntityType, position Vector3) *Entity {
	return &Entity{
		ID:         id,
		Type:       entityType,
		Position:   position,
		Rotation:   0.0,
		components: make(map[ComponentType]interface{}),
	}
}

// ============================= 组件操作方法 =============================

// AddComponent 添加组件到实体
// 注意：此操作是线程安全的，使用写锁保护
func (e *Entity) AddComponent(componentType ComponentType, component interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.components[componentType] = component
}

// RemoveComponent 从实体移除组件
func (e *Entity) RemoveComponent(componentType ComponentType) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.components, componentType)
}

// GetComponent 获取实体组件
// 注意：此操作是线程安全的，使用读锁保护
// 返回组件实例和是否存在标记
func (e *Entity) GetComponent(componentType ComponentType) (interface{}, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	component, exists := e.components[componentType]
	return component, exists
}

// HasComponent 检查实体是否拥有指定类型的组件
func (e *Entity) HasComponent(componentType ComponentType) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.components[componentType]
	return exists
}

// ============================= 玩家实体定义 =============================

// PlayerConnection 玩家连接组件
// 存储玩家的网络连接信息
type PlayerConnection struct {
	ConnID   uint64 // Zinx连接ID
	PlayerID uint32 // 游戏内玩家ID（不同于连接ID）
	// 可以添加更多连接相关字段，如最后心跳时间、IP地址等
}

// Player 玩家实体结构体
// 玩家实体 = 基础Entity + ConnectionComponent + HealthComponent
type Player struct {
	*Entity                      // 嵌入基础实体（组合模式）
	Connection *PlayerConnection // 网络连接组件
	Health     int32             // 生命值组件（简单实现，未来可扩展为HealthComponent）
	// 注意：Player结构体本身不存储位置等信息，这些信息在嵌入的Entity中
}

// NewPlayer 创建新玩家实体
// 参数：实体ID、玩家ID、连接ID、初始位置
func NewPlayer(entityID uint64, playerID uint32, connID uint64, position Vector3) *Player {
	// 创建基础实体
	entity := NewEntity(entityID, EntityTypePlayer, position)

	// 创建连接组件
	connComponent := &PlayerConnection{
		ConnID:   connID,
		PlayerID: playerID,
	}

	// 创建玩家实体
	player := &Player{
		Entity:     entity,
		Connection: connComponent,
		Health:     100, // 默认生命值100
	}

	// 将连接组件添加到实体的组件映射中（便于统一管理）
	entity.AddComponent(ConnectionComponent, connComponent)

	// 注意：生命值目前作为Player的字段，未来可以改为HealthComponent
	// entity.AddComponent(HealthComponent, healthComponent)

	return player
}

// ============================= 辅助函数 =============================

// DistanceTo 计算两个实体之间的距离
// 使用三维欧几里得距离公式
func (e *Entity) DistanceTo(other *Entity) float32 {
	dx := e.Position.X - other.Position.X
	dy := e.Position.Y - other.Position.Y
	dz := e.Position.Z - other.Position.Z

	// 注意：实际游戏中可能使用曼哈顿距离或切比雪夫距离以提高性能
	// 这里使用欧几里得距离作为通用实现
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// MoveTo 移动实体到指定位置
// 注意：此方法只更新位置，不处理碰撞检测等游戏逻辑
func (e *Entity) MoveTo(position Vector3) {
	e.Position = position
}

// LookAt 设置实体朝向
func (e *Entity) LookAt(rotation float32) {
	e.Rotation = rotation
}
