package world

import (
	"sync"
)

// ============================= 实体管理器接口定义 =============================

// IEntity 实体接口定义
// 这是所有游戏实体的抽象接口，采用接口隔离原则，只暴露必要的方法
// 在Go语言中，接口是隐式实现的，只要结构体实现了这些方法就自动满足接口
type IEntity interface {
	// GetID 返回实体的全局唯一ID
	// 注意：ID必须是全局唯一的，用于在EntityManager中快速查找实体
	GetID() uint64

	// GetType 返回实体类型（玩家、物品、怪物等）
	// 实体类型使用预定义的EntityType枚举，便于分类处理
	GetType() EntityType

	// GetPosition 返回实体的三维坐标
	// 坐标使用Vector3结构体，与Protobuf的Vector3消息兼容
	GetPosition() Vector3

	// SetPosition 设置实体的三维坐标
	// 注意：这个方法不是必须的，但为了方便实体管理，这里提供基础设置方法
	// 实际游戏中可能需要更复杂的移动逻辑，但基础设置仍然有用
	SetPosition(position Vector3)

	// GetRotation 返回实体的朝向（弧度）
	GetRotation() float32

	// SetRotation 设置实体的朝向
	SetRotation(rotation float32)

	// 注意：这里没有定义组件相关的方法，因为组件管理是Entity内部实现细节
	// 实体管理器不需要知道实体的组件结构，这符合接口隔离原则
}

// ============================= 实体管理器实现 =============================

// EntityManager 实体管理器
// 负责管理游戏中所有实体的生命周期，提供线程安全的实体增删改查
// 采用NASA级安全军规：极度安全的并发设计
type EntityManager struct {
	// allEntities 存储所有实体的映射表
	// key: 实体ID (uint64), value: 实体接口 (IEntity)
	// 使用map提供O(1)的查找性能，但map本身不是并发安全的
	allEntities map[uint64]IEntity

	// mu 保护allEntities的读写锁
	// 使用RWMutex而不是Mutex，因为读操作远多于写操作
	// 读锁（RLock）允许多个goroutine同时读取，写锁（Lock）独占访问
	mu sync.RWMutex
}

// NewEntityManager 创建新的实体管理器
// 注意：必须调用此函数初始化EntityManager，否则map为nil会导致panic
func NewEntityManager() *EntityManager {
	return &EntityManager{
		allEntities: make(map[uint64]IEntity),
		// mu 使用零值初始化即可，sync.RWMutex不需要额外初始化
	}
}

// AddEntity 添加实体到管理器
// 线程安全：使用写锁保护map的插入操作
func (em *EntityManager) AddEntity(entity IEntity) bool {
	if entity == nil {
		// 防御性编程：不允许添加nil实体
		return false
	}

	entityID := entity.GetID()

	em.mu.Lock()
	defer em.mu.Unlock()

	// 检查实体是否已存在
	if _, exists := em.allEntities[entityID]; exists {
		// 实体已存在，不重复添加
		return false
	}

	em.allEntities[entityID] = entity
	return true
}

// RemoveEntity 从管理器移除实体
// 线程安全：使用写锁保护map的删除操作
func (em *EntityManager) RemoveEntity(entityID uint64) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	// 检查实体是否存在
	if _, exists := em.allEntities[entityID]; !exists {
		return false
	}

	delete(em.allEntities, entityID)
	return true
}

// GetEntity 获取实体
// 线程安全：使用读锁保护map的查找操作
func (em *EntityManager) GetEntity(entityID uint64) (IEntity, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	entity, exists := em.allEntities[entityID]
	return entity, exists
}

// GetAllEntities 获取所有实体的副本
// 注意：返回的是实体ID到实体的映射副本，避免外部直接修改内部map
// 线程安全：使用读锁保护map的遍历操作
func (em *EntityManager) GetAllEntities() map[uint64]IEntity {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// 创建副本
	copyMap := make(map[uint64]IEntity, len(em.allEntities))
	for id, entity := range em.allEntities {
		copyMap[id] = entity
	}

	return copyMap
}

// GetEntitiesByType 获取指定类型的所有实体
// 线程安全：使用读锁保护map的遍历操作
func (em *EntityManager) GetEntitiesByType(entityType EntityType) []IEntity {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var entities []IEntity
	for _, entity := range em.allEntities {
		if entity.GetType() == entityType {
			entities = append(entities, entity)
		}
	}

	return entities
}

// EntityCount 返回实体数量
// 线程安全：使用读锁保护map的长度获取
func (em *EntityManager) EntityCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return len(em.allEntities)
}

// Clear 清空所有实体
// 线程安全：使用写锁保护map的清空操作
// 注意：这个方法主要用于测试，生产环境中谨慎使用
func (em *EntityManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()

	// 清空map
	// Go语言中，清空map的正确方法是重新分配一个新的map
	// 或者遍历删除所有key，这里选择重新分配以释放内存
	em.allEntities = make(map[uint64]IEntity)
}
