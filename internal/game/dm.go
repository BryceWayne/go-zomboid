package game

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// LootDropItem represents an item entry with relative weighting in the loot drop table.
type LootDropItem struct {
	Type   string
	Weight int
}

// DefaultLootTable defines the 8 weighted item drops upon zombie death or ambient restock.
// ammo: 30%, food: 25%, water: 20%, weapon: 10%, antidote: 8%, axe: 4%, armor: 2%, shotgun: 1%.
var DefaultLootTable = []LootDropItem{
	{Type: "ammo", Weight: 30},
	{Type: "food", Weight: 25},
	{Type: "water", Weight: 20},
	{Type: "weapon", Weight: 10},
	{Type: "antidote", Weight: 8},
	{Type: "axe", Weight: 4},
	{Type: "armor", Weight: 2},
	{Type: "shotgun", Weight: 1},
}

// DungeonMasterConfig holds configuration parameters for the Dungeon Master simulation.
type DungeonMasterConfig struct {
	// Wave Spawning Config
	WaveIntervalTicks      int64   // e.g. 1800 frames = 30.0s between wave evaluations
	BaseZombiesPerWave     int     // Base wave size (default 3)
	MaxLivingZombies       int     // Cap on concurrent living zombies (default 140)
	MinSpawnDistance       float64 // 700.0px (outside immediate viewport)
	MaxSpawnDistance       float64 // 1600.0px (within active neighborhood)
	DayRunnerProbability   float64 // 0.15 (15% runners by day)
	NightRunnerProbability float64 // 0.45 (45% runners at night)
	SpawnRetryLimit        int     // Max candidate search attempts per zombie (default 50)

	// Loot Drop Config
	ZombieDropChance   float64        // 0.25 (25% chance of item drop on kill)
	MaxMapItems        int            // Cap on concurrent ground items (default 60)
	SupplyDropInterval int64          // 3600 frames = 60.0s between ambient supply drops
	MinSupplyPerDrop   int            // Minimum items per ambient supply roll (default 2)
	MaxSupplyPerDrop   int            // Maximum items per ambient supply roll (default 4)
	LootTable          []LootDropItem // Active loot drop table

	// Day/Night & Aggression Config
	DayCycleMinutes       float64 // 5.0 real minutes per 24 in-game hours
	NightSpeedMultiplier  float64 // 1.25 base (+25% speed at night, up to 1.35 at midnight)
	NightNoiseMultiplier  float64 // 1.50 base (+50% hearing at night, up to 1.75 at midnight)
	NightVisionMultiplier float64 // 1.25 base (+25% vision at night, up to 1.35 at midnight)
}

// DefaultDungeonMasterConfig returns standard balanced configuration for the DM simulation.
func DefaultDungeonMasterConfig() DungeonMasterConfig {
	return DungeonMasterConfig{
		WaveIntervalTicks:      1800,
		BaseZombiesPerWave:     3,
		MaxLivingZombies:       140,
		MinSpawnDistance:       700.0,
		MaxSpawnDistance:       1600.0,
		DayRunnerProbability:   0.15,
		NightRunnerProbability: 0.45,
		SpawnRetryLimit:        50,

		ZombieDropChance:   0.25,
		MaxMapItems:        60,
		SupplyDropInterval: 3600,
		MinSupplyPerDrop:   2,
		MaxSupplyPerDrop:   4,
		LootTable:          DefaultLootTable,

		DayCycleMinutes:       5.0,
		NightSpeedMultiplier:  1.25,
		NightNoiseMultiplier:  1.50,
		NightVisionMultiplier: 1.25,
	}
}

// DungeonMaster orchestrates dynamic wave spawning, difficulty threat scaling, dynamic loot drops,
// ambient supply restock, and day/night aggression & lighting modifiers.
type DungeonMaster struct {
	world   *arkecs.World
	gameMap *world.Map
	config  DungeonMasterConfig

	// State trackers
	TotalTicks         int64
	DayCount           int
	WaveCount          int
	NextWaveTick       int64
	NextSupplyDropTick int64
	LastSpawnTick      int64
	LastTimeOfDay      float64

	// Random generator
	rng *rand.Rand

	// ECS Entity Mappers
	zombieMap *arkecs.Map5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider]
	itemMap   *arkecs.Map2[ecs.Item, ecs.Position]
	playerMap *arkecs.Map2[ecs.Player, ecs.Position]

	// ECS Query Filters
	zombieFilter *arkecs.Filter1[ecs.Zombie]
	itemFilter   *arkecs.Filter1[ecs.Item]
	playerFilter *arkecs.Filter2[ecs.Player, ecs.Position]
}

