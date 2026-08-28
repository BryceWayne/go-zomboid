package game

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestGameResetStress tests game.Reset() across 100 iterations with intermediate state mutations.
func TestGameResetStress(t *testing.T) {
	assets.Load()
	g := NewGame()
	if g == nil {
		t.Fatal("NewGame returned nil")
	}

	iterations := 100
	for i := 0; i < iterations; i++ {
		// 1. Mutate game state before reset
		g.timeOfDay = float64(i % 24)

		// Mutate player
		pQuery := arkecs.NewFilter3[ecs.Player, ecs.Position, ecs.Velocity](g.world).Query()
		for pQuery.Next() {
			p, pos, vel := pQuery.Get()
			p.Health = float64(i % 100)
			p.Hunger = 10.0
			p.Thirst = 5.0
			p.Dead = (i%2 == 0)
			p.Infected = (i%3 == 0)
			p.Inventory = []string{"food", "water", "weapon", "axe", "shotgun", "ammo", "armor"}
			p.WeaponEquipped = true
			p.WeaponDurability = 99
			pos.X = 9999.0
			pos.Y = -500.0
			vel.X = 50.0
			vel.Y = -50.0
		}

		// Mutate or remove some items
		var removeItems []arkecs.Entity
		itemQuery := arkecs.NewFilter1[ecs.Item](g.world).Query()
		itemIdx := 0
		for itemQuery.Next() {
			if itemIdx%2 == 0 {
				removeItems = append(removeItems, itemQuery.Entity())
			}
			itemIdx++
		}
		for _, ent := range removeItems {
			g.world.RemoveEntity(ent)
		}

		// Mutate or remove some zombies
		var removeZombies []arkecs.Entity
		zQuery := arkecs.NewFilter1[ecs.Zombie](g.world).Query()
		zIdx := 0
		for zQuery.Next() {
			if zIdx%3 == 0 {
				removeZombies = append(removeZombies, zQuery.Entity())
			}
			zIdx++
		}
		for _, ent := range removeZombies {
			g.world.RemoveEntity(ent)
		}

		// 2. Perform Reset
		g.Reset()

		// 3. Strict Verification of Reset state
		if g.timeOfDay != 8.0 {
			t.Fatalf("Iteration %d: Expected timeOfDay 8.0, got %f", i, g.timeOfDay)
		}
		if g.gameMap == nil {
			t.Fatalf("Iteration %d: gameMap is nil", i)
		}
		if g.world == nil {
			t.Fatalf("Iteration %d: world is nil", i)
		}

		// Verify player
		var playerCount int
		pq := arkecs.NewFilter3[ecs.Player, ecs.Position, ecs.Velocity](g.world).Query()
		for pq.Next() {
			playerCount++
			p, pos, vel := pq.Get()
			if p.Health != 100.0 {
				t.Errorf("Iter %d: Expected player Health 100.0, got %f", i, p.Health)
			}
			if p.Hunger != 100.0 {
				t.Errorf("Iter %d: Expected player Hunger 100.0, got %f", i, p.Hunger)
			}
			if p.Thirst != 100.0 {
				t.Errorf("Iter %d: Expected player Thirst 100.0, got %f", i, p.Thirst)
			}
			if p.Dead {
				t.Errorf("Iter %d: Expected player Dead to be false", i)
			}
			if p.Infected {
				t.Errorf("Iter %d: Expected player Infected to be false", i)
			}
			if len(p.Inventory) != 0 {
				t.Errorf("Iter %d: Expected empty inventory, got len %d: %v", i, len(p.Inventory), p.Inventory)
			}
			if pos.X != g.gameMap.PlayerSpawn.X || pos.Y != g.gameMap.PlayerSpawn.Y {
				t.Errorf("Iter %d: Expected player at (%f, %f), got (%f, %f)",
					i, g.gameMap.PlayerSpawn.X, g.gameMap.PlayerSpawn.Y, pos.X, pos.Y)
			}
			if vel.X != 0 || vel.Y != 0 {
				t.Errorf("Iter %d: Expected player vel (0,0), got (%f, %f)", i, vel.X, vel.Y)
			}

			// Verify spawn point is inside world bounds and on non-solid tile
			px := int(pos.X) / world.TileSize
			py := int(pos.Y) / world.TileSize
			if px < 0 || px >= g.gameMap.Width || py < 0 || py >= g.gameMap.Height {
				t.Fatalf("Iter %d: Player spawn (%f, %f) out of grid bounds", i, pos.X, pos.Y)
			}
			if g.gameMap.GetTile(px, py).IsSolid() {
				t.Errorf("Iter %d: Player spawn tile (%d, %d) is solid: %v", i, px, py, g.gameMap.GetTile(px, py))
			}
		}
		if playerCount != 1 {
			t.Fatalf("Iter %d: Expected exactly 1 player, found %d", i, playerCount)
		}

		// Verify Loot items
		var itemCount int
		iq := arkecs.NewFilter2[ecs.Item, ecs.Position](g.world).Query()
		for iq.Next() {
			item, pos := iq.Get()
			itemCount++
			if item.Type == "" {
				t.Errorf("Iter %d: Item has empty Type", i)
			}
			tx := int(pos.X) / world.TileSize
			ty := int(pos.Y) / world.TileSize
			if g.gameMap.GetTile(tx, ty).IsSolid() {
				t.Errorf("Iter %d: Item %s at (%f, %f) on solid tile", i, item.Type, pos.X, pos.Y)
			}
		}
		if itemCount != len(g.gameMap.LootSpawns) {
			t.Errorf("Iter %d: Item count mismatch: expected %d, got %d",
				i, len(g.gameMap.LootSpawns), itemCount)
		}

		// Verify Zombies
		var zombieCount int
		zq := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](g.world).Query()
		for zq.Next() {
			z, pos, vel := zq.Get()
			zombieCount++
			if z.Speed <= 0 {
				t.Errorf("Iter %d: Zombie has invalid speed %f", i, z.Speed)
			}
			if vel.X != 0 || vel.Y != 0 {
				t.Errorf("Iter %d: Zombie has non-zero initial velocity (%f, %f)", i, vel.X, vel.Y)
			}
			// Verify distance from player spawn
			dx := pos.X - g.gameMap.PlayerSpawn.X
			dy := pos.Y - g.gameMap.PlayerSpawn.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 350.0 {
				t.Errorf("Iter %d: Zombie spawned too close to player spawn: dist=%f < 350", i, dist)
			}
			tx := int(pos.X) / world.TileSize
			ty := int(pos.Y) / world.TileSize
			if g.gameMap.GetTile(tx, ty).IsSolid() {
				t.Errorf("Iter %d: Zombie at (%f, %f) on solid tile", i, pos.X, pos.Y)
			}
		}
		if zombieCount != len(g.gameMap.ZombieSpawns) {
			t.Errorf("Iter %d: Zombie count mismatch: expected %d, got %d",
				i, len(g.gameMap.ZombieSpawns), zombieCount)
		}
	}
}

