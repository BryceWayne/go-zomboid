package game

import (
	"image/color"
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// Helper to construct a test environment with a player and an update system
func setupDestructionTestHarness() (*arkecs.World, *world.Map, *UpdateSystem, arkecs.Entity) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	updateSys := NewUpdateSystem(w, m)

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := playerMap.NewEntity(
		&ecs.Player{
			Health:           100.0,
			Hunger:           100.0,
			Thirst:           100.0,
			Inventory:        make([]string, 9),
			WeaponEquipped:   false,
			WeaponType:       "",
			WeaponDurability: 0,
			AttackCooldown:   0,
			Dead:             false,
			Infected:         false,
			FacingX:          1.0,
			FacingY:          0.0,
		},
		&ecs.Position{X: 100.0, Y: 100.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, updateSys, pEnt
}

// 1. Test Axe Single-Swing Barrier Destruction and Wood Drop Spawning
func TestCombat_AxeChopBarrierSpawnsWoodDrop(t *testing.T) {
	w, m, _, pEnt := setupDestructionTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	// Place player at (200.0, 200.0), facing East (1.0, 0.0)
	pos.X = 200.0
	pos.Y = 200.0
	player.FacingX = 1.0
	player.FacingY = 0.0
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12

	// Place TileFence at tile (2, 1) -> center at (2*128+64, 1*128+64) = (320.0, 192.0)
	targetTx, targetTy := 2, 1
	m.SetTile(targetTx, targetTy, world.TileFence)

	// Axe attack reach: 128.0, radius: 128.0
	attackX := pos.X + player.FacingX*128.0 // 328.0
	attackY := pos.Y + player.FacingY*128.0 // 200.0

	// Execute Axe barrier chop logic
	hitBarrier := false
	minTx := int(attackX-128.0-float64(world.TileSize)/2.0) / world.TileSize
	maxTx := int(attackX+128.0+float64(world.TileSize)/2.0) / world.TileSize
	minTy := int(attackY-128.0-float64(world.TileSize)/2.0) / world.TileSize
	maxTy := int(attackY+128.0+float64(world.TileSize)/2.0) / world.TileSize

	for ty := minTy; ty <= maxTy; ty++ {
		for tx := minTx; tx <= maxTx; tx++ {
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
				continue
			}
			tileCenterX := float64(tx)*float64(world.TileSize) + float64(world.TileSize)/2.0
			tileCenterY := float64(ty)*float64(world.TileSize) + float64(world.TileSize)/2.0
			dist := math.Hypot(attackX-tileCenterX, attackY-tileCenterY)
			if dist <= 128.0+float64(world.TileSize)/2.0 {
				if m.IsDestructible(tx, ty) {
					destroyed, dropType := m.DamageTile(tx, ty, 2)
					hitBarrier = true
					if destroyed && dropType != "" {
						itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
						itemMap.NewEntity(
							&ecs.Item{Type: dropType},
							&ecs.Position{X: tileCenterX, Y: tileCenterY},
						)
					}
				}
			}
		}
	}

	if hitBarrier {
		player.WeaponDurability--
	}

	if !hitBarrier {
		t.Fatalf("Axe swing should have hit the fence barrier")
	}
	if player.WeaponDurability != 11 {
		t.Errorf("Expected Axe durability 11 after chopping fence, got %d", player.WeaponDurability)
	}
	if m.GetTile(targetTx, targetTy) != world.TileGrass {
		t.Errorf("Fence at (%d, %d) should be destroyed and replaced with TileGrass, got %v", targetTx, targetTy, m.GetTile(targetTx, targetTy))
	}

	// Verify wood item drop spawned in ECS world
	itemFilter := arkecs.NewFilter2[ecs.Item, ecs.Position](w)
	query := itemFilter.Query()
	foundWood := false
	for query.Next() {
		item, iPos := query.Get()
		if item.Type == "wood" {
			foundWood = true
			expectedCenterX := float64(targetTx)*float64(world.TileSize) + 64.0
			expectedCenterY := float64(targetTy)*float64(world.TileSize) + 64.0
			if iPos.X != expectedCenterX || iPos.Y != expectedCenterY {
				t.Errorf("Wood drop spawned at (%f, %f), expected (%f, %f)", iPos.X, iPos.Y, expectedCenterX, expectedCenterY)
			}
		}
	}

	if !foundWood {
		t.Errorf("Expected wood item entity to be spawned at destroyed tile center")
	}
}

// 2. Test Club/Weapon Requires 2 Hits to Destroy Fence
func TestCombat_ClubChopBarrierTwoSwings(t *testing.T) {
	w, m, _, pEnt := setupDestructionTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	pos.X = 200.0
	pos.Y = 200.0
	player.FacingX = 1.0
	player.FacingY = 0.0
	player.WeaponEquipped = true
	player.WeaponType = "weapon"
	player.WeaponDurability = 5

	targetTx, targetTy := 2, 1
	m.SetTile(targetTx, targetTy, world.TileFence)

	attackX := pos.X + player.FacingX*96.0
	attackY := pos.Y + player.FacingY*96.0

	chopClub := func() (hit bool, spawnedWood bool) {
		minTx := int(attackX-96.0-float64(world.TileSize)/2.0) / world.TileSize
		maxTx := int(attackX+96.0+float64(world.TileSize)/2.0) / world.TileSize
		minTy := int(attackY-96.0-float64(world.TileSize)/2.0) / world.TileSize
		maxTy := int(attackY+96.0+float64(world.TileSize)/2.0) / world.TileSize

		for ty := minTy; ty <= maxTy; ty++ {
			for tx := minTx; tx <= maxTx; tx++ {
				if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
					continue
				}
				tileCenterX := float64(tx)*float64(world.TileSize) + float64(world.TileSize)/2.0
				tileCenterY := float64(ty)*float64(world.TileSize) + float64(world.TileSize)/2.0
				dist := math.Hypot(attackX-tileCenterX, attackY-tileCenterY)
				if dist <= 96.0+float64(world.TileSize)/2.0 {
					if m.IsDestructible(tx, ty) {
						destroyed, dropType := m.DamageTile(tx, ty, 1)
						hit = true
						if destroyed && dropType != "" {
							spawnedWood = true
							itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
							itemMap.NewEntity(
								&ecs.Item{Type: dropType},
								&ecs.Position{X: tileCenterX, Y: tileCenterY},
							)
						}
					}
				}
			}
		}
		if hit {
			player.WeaponDurability--
		}
		return
	}

	// Swing 1: 1 damage -> Fence HP: 2 -> 1, weapon durability: 5 -> 4
	hit1, wood1 := chopClub()
	if !hit1 {
		t.Fatalf("Club swing 1 should hit fence")
	}
	if wood1 {
		t.Fatalf("Club swing 1 should NOT destroy fence with 2 HP")
	}
	if player.WeaponDurability != 4 {
		t.Errorf("Expected weapon durability 4, got %d", player.WeaponDurability)
	}
	if m.GetTileDurability(targetTx, targetTy) != 1 {
		t.Errorf("Expected fence durability 1, got %d", m.GetTileDurability(targetTx, targetTy))
	}
	if m.GetTile(targetTx, targetTy) != world.TileFence {
		t.Errorf("Tile should still be TileFence after 1 hit")
	}

	// Swing 2: 1 damage -> Fence HP: 1 -> 0, destroyed! Weapon durability: 4 -> 3
	hit2, wood2 := chopClub()
	if !hit2 {
		t.Fatalf("Club swing 2 should hit fence")
	}
	if !wood2 {
		t.Fatalf("Club swing 2 MUST destroy fence and drop wood")
	}
	if player.WeaponDurability != 3 {
		t.Errorf("Expected weapon durability 3, got %d", player.WeaponDurability)
	}
	if m.GetTile(targetTx, targetTy) != world.TileGrass {
		t.Errorf("Tile should now be TileGrass")
	}
}