// NewDungeonMaster constructs and initializes a new DungeonMaster instance.
func NewDungeonMaster(w *arkecs.World, m *world.Map, cfgs ...DungeonMasterConfig) *DungeonMaster {
	cfg := DefaultDungeonMasterConfig()
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}

	dm := &DungeonMaster{
		world:              w,
		gameMap:            m,
		config:             cfg,
		TotalTicks:         0,
		DayCount:           1,
		WaveCount:          0,
		NextWaveTick:       cfg.WaveIntervalTicks,
		NextSupplyDropTick: cfg.SupplyDropInterval,
		LastSpawnTick:      0,
		LastTimeOfDay:      8.0,
		rng:                rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if w != nil {
		dm.zombieMap = arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		dm.itemMap = arkecs.NewMap2[ecs.Item, ecs.Position](w)
		dm.playerMap = arkecs.NewMap2[ecs.Player, ecs.Position](w)
		dm.zombieFilter = arkecs.NewFilter1[ecs.Zombie](w)
		dm.itemFilter = arkecs.NewFilter1[ecs.Item](w)
		dm.playerFilter = arkecs.NewFilter2[ecs.Player, ecs.Position](w)
	}

	return dm
}

// SetRNG sets a custom random generator for deterministic testing.
func (dm *DungeonMaster) SetRNG(r *rand.Rand) {
	dm.rng = r
}

// GetConfig returns the current DungeonMasterConfig.
func (dm *DungeonMaster) GetConfig() DungeonMasterConfig {
	return dm.config
}

// SetConfig sets the current DungeonMasterConfig.
func (dm *DungeonMaster) SetConfig(cfg DungeonMasterConfig) {
	dm.config = cfg
}

// CountZombies returns the number of active living zombie entities in the ECS world.
func (dm *DungeonMaster) CountZombies() int {
	if dm.world == nil || dm.zombieFilter == nil {
		return 0
	}
	q := dm.zombieFilter.Query()
	defer q.Close()
	return q.Count()
}

// CountItems returns the number of active ground items in the ECS world.
func (dm *DungeonMaster) CountItems() int {
	if dm.world == nil || dm.itemFilter == nil {
		return 0
	}
	q := dm.itemFilter.Query()
	defer q.Close()
	return q.Count()
}

// CalculateThreat computes the dynamic threat scaling factor based on elapsed ticks, day count, and night status.
// Threat(t) = 1.0 + (TotalTicks / (60 * 180)) + 0.25 * (DayCount - 1) + (0.50 if Night else 0.0)
func (dm *DungeonMaster) CalculateThreat(timeOfDay float64) float64 {
	threat := 1.0 + float64(dm.TotalTicks)/float64(60*180)
	if dm.DayCount > 1 {
		threat += 0.25 * float64(dm.DayCount-1)
	}

	// Night bonus threat
	normalizedTime := dm.normalizeTime(timeOfDay)
	if normalizedTime < 5.0 || normalizedTime >= 20.0 {
		threat += 0.50
	}

	return threat
}

// CalculateWaveSize computes the number of zombies to spawn in a wave, clamped between 3 and 16.
func (dm *DungeonMaster) CalculateWaveSize(timeOfDay float64) int {
	threat := dm.CalculateThreat(timeOfDay)
	size := int(math.Floor(float64(dm.config.BaseZombiesPerWave) * threat))
	if size < 3 {
		size = 3
	}
	if size > 16 {
		size = 16
	}
	return size
}

// GetRunnerProbability returns the probability of a spawned zombie being a runner.
// Daytime (08:00 - 17:00): 15% (0.15)
// Night (20:00 - 05:00): 45% (0.45)
// Transitions interpolate smoothly across Dawn and Dusk.
func (dm *DungeonMaster) GetRunnerProbability(timeOfDay float64) float64 {
	t := dm.normalizeTime(timeOfDay)
	if t >= 8.0 && t < 17.0 {
		return dm.config.DayRunnerProbability
	} else if t >= 5.0 && t < 8.0 {
		// Dawn: transitions from NightRunnerProbability (at 5:00) to DayRunnerProbability (at 8:00)
		blend := (t - 5.0) / 3.0
		return dm.config.NightRunnerProbability - blend*(dm.config.NightRunnerProbability-dm.config.DayRunnerProbability)
	} else if t >= 17.0 && t < 20.0 {
		// Dusk: transitions from DayRunnerProbability (at 17:00) to NightRunnerProbability (at 20:00)
		blend := (t - 17.0) / 3.0
		return dm.config.DayRunnerProbability + blend*(dm.config.NightRunnerProbability-dm.config.DayRunnerProbability)
	} else {
		return dm.config.NightRunnerProbability
	}
}

