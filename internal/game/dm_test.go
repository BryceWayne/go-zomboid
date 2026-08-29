package game

import (
	"math"
	"math/rand"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

func TestDungeonMaster_WaveScaling(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)

	// Test 1: Day 1, Noon (Daytime base)
	dm.TotalTicks = 0
	dm.DayCount = 1
	threatDay1 := dm.CalculateThreat(12.0)
	if threatDay1 != 1.0 {
		t.Errorf("Day 1 Daytime Threat = %f, want 1.0", threatDay1)
	}
	waveDay1 := dm.CalculateWaveSize(12.0)
	if waveDay1 != 3 {
		t.Errorf("Day 1 Daytime WaveSize = %d, want 3", waveDay1)
	}

	// Test 2: Day 1, Midnight (Night bonus +0.50)
	threatNight1 := dm.CalculateThreat(0.0)
	if threatNight1 != 1.5 {
		t.Errorf("Day 1 Midnight Threat = %f, want 1.5", threatNight1)
	}
	waveNight1 := dm.CalculateWaveSize(0.0)
	if waveNight1 != 4 { // floor(3 * 1.5) = 4
		t.Errorf("Day 1 Midnight WaveSize = %d, want 4", waveNight1)
	}

	// Test 3: Day 3, Noon (+0.50 from day count)
	dm.DayCount = 3
	threatDay3 := dm.CalculateThreat(12.0)
	if threatDay3 != 1.5 {
		t.Errorf("Day 3 Daytime Threat = %f, want 1.5", threatDay3)
	}
	waveDay3 := dm.CalculateWaveSize(12.0)
	if waveDay3 != 4 {
		t.Errorf("Day 3 Daytime WaveSize = %d, want 4", waveDay3)
	}

	// Test 4: Day 3, Midnight (+0.50 day count + 0.50 night bonus = 2.0)
	threatNight3 := dm.CalculateThreat(0.0)
	if threatNight3 != 2.0 {
		t.Errorf("Day 3 Midnight Threat = %f, want 2.0", threatNight3)
	}
	waveNight3 := dm.CalculateWaveSize(0.0)
	if waveNight3 != 6 { // floor(3 * 2.0) = 6
		t.Errorf("Day 3 Midnight WaveSize = %d, want 6", waveNight3)
	}

	// Test 5: High ticks elapsed threat progression and max clamping (16)
	dm.TotalTicks = 60 * 180 * 10 // 10 units of tick threat = +10.0
	dm.DayCount = 10              // +2.25
	threatHigh := dm.CalculateThreat(0.0)
	waveHigh := dm.CalculateWaveSize(0.0)
	if threatHigh < 13.0 {
		t.Errorf("High progression threat = %f, want >= 13.0", threatHigh)
	}
	if waveHigh != 16 {
		t.Errorf("Clamped high wave size = %d, want 16", waveHigh)
	}
}

func TestDungeonMaster_PerimeterSpawnValidity(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)
	dm.SetRNG(rand.New(rand.NewSource(42)))

	playerX := m.PlayerSpawn.X
	playerY := m.PlayerSpawn.Y

	// Spawn 50 candidate zombies around player
	for i := 0; i < 50; i++ {
		success := dm.SpawnPerimeterZombie(playerX, playerY, 12.0)
		if !success {
			t.Fatalf("SpawnPerimeterZombie failed at iteration %d", i)
		}
	}

	// Verify all spawned zombies meet invariants
	zFilter := arkecs.NewFilter2[ecs.Zombie, ecs.Position](w)
	q := zFilter.Query()
	count := 0
	for q.Next() {
		count++
		_, pos := q.Get()

		// 1. Distance check [700px, 1600px]
		dist := math.Hypot(pos.X-playerX, pos.Y-playerY)
		if dist < 699.9 || dist > 1600.1 {
			t.Errorf("Zombie #%d spawned at distance %f px from player (want [700, 1600])", count, dist)
		}

		// 2. Tile bounds check
		tx := int(pos.X / float64(world.TileSize))
		ty := int(pos.Y / float64(world.TileSize))
		if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
			t.Errorf("Zombie #%d spawned out of map bounds: tile (%d, %d)", count, tx, ty)
			continue
		}

		// 3. Tile solid check (must be non-solid walkable)
		tile := m.GetTile(tx, ty)
		if tile.IsSolid() {
			t.Errorf("Zombie #%d spawned on solid tile %v at (%d, %d)", count, tile, tx, ty)
		}

		// 4. AABB collision check
		if m.IsColliding(pos.X-32.0, pos.Y-32.0, 64.0, 64.0) {
			t.Errorf("Zombie #%d collides with solid obstacle at (%f, %f)", count, pos.X, pos.Y)
		}
	}

	if count != 50 {
		t.Errorf("Total spawned zombies = %d, want 50", count)
	}
}

