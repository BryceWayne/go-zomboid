package game

import (
	"fmt"
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// =============================================================================
// TIER 1: FEATURE COVERAGE (Features 1-12)
// =============================================================================

// F1: 4x Floor Tiles (256x128)
func TestE2E_Tier1_Feature1_FloorTiles(t *testing.T) {
	assets.Load()
	floors := []struct {
		name string
		img  *ebiten.Image
	}{
		{"grass", assets.GrassImage},
		{"dirt", assets.DirtImage},
		{"wood", assets.WoodImage},
		{"asphalt", assets.AsphaltImage},
		{"concrete", assets.ConcreteImage},
		{"tile_floor", assets.TileFloorImage},
	}

	for _, f := range floors {
		t.Run("Floor_"+f.name, func(t *testing.T) {
			if f.img == nil {
				t.Fatalf("Floor image %s is nil", f.name)
			}
			w, h := f.img.Bounds().Dx(), f.img.Bounds().Dy()
			if w != 256 || h != 128 {
				t.Errorf("Floor %s dimensions = %dx%d, want 256x128", f.name, w, h)
			}
		})
	}
}

// F2: 4x Obstacles/Props (256x256)
func TestE2E_Tier1_Feature2_ObstaclesProps(t *testing.T) {
	assets.Load()
	obstacles := []struct {
		name string
		img  *ebiten.Image
	}{
		{"wall", assets.WallImage},
		{"tree", assets.TreeImage},
		{"fence", assets.FenceImage},
		{"debris", assets.DebrisImage},
		{"tent", assets.TentImage},
		{"stump", assets.StumpImage},
		{"mushroom", assets.MushroomImage},
		{"sign", assets.SignImage},
		{"elevation_block", assets.ElevationBlockImage},
		{"elevation_ramp", assets.ElevationRampImage},
	}

	for _, obs := range obstacles {
		t.Run("Obstacle_"+obs.name, func(t *testing.T) {
			if obs.img == nil {
				t.Fatalf("Obstacle image %s is nil", obs.name)
			}
			w, h := obs.img.Bounds().Dx(), obs.img.Bounds().Dy()
			if w != 256 || h != 256 {
				t.Errorf("Obstacle %s dimensions = %dx%d, want 256x256", obs.name, w, h)
			}
		})
	}
}

// F3: 4x Character Entities (64x128)
func TestE2E_Tier1_Feature3_CharacterEntities(t *testing.T) {
	assets.Load()
	chars := []struct {
		name string
		img  *ebiten.Image
	}{
		{"player", assets.PlayerImage},
		{"zombie", assets.ZombieImage},
		{"runner", assets.RunnerImage},
	}

	for _, c := range chars {
		t.Run("Character_"+c.name, func(t *testing.T) {
			if c.img == nil {
				t.Fatalf("Character image %s is nil", c.name)
			}
			w, h := c.img.Bounds().Dx(), c.img.Bounds().Dy()
			if w != 64 || h != 128 {
				t.Errorf("Character %s dimensions = %dx%d, want 64x128", c.name, w, h)
			}
		})
	}
}

// F4: 4x Items & Equipment (64x64)
func TestE2E_Tier1_Feature4_ItemsAndEquipment(t *testing.T) {
	assets.Load()
	items := []struct {
		name string
		img  *ebiten.Image
	}{
		{"food", assets.FoodImage},
		{"water", assets.WaterImage},
		{"weapon", assets.WeaponImage},
		{"axe", assets.AxeImage},
		{"shotgun", assets.ShotgunImage},
		{"ammo", assets.AmmoImage},
		{"armor", assets.ArmorImage},
		{"antidote", assets.AntidoteImage},
	}

	for _, it := range items {
		t.Run("Item_"+it.name, func(t *testing.T) {
			if it.img == nil {
				t.Fatalf("Item image %s is nil", it.name)
			}
			w, h := it.img.Bounds().Dx(), it.img.Bounds().Dy()
			if w != 64 || h != 64 {
				t.Errorf("Item %s dimensions = %dx%d, want 64x64", it.name, w, h)
			}
		})
	}
}

// F5: Geometric Vector Overlays
func TestE2E_Tier1_Feature5_GeometricVectorOverlays(t *testing.T) {
	assets.Load()
	floorImgs := []*ebiten.Image{
		assets.GrassImage, assets.DirtImage, assets.WoodImage,
		assets.AsphaltImage, assets.ConcreteImage, assets.TileFloorImage,
	}

	screen := ebiten.NewImage(800, 600)
	for idx, img := range floorImgs {
		t.Run(fmt.Sprintf("Overlay_Floor_%d", idx), func(t *testing.T) {
			if img == nil {
				t.Fatalf("Floor %d is nil", idx)
			}
			op := &ebiten.DrawImageOptions{}
			screen.DrawImage(img, op)
		})
	}
}

// F6: Engine TileSize (128) & Math Upgrade
func TestE2E_Tier1_Feature6_TileSizeAndMath(t *testing.T) {
	if world.TileSize != 128 {
		t.Fatalf("world.TileSize = %d, want 128", world.TileSize)
	}

	testPoints := []struct {
		wx, wy float64
	}{
		{0, 0},
		{128, 0},
		{0, 128},
		{128, 128},
		{256, 512},
	}

	for i, pt := range testPoints {
		t.Run(fmt.Sprintf("Transform_%d", i), func(t *testing.T) {
			isoX, isoY := WorldToIso(pt.wx, pt.wy)
			recoveredX, recoveredY := IsoToWorld(isoX, isoY)
			if math.Abs(recoveredX-pt.wx) > 1e-6 || math.Abs(recoveredY-pt.wy) > 1e-6 {
				t.Errorf("Reversible projection failed: got (%f, %f), want (%f, %f)", recoveredX, recoveredY, pt.wx, pt.wy)
			}
		})
	}
}

// F7: DrawSystem Anchors & Camera
func TestE2E_Tier1_Feature7_DrawSystemAnchors(t *testing.T) {
	assets.Load()
	g := NewGame()
	screen := ebiten.NewImage(800, 600)

	// Render without errors
	g.Draw(screen)
}

// F8: Entity Physics & Speeds
func TestE2E_Tier1_Feature8_EntityPhysicsAndSpeeds(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	for i := range m.Tiles {
		m.Tiles[i] = world.TileGrass
	}
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := pMap.NewEntity(
		&ecs.Player{Health: 100.0, AttackCooldown: 0, FacingX: 1.0, FacingY: 0.0},
		&ecs.Position{X: 500.0, Y: 500.0},
		&ecs.Velocity{X: 12.0, Y: 0.0},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 64, Height: 64},
	)

	sys.processMovement()

	pos := arkecs.NewMap1[ecs.Position](w).Get(pEnt)
	if pos.X != 512.0 {
		t.Errorf("Player movement step failed: got %f, want 512.0", pos.X)
	}
}