// GetAggressionModifiers returns multipliers for zombie speed, noise detection radius, and vision radius.
// Daytime (08:00 - 17:00): 1.0, 1.0, 1.0
// Night (20:00 - 05:00): speedMult >= 1.25 (up to 1.35 at midnight), noiseMult >= 1.50 (up to 1.75), visionMult >= 1.25 (up to 1.35).
func (dm *DungeonMaster) GetAggressionModifiers(timeOfDay float64) (speedMult, noiseMult, visionMult float64) {
	t := dm.normalizeTime(timeOfDay)

	if t >= 8.0 && t < 17.0 {
		return 1.0, 1.0, 1.0
	} else if t >= 5.0 && t < 8.0 {
		// Dawn: Blend from night down to day (1.0)
		blend := (t - 5.0) / 3.0 // 0.0 at 5:00, 1.0 at 8:00
		speed := 1.25 - blend*0.25
		noise := 1.50 - blend*0.50
		vision := 1.25 - blend*0.25
		return speed, noise, vision
	} else if t >= 17.0 && t < 20.0 {
		// Dusk: Blend from day (1.0) up to night
		blend := (t - 17.0) / 3.0 // 0.0 at 17:00, 1.0 at 20:00
		speed := 1.0 + blend*0.25
		noise := 1.0 + blend*0.50
		vision := 1.0 + blend*0.25
		return speed, noise, vision
	} else {
		// Night: Check midnight proximity peak (between 22:00 and 03:00)
		if t >= 22.0 || t <= 3.0 {
			return 1.35, 1.75, 1.35
		}
		return 1.25, 1.50, 1.25
	}
}

// GetAmbientLighting returns the ambient tint color and darkness overlay alpha.
// Dawn (05:00 - 08:00): Warm rose/gold tint, alpha transitions 0.55 -> 0.0.
// Day (08:00 - 17:00): Clear sunlight (alpha 0.0).
// Dusk (17:00 - 20:00): Amber twilight tint, alpha transitions 0.0 -> 0.60.
// Night (20:00 - 05:00): Midnight navy tint peaking at alpha ~0.85 - 0.90 at Midnight (00:00).
func (dm *DungeonMaster) GetAmbientLighting(timeOfDay float64) (color.RGBA, float64) {
	t := dm.normalizeTime(timeOfDay)

	if t >= 8.0 && t < 17.0 {
		// Day: clear sunlight
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}, 0.0
	} else if t >= 5.0 && t < 8.0 {
		// Dawn: Warm rose/gold tint fading out
		blend := (8.0 - t) / 3.0 // 1.0 at 5:00, 0.0 at 8:00
		alpha := blend * 0.55
		tint := color.RGBA{R: 180, G: 140, B: 80, A: 255}
		return tint, alpha
	} else if t >= 17.0 && t < 20.0 {
		// Dusk: Amber twilight tint fading in
		blend := (t - 17.0) / 3.0 // 0.0 at 17:00, 1.0 at 20:00
		alpha := blend * 0.60
		tint := color.RGBA{R: 140, G: 60, B: 50, A: 255}
		return tint, alpha
	} else {
		// Night: Midnight navy tint peaking at midnight (0.0 / 24.0)
		distFromMidnight := t
		if distFromMidnight > 12.0 {
			distFromMidnight = 24.0 - distFromMidnight
		}
		// distFromMidnight is in [0.0, 5.0]
		// At 0.0 (midnight): alpha = 0.88
		// At 4.0 (20:00 / 04:00): alpha ~ 0.65
		// At 5.0 (05:00): alpha = 0.55
		alpha := 0.88 - (distFromMidnight/5.0)*0.33
		if alpha < 0.55 {
			alpha = 0.55
		}
		if alpha > 0.90 {
			alpha = 0.90
		}
		tint := color.RGBA{R: 5, G: 10, B: 35, A: 255}
		return tint, alpha
	}
}

// RollLootItem selects a randomized loot item from the active loot table according to weights.
func (dm *DungeonMaster) RollLootItem() string {
	table := dm.config.LootTable
	if len(table) == 0 {
		table = DefaultLootTable
	}

	totalWeight := 0
	for _, item := range table {
		totalWeight += item.Weight
	}
	if totalWeight <= 0 {
		return "food"
	}

	roll := dm.rng.Intn(totalWeight)
	accum := 0
	for _, item := range table {
		accum += item.Weight
		if roll < accum {
			return item.Type
		}
	}
	return table[0].Type
}

