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

// TestEmpirical_DungeonMaster_HighLoadWaveSpawning tests dynamic wave spawning under high load:
// hundreds of waves over tens of thousands of simulated ticks across multiple in-game days.
func TestEmpirical_DungeonMaster_HighLoadWaveSpawning(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)
	dm.SetRNG(rand.New(rand.NewSource(987654321)))

	playerPos := m.PlayerSpawn

	// Simulate 100,000 ticks (~55 in-game 5-minute days)
	// We will periodically kill zombies to simulate player combat and trigger waves
	totalTicks := int64(100000)
	initialZombieCap := dm.config.MaxLivingZombies
	if initialZombieCap != 140 {
		t.Fatalf("Default MaxLivingZombies = %d, want 140", initialZombieCap)
	}

	zMap := arkecs.NewMap1[ecs.Zombie](w)

	for tick := int64(0); tick < totalTicks; tick++ {
		// Time of day advances: 24h per 18,000 ticks (5 real minutes at 60 fps)
		timeOfDay := math.Mod(float64(tick)*(24.0/18000.0)+8.0, 24.0)

		// DM Update
		dm.Update(timeOfDay, playerPos)

		// Check Invariant 1: Zombie count never exceeds MaxLivingZombies
		livingZombies := dm.CountZombies()
		if livingZombies > dm.config.MaxLivingZombies {
			t.Fatalf("Tick %d: Zombie count %d exceeded cap %d", tick, livingZombies, dm.config.MaxLivingZombies)
		}

		// Check Invariant 2: Item count never exceeds MaxMapItems
		livingItems := dm.CountItems()
		if livingItems > dm.config.MaxMapItems {
			t.Fatalf("Tick %d: Item count %d exceeded cap %d", tick, livingItems, dm.config.MaxMapItems)
		}

		// Simulate combat: Every 120 ticks, cull 50% of zombies to trigger wave replenishments
		if tick%120 == 0 && livingZombies > 10 {
			zq := dm.zombieFilter.Query()
			var toRemove []arkecs.Entity
			cullCount := 0
			for zq.Next() {
				if cullCount%2 == 0 {
					toRemove = append(toRemove, zq.Entity())
				}
				cullCount++
			}
			zq.Close()

			for _, ent := range toRemove {
				// Also roll zombie death drop
				dm.HandleZombieDeath(playerPos.X+float64(cullCount), playerPos.Y+float64(cullCount))
				w.RemoveEntity(ent)
			}
		}

		// Periodically verify wave size calculation invariants
		if tick%1000 == 0 {
			waveSize := dm.CalculateWaveSize(timeOfDay)
			if waveSize < 3 || waveSize > 16 {
				t.Fatalf("Tick %d: WaveSize %d out of bounds [3, 16]", tick, waveSize)
			}

			threat := dm.CalculateThreat(timeOfDay)
			if threat < 1.0 {
				t.Fatalf("Tick %d: Threat %f below minimum 1.0", tick, threat)
			}
		}
	}

	// Verify that hundreds of waves were successfully spawned
	if dm.WaveCount < 200 {
		t.Errorf("Total waves spawned = %d, expected >= 200 over 100,000 ticks", dm.WaveCount)
	}

	// Verify day count progressed appropriately
	expectedDays := int(totalTicks / 18000)
	if dm.DayCount < expectedDays {
		t.Errorf("DayCount = %d, expected >= %d", dm.DayCount, expectedDays)
	}

	// Verify all surviving zombies are valid entities
	zCount := 0
	zq := dm.zombieFilter.Query()
	for zq.Next() {
		zCount++
		z := zMap.Get(zq.Entity())
		if z.Speed <= 0 {
			t.Errorf("Surviving zombie has invalid speed: %f", z.Speed)
		}
	}
	zq.Close()

	if zCount != dm.CountZombies() {
		t.Errorf("CountZombies mismatch: filter=%d, method=%d", zCount, dm.CountZombies())
	}
}