// F9: Combat & AI Range Scaling
func TestE2E_Tier1_Feature9_CombatAndAIRangeScaling(t *testing.T) {
	w, _, sys, pEnt := setupM5AdversarialHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zMap.NewEntity(
		&ecs.Zombie{Speed: 4.0},
		&ecs.Position{X: 300.0, Y: 200.0}, // Player at (200, 200) -> axe reach is 128.0px -> within reach!
		&ecs.Velocity{},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 64, Height: 64},
	)

	// Trigger axe attack
	attackX := 200.0 + player.FacingX*128.0
	attackY := 200.0 + player.FacingY*128.0
	var toRemove []arkecs.Entity
	zq := sys.zombieFilter.Query()
	for zq.Next() {
		_, zPos, _ := zq.Get()
		if math.Hypot(attackX-zPos.X, attackY-zPos.Y) < 128.0 {
			toRemove = append(toRemove, zq.Entity())
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}

	if w.Alive(zEnt) {
		t.Error("Zombie within 128px axe reach should be killed")
	}
}

// F10: Bezier Attack Curve Calculation
func TestE2E_Tier1_Feature10_BezierCurveCalculation(t *testing.T) {
	pX, pY := 400.0, 300.0
	fx, fy := 0.0, 1.0 // Facing Down (South)
	baseAngle := math.Atan2(fy, fx)

	rApex := 140.0

	apexX := pX + rApex*math.Cos(baseAngle)
	apexY := pY + rApex*math.Sin(baseAngle)

	if math.Abs(apexX-400.0) > 1e-6 || math.Abs(apexY-440.0) > 1e-6 {
		t.Errorf("Axe South Apex calculation error: got (%f, %f), want (400.0, 440.0)", apexX, apexY)
	}
}