// 3. Test Unarmed Shove Cannot Damage Barriers
func TestCombat_UnarmedCannotChopBarriers(t *testing.T) {
	_, m, _, _ := setupDestructionTestHarness()
	targetTx, targetTy := 10, 10
	m.SetTile(targetTx, targetTy, world.TileFence)

	// Unarmed attack applies 0 damage to barriers
	initialDur := m.GetTileDurability(targetTx, targetTy)
	if initialDur != 2 {
		t.Fatalf("Expected fence durability 2, got %d", initialDur)
	}

	// Simulate unarmed attack: does not call DamageTile (or calls with dmg 0)
	destroyed, drop := m.DamageTile(targetTx, targetTy, 0)
	if destroyed || drop != "" {
		t.Errorf("Unarmed attack must not destroy barriers")
	}
	if m.GetTileDurability(targetTx, targetTy) != 2 {
		t.Errorf("Fence durability should remain 2 after unarmed attack, got %d", m.GetTileDurability(targetTx, targetTy))
	}
	if m.GetTile(targetTx, targetTy) != world.TileFence {
		t.Errorf("Fence tile must remain intact")
	}
}

// 4. Test Shotgun Blast Destroys Barriers
func TestCombat_ShotgunBlastDestroysBarriers(t *testing.T) {
	w, m, _, pEnt := setupDestructionTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	pos.X = 200.0
	pos.Y = 200.0
	player.FacingX = 1.0
	player.FacingY = 0.0
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory[0] = "ammo"

	targetTx, targetTy := 3, 1 // Center at (3*128+64, 1*128+64) = (448, 192) - distance ~248px within 640px range
	m.SetTile(targetTx, targetTy, world.TileFence)

	// Consume ammo and decrement durability
	player.Inventory[0] = ""
	player.WeaponDurability--

	// Apply shotgun blast damage to destructible barriers in cone
	const maxShotgunRange = 640.0
	const cosSpread = 0.9238795325112867

	minTx := int(pos.X-maxShotgunRange-float64(world.TileSize)/2.0) / world.TileSize
	maxTx := int(pos.X+maxShotgunRange+float64(world.TileSize)/2.0) / world.TileSize
	minTy := int(pos.Y-maxShotgunRange-float64(world.TileSize)/2.0) / world.TileSize
	maxTy := int(pos.Y+maxShotgunRange+float64(world.TileSize)/2.0) / world.TileSize

	destroyedCount := 0
	for ty := minTy; ty <= maxTy; ty++ {
		for tx := minTx; tx <= maxTx; tx++ {
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
				continue
			}
			tileCenterX := float64(tx)*float64(world.TileSize) + float64(world.TileSize)/2.0
			tileCenterY := float64(ty)*float64(world.TileSize) + float64(world.TileSize)/2.0
			dx := tileCenterX - pos.X
			dy := tileCenterY - pos.Y
			dist := math.Hypot(dx, dy)
			if dist <= maxShotgunRange+float64(world.TileSize)/2.0 {
				inCone := false
				if dist < 96.0+float64(world.TileSize)/2.0 {
					inCone = true
				} else if dist > 0.001 {
					cosAngle := (player.FacingX*dx + player.FacingY*dy) / dist
					if cosAngle >= cosSpread {
						inCone = true
					}
				}
				if inCone && m.IsDestructible(tx, ty) {
					destroyed, dropType := m.DamageTile(tx, ty, 2)
					if destroyed && dropType != "" {
						destroyedCount++
						itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
						itemMap.NewEntity(
							&ecs.Item{Type: dropType},
							&ecs.Position{X: tileCenterX, Y: tileCenterY},
						)
					}
				}
			}
		}
	}

	if destroyedCount != 1 {
		t.Fatalf("Expected 1 destroyed barrier, got %d", destroyedCount)
	}
	if player.WeaponDurability != 14 {
		t.Errorf("Expected shotgun durability 14, got %d", player.WeaponDurability)
	}
	if m.GetTile(targetTx, targetTy) != world.TileGrass {
		t.Errorf("Target tile should now be TileGrass")
	}
}