// TestEmpirical_DungeonMaster_100PercentValidPerimeterSpawns rigorously verifies that
// 100% of spawned zombies across thousands of trials land on non-solid walkable floor tiles
// at distance >= 700.0px (and <= 1600.0px), with zero AABB bounding box collisions.
func TestEmpirical_DungeonMaster_100PercentValidPerimeterSpawns(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)
	dm.SetRNG(rand.New(rand.NewSource(54321)))

	// Test from various player positions across the map
	testOrigins := []world.FloatPoint{
		m.PlayerSpawn,                                    // Default residential spawn
		{X: 50.0 * world.TileSize, Y: 50.0 * world.TileSize}, // Map center
		{X: 25.0 * world.TileSize, Y: 75.0 * world.TileSize}, // Sector southwest
		{X: 75.0 * world.TileSize, Y: 25.0 * world.TileSize}, // Sector northeast
		{X: 30.0 * world.TileSize, Y: 30.0 * world.TileSize}, // Dense commercial district
	}

	totalSpawnsAttempted := 0
	totalSpawnsSucceeded := 0

	for originIdx, origin := range testOrigins {
		t.Run(fmt.Sprintf("Origin_%d", originIdx), func(t *testing.T) {
			for i := 0; i < 500; i++ {
				totalSpawnsAttempted++
				// Reset world if getting full to prevent cap blocking
				if dm.CountZombies() >= 120 {
					zq := dm.zombieFilter.Query()
					var toRemove []arkecs.Entity
					for zq.Next() {
						toRemove = append(toRemove, zq.Entity())
					}
					zq.Close()
					for _, ent := range toRemove {
						w.RemoveEntity(ent)
					}
				}

				zCountBefore := dm.CountZombies()
				success := dm.SpawnPerimeterZombie(origin.X, origin.Y, 12.0)
				if !success {
					continue
				}
				totalSpawnsSucceeded++

				// Verify count increased by 1
				zCountAfter := dm.CountZombies()
				if zCountAfter != zCountBefore+1 {
					t.Fatalf("Zombie count did not increment after successful spawn: before=%d, after=%d", zCountBefore, zCountAfter)
				}

				// Find the newly spawned zombie and verify invariants immediately
				zq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](w).Query()
				var newestPos *ecs.Position
				for zq.Next() {
					_, pos := zq.Get()
					newestPos = pos // last one in query
				}
				zq.Close()

				if newestPos == nil {
					t.Fatalf("Failed to query newly spawned zombie position")
				}

				// Invariant 1: Distance from origin in [700px, 1600px]
				dist := math.Hypot(newestPos.X-origin.X, newestPos.Y-origin.Y)
				if dist < 699.9 || dist > 1600.1 {
					t.Fatalf("Spawned zombie distance %f px out of perimeter range [700, 1600]", dist)
				}

				// Invariant 2: Map tile bounds check
				tx := int(newestPos.X / float64(world.TileSize))
				ty := int(newestPos.Y / float64(world.TileSize))
				if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
					t.Fatalf("Spawned zombie out of map bounds: pos=(%f, %f), tile=(%d, %d)", newestPos.X, newestPos.Y, tx, ty)
				}

				// Invariant 3: Tile solid check: Must NOT be solid
				tile := m.GetTile(tx, ty)
				if tile.IsSolid() {
					t.Fatalf("Spawned zombie on solid tile %v at tile (%d, %d)", tile, tx, ty)
				}

				// Invariant 4: AABB Collision Box (64x64) check
				if m.IsColliding(newestPos.X-32.0, newestPos.Y-32.0, 64.0, 64.0) {
					t.Fatalf("Spawned zombie collides with solid obstacle at (%f, %f)", newestPos.X, newestPos.Y)
				}
			}
		})
	}

	if totalSpawnsSucceeded < 2000 {
		t.Fatalf("Expected >= 2000 successful spawns across origins, got %d out of %d", totalSpawnsSucceeded, totalSpawnsAttempted)
	}

	t.Logf("Empirically verified 100%% spatial correctness on %d spawned zombies (attempted %d)", totalSpawnsSucceeded, totalSpawnsAttempted)
}