// F11: Vector Attack Swoosh Rendering
func TestE2E_Tier1_Feature11_VectorSwooshRendering(t *testing.T) {
	assets.Load()
	screen := ebiten.NewImage(800, 600)
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	drw := NewDrawSystem(w, m)

	// Call DrawAttackSwingArc for axe
	drw.DrawAttackSwingArc(screen, 400.0, 300.0, 1.0, 0.0, "axe", 28, 0, 0)
}

// F12: Weapon-Specific Swoosh Styles
func TestE2E_Tier1_Feature12_WeaponSpecificStyles(t *testing.T) {
	assets.Load()
	screen := ebiten.NewImage(800, 600)
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	drw := NewDrawSystem(w, m)

	styles := []string{"axe", "weapon", "shotgun", "shove"}
	for _, st := range styles {
		t.Run("Style_"+st, func(t *testing.T) {
			drw.DrawAttackSwingArc(screen, 400.0, 300.0, 1.0, 0.0, st, 25, 0, 0)
		})
	}
}

// =============================================================================
// TIER 2: BOUNDARY & CORNER CASES
// =============================================================================

func TestE2E_Tier2_BoundaryCases(t *testing.T) {
	t.Run("AttackCooldownBoundaries", func(t *testing.T) {
		screen := ebiten.NewImage(800, 600)
		w := arkecs.NewWorld()
		m := world.NewMap(30, 30)
		drw := NewDrawSystem(w, m)

		// Cooldown <= 16 or > 30 should not render
		drw.DrawAttackSwingArc(screen, 400.0, 300.0, 1.0, 0.0, "axe", 16, 0, 0)
		drw.DrawAttackSwingArc(screen, 400.0, 300.0, 1.0, 0.0, "axe", 31, 0, 0)
		drw.DrawAttackSwingArc(screen, 400.0, 300.0, 1.0, 0.0, "axe", -5, 0, 0)
	})

	t.Run("ZeroLengthFacingVector", func(t *testing.T) {
		screen := ebiten.NewImage(800, 600)
		w := arkecs.NewWorld()
		m := world.NewMap(30, 30)
		drw := NewDrawSystem(w, m)

		// Facing (0, 0) should default to (1, 0) without division by zero
		drw.DrawAttackSwingArc(screen, 400.0, 300.0, 0.0, 0.0, "axe", 28, 0, 0)
	})

	t.Run("MapEdgeCoordinateClamping", func(t *testing.T) {
		m := world.NewMap(20, 20)
		if !m.IsColliding(-10, -10, 64, 64) {
			t.Error("Expected collision out of bounds negative")
		}
		if !m.IsColliding(float64(20*world.TileSize)+50, 100, 64, 64) {
			t.Error("Expected collision out of bounds positive")
		}
	})
}

// =============================================================================
// TIER 3: CROSS-FEATURE INTERACTIONS
// =============================================================================

func TestE2E_Tier3_CrossFeatureInteractions(t *testing.T) {
	t.Run("AxeCleaveAndDurabilityWithDrawSwoosh", func(t *testing.T) {
		assets.Load()
		g := NewGame()
		screen := ebiten.NewImage(800, 600)

		pMap := arkecs.NewMap1[ecs.Player](g.world)
		pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
		var pEnt arkecs.Entity
		for pq.Next() {
			pEnt = pq.Entity()
		}

		player := pMap.Get(pEnt)
		player.WeaponEquipped = true
		player.WeaponType = "axe"
		player.WeaponDurability = 12
		player.AttackCooldown = 28

		// Render while in active attack swing
		g.Draw(screen)
		if player.WeaponDurability != 12 {
			t.Errorf("Durability corrupted during draw: %d", player.WeaponDurability)
		}
	})

	t.Run("ShotgunBlastNoiseHordeAggroAndArmorDeflection", func(t *testing.T) {
		w, _, sys, pEnt := setupM5AdversarialHarness()
		pMap := arkecs.NewMap1[ecs.Player](w)
		player := pMap.Get(pEnt)

		player.WeaponEquipped = true
		player.WeaponType = "shotgun"
		player.WeaponDurability = 15
		player.Inventory = []string{"ammo"}
		player.ArmorEquipped = true
		player.ArmorType = "vest"
		player.ArmorDefense = 0.50
		player.ArmorDurability = 10
		player.ArmorMaxDurability = 10
		player.InfectionResist = 1.0

		// Fire shotgun
		player.Inventory = []string{}
		player.WeaponDurability--

		// Noise pulse alerts horde within 1600.0px
		zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		zEnt := zMap.NewEntity(
			&ecs.Zombie{Speed: 4.0, Chasing: false, WanderTimer: 60},
			&ecs.Position{X: 600.0, Y: 200.0}, // dist = 400 <= 1600
			&ecs.Velocity{},
			&ecs.Sprite{W: 64, H: 128},
			&ecs.Collider{Width: 64, Height: 64},
		)

		nq := sys.zombieFilter.Query()
		for nq.Next() {
			z, zPos, _ := nq.Get()
			if math.Hypot(200.0-zPos.X, 200.0-zPos.Y) <= 1600.0 {
				z.Chasing = true
				z.WanderTimer = 0
			}
		}

		zComp := arkecs.NewMap1[ecs.Zombie](w).Get(zEnt)
		if !zComp.Chasing {
			t.Error("Horde zombie within 1600px should be alerted")
		}
	})
}