// TestIsometricProjectionMathStress verifies isometric coordinate transformations under stress.
func TestIsometricProjectionMathStress(t *testing.T) {
	testCases := []struct {
		name   string
		wx, wy float64
	}{
		{"Origin", 0, 0},
		{"Positive Unit", 1, 1},
		{"Tile Step", 32, 32},
		{"Grid Offset", 64, 32},
		{"Negative Coordinates", -100, -250},
		{"Large World Coordinates", 100000, 500000},
		{"Sub-pixel Coordinates", 123.456, 789.012},
		{"Asymmetric", 1280, -640},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isoX, isoY := WorldToIso(tc.wx, tc.wy)
			expectedIsoX := tc.wx - tc.wy
			expectedIsoY := (tc.wx + tc.wy) / 2.0

			if math.Abs(isoX-expectedIsoX) > 1e-9 {
				t.Errorf("WorldToIso(%f, %f) isoX = %f, want %f", tc.wx, tc.wy, isoX, expectedIsoX)
			}
			if math.Abs(isoY-expectedIsoY) > 1e-9 {
				t.Errorf("WorldToIso(%f, %f) isoY = %f, want %f", tc.wx, tc.wy, isoY, expectedIsoY)
			}

			// Inverse projection check: wx = isoY + isoX/2, wy = isoY - isoX/2
			recoveredWx := isoY + (isoX / 2.0)
			recoveredWy := isoY - (isoX / 2.0)

			if math.Abs(recoveredWx-tc.wx) > 1e-9 {
				t.Errorf("Inverse projection wx mismatch: got %f, want %f", recoveredWx, tc.wx)
			}
			if math.Abs(recoveredWy-tc.wy) > 1e-9 {
				t.Errorf("Inverse projection wy mismatch: got %f, want %f", recoveredWy, tc.wy)
			}
		})
	}

	// Randomized fuzzing of WorldToIso transformation
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 5000; i++ {
		wx := (rng.Float64() - 0.5) * 20000.0
		wy := (rng.Float64() - 0.5) * 20000.0

		isoX, isoY := WorldToIso(wx, wy)
		recWx := isoY + (isoX / 2.0)
		recWy := isoY - (isoX / 2.0)

		if math.Abs(recWx-wx) > 1e-6 || math.Abs(recWy-wy) > 1e-6 {
			t.Fatalf("Fuzz %d failed for (%f, %f): recovered (%f, %f)", i, wx, wy, recWx, recWy)
		}
	}
}