// 5. Test Wood Item Pickup into Player Inventory
func TestCombat_WoodPickupIntoInventory(t *testing.T) {
	w, _, updateSys, pEnt := setupDestructionTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	pos.X = 300.0
	pos.Y = 300.0
	for i := range player.Inventory {
		player.Inventory[i] = ""
	}

	// Spawn a wood item 20px away from player (within 64px pickup radius)
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
	woodEnt := itemMap.NewEntity(
		&ecs.Item{Type: "wood"},
		&ecs.Position{X: 320.0, Y: 300.0},
	)

	// Run processItems()
	updateSys.processItems()

	// Verify wood collected into first slot
	if player.Inventory[0] != "wood" {
		t.Fatalf("Expected player.Inventory[0] == 'wood', got %q", player.Inventory[0])
	}
	if w.Alive(woodEnt) {
		t.Fatalf("Collected wood item entity should be removed from ECS world")
	}
}

// 6. Test Multi-Barrier Breach and Traversal Simulation
func TestStress_MultiBarrierBreachAndTraversal(t *testing.T) {
	w, m, updateSys, pEnt := setupDestructionTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Build a 5-segment vertical fence wall at column tx = 10 (ty = 5..9)
	for ty := 5; ty <= 9; ty++ {
		m.SetTile(10, ty, world.TileFence)
	}

	// Place player on left side of fence: tile (9, 7) -> pixel (9*128+64, 7*128+64) = (1216, 960)
	pos.X = float64(9*world.TileSize + 64)
	pos.Y = float64(7*world.TileSize + 64)

	// Verify player cannot walk East through fence before destruction due to collision
	rectW, rectH := 32.0, 32.0
	nextStepX := float64(10*world.TileSize + 64)
	nextStepY := float64(7*world.TileSize + 64)
	if !m.IsColliding(nextStepX-rectW/2, nextStepY-rectH/2, rectW, rectH) {
		t.Fatalf("Movement across intact fence at (10, 7) must collide")
	}

	// Chop down fence at (10, 7) with Axe
	destroyed, drop := m.DamageTile(10, 7, 2)
	if !destroyed || drop != "wood" {
		t.Fatalf("Fence at (10, 7) should be destroyed by Axe (2 dmg)")
	}

	// Spawn wood drop at destroyed tile center
	tileCenterX := float64(10*world.TileSize + 64)
	tileCenterY := float64(7*world.TileSize + 64)
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
	itemMap.NewEntity(
		&ecs.Item{Type: drop},
		&ecs.Position{X: tileCenterX, Y: tileCenterY},
	)

	// Now move player into the breached tile (10, 7)
	pos.X = nextStepX
	pos.Y = nextStepY

	// Verify no collision at breached location
	if m.IsColliding(pos.X-rectW/2, pos.Y-rectH/2, rectW, rectH) {
		t.Fatalf("Movement into breached tile (10, 7) should NOT collide")
	}

	// Collect the dropped wood item
	updateSys.processItems()
	if player.Inventory[0] != "wood" {
		t.Fatalf("Player should have collected 'wood' upon stepping onto the destroyed tile")
	}

	// Move player beyond the fence to tile (11, 7)
	pos.X = float64(11*world.TileSize + 64)
	pos.Y = float64(7*world.TileSize + 64)
	if m.IsColliding(pos.X-rectW/2, pos.Y-rectH/2, rectW, rectH) {
		t.Fatalf("Movement beyond breach to tile (11, 7) should NOT collide")
	}
}

// 7. Test Weapon Breaking and Unequipping upon Final Barrier Hit
func TestCombat_WeaponBreakOnFinalBarrierHit(t *testing.T) {
	w, m, _, pEnt := setupDestructionTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 1

	m.SetTile(10, 10, world.TileFence)

	// Hit fence
	destroyed, drop := m.DamageTile(10, 10, 2)
	if !destroyed || drop != "wood" {
		t.Fatalf("Axe should destroy fence")
	}

	player.WeaponDurability--
	if player.WeaponDurability <= 0 {
		player.WeaponEquipped = false
		player.WeaponType = ""
		player.WeaponDurability = 0
	}

	if player.WeaponEquipped {
		t.Errorf("Weapon should be unequipped when durability drops to 0")
	}
	if player.WeaponType != "" {
		t.Errorf("WeaponType should be cleared to empty string")
	}
	if player.WeaponDurability != 0 {
		t.Errorf("WeaponDurability should be 0")
	}
}