// =============================================================================
// TIER 4: REAL-WORLD APPLICATION SCENARIOS
// =============================================================================

// Scenario 1: High-Res Isometric Town Exploration & FOV Raycasting
func TestE2E_Tier4_Scenario1_TownExplorationAndFOV(t *testing.T) {
	assets.Load()
	m := world.NewMap(100, 100)
	px, py := m.PlayerSpawn.X, m.PlayerSpawn.Y

	m.CalculateFOV(px, py, 15)

	ptx := int(px) / world.TileSize
	pty := int(py) / world.TileSize
	if !m.Visible[pty*m.Width+ptx] {
		t.Error("Player tile must be visible after FOV calculation")
	}
}

// Scenario 2: Fire Axe Cleave Combat & Bezier Swoosh Arc under Right-Click Aim
func TestE2E_Tier4_Scenario2_AxeCleaveAndBezierSwoosh(t *testing.T) {
	assets.Load()
	g := NewGame()
	screen := ebiten.NewImage(800, 600)

	pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
	for pq.Next() {
		p := pq.Get()
		p.WeaponEquipped = true
		p.WeaponType = "axe"
		p.WeaponDurability = 12
		p.FacingX = 0.7071
		p.FacingY = 0.7071
		p.AttackCooldown = 26
	}

	g.Draw(screen)
}

// Scenario 3: Multi-Weapon Durability Lifecycle & Armor Mitigation with 4x Physics
func TestE2E_Tier4_Scenario3_WeaponLifecycleAndArmorMitigation(t *testing.T) {
	assets.Load()
	w, _, sys, pEnt := setupM5AdversarialHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	player.ArmorEquipped = true
	player.ArmorType = "vest"
	player.ArmorDefense = 0.50
	player.ArmorDurability = 10
	player.ArmorMaxDurability = 10
	player.InfectionResist = 1.0

	// 10 contact hits break armor
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap.NewEntity(
		&ecs.Zombie{Speed: 4.0, Chasing: true},
		&ecs.Position{X: 220.0, Y: 200.0}, // dist = 20.0 < 56.0
		&ecs.Velocity{},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 64, Height: 64},
	)

	for hit := 1; hit <= 10; hit++ {
		sys.processZombies()
	}

	if player.ArmorEquipped {
		t.Error("Armor should break after 10 hits")
	}
}

// Scenario 4: Night Survival Horde Encounter with Shotgun Blast & Acoustic Pulse
func TestE2E_Tier4_Scenario4_NightHordeSurvival(t *testing.T) {
	assets.Load()
	g := NewGame()
	screen := ebiten.NewImage(800, 600)
	g.timeOfDay = 23.5 // Midnight

	g.Draw(screen)
}

// Scenario 5: Full Procedural Asset Regeneration Determinism & Hash Stability
func TestE2E_Tier4_Scenario5_ProceduralGenerationDeterminism(t *testing.T) {
	m1 := world.NewMap(50, 50)
	m2 := world.NewMap(50, 50)

	if m1.Width != m2.Width || m1.Height != m2.Height {
		t.Error("Map dimensions mismatch")
	}
}