// HandleZombieDeath processes loot drops when a zombie is killed at (wx, wy).
// 25% chance of item drop upon kill, weighted drop table across 8 items.
func (dm *DungeonMaster) HandleZombieDeath(wx, wy float64) bool {
	if dm.world == nil {
		return false
	}

	// Check ground item cap
	if dm.CountItems() >= dm.config.MaxMapItems {
		return false
	}

	// Roll drop chance
	if dm.rng.Float64() < dm.config.ZombieDropChance {
		itemType := dm.RollLootItem()
		dm.itemMap.NewEntity(
			&ecs.Item{Type: itemType},
			&ecs.Position{X: wx, Y: wy},
		)
		return true
	}

	return false
}

// SpawnPerimeterZombie attempts to spawn a single candidate zombie at perimeter distance (700px - 1600px)
// from the player on a valid, non-solid walkable tile.
func (dm *DungeonMaster) SpawnPerimeterZombie(playerX, playerY, timeOfDay float64) bool {
	if dm.world == nil || dm.gameMap == nil {
		return false
	}

	minDist := dm.config.MinSpawnDistance
	maxDist := dm.config.MaxSpawnDistance
	retries := dm.config.SpawnRetryLimit
	if retries <= 0 {
		retries = 50
	}

	for i := 0; i < retries; i++ {
		angle := dm.rng.Float64() * 2 * math.Pi
		dist := minDist + dm.rng.Float64()*(maxDist-minDist)

		candX := playerX + math.Cos(angle)*dist
		candY := playerY + math.Sin(angle)*dist

		tx := int(candX / float64(world.TileSize))
		ty := int(candY / float64(world.TileSize))

		// Check bounds within map interior
		if tx < 2 || tx >= dm.gameMap.Width-2 || ty < 2 || ty >= dm.gameMap.Height-2 {
			continue
		}

		// Check non-solid walkable tile
		if dm.gameMap.GetTile(tx, ty).IsSolid() {
			continue
		}

		// Check AABB collision box (64x64 centered at candX, candY)
		if dm.gameMap.IsColliding(candX-32.0, candY-32.0, 64.0, 64.0) {
			continue
		}

		// Valid candidate tile found! Determine runner status and speed
		isRunner := dm.rng.Float64() < dm.GetRunnerProbability(timeOfDay)
		speed := 4.0 + dm.rng.Float64()*2.0
		if isRunner {
			speed = 8.8 + dm.rng.Float64()*1.6
		}

		dm.zombieMap.NewEntity(
			&ecs.Zombie{
				Speed:       speed,
				IsRunner:    isRunner,
				WanderTimer: dm.rng.Intn(120),
			},
			&ecs.Position{X: candX, Y: candY},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{
				Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
				W:     64,
				H:     128,
			},
			&ecs.Collider{Width: 64, Height: 64},
		)
		return true
	}

	return false
}

// SpawnWave executes a wave spawn of candidate zombies around player position.
func (dm *DungeonMaster) SpawnWave(timeOfDay float64, playerPos ...world.FloatPoint) int {
	if dm.world == nil || dm.gameMap == nil {
		return 0
	}

	livingZombies := dm.CountZombies()
	if livingZombies >= dm.config.MaxLivingZombies {
		return 0
	}

	waveSize := dm.CalculateWaveSize(timeOfDay)
	availableCap := dm.config.MaxLivingZombies - livingZombies
	toSpawn := waveSize
	if toSpawn > availableCap {
		toSpawn = availableCap
	}

	var px, py float64
	if len(playerPos) > 0 {
		px, py = playerPos[0].X, playerPos[0].Y
	} else if dm.playerFilter != nil {
		pq := dm.playerFilter.Query()
		if pq.Next() {
			_, pPos := pq.Get()
			px, py = pPos.X, pPos.Y
		} else if dm.gameMap != nil {
			px, py = dm.gameMap.PlayerSpawn.X, dm.gameMap.PlayerSpawn.Y
		}
		pq.Close()
	}

	spawned := 0
	for i := 0; i < toSpawn; i++ {
		if dm.SpawnPerimeterZombie(px, py, timeOfDay) {
			spawned++
		}
	}

	if spawned > 0 {
		dm.WaveCount++
		dm.LastSpawnTick = dm.TotalTicks
	}

	return spawned
}