func TestDungeonMaster_DeathDrops(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)
	dm.SetRNG(rand.New(rand.NewSource(12345)))

	// Test 1: Guaranteed 100% drop rate
	dm.config.ZombieDropChance = 1.0
	for i := 0; i < 50; i++ {
		deadX := 500.0 + float64(i)*10.0
		deadY := 600.0 + float64(i)*10.0
		dropped := dm.HandleZombieDeath(deadX, deadY)
		if !dropped {
			t.Errorf("HandleZombieDeath failed at iteration %d with 100%% drop chance", i)
		}
	}
	if dm.CountItems() != 50 {
		t.Errorf("Total dropped items = %d, want 50", dm.CountItems())
	}

	// Test 2: Statistical 25% drop rate
	w2 := arkecs.NewWorld()
	dm2 := NewDungeonMaster(w2, m)
	dm2.SetRNG(rand.New(rand.NewSource(999)))
	dm2.config.ZombieDropChance = 0.25
	dm2.config.MaxMapItems = 5000

	drops := 0
	totalKills := 2000
	for i := 0; i < totalKills; i++ {
		if dm2.HandleZombieDeath(float64(i), float64(i)) {
			drops++
		}
	}
	rate := float64(drops) / float64(totalKills)
	if rate < 0.20 || rate > 0.30 {
		t.Errorf("Observed drop rate = %f (want ~0.25 +/- 0.05)", rate)
	}

	// Test 3: Weighted drop table across all 8 items
	counts := make(map[string]int)
	rolls := 20000
	for i := 0; i < rolls; i++ {
		item := dm.RollLootItem()
		counts[item]++
	}

	expectedItems := []string{"ammo", "food", "water", "weapon", "antidote", "axe", "armor", "shotgun"}
	for _, it := range expectedItems {
		if counts[it] == 0 {
			t.Errorf("Item %s never dropped in %d rolls", it, rolls)
		}
	}

	// Verify relative frequencies: ammo > food > water > weapon > antidote > axe > armor > shotgun
	if counts["ammo"] <= counts["food"] ||
		counts["food"] <= counts["water"] ||
		counts["water"] <= counts["weapon"] ||
		counts["weapon"] <= counts["antidote"] ||
		counts["antidote"] <= counts["axe"] ||
		counts["axe"] <= counts["armor"] ||
		counts["armor"] <= counts["shotgun"] {
		t.Errorf("Drop frequencies do not match expected rarity ordering: %+v", counts)
	}
}

func TestDungeonMaster_AmbientRestock(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)
	dm.SetRNG(rand.New(rand.NewSource(777)))

	if dm.CountItems() != 0 {
		t.Fatalf("Initial ground items = %d, want 0", dm.CountItems())
	}

	// Advance DM simulation by SupplyDropInterval ticks
	for i := int64(0); i <= dm.config.SupplyDropInterval; i++ {
		dm.Update(12.0, m.PlayerSpawn)
	}

	itemCount := dm.CountItems()
	if itemCount < 2 {
		t.Errorf("Ambient restock spawned %d items, want >= 2", itemCount)
	}

	// Verify all spawned items are on non-solid floor tiles
	iFilter := arkecs.NewFilter2[ecs.Item, ecs.Position](w)
	q := iFilter.Query()
	for q.Next() {
		_, pos := q.Get()
		tx := int(pos.X / float64(world.TileSize))
		ty := int(pos.Y / float64(world.TileSize))
		tile := m.GetTile(tx, ty)
		if tile.IsSolid() {
			t.Errorf("Ambient supply spawned on solid tile %v at (%d, %d)", tile, tx, ty)
		}
	}
}