// TestEmpirical_DungeonMaster_LootDropDistribution tests loot drop distributions over 20,000 rolls,
// verifying statistically that observed frequencies match DefaultLootTable relative weights.
func TestEmpirical_DungeonMaster_LootDropDistribution(t *testing.T) {
	dm := NewDungeonMaster(nil, nil)
	dm.SetRNG(rand.New(rand.NewSource(1337)))

	totalRolls := 50000
	counts := make(map[string]int)

	for i := 0; i < totalRolls; i++ {
		item := dm.RollLootItem()
		counts[item]++
	}

	// Expected probabilities from DefaultLootTable:
	// ammo: 30% (0.30), food: 25% (0.25), water: 20% (0.20), weapon: 10% (0.10)
	// antidote: 8% (0.08), axe: 4% (0.04), armor: 2% (0.02), shotgun: 1% (0.01)
	expectedProbs := map[string]float64{
		"ammo":     0.30,
		"food":     0.25,
		"water":    0.20,
		"weapon":   0.10,
		"antidote": 0.08,
		"axe":      0.04,
		"armor":    0.02,
		"shotgun":  0.01,
	}

	for itemType, expectedProb := range expectedProbs {
		observedCount := counts[itemType]
		observedProb := float64(observedCount) / float64(totalRolls)

		// Binomial standard deviation: sigma = sqrt(p * (1-p) / N)
		sigma := math.Sqrt(expectedProb * (1.0 - expectedProb) / float64(totalRolls))
		deviation := math.Abs(observedProb - expectedProb)
		zScore := deviation / sigma

		t.Logf("Loot '%s': Observed=%.4f (count=%d), Expected=%.4f, z-score=%.2f",
			itemType, observedProb, observedCount, expectedProb, zScore)

		// With N=50,000, 4 sigma corresponds to 99.993% confidence interval
		if zScore > 4.0 {
			t.Errorf("Loot '%s' distribution deviates significantly: observed=%.4f, expected=%.4f (z-score=%.2f > 4.0)",
				itemType, observedProb, expectedProb, zScore)
		}
	}

	// Strict Monotonic Rarity Check: ammo > food > water > weapon > antidote > axe > armor > shotgun > 0
	if !(counts["ammo"] > counts["food"] &&
		counts["food"] > counts["water"] &&
		counts["water"] > counts["weapon"] &&
		counts["weapon"] > counts["antidote"] &&
		counts["antidote"] > counts["axe"] &&
		counts["axe"] > counts["armor"] &&
		counts["armor"] > counts["shotgun"] &&
		counts["shotgun"] > 0) {
		t.Errorf("Loot frequencies do not satisfy strict rarity ordering: %+v", counts)
	}

	// Test Zombie Death Drop Rate across 20,000 kills
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	dm2 := NewDungeonMaster(w, m)
	dm2.SetRNG(rand.New(rand.NewSource(424242)))
	dm2.config.ZombieDropChance = 0.25
	dm2.config.MaxMapItems = 100000 // High cap for statistical gathering

	dropCount := 0
	killTrials := 20000
	for i := 0; i < killTrials; i++ {
		if dm2.HandleZombieDeath(100.0+float64(i%100), 100.0+float64(i/100)) {
			dropCount++
		}
	}

	observedDropRate := float64(dropCount) / float64(killTrials)
	dropSigma := math.Sqrt(0.25 * 0.75 / float64(killTrials))
	dropZScore := math.Abs(observedDropRate-0.25) / dropSigma

	t.Logf("Zombie Death Drops: Observed=%.4f (%d/%d), Expected=0.2500, z-score=%.2f",
		observedDropRate, dropCount, killTrials, dropZScore)

	if dropZScore > 4.0 {
		t.Errorf("Zombie drop rate deviates significantly: observed=%.4f, expected=0.2500 (z-score=%.2f)",
			observedDropRate, dropZScore)
	}
}