// SpawnAmbientSupplies injects ambient supply loot items into valid building rooms or walkable tiles.
func (dm *DungeonMaster) SpawnAmbientSupplies(targetCount int) int {
	if dm.world == nil || dm.gameMap == nil || targetCount <= 0 {
		return 0
	}

	curItems := dm.CountItems()
	if curItems >= dm.config.MaxMapItems {
		return 0
	}

	maxCanSpawn := dm.config.MaxMapItems - curItems
	if targetCount > maxCanSpawn {
		targetCount = maxCanSpawn
	}

	spawned := 0
	buildings := dm.gameMap.Buildings

	for i := 0; i < targetCount; i++ {
		var spawnX, spawnY float64
		var found bool

		// Attempt 1: Pick a room in a building
		if len(buildings) > 0 {
			for attempt := 0; attempt < 20; attempt++ {
				bIdx := dm.rng.Intn(len(buildings))
				b := buildings[bIdx]
				if len(b.Rooms) == 0 {
					continue
				}
				rIdx := dm.rng.Intn(len(b.Rooms))
				rm := b.Rooms[rIdx]

				tx := rm.X + 1 + dm.rng.Intn(rm.W-2)
				ty := rm.Y + 1 + dm.rng.Intn(rm.H-2)

				if tx >= 0 && tx < dm.gameMap.Width && ty >= 0 && ty < dm.gameMap.Height {
					tile := dm.gameMap.GetTile(tx, ty)
					if !tile.IsSolid() && tile.IsFloor() {
						spawnX = float64(tx)*float64(world.TileSize) + 64.0
						spawnY = float64(ty)*float64(world.TileSize) + 64.0
						found = true
						break
					}
				}
			}
		}

		// Fallback: Pick any random walkable non-solid tile
		if !found {
			for attempt := 0; attempt < 30; attempt++ {
				tx := 3 + dm.rng.Intn(dm.gameMap.Width-6)
				ty := 3 + dm.rng.Intn(dm.gameMap.Height-6)
				tile := dm.gameMap.GetTile(tx, ty)
				if !tile.IsSolid() && tile.IsFloor() {
					spawnX = float64(tx)*float64(world.TileSize) + 64.0
					spawnY = float64(ty)*float64(world.TileSize) + 64.0
					found = true
					break
				}
			}
		}

		if found {
			itemType := dm.RollLootItem()
			dm.itemMap.NewEntity(
				&ecs.Item{Type: itemType},
				&ecs.Position{X: spawnX, Y: spawnY},
			)
			spawned++
		}
	}

	return spawned
}

// Update advances the DungeonMaster simulation tick, evaluating wave spawns and ambient supply rolls.
func (dm *DungeonMaster) Update(timeOfDay float64, playerPos ...world.FloatPoint) {
	dm.TotalTicks++

	// Track day count progression
	normTime := dm.normalizeTime(timeOfDay)
	if normTime < dm.LastTimeOfDay && dm.LastTimeOfDay-normTime > 12.0 {
		dm.DayCount++
	}
	dm.LastTimeOfDay = normTime

	// 1. Dynamic Zombie Wave Spawning
	// Periodic evaluation or reinforcement trigger when zombie population drops below 15
	livingZombies := dm.CountZombies()
	isPeriodic := dm.TotalTicks >= dm.NextWaveTick
	isLowPopulation := livingZombies < 15 && dm.TotalTicks >= dm.LastSpawnTick+300

	if isPeriodic || isLowPopulation {
		dm.SpawnWave(timeOfDay, playerPos...)
		dm.NextWaveTick = dm.TotalTicks + dm.config.WaveIntervalTicks
	}

	// 2. Periodic Ambient Supply Drops
	if dm.TotalTicks >= dm.NextSupplyDropTick {
		dm.NextSupplyDropTick = dm.TotalTicks + dm.config.SupplyDropInterval
		if dm.CountItems() < dm.config.MaxMapItems {
			numDrops := dm.config.MinSupplyPerDrop
			if dm.config.MaxSupplyPerDrop > dm.config.MinSupplyPerDrop {
				numDrops += dm.rng.Intn(dm.config.MaxSupplyPerDrop - dm.config.MinSupplyPerDrop + 1)
			}
			dm.SpawnAmbientSupplies(numDrops)
		}
	}
}

// normalizeTime constrains timeOfDay into [0.0, 24.0).
func (dm *DungeonMaster) normalizeTime(t float64) float64 {
	for t >= 24.0 {
		t -= 24.0
	}
	for t < 0.0 {
		t += 24.0
	}
	return t
}