func TestDungeonMaster_DayNightLighting(t *testing.T) {
	dm := NewDungeonMaster(nil, nil)

	// 1. Daytime (08:00 - 17:00): Alpha must be 0.0 (clear sunlight)
	dayHours := []float64{8.0, 10.0, 12.0, 14.0, 16.99}
	for _, h := range dayHours {
		_, alpha := dm.GetAmbientLighting(h)
		if alpha != 0.0 {
			t.Errorf("Hour %f: ambient alpha = %f, want 0.0", h, alpha)
		}
	}

	// 2. Dawn (05:00 - 08:00): Warm rose/gold tint, alpha in (0, 0.55]
	dawnTint, dawnAlpha := dm.GetAmbientLighting(6.0)
	if dawnAlpha <= 0.0 || dawnAlpha > 0.55 {
		t.Errorf("Dawn (06:00): alpha = %f, want in (0.0, 0.55]", dawnAlpha)
	}
	if dawnTint.R != 180 || dawnTint.G != 140 || dawnTint.B != 80 {
		t.Errorf("Dawn tint = %+v, want {180, 140, 80}", dawnTint)
	}

	// 3. Dusk (17:00 - 20:00): Amber twilight tint, alpha in (0, 0.60]
	duskTint, duskAlpha := dm.GetAmbientLighting(18.5)
	if duskAlpha <= 0.0 || duskAlpha > 0.60 {
		t.Errorf("Dusk (18:30): alpha = %f, want in (0.0, 0.60]", duskAlpha)
	}
	if duskTint.R != 140 || duskTint.G != 60 || duskTint.B != 50 {
		t.Errorf("Dusk tint = %+v, want {140, 60, 50}", duskTint)
	}

	// 4. Midnight (00:00): Midnight navy tint peaking at alpha ~0.85-0.90
	midTint, midAlpha := dm.GetAmbientLighting(0.0)
	if midAlpha < 0.85 || midAlpha > 0.90 {
		t.Errorf("Midnight (00:00): alpha = %f, want in [0.85, 0.90]", midAlpha)
	}
	if midTint.R != 5 || midTint.G != 10 || midTint.B != 35 {
		t.Errorf("Midnight tint = %+v, want {5, 10, 35}", midTint)
	}

	// 5. Deep night hours
	nightHours := []float64{20.0, 22.0, 24.0, 2.0, 4.5}
	for _, h := range nightHours {
		_, alpha := dm.GetAmbientLighting(h)
		if alpha < 0.55 || alpha > 0.90 {
			t.Errorf("Night hour %f: alpha = %f, want in [0.55, 0.90]", h, alpha)
		}
	}
}

func TestDungeonMaster_NightAggressionScaling(t *testing.T) {
	dm := NewDungeonMaster(nil, nil)

	// 1. Daytime (08:00 - 17:00): Must return 1.0, 1.0, 1.0
	dayHours := []float64{8.0, 9.5, 12.0, 15.0, 16.99}
	for _, h := range dayHours {
		s, n, v := dm.GetAggressionModifiers(h)
		if s != 1.0 || n != 1.0 || v != 1.0 {
			t.Errorf("Daytime hour %f: got (%f, %f, %f), want (1.0, 1.0, 1.0)", h, s, n, v)
		}
	}

	// 2. Midnight (00:00): speedMult >= 1.25, noiseMult >= 1.50, visionMult >= 1.25
	midSpeed, midNoise, midVision := dm.GetAggressionModifiers(0.0)
	if midSpeed < 1.25 || midNoise < 1.50 || midVision < 1.25 {
		t.Errorf("Midnight aggression = (%f, %f, %f), want >= (1.25, 1.50, 1.25)", midSpeed, midNoise, midVision)
	}
	if midSpeed > 1.35 || midNoise > 1.75 || midVision > 1.35 {
		t.Errorf("Midnight aggression exceeds ceiling: (%f, %f, %f)", midSpeed, midNoise, midVision)
	}

	// 3. Early night (21:00)
	nightSpeed, nightNoise, nightVision := dm.GetAggressionModifiers(21.0)
	if nightSpeed < 1.25 || nightNoise < 1.50 || nightVision < 1.25 {
		t.Errorf("Night 21:00 aggression = (%f, %f, %f), want >= (1.25, 1.50, 1.25)", nightSpeed, nightNoise, nightVision)
	}
}

