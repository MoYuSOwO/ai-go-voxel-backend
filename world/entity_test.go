package world

import (
	"testing"
)

// TestEntityCreation 测试实体创建
func TestEntityCreation(t *testing.T) {
	entity := NewEntity(1001, EntityTypePlayer, Vector3{X: 10, Y: 20, Z: 30})

	if entity.ID != 1001 {
		t.Errorf("实体ID期望为1001，实际为%d", entity.ID)
	}
	if entity.Type != EntityTypePlayer {
		t.Errorf("实体类型期望为EntityTypePlayer，实际为%d", entity.Type)
	}
	if entity.Position.X != 10 || entity.Position.Y != 20 || entity.Position.Z != 30 {
		t.Errorf("实体位置期望为(10,20,30)，实际为(%v,%v,%v)",
			entity.Position.X, entity.Position.Y, entity.Position.Z)
	}
	if entity.Rotation != 0.0 {
		t.Errorf("实体朝向期望为0.0，实际为%v", entity.Rotation)
	}
}

// TestEntityComponents 测试实体组件操作
func TestEntityComponents(t *testing.T) {
	entity := NewEntity(1002, EntityTypePlayer, Vector3{})

	// 测试添加组件
	entity.AddComponent(HealthComponent, 100)
	entity.AddComponent(ConnectionComponent, "connection_data")

	// 测试获取组件
	health, exists := entity.GetComponent(HealthComponent)
	if !exists {
		t.Error("HealthComponent应该存在")
	}
	if health != 100 {
		t.Errorf("HealthComponent期望为100，实际为%v", health)
	}

	// 测试HasComponent
	if !entity.HasComponent(HealthComponent) {
		t.Error("HasComponent(HealthComponent)应该返回true")
	}
	if !entity.HasComponent(ConnectionComponent) {
		t.Error("HasComponent(ConnectionComponent)应该返回true")
	}
	if entity.HasComponent(InventoryComponent) {
		t.Error("HasComponent(InventoryComponent)应该返回false")
	}

	// 测试移除组件
	entity.RemoveComponent(HealthComponent)
	if entity.HasComponent(HealthComponent) {
		t.Error("移除HealthComponent后HasComponent应该返回false")
	}

	// 测试并发安全（简单测试）
	// 在实际项目中应该进行更严格的并发测试
	done := make(chan bool)
	go func() {
		entity.AddComponent(MovementComponent, "movement")
		done <- true
	}()
	go func() {
		entity.HasComponent(MovementComponent)
		done <- true
	}()
	<-done
	<-done
}

// TestPlayerCreation 测试玩家实体创建
func TestPlayerCreation(t *testing.T) {
	player := NewPlayer(2001, 3001, 4001, nil, Vector3{X: 5, Y: 0, Z: 5})

	if player.Entity.ID != 2001 {
		t.Errorf("玩家实体ID期望为2001，实际为%d", player.Entity.ID)
	}
	if player.Entity.Type != EntityTypePlayer {
		t.Errorf("玩家实体类型期望为EntityTypePlayer，实际为%d", player.Entity.Type)
	}
	if player.Connection.PlayerID != 3001 {
		t.Errorf("玩家ID期望为3001，实际为%d", player.Connection.PlayerID)
	}
	if player.Connection.ConnID != 4001 {
		t.Errorf("连接ID期望为4001，实际为%d", player.Connection.ConnID)
	}
	if player.Health != 100 {
		t.Errorf("玩家生命值期望为100，实际为%d", player.Health)
	}

	// 测试组件是否已添加到基础实体
	connComp, exists := player.Entity.GetComponent(ConnectionComponent)
	if !exists {
		t.Error("ConnectionComponent应该已添加到基础实体")
	}
	if connComp != player.Connection {
		t.Error("获取的ConnectionComponent应该与player.Connection相同")
	}
}

// TestEntityMovement 测试实体移动和朝向
func TestEntityMovement(t *testing.T) {
	entity := NewEntity(1003, EntityTypePlayer, Vector3{X: 0, Y: 0, Z: 0})

	// 测试移动
	newPos := Vector3{X: 10, Y: 20, Z: 30}
	entity.MoveTo(newPos)

	if entity.Position.X != 10 || entity.Position.Y != 20 || entity.Position.Z != 30 {
		t.Errorf("移动后位置期望为(10,20,30)，实际为(%v,%v,%v)",
			entity.Position.X, entity.Position.Y, entity.Position.Z)
	}

	// 测试朝向
	entity.LookAt(1.57) // 约90度
	if entity.Rotation != 1.57 {
		t.Errorf("朝向期望为1.57，实际为%v", entity.Rotation)
	}
}

// TestDistanceTo 测试实体距离计算
func TestDistanceTo(t *testing.T) {
	entity1 := NewEntity(1004, EntityTypePlayer, Vector3{X: 0, Y: 0, Z: 0})
	entity2 := NewEntity(1005, EntityTypePlayer, Vector3{X: 3, Y: 4, Z: 0})

	distance := entity1.DistanceTo(entity2)
	// 距离应为5（勾股定理：3-4-5三角形）
	expected := float32(5.0)
	if distance != expected {
		// 浮点数比较允许微小误差
		if distance < expected-0.001 || distance > expected+0.001 {
			t.Errorf("距离期望为%v，实际为%v", expected, distance)
		}
	}

	// 测试零距离
	entity3 := NewEntity(1006, EntityTypePlayer, Vector3{X: 10, Y: 20, Z: 30})
	distance = entity3.DistanceTo(entity3)
	if distance != 0 {
		t.Errorf("自身距离期望为0，实际为%v", distance)
	}
}

// TestEntityTypeConstants 测试实体类型常量
func TestEntityTypeConstants(t *testing.T) {
	if EntityTypePlayer != 1 {
		t.Errorf("EntityTypePlayer期望为1，实际为%d", EntityTypePlayer)
	}
	if EntityTypeItem != 2 {
		t.Errorf("EntityTypeItem期望为2，实际为%d", EntityTypeItem)
	}
	if EntityTypeMonster != 3 {
		t.Errorf("EntityTypeMonster期望为3，实际为%d", EntityTypeMonster)
	}
}

// TestComponentTypeConstants 测试组件类型常量
func TestComponentTypeConstants(t *testing.T) {
	if HealthComponent != 1 {
		t.Errorf("HealthComponent期望为1，实际为%d", HealthComponent)
	}
	if ConnectionComponent != 2 {
		t.Errorf("ConnectionComponent期望为2，实际为%d", ConnectionComponent)
	}
	if InventoryComponent != 3 {
		t.Errorf("InventoryComponent期望为3，实际为%d", InventoryComponent)
	}
	if MovementComponent != 4 {
		t.Errorf("MovementComponent期望为4，实际为%d", MovementComponent)
	}
}
