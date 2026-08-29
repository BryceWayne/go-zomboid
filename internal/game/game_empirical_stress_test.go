package game

import (
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestEmpirical_ECSMovementAABBCollision blocks movement into solid tiles and permits floors
func TestEmpirical_ECSMovementAABBCollision(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)

	// Clear area (10, 10) to (20, 20) with Grass
	for y := 10; y <= 20; y++ {
		for x := 10; x <= 20; x++ {
			m.SetTile(x, y, world.TileGrass)
		}
	}

	// Place solid obstacles at (15, 10)=Wall, (15, 12)=Tree, (15, 14)=Fence, (15, 16)=Debris
	m.SetTile(15, 10, world.TileWall)
	m.SetTile(15, 12, world.TileTree)
	m.SetTile(15, 14, world.TileFence)
	m.SetTile(15, 16, world.TileDebris)

	sys := NewUpdateSystem(w, m)
	posMap := arkecs.NewMap3[ecs.Position, ecs.Velocity, ecs.Collider](w)

	testCases := []struct {
		name       string
		startX     float64
		startY     float64
		velX       float64
		velY       float64
		shouldMove bool
	}{
		// Moving right towards Wall at x=15 (pixel 1920). Start at x=1850 (box x+w=1914, clear), velX=12 (box x+w=1926, collides with wall 1920)
		{"BlockedByWall", 1850, 10.0 * world.TileSize, 12.0, 0, false},
		// Moving right towards Tree at x=15 (pixel 1920). Start at x=1850, velX=12 (collides with tree)
		{"BlockedByTree", 1850, 12.0 * world.TileSize, 12.0, 0, false},
		// Moving right towards Fence at x=15 (pixel 1920). Start at x=1850, velX=12 (collides with fence)
		{"BlockedByFence", 1850, 14.0 * world.TileSize, 12.0, 0, false},
		// Moving right towards Debris at x=15 (pixel 1920). Start at x=1850, velX=12 (collides with debris)
		{"BlockedByDebris", 1850, 16.0 * world.TileSize, 12.0, 0, false},
		// Moving right on grass at y=11 (pixel 1408). Start at x=1850, velX=12. Free grass passage.
		{"AllowedOnGrass", 1850, 11.0 * world.TileSize, 12.0, 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ent := posMap.NewEntity(
				&ecs.Position{X: tc.startX, Y: tc.startY},
				&ecs.Velocity{X: tc.velX, Y: tc.velY},
				&ecs.Collider{Width: 64, Height: 64},
			)

			sys.processMovement()

			pos := arkecs.NewMap1[ecs.Position](w).Get(ent)
			if tc.shouldMove {
				if pos.X == tc.startX {
					t.Errorf("FAIL %s: Entity failed to move on walkable floor", tc.name)
				}
			} else {
				if pos.X != tc.startX {
					t.Errorf("FAIL %s: Entity moved into solid obstacle! Start=%f, Got=%f", tc.name, tc.startX, pos.X)
				}
			}

			w.RemoveEntity(ent)
		})
	}
}

// TestEmpirical_FOVInGameUpdate verifies FOV updates as player moves
func TestEmpirical_FOVInGameUpdate(t *testing.T) {
	assets.Load()
	g := NewGame()

	// Initial player position
	var pPos *ecs.Position
	pq := arkecs.NewFilter2[ecs.Player, ecs.Position](g.world).Query()
	for pq.Next() {
		_, pPos = pq.Get()
	}
	if pPos == nil {
		t.Fatal("Player not found")
	}

	// Update game system once
	g.updateSys.Update(-1)

	px := int(pPos.X) / world.TileSize
	py := int(pPos.Y) / world.TileSize

	if !g.gameMap.Visible[py*g.gameMap.Width+px] {
		t.Errorf("FAIL: Player tile (%d,%d) is not marked visible after Update()", px, py)
	}
	if !g.gameMap.Explored[py*g.gameMap.Width+px] {
		t.Errorf("FAIL: Player tile (%d,%d) is not marked explored after Update()", px, py)
	}
}

// TestEmpirical_ZombieAndPlayerSpawnInvariants verifies all spawns across 20 Game resets
func TestEmpirical_ZombieAndPlayerSpawnInvariants(t *testing.T) {
	assets.Load()
	g := NewGame()

	for iter := 0; iter < 10; iter++ {
		g.Reset()

		// 1. Exactly 1 player
		var pPos *ecs.Position
		playerCount := 0
		pq := arkecs.NewFilter2[ecs.Player, ecs.Position](g.world).Query()
		for pq.Next() {
			playerCount++
			_, pPos = pq.Get()
		}
		if playerCount != 1 {
			t.Fatalf("FAIL: Iter %d: Expected 1 player, got %d", iter, playerCount)
		}

		pTileX := int(pPos.X) / world.TileSize
		pTileY := int(pPos.Y) / world.TileSize
		if g.gameMap.GetTile(pTileX, pTileY).IsSolid() {
			t.Fatalf("FAIL: Iter %d: Player spawned on solid tile %v", iter, g.gameMap.GetTile(pTileX, pTileY))
		}

		// 2. Zombies
		zCount := 0
		zq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](g.world).Query()
		for zq.Next() {
			zCount++
			_, zPos := zq.Get()

			zTileX := int(zPos.X) / world.TileSize
			zTileY := int(zPos.Y) / world.TileSize

			if g.gameMap.GetTile(zTileX, zTileY).IsSolid() {
				t.Fatalf("FAIL: Iter %d: Zombie spawned on solid tile %v at (%f,%f)", iter, g.gameMap.GetTile(zTileX, zTileY), zPos.X, zPos.Y)
			}

			dx := zPos.X - pPos.X
			dy := zPos.Y - pPos.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 1400.0 {
				t.Fatalf("FAIL: Iter %d: Zombie spawned too close to player: dist=%.2f < 1400", iter, dist)
			}
		}

		if zCount == 0 {
			t.Fatalf("FAIL: Iter %d: 0 zombies spawned in world", iter)
		}
	}
}