func TestDungeonMaster_RunnerScaling(t *testing.T) {
	dm := NewDungeonMaster(nil, nil)

	dayProb := dm.GetRunnerProbability(12.0)
	if dayProb != 0.15 {
		t.Errorf("Daytime runner probability = %f, want 0.15", dayProb)
	}

	nightProb := dm.GetRunnerProbability(0.0)
	if nightProb != 0.45 {
		t.Errorf("Night runner probability = %f, want 0.45", nightProb)
	}

	// Statistical test
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dmWorld := NewDungeonMaster(w, m)
	dmWorld.SetRNG(rand.New(rand.NewSource(42)))

	dayRunners := 0
	samples := 2000
	for i := 0; i < samples; i++ {
		prob := dmWorld.GetRunnerProbability(12.0)
		if dmWorld.rng.Float64() < prob {
			dayRunners++
		}
	}
	dayRatio := float64(dayRunners) / float64(samples)
	if math.Abs(dayRatio-0.15) > 0.03 {
		t.Errorf("Empirical day runner ratio = %f, want ~0.15", dayRatio)
	}

	nightRunners := 0
	for i := 0; i < samples; i++ {
		prob := dmWorld.GetRunnerProbability(0.0)
		if dmWorld.rng.Float64() < prob {
			nightRunners++
		}
	}
	nightRatio := float64(nightRunners) / float64(samples)
	if math.Abs(nightRatio-0.45) > 0.03 {
		t.Errorf("Empirical night runner ratio = %f, want ~0.45", nightRatio)
	}
}

func TestDungeonMaster_MaxCapEnforcement(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)
	dm.config.MaxLivingZombies = 10
	dm.config.MaxMapItems = 5

	// 1. Fill zombie cap
	for i := 0; i < 10; i++ {
		dm.SpawnPerimeterZombie(m.PlayerSpawn.X, m.PlayerSpawn.Y, 12.0)
	}
	if dm.CountZombies() != 10 {
		t.Fatalf("Zombie count = %d, want 10", dm.CountZombies())
	}

	// Attempt wave spawn when at cap
	spawned := dm.SpawnWave(12.0, m.PlayerSpawn)
	if spawned != 0 {
		t.Errorf("SpawnWave at cap returned %d, want 0", spawned)
	}
	if dm.CountZombies() != 10 {
		t.Errorf("Zombie count exceeded cap: %d", dm.CountZombies())
	}

	// 2. Fill item cap
	for i := 0; i < 5; i++ {
		dm.itemMap.NewEntity(&ecs.Item{Type: "food"}, &ecs.Position{X: 100, Y: 100})
	}
	if dm.CountItems() != 5 {
		t.Fatalf("Item count = %d, want 5", dm.CountItems())
	}

	// Attempt death drop and ambient drop when at cap
	dm.config.ZombieDropChance = 1.0
	dropped := dm.HandleZombieDeath(100, 100)
	if dropped {
		t.Errorf("HandleZombieDeath succeeded when at item cap")
	}
	ambientSpawned := dm.SpawnAmbientSupplies(5)
	if ambientSpawned != 0 {
		t.Errorf("SpawnAmbientSupplies spawned %d items when at cap, want 0", ambientSpawned)
	}
	if dm.CountItems() != 5 {
		t.Errorf("Item count exceeded cap: %d", dm.CountItems())
	}
}

func TestDungeonMaster_GameLoopIntegration(t *testing.T) {
	assets.Load()
	g := NewGame()
	if g.dm == nil {
		t.Fatal("Game.dm is nil after NewGame()")
	}
	if g.updateSys.dm == nil {
		t.Fatal("UpdateSystem.dm is nil after NewGame()")
	}
	if g.drawSys.dm == nil {
		t.Fatal("DrawSystem.dm is nil after NewGame()")
	}

	initialTicks := g.dm.TotalTicks
	// Run game update for 60 frames
	for i := 0; i < 60; i++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Game.Update() returned error: %v", err)
		}
	}

	if g.dm.TotalTicks != initialTicks+60 {
		t.Errorf("DM TotalTicks = %d, want %d", g.dm.TotalTicks, initialTicks+60)
	}
}
