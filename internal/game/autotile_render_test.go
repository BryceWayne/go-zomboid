package game

import (
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestAutotiling_DrawAll16WallAndFenceBitmasks verifies that all 16 connected wall
// and fence bitmasks render cleanly through DrawSystem without panicking or missing assets.
func TestAutotiling_DrawAll16WallAndFenceBitmasks(t *testing.T) {
	assets.Load()

	w := arkecs.NewWorld()
	m := world.NewMap(20, 20)

	// Clear to grass
	for i := range m.Tiles {
		m.Tiles[i] = world.TileGrass
		m.Visible[i] = true
		m.Explored[i] = true
	}

	// Layout all 16 wall configurations and 16 fence configurations
	for mask := 0; mask < 16; mask++ {
		// Wall at (1 + (mask%4)*4, 1 + (mask/4)*4)
		wx := 1 + (mask%4)*4
		wy := 1 + (mask/4)*4
		m.SetTile(wx, wy, world.TileWall)

		hasN := (mask & (1 << 0)) != 0
		hasE := (mask & (1 << 1)) != 0
		hasS := (mask & (1 << 2)) != 0
		hasW := (mask & (1 << 3)) != 0

		if hasN {
			m.SetTile(wx, wy-1, world.TileWall)
		}
		if hasE {
			m.SetTile(wx+1, wy, world.TileWall)
		}
		if hasS {
			m.SetTile(wx, wy+1, world.TileWall)
		}
		if hasW {
			m.SetTile(wx-1, wy, world.TileWall)
		}

		// Verify asset lookup
		wallImg := assets.GetWallAutotileImage(uint8(mask))
		if wallImg == nil {
			t.Fatalf("Wall autotile image nil for mask %d", mask)
		}
		fenceImg := assets.GetFenceAutotileImage(uint8(mask))
		if fenceImg == nil {
			t.Fatalf("Fence autotile image nil for mask %d", mask)
		}
	}

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{Health: 100, Hunger: 100, Thirst: 100},
		&ecs.Position{X: 10 * world.TileSize, Y: 10 * world.TileSize},
		&ecs.Velocity{},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 16, Height: 16},
	)

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	drawSys.Draw(screen, 12.0, -1)
}

// TestAutotiling_TerrainBlendingOverlayRendering tests multi-layered terrain blending
// (Dirt, Grass, Concrete, Asphalt, WoodFloor, TileFloor) in a dense mosaic pattern.
func TestAutotiling_TerrainBlendingOverlayRendering(t *testing.T) {
	assets.Load()

	w := arkecs.NewWorld()
	m := world.NewMap(30, 30)

	// Create a dense terrain patchwork
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			switch (x + y) % 6 {
			case 0:
				m.SetTile(x, y, world.TileDirt)
			case 1:
				m.SetTile(x, y, world.TileGrass)
			case 2:
				m.SetTile(x, y, world.TileConcrete)
			case 3:
				m.SetTile(x, y, world.TileAsphalt)
			case 4:
				m.SetTile(x, y, world.TileWoodFloor)
			case 5:
				m.SetTile(x, y, world.TileTileFloor)
			}
			m.Visible[y*30+x] = true
			m.Explored[y*30+x] = true
		}
	}

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{Health: 100, Hunger: 100, Thirst: 100},
		&ecs.Position{X: 15 * world.TileSize, Y: 15 * world.TileSize},
		&ecs.Velocity{},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 16, Height: 16},
	)

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	// Draw across day and night cycle
	times := []float64{0.0, 6.0, 12.0, 18.0, 22.0}
	for _, tod := range times {
		drawSys.Draw(screen, tod, -1)
	}
}

// TestAutotiling_ProceduralTownRenderingMultiFrame verifies that 10 full procedurally
// generated 100x100 towns render with full autotiling without panics.
func TestAutotiling_ProceduralTownRenderingMultiFrame(t *testing.T) {
	assets.Load()

	for iter := 0; iter < 10; iter++ {
		w := arkecs.NewWorld()
		m := world.NewMap(100, 100)

		// Set player FOV
		m.CalculateFOV(m.PlayerSpawn.X, m.PlayerSpawn.Y, 15)

		pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		pMap.NewEntity(
			&ecs.Player{Health: 100, Hunger: 100, Thirst: 100},
			&ecs.Position{X: m.PlayerSpawn.X, Y: m.PlayerSpawn.Y},
			&ecs.Velocity{},
			&ecs.Sprite{W: 64, H: 128},
			&ecs.Collider{Width: 16, Height: 16},
		)

		drawSys := NewDrawSystem(w, m)
		screen := ebiten.NewImage(1280, 720)

		drawSys.Draw(screen, 14.0, -1)
	}
}