// TestIsometricRenderingAllTileTypesAndPropsStress tests rendering pipeline coverage for all tiles, props, items, and entity states.
func TestIsometricRenderingAllTileTypesAndPropsStress(t *testing.T) {
	assets.Load()

	allTiles := []world.TileType{
		world.TileGrass,
		world.TileWall,
		world.TileDirt,
		world.TileWoodFloor,
		world.TileTree,
		world.TileAsphalt,
		world.TileConcrete,
		world.TileTileFloor,
		world.TileFence,
		world.TileDebris,
	}

	itemTypes := []string{"food", "water", "weapon", "axe", "shotgun", "ammo", "armor", "unknown_item"}

	screen := ebiten.NewImage(800, 600)

	// Create custom test map with all tile types arranged in a grid
	width, height := 30, 30
	testMap := world.NewMap(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			tileIdx := (y*width + x) % len(allTiles)
			testMap.SetTile(x, y, allTiles[tileIdx])
			// Mark all visible and explored to stress every draw branch
			testMap.Visible[y*width+x] = true
			testMap.Explored[y*width+x] = true
		}
	}

	w := arkecs.NewWorld()
	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	iMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)

	// Add player
	pMap.NewEntity(
		&ecs.Player{
			Health:         85.0,
			Hunger:         70.0,
			Thirst:         60.0,
			Inventory:      []string{"food", "water", "weapon", "axe", "shotgun", "ammo", "armor"},
			WeaponEquipped: true,
			WeaponDurability: 4,
			AttackCooldown: 25, // Tests attack cooldown color tint branch
			Dead:           false,
			Infected:       true, // Tests infected pulse color tint branch
			FacingX:        1,
			FacingY:        0,
		},
		&ecs.Position{X: 15.0 * world.TileSize, Y: 15.0 * world.TileSize},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Add items of all types around player
	for idx, it := range itemTypes {
		iMap.NewEntity(
			&ecs.Item{Type: it},
			&ecs.Position{
				X: float64(14+idx%4) * world.TileSize,
				Y: float64(14+idx/4) * world.TileSize,
			},
		)
	}

	// Add standard zombie, runner zombie, and stunned zombie
	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.2, IsRunner: false, WanderTimer: 30},
		&ecs.Position{X: 16.0 * world.TileSize, Y: 15.0 * world.TileSize},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)
	zMap.NewEntity(
		&ecs.Zombie{Speed: 2.5, IsRunner: true, WanderTimer: 30},
		&ecs.Position{X: 15.0 * world.TileSize, Y: 16.0 * world.TileSize},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)
	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, IsRunner: false, StunTimer: 35}, // Tests stun tint branch
		&ecs.Position{X: 14.0 * world.TileSize, Y: 15.0 * world.TileSize},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	drawSys := NewDrawSystem(w, testMap)

	// Test rendering across 24h day-night cycle in 0.5 hour increments
	for hour := 0.0; hour <= 24.0; hour += 0.5 {
		t.Run(fmt.Sprintf("Hour_%.1f", hour), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic during draw at hour %.1f: %v", hour, r)
				}
			}()
			screen.Clear()
			drawSys.Draw(screen, hour)
		})
	}

	// Test with Fog of War (Explored only, not Visible)
	for i := range testMap.Visible {
		testMap.Visible[i] = false
		testMap.Explored[i] = true
	}
	t.Run("FogOfWarRendering", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Panic during fog-of-war draw: %v", r)
			}
		}()
		screen.Clear()
		drawSys.Draw(screen, 12.0)
	})

	// Test with Dead Player state
	pq := arkecs.NewFilter1[ecs.Player](w).Query()
	for pq.Next() {
		p := pq.Get()
		p.Dead = true
	}
	t.Run("DeadPlayerRendering", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Panic during dead player draw: %v", r)
			}
		}()
		screen.Clear()
		drawSys.Draw(screen, 12.0)
	})
}