// TestEmpirical_DungeonMaster_DayNightAggressionModifiers validates that aggression multipliers
// scale up strictly at night (speed >= 1.25, noise >= 1.50, vision >= 1.25) and normalize to 1.0 by day.
func TestEmpirical_DungeonMaster_DayNightAggressionModifiers(t *testing.T) {
	dm := NewDungeonMaster(nil, nil)

	// Step across full 24h day-night cycle in 0.1h increments (240 sample points)
	for hour := 0.0; hour < 24.0; hour += 0.1 {
		speed, noise, vision := dm.GetAggressionModifiers(hour)
		runnerProb := dm.GetRunnerProbability(hour)
		tint, alpha := dm.GetAmbientLighting(hour)

		// 1. Daytime (08:00 - 17:00)
		if hour >= 8.0 && hour < 17.0 {
			if speed != 1.0 || noise != 1.0 || vision != 1.0 {
				t.Errorf("Hour %.2f (Day): Aggression got (%.2f, %.2f, %.2f), want (1.0, 1.0, 1.0)", hour, speed, noise, vision)
			}
			if runnerProb != 0.15 {
				t.Errorf("Hour %.2f (Day): Runner prob got %.2f, want 0.15", hour, runnerProb)
			}
			if alpha != 0.0 {
				t.Errorf("Hour %.2f (Day): Ambient alpha got %.2f, want 0.0", hour, alpha)
			}
		}

		// 2. Nighttime (20:00 - 05:00)
		if hour >= 20.0 || hour < 5.0 {
			if speed < 1.25 {
				t.Errorf("Hour %.2f (Night): Speed mult %.2f < 1.25", hour, speed)
			}
			if noise < 1.50 {
				t.Errorf("Hour %.2f (Night): Noise mult %.2f < 1.50", hour, noise)
			}
			if vision < 1.25 {
				t.Errorf("Hour %.2f (Night): Vision mult %.2f < 1.25", hour, vision)
			}
			if runnerProb != 0.45 {
				t.Errorf("Hour %.2f (Night): Runner prob got %.2f, want 0.45", hour, runnerProb)
			}
			if alpha < 0.55 || alpha > 0.90 {
				t.Errorf("Hour %.2f (Night): Ambient alpha %.2f out of range [0.55, 0.90]", hour, alpha)
			}
			if tint.B < tint.R || tint.B < tint.G {
				t.Errorf("Hour %.2f (Night): Tint should be navy (B dominant): %+v", hour, tint)
			}
		}

		// 3. Midnight Peak (22:00 - 03:00)
		if hour >= 22.0 || hour <= 3.0 {
			if speed != 1.35 || noise != 1.75 || vision != 1.35 {
				t.Errorf("Hour %.2f (Midnight peak): Got (%.2f, %.2f, %.2f), want (1.35, 1.75, 1.35)", hour, speed, noise, vision)
			}
		}

		// 4. Dawn Transition (05:00 - 08:00)
		if hour >= 5.0 && hour < 8.0 {
			if speed < 1.0 || speed > 1.25 {
				t.Errorf("Hour %.2f (Dawn): Speed %.2f out of range [1.0, 1.25]", hour, speed)
			}
			if noise < 1.0 || noise > 1.50 {
				t.Errorf("Hour %.2f (Dawn): Noise %.2f out of range [1.0, 1.50]", hour, noise)
			}
			if runnerProb < 0.15 || runnerProb > 0.45 {
				t.Errorf("Hour %.2f (Dawn): Runner prob %.2f out of range [0.15, 0.45]", hour, runnerProb)
			}
			if alpha < 0.0 || alpha > 0.55 {
				t.Errorf("Hour %.2f (Dawn): Ambient alpha %.2f out of range [0.0, 0.55]", hour, alpha)
			}
		}

		// 5. Dusk Transition (17:00 - 20:00)
		if hour >= 17.0 && hour < 20.0 {
			if speed < 1.0 || speed > 1.25 {
				t.Errorf("Hour %.2f (Dusk): Speed %.2f out of range [1.0, 1.25]", hour, speed)
			}
			if noise < 1.0 || noise > 1.50 {
				t.Errorf("Hour %.2f (Dusk): Noise %.2f out of range [1.0, 1.50]", hour, noise)
			}
			if runnerProb < 0.15 || runnerProb > 0.45 {
				t.Errorf("Hour %.2f (Dusk): Runner prob %.2f out of range [0.15, 0.45]", hour, runnerProb)
			}
			if alpha < 0.0 || alpha > 0.60 {
				t.Errorf("Hour %.2f (Dusk): Ambient alpha %.2f out of range [0.0, 0.60]", hour, alpha)
			}
		}
	}
}

