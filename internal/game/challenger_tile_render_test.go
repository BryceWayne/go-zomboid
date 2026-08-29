package game

import (
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestChallenger_All16ObstaclesPropsAnd6FloorsRenderNoPanic verifies that all 22 tile types
// (6 floor tiles, 10 legacy obstacle tiles, and 6 new prop tiles) render cleanly through DrawSystem without panicking.
func TestChallenger_All16ObstaclesPropsAnd6FloorsRenderNoPanic(t *testing.T) {
	assets.Load()

	w := arkecs.NewWorld()
	m := world.NewMap(30, 30)

	// Populate map with all 22 tile types in a grid
	allTileTypes := []world.TileType{
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
		world.TileTent,
		world.TileElevationBlock,
		world.TileRamp,
		world.TileStump,
		world.TileMushroom,
		world.TileSign,
		world.TileBench,
		world.TileChest,
		world.TileSculpture,
		world.TileBush,
		world.TileFlower,
		world.TileStone,
	}

	for idx, tt := range allTileTypes {
		x := (idx % 5) * 5 + 2
		y := (idx / 5) * 5 + 2
		m.SetTile(x, y, tt)
		// Mark visible and explored
		m.Visible[y*m.Width+x] = true
		m.Explored[y*m.Width+x] = true
	}

	// Create player entity
	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{
			Health:         100.0,
			Hunger:         100.0,
			Thirst:         100.0,
			FacingX:        1.0,
			FacingY:        0.0,
			WeaponEquipped: true,
			WeaponType:     "axe",
		},
		&ecs.Position{X: 15.0 * world.TileSize, Y: 15.0 * world.TileSize},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 16.0, Height: 16.0},
	)

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	timesOfDay := []float64{0.0, 6.0, 12.0, 18.0, 23.5}
	for _, tod := range timesOfDay {
		// Test fully explored & visible
		for i := range m.Visible {
			m.Visible[i] = true
			m.Explored[i] = true
		}
		drawSys.Draw(screen, tod, -1)

		// Test explored but not currently visible (fog of war / memory tint)
		for i := range m.Visible {
			m.Visible[i] = false
			m.Explored[i] = true
		}
		drawSys.Draw(screen, tod, -1)

		// Test unexplored & not visible
		for i := range m.Visible {
			m.Visible[i] = false
			m.Explored[i] = false
		}
		drawSys.Draw(screen, tod, -1)
	}
}

// TestChallenger_ProceduralMapContainsAll10LegacyAnd6Props verifies that 50 procedurally
// generated 100x100 maps reliably contain all 10 legacy tiles and all 6 new prop tile types.
func TestChallenger_ProceduralMapContainsAll10LegacyAnd6Props(t *testing.T) {
	legacyTiles := []world.TileType{
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

	newPropTypes := []world.TileType{
		world.TileBench,
		world.TileChest,
		world.TileSculpture,
		world.TileBush,
		world.TileFlower,
		world.TileStone,
	}

	for iter := 0; iter < 50; iter++ {
		m := world.NewMap(100, 100)
		counts := make(map[world.TileType]int)
		for _, tile := range m.Tiles {
			counts[tile]++
		}

		for _, lt := range legacyTiles {
			if counts[lt] == 0 {
				t.Fatalf("Map iter %d: Missing legacy tile %v (%s)", iter, lt, lt.String())
			}
		}

		for _, pt := range newPropTypes {
			if counts[pt] == 0 {
				t.Fatalf("Map iter %d: Missing new prop tile %v (%s)", iter, pt, pt.String())
			}
		}
	}
}

// TestChallenger_GameLoopMultiFrameExecution tests the complete lifecycle of NewGame(),
// running 120 ticks of Update() and Draw() without errors or panics.
func TestChallenger_GameLoopMultiFrameExecution(t *testing.T) {
	assets.Load()

	g := NewGame()
	if g == nil {
		t.Fatal("NewGame returned nil")
	}

	screen := ebiten.NewImage(1280, 720)

	for frame := 0; frame < 120; frame++ {
		if err := g.Update(); err != nil {
			t.Fatalf("Update() failed on frame %d: %v", frame, err)
		}
		g.Draw(screen)
	}
}