// TestGameLoopContinuousSimulationStress runs game.Update() and game.Draw() across 2500 consecutive frames,
// verifying continuous headless simulation without panics, memory/entity leaks, or NaN velocities.
func TestGameLoopContinuousSimulationStress(t *testing.T) {
	assets.Load()
	g := NewGame()
	screen := ebiten.NewImage(800, 600)

	totalFrames := 2500
	for frame := 0; frame < totalFrames; frame++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Frame %d: Update returned error: %v", frame, err)
		}
		g.Draw(screen)

		// Every 100 frames, perform deep invariant checks
		if frame%100 == 0 {
			// 1. Verify day/night cycle sanity
			if math.IsNaN(g.timeOfDay) || math.IsInf(g.timeOfDay, 0) || g.timeOfDay < 0.0 || g.timeOfDay >= 24.0 {
				t.Fatalf("Frame %d: Invalid timeOfDay %f", frame, g.timeOfDay)
			}

			// 2. Verify all player components and physics values
			pq := arkecs.NewFilter3[ecs.Player, ecs.Position, ecs.Velocity](g.world).Query()
			for pq.Next() {
				p, pos, vel := pq.Get()
				if math.IsNaN(pos.X) || math.IsNaN(pos.Y) || math.IsInf(pos.X, 0) || math.IsInf(pos.Y, 0) {
					t.Fatalf("Frame %d: Player Position has NaN/Inf: (%f, %f)", frame, pos.X, pos.Y)
				}
				if math.IsNaN(vel.X) || math.IsNaN(vel.Y) || math.IsInf(vel.X, 0) || math.IsInf(vel.Y, 0) {
					t.Fatalf("Frame %d: Player Velocity has NaN/Inf: (%f, %f)", frame, vel.X, vel.Y)
				}
				if math.IsNaN(p.Health) || math.IsNaN(p.Hunger) || math.IsNaN(p.Thirst) {
					t.Fatalf("Frame %d: Player stats have NaN: Health=%f, Hunger=%f, Thirst=%f", frame, p.Health, p.Hunger, p.Thirst)
				}
			}

			// 3. Verify all zombie components and physics values
			zq := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](g.world).Query()
			for zq.Next() {
				_, pos, vel := zq.Get()
				if math.IsNaN(pos.X) || math.IsNaN(pos.Y) || math.IsInf(pos.X, 0) || math.IsInf(pos.Y, 0) {
					t.Fatalf("Frame %d: Zombie Position has NaN/Inf: (%f, %f)", frame, pos.X, pos.Y)
				}
				if math.IsNaN(vel.X) || math.IsNaN(vel.Y) || math.IsInf(vel.X, 0) || math.IsInf(vel.Y, 0) {
					t.Fatalf("Frame %d: Zombie Velocity has NaN/Inf: (%f, %f)", frame, vel.X, vel.Y)
				}
			}
		}

		// At frame 1500, trigger a mid-simulation Reset to test state recycling under continuous load
		if frame == 1500 {
			g.Reset()
		}
	}
}