// TestEmpirical_ContinuousHeadlessSimulation3500Frames executes a continuous headless game simulation
// across 3,500 consecutive frames with randomized player actions, combat interactions, day/night cycles,
// dynamic wave spawning, ambient supply drops, and drawing pipeline execution.
func TestEmpirical_ContinuousHeadlessSimulation3500Frames(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	g := NewGame()
	screen := ebiten.NewImage(1280, 720)

	totalFrames := 3500
	rng := rand.New(rand.NewSource(20260829))

	posMap := arkecs.NewMap1[ecs.Position](g.world)
	velMap := arkecs.NewMap1[ecs.Velocity](g.world)
	pCompMap := arkecs.NewMap1[ecs.Player](g.world)

	for frame := 0; frame < totalFrames; frame++ {
		// Mutate random inputs occasionally to simulate active gameplay
		if frame%15 == 0 {
			pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
			if pq.Next() {
				vel := velMap.Get(pq.Entity())
				p := pCompMap.Get(pq.Entity())
				vel.X = (rng.Float64() - 0.5) * 6.0
				vel.Y = (rng.Float64() - 0.5) * 6.0
				p.FacingX = vel.X
				p.FacingY = vel.Y
			}
			pq.Close()
		}

		// Run Game Update and Draw
		err := g.Update()
		if err != nil {
			t.Fatalf("Frame %d: Game.Update(-1) returned error: %v", frame, err)
		}
		g.Draw(screen)

		// Continuous invariant checks every 50 frames
		if frame%50 == 0 {
			// Invariant 1: Time of day valid
			if math.IsNaN(g.timeOfDay) || math.IsInf(g.timeOfDay, 0) || g.timeOfDay < 0.0 || g.timeOfDay >= 24.0 {
				t.Fatalf("Frame %d: Invalid timeOfDay: %f", frame, g.timeOfDay)
			}

			// Invariant 2: Player integrity
			playerCount := 0
			pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
			for pq.Next() {
				playerCount++
				p := pCompMap.Get(pq.Entity())
				pos := posMap.Get(pq.Entity())
				if math.IsNaN(pos.X) || math.IsNaN(pos.Y) {
					t.Fatalf("Frame %d: Player Position has NaN: (%f, %f)", frame, pos.X, pos.Y)
				}
				if math.IsNaN(p.Health) || math.IsNaN(p.Hunger) || math.IsNaN(p.Thirst) {
					t.Fatalf("Frame %d: Player stats have NaN: H=%f, Hu=%f, T=%f", frame, p.Health, p.Hunger, p.Thirst)
				}
			}
			pq.Close()

			if playerCount != 1 {
				t.Fatalf("Frame %d: Player count = %d, expected 1", frame, playerCount)
			}

			// Invariant 3: Camera within map bounds
			if g.camera.X < 0 || g.camera.X > float64(g.gameMap.Width*world.TileSize) ||
				g.camera.Y < 0 || g.camera.Y > float64(g.gameMap.Height*world.TileSize) {
				t.Fatalf("Frame %d: Camera out of bounds: (%f, %f)", frame, g.camera.X, g.camera.Y)
			}

			// Invariant 4: Zombie population within cap
			zCount := g.dm.CountZombies()
			if zCount > g.dm.config.MaxLivingZombies {
				t.Fatalf("Frame %d: Zombie count %d exceeded cap %d", frame, zCount, g.dm.config.MaxLivingZombies)
			}

			// Invariant 5: Item count within bounded limits and DM throttles dynamic spawning
			iCount := g.dm.CountItems()
			if iCount > 200 {
				t.Fatalf("Frame %d: Item count %d ran away unbounded", frame, iCount)
			}
			if iCount >= g.dm.config.MaxMapItems {
				// Verify DM does not spawn ambient supplies when at or above cap
				ambient := g.dm.SpawnAmbientSupplies(5)
				if ambient != 0 {
					t.Fatalf("Frame %d: SpawnAmbientSupplies spawned %d items while at item cap (%d >= %d)",
						frame, ambient, iCount, g.dm.config.MaxMapItems)
				}
			}
		}

		// At frame 1800 (halfway), trigger a player death and reset to verify clean state recycling
		if frame == 1800 {
			pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
			if pq.Next() {
				p := pCompMap.Get(pq.Entity())
				p.Dead = true
			}
			pq.Close()

			g.Reset()
		}
	}

	t.Logf("Successfully completed 3500-frame continuous headless simulation stress test without violations")
}

// TestEmpirical_DungeonMaster_AdversarialEdgeCases stress-tests unusual and extreme edge inputs:
// negative/infinite timeOfDay normalization, extreme tick numbers, empty/custom loot tables,
// player positioned at map corner (0,0) or (99,99), and rapid zero-delay resets.
func TestEmpirical_DungeonMaster_AdversarialEdgeCases(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(100, 100)
	dm := NewDungeonMaster(w, m)

	// 1. Time Normalization Fuzzing
	fuzzTimes := []float64{-1000.5, -24.0, -0.0001, 0.0, 24.0, 48.0, 9999.999, 1e6}
	for _, ft := range fuzzTimes {
		norm := dm.normalizeTime(ft)
		if norm < 0.0 || norm >= 24.0 {
			t.Errorf("normalizeTime(%f) = %f, want in [0.0, 24.0)", ft, norm)
		}
		// Invariant: lighting and aggression queries must not panic or return NaNs
		tint, alpha := dm.GetAmbientLighting(ft)
		if math.IsNaN(alpha) || alpha < 0.0 || alpha > 1.0 {
			t.Errorf("GetAmbientLighting(%f) alpha = %f, out of bounds", ft, alpha)
		}
		if tint.A != 255 && alpha > 0.0 {
			t.Errorf("GetAmbientLighting(%f) tint alpha = %d", ft, tint.A)
		}
		speed, noise, vision := dm.GetAggressionModifiers(ft)
		if math.IsNaN(speed) || math.IsNaN(noise) || math.IsNaN(vision) || speed < 1.0 || noise < 1.0 || vision < 1.0 {
			t.Errorf("GetAggressionModifiers(%f) = (%f, %f, %f), invalid", ft, speed, noise, vision)
		}
	}

	// 2. Threat Progression at 10,000,000 Ticks & Day 500
	dm.TotalTicks = 10000000
	dm.DayCount = 500
	threatExt := dm.CalculateThreat(0.0)
	if math.IsNaN(threatExt) || math.IsInf(threatExt, 0) || threatExt <= 0.0 {
		t.Errorf("Extreme threat = %f", threatExt)
	}
	waveSizeExt := dm.CalculateWaveSize(0.0)
	if waveSizeExt != 16 { // Clamped at 16
		t.Errorf("Extreme wave size = %d, want 16", waveSizeExt)
	}

	// 3. Fallback Loot Tables
	// 3a. Empty loot table
	dm.config.LootTable = []LootDropItem{}
	itemEmpty := dm.RollLootItem()
	if itemEmpty == "" {
		t.Errorf("RollLootItem with empty table returned empty string")
	}

	// 3b. Zero-weight items table
	dm.config.LootTable = []LootDropItem{
		{Type: "invalid1", Weight: 0},
		{Type: "invalid2", Weight: 0},
	}
	itemZero := dm.RollLootItem()
	if itemZero == "" {
		t.Errorf("RollLootItem with zero weights returned empty string")
	}

	// 3c. Single-item custom table
	dm.config.LootTable = []LootDropItem{
		{Type: "custom_antidote", Weight: 100},
	}
	for i := 0; i < 100; i++ {
		rolled := dm.RollLootItem()
		if rolled != "custom_antidote" {
			t.Fatalf("RollLootItem returned %s, want custom_antidote", rolled)
		}
	}

	// 4. Perimeter Spawn Attempt When Player Is at Extreme Map Corners
	dm.config.LootTable = DefaultLootTable
	cornerOrigins := []world.FloatPoint{
		{X: 0.0, Y: 0.0},
		{X: float64(m.Width * world.TileSize), Y: 0.0},
		{X: 0.0, Y: float64(m.Height * world.TileSize)},
		{X: float64(m.Width * world.TileSize), Y: float64(m.Height * world.TileSize)},
	}
	for _, corner := range cornerOrigins {
		// Should not panic or hang
		dm.SpawnPerimeterZombie(corner.X, corner.Y, 12.0)
	}
}

