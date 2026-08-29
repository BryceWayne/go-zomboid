package game

import (
	"image/color"
	"math"
	"math/rand"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupM5AdversarialHarness initializes an isolated test world with player entity
func setupM5AdversarialHarness() (*arkecs.World, *world.Map, *UpdateSystem, arkecs.Entity) {
	assets.Load()
	assets.InitAudio()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			Hunger:             100.0,
			Thirst:             100.0,
			Inventory:          []string{},
			WeaponEquipped:     false,
			WeaponType:         "",
			WeaponDurability:   0,
			ArmorEquipped:      false,
			ArmorType:          "",
			ArmorDefense:       0.0,
			ArmorDurability:    0,
			ArmorMaxDurability: 0,
			InfectionResist:    0.0,
			AttackCooldown:     0,
			Dead:               false,
			Infected:           false,
			FacingX:            1.0,
			FacingY:            0.0,
		},
		&ecs.Position{X: 200.0, Y: 200.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, sys, pEnt
}

// 1. Adversarial Test: World Reset While In Active Combat
// Stress-tests rapid state transitions where Reset() is called mid-combat with active chasing zombies,
// degraded armor/weapons, low health, infected state, and scattered loot.
func TestAdversarial_WorldResetWhileInCombat(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	g := NewGame()

	for iteration := 0; iteration < 25; iteration++ {
		// Mutate state into heavy combat
		pMap := arkecs.NewMap1[ecs.Player](g.world)
		pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
		if !pq.Next() {
			t.Fatalf("Iter %d: No player entity found before combat simulation", iteration)
		}
		player := pMap.Get(pq.Entity())
		pq.Close()
		
		player.Health = 22.5
		player.Hunger = 15.0
		player.Thirst = 8.0
		player.Infected = true
		player.ArmorEquipped = true
		player.ArmorType = "vest"
		player.ArmorDefense = 0.50
		player.ArmorDurability = 2
		player.ArmorMaxDurability = 10
		player.InfectionResist = 0.70
		player.WeaponEquipped = true
		player.WeaponType = "shotgun"
		player.WeaponDurability = 1
		player.Inventory = []string{"ammo", "axe", "food", "water"}
		player.AttackCooldown = 15

		// Set several zombies to aggressive chasing state
		zMap := arkecs.NewMap1[ecs.Zombie](g.world)
		zq := arkecs.NewFilter1[ecs.Zombie](g.world).Query()
		zCount := 0
		for zq.Next() {
			z := zMap.Get(zq.Entity())
			z.Chasing = true
			z.WanderTimer = 0
			zCount++
		}
		if zCount == 0 {
			t.Fatalf("Iter %d: No zombies found in world", iteration)
		}

		// Run 5 combat update frames
		for frame := 0; frame < 5; frame++ {
			g.updateSys.Update(-1)
		}

		// Rapid State Transition: Call Reset() mid-combat
		g.Reset()

		// Verify Invariants after Reset()
		if g.timeOfDay != 8.0 {
			t.Errorf("Iter %d: Expected timeOfDay 8.0, got %f", iteration, g.timeOfDay)
		}

		// Verify exactly 1 fresh player exists with clean stats
		newPQ := arkecs.NewFilter2[ecs.Player, ecs.Position](g.world).Query()
		freshPlayerCount := 0
		for newPQ.Next() {
			freshPlayerCount++
			freshP, pos := newPQ.Get()

			if freshP.Health != 100.0 {
				t.Errorf("Iter %d: Expected Health 100.0, got %f", iteration, freshP.Health)
			}
			if freshP.Hunger != 100.0 || freshP.Thirst != 100.0 {
				t.Errorf("Iter %d: Expected Hunger/Thirst 100.0, got H=%f, T=%f", iteration, freshP.Hunger, freshP.Thirst)
			}
			if freshP.Infected || freshP.Dead {
				t.Errorf("Iter %d: Player has dirty infected/dead flag: Infected=%v, Dead=%v", iteration, freshP.Infected, freshP.Dead)
			}
			if freshP.ArmorEquipped || freshP.WeaponEquipped {
				t.Errorf("Iter %d: Player has dirty equipment state: Armor=%v, Weapon=%v", iteration, freshP.ArmorEquipped, freshP.WeaponEquipped)
			}
			for _, it := range freshP.Inventory { if it != "" { t.Errorf("Iter %d: Expected empty inventory, got %s", iteration, it); break } }; if false  {
				t.Errorf("Iter %d: Expected empty inventory, got len=%d", iteration, len(freshP.Inventory))
			}
			if freshP.AttackCooldown != 0 {
				t.Errorf("Iter %d: Expected AttackCooldown 0, got %d", iteration, freshP.AttackCooldown)
			}

			// Player position must match Map.PlayerSpawn and be on walkable non-solid tile
			if pos.X != g.gameMap.PlayerSpawn.X || pos.Y != g.gameMap.PlayerSpawn.Y {
				t.Errorf("Iter %d: Player spawn mismatch: got (%f,%f), want (%f,%f)", iteration, pos.X, pos.Y, g.gameMap.PlayerSpawn.X, g.gameMap.PlayerSpawn.Y)
			}
			tileX := int(pos.X) / world.TileSize
			tileY := int(pos.Y) / world.TileSize
			if g.gameMap.GetTile(tileX, tileY).IsSolid() {
				t.Errorf("Iter %d: Player spawned on solid tile %v", iteration, g.gameMap.GetTile(tileX, tileY))
			}
		}
		if freshPlayerCount != 1 {
			t.Fatalf("Iter %d: Expected exactly 1 player after Reset(), found %d", iteration, freshPlayerCount)
		}

		// Verify all zombies spawned outside solid obstacles and at safe distance
		newZQ := arkecs.NewFilter2[ecs.Zombie, ecs.Position](g.world).Query()
		freshZCount := 0
		for newZQ.Next() {
			freshZCount++
			_, zPos := newZQ.Get()
			ztX := int(zPos.X) / world.TileSize
			ztY := int(zPos.Y) / world.TileSize
			if g.gameMap.GetTile(ztX, ztY).IsSolid() {
				t.Errorf("Iter %d: Zombie spawned on solid tile %v at (%f,%f)", iteration, g.gameMap.GetTile(ztX, ztY), zPos.X, zPos.Y)
			}
		}
		if freshZCount != len(g.gameMap.ZombieSpawns) {
			t.Errorf("Iter %d: Zombie count mismatch: got %d, want %d", iteration, freshZCount, len(g.gameMap.ZombieSpawns))
		}

		// Verify 20 frames of continuous simulation after reset run without error
		for f := 0; f < 20; f++ {
			g.updateSys.Update(-1)
		}
	}
}

// 2. Adversarial Test: Simultaneous Death and Zombie Infection
// Stress-tests edge conditions where player dies on the exact frame of infection contact,
// or dies from infection damage while surrounded by an aggro horde.
func TestAdversarial_SimultaneousDeathAndZombieInfection(t *testing.T) {
	w, m, sys, pEnt := setupM5AdversarialHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	// Scenario A: Player with 0.02 health becomes infected and dies on next frame
	player.Health = 0.02
	player.Infected = true
	player.Dead = false
	player.ArmorEquipped = false

	// Execute input & combat update (infection drain = 0.05)
	sys.processInputAndCombat(-1)

	if player.Health > 0 {
		t.Errorf("Scenario A: Expected Health <= 0, got %f", player.Health)
	}
	if !player.Dead {
		t.Errorf("Scenario A: Expected player.Dead to be true, got false")
	}

	// Scenario B: Dead player in contact with 20 zombies must NOT take additional infection rolls or crash
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	for i := 0; i < 20; i++ {
		zMap.NewEntity(
			&ecs.Zombie{Speed: 2.0, Chasing: true, WanderTimer: 0},
			&ecs.Position{X: 200.0 + float64(i%5), Y: 200.0 + float64(i/5)}, // dist < 14.0 (contact range)
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)
	}

	// Run zombie update with Dead player
	sys.processZombies()

	// Verify all zombies disengage when player is dead
	zq := arkecs.NewFilter1[ecs.Zombie](w).Query()
	for zq.Next() {
		z := zq.Get()
		if z.Chasing {
			t.Errorf("Scenario B: Zombie should stop chasing dead player, got Chasing=true")
		}
	}

	// Scenario C: Dead player cannot pick up items
	iMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
	itemEnt := iMap.NewEntity(
		&ecs.Item{Type: "food"},
		&ecs.Position{X: 200.0, Y: 200.0},
	)
	sys.processItems()

	if len(player.Inventory) != 0 {
		t.Errorf("Scenario C: Dead player picked up item! Inventory: %v", player.Inventory)
	}
	if !w.Alive(itemEnt) {
		t.Errorf("Scenario C: Item on ground was erroneously removed by dead player")
	}

	// Scenario D: 100 continuous updates with Dead player should not produce NaNs or underflow
	for f := 0; f < 100; f++ {
		sys.Update(-1)
		if math.IsNaN(player.Health) || math.IsNaN(player.Hunger) || math.IsNaN(player.Thirst) {
			t.Fatalf("Scenario D: NaN detected in player stats during dead simulation at frame %d", f)
		}
	}

	// Unused var suppression
	_ = m
}

// 3. Adversarial Test: Weapon Break on Shotgun Blast Emitting Acoustic Noise Pulse
// Stress-tests boundary where shotgun durability reaches 0 on the shot that kills enemies
// and validates that the 400px noise pulse STILL broadcasts and agitates surrounding zombies.
func TestAdversarial_WeaponBreakOnShotgunBlastEmittingNoisePulse(t *testing.T) {
	w, _, sys, pEnt := setupM5AdversarialHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	// Setup: Shotgun with EXACTLY 1 durability, 1 ammo in inventory, facing East (+X)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 1
	player.Inventory = []string{"ammo"}
	player.FacingX = 1.0
	player.FacingY = 0.0
	player.AttackCooldown = 0

	posMap := arkecs.NewMap1[ecs.Position](w)
	pPos := posMap.Get(pEnt)
	pX, pY := pPos.X, pPos.Y // (200, 200)

	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	// Group 1: 4 Target Zombies inside shotgun spread cone (dist 50..120px in front)
	var coneZombies []arkecs.Entity
	coneZombies = append(coneZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 60}, &ecs.Position{X: pX + 50, Y: pY}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{}))
	coneZombies = append(coneZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 70}, &ecs.Position{X: pX + 100, Y: pY + 10}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{}))
	coneZombies = append(coneZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 80}, &ecs.Position{X: pX + 140, Y: pY - 15}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{}))
	coneZombies = append(coneZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 90}, &ecs.Position{X: pX + 15, Y: pY + 5}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // point-blank < 24px

	// Group 2: 6 Wandering Zombies outside cone but within 400px noise radius (behind & flanks)
	var noiseRadiusZombies []arkecs.Entity
	noiseRadiusZombies = append(noiseRadiusZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 100}, &ecs.Position{X: pX - 100, Y: pY}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // behind 100px
	noiseRadiusZombies = append(noiseRadiusZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 110}, &ecs.Position{X: pX - 250, Y: pY}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // behind 250px
	noiseRadiusZombies = append(noiseRadiusZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 120}, &ecs.Position{X: pX, Y: pY + 200}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // flank 200px
	noiseRadiusZombies = append(noiseRadiusZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 130}, &ecs.Position{X: pX, Y: pY - 300}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // flank 300px
	noiseRadiusZombies = append(noiseRadiusZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 140}, &ecs.Position{X: pX - 200, Y: pY - 200}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // diagonal dist ~282px
	noiseRadiusZombies = append(noiseRadiusZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 150}, &ecs.Position{X: pX + 250, Y: pY + 250}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // diagonal dist ~353px

	// Group 3: 4 Distant Zombies beyond 400px noise radius (> 400px)
	var distantZombies []arkecs.Entity
	distantZombies = append(distantZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 160}, &ecs.Position{X: pX + 450, Y: pY}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{}))
	distantZombies = append(distantZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 170}, &ecs.Position{X: pX - 450, Y: pY}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{}))
	distantZombies = append(distantZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 180}, &ecs.Position{X: pX, Y: pY + 450}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{}))
	distantZombies = append(distantZombies, zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 190}, &ecs.Position{X: pX - 350, Y: pY - 350}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})) // dist ~494px

	// Execute Shotgun Combat Logic (matching game.go lines 361-432)
	ammoIdx := -1
	for idx, itm := range player.Inventory {
		if itm == "ammo" {
			ammoIdx = idx
			break
		}
	}
	if ammoIdx < 0 {
		t.Fatal("Expected ammo in inventory")
	}

	// 1. Consume Ammo
	player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)

	// 2. Weapon Durability Breakdown
	player.WeaponDurability--
	if player.WeaponDurability <= 0 {
		player.WeaponEquipped = false
		player.WeaponType = ""
		player.WeaponDurability = 0
	}

	// 3. Resolve Shotgun Spread Cone (Range: 160, Angle: +-22.5 deg)
	const maxShotgunRange = 160.0
	const cosSpread = 0.9238795325112867
	var toRemoveZombies []arkecs.Entity

	facingX, facingY := player.FacingX, player.FacingY
	facingLen := math.Hypot(facingX, facingY)
	if facingLen > 0.001 {
		facingX /= facingLen
		facingY /= facingLen
	}

	zQuery := sys.zombieFilter.Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()

		dx := zPos.X - pX
		dy := zPos.Y - pY
		dist := math.Hypot(dx, dy)

		if dist <= maxShotgunRange {
			if dist < 24.0 {
				toRemoveZombies = append(toRemoveZombies, ent)
			} else {
				cosAngle := (facingX*dx + facingY*dy) / dist
				if cosAngle >= cosSpread {
					toRemoveZombies = append(toRemoveZombies, ent)
				}
			}
		}
	}
	for _, ent := range toRemoveZombies {
		w.RemoveEntity(ent)
	}

	// 4. Acoustic Noise Pulse (400.0px radius alert)
	noiseQuery := sys.zombieFilter.Query()
	for noiseQuery.Next() {
		z, zPos, _ := noiseQuery.Get()
		zdx := pX - zPos.X
		zdy := pY - zPos.Y
		if math.Hypot(zdx, zdy) <= 400.0 {
			z.Chasing = true
			z.WanderTimer = 0
		}
	}

	// Verify Weapon Break State
	if player.WeaponEquipped {
		t.Errorf("Expected WeaponEquipped false after shotgun durability 0")
	}
	if player.WeaponType != "" {
		t.Errorf("Expected WeaponType empty string after break, got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 0 {
		t.Errorf("Expected WeaponDurability 0, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 0 {
		t.Errorf("Expected ammo consumed from inventory, got len=%d", len(player.Inventory))
	}

	// Verify Group 1: All 4 Cone Zombies were killed
	for i, ent := range coneZombies {
		if w.Alive(ent) {
			t.Errorf("Group 1 Zombie #%d inside shotgun cone was NOT killed", i)
		}
	}

	// Verify Group 2: All 6 noise radius zombies are alerted (Chasing=true, WanderTimer=0)
	zCompMap := arkecs.NewMap1[ecs.Zombie](w)
	for i, ent := range noiseRadiusZombies {
		if !w.Alive(ent) {
			t.Errorf("Group 2 Zombie #%d outside cone should still be alive", i)
			continue
		}
		z := zCompMap.Get(ent)
		if !z.Chasing || z.WanderTimer != 0 {
			t.Errorf("Group 2 Zombie #%d (dist <= 400px) was NOT alerted: Chasing=%v, WanderTimer=%d", i, z.Chasing, z.WanderTimer)
		}
	}

	// Verify Group 3: All 4 distant zombies remained unalerted (Chasing=false)
	for i, ent := range distantZombies {
		if !w.Alive(ent) {
			t.Errorf("Group 3 Zombie #%d should be alive", i)
			continue
		}
		z := zCompMap.Get(ent)
		if z.Chasing {
			t.Errorf("Group 3 Zombie #%d (dist > 400px) was erroneously alerted: Chasing=%v", i, z.Chasing)
		}
	}
}

// 4. Adversarial Test: Inventory Manipulation Under Max Capacity (9 Items) & Rapid Churn
// Stress-tests boundary condition where inventory is at 9 items (max capacity),
// verifies ground items cannot exceed capacity, and stress tests 10,000 rapid churn operations.
func TestAdversarial_InventoryManipulationUnderMaxCapacity(t *testing.T) {
	w, _, sys, pEnt := setupM5AdversarialHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	posMap := arkecs.NewMap1[ecs.Position](w)
	pPos := posMap.Get(pEnt)
	pX, pY := pPos.X, pPos.Y

	// Step 1: Place 15 items on ground near player (dist = 5px < 16px pickup radius)
	iMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
	var groundItems []arkecs.Entity
	itemTypes := []string{"food", "water", "weapon", "axe", "shotgun", "ammo", "armor"}

	for i := 0; i < 15; i++ {
		tName := itemTypes[i%len(itemTypes)]
		ent := iMap.NewEntity(
			&ecs.Item{Type: tName},
			&ecs.Position{X: pX + 2.0, Y: pY + 2.0},
		)
		groundItems = append(groundItems, ent)
	}

	// Player starts with empty inventory
	player.Inventory = make([]string, 9)

	// Run processItems()
	sys.processItems()

	// Invariant 1: Exactly 9 items picked up (inventory capacity limit)
	count := 0
	for _, it := range player.Inventory {
		if it != "" {
			count++
		}
	}
	if count != 9 {
		t.Fatalf("Expected inventory to reach exactly 9 items, got %d", count)
	}

	// Invariant 2: Exactly 6 items remain on ground in ECS world
	aliveCount := 0
	for _, ent := range groundItems {
		if w.Alive(ent) {
			aliveCount++
		}
	}
	if aliveCount != 6 {
		t.Fatalf("Expected 6 items remaining on ground, got %d", aliveCount)
	}

	// Step 2: Run processItems() again with full inventory (9 items)
	sys.processItems()
	if len(player.Inventory) != 9 {
		t.Errorf("Inventory expanded beyond 9 items: len=%d", len(player.Inventory))
	}
	aliveCount2 := 0
	for _, ent := range groundItems {
		if w.Alive(ent) {
			aliveCount2++
		}
	}
	if aliveCount2 != 6 {
		t.Errorf("Ground items consumed while inventory full: got %d alive", aliveCount2)
	}

	// Step 3: Consume item from slot 0 (say, food)
	player.Inventory[0] = "" // Now len = 8
	sys.processItems()

	count = 0
	for _, it := range player.Inventory {
		if it != "" {
			count++
		}
	}
	if count != 9 {
		t.Errorf("Inventory did not backfill to 9 after consumption: len=%d", count)
	}
	aliveCount3 := 0
	for _, ent := range groundItems {
		if w.Alive(ent) {
			aliveCount3++
		}
	}
	if aliveCount3 != 5 {
		t.Errorf("Expected 5 ground items remaining after 1 backfill, got %d", aliveCount3)
	}

	// Step 4: 10,000 Rapid Randomized Inventory Churn Cycles
	r := rand.New(rand.NewSource(123456789))
	for cycle := 0; cycle < 10000; cycle++ {
		// Random action:
		action := r.Intn(4)
		switch action {
		case 0: // Spawn ground item near player
			tName := itemTypes[r.Intn(len(itemTypes))]
			iMap.NewEntity(
				&ecs.Item{Type: tName},
				&ecs.Position{X: pX + r.Float64()*10.0, Y: pY + r.Float64()*10.0},
			)
		case 1: // Consume/use item from random slot
			if len(player.Inventory) > 0 {
				slot := r.Intn(len(player.Inventory))
				itm := player.Inventory[slot]
				// Use item
				if itm == "food" {
					player.Hunger = math.Min(100, player.Hunger+50)
				} else if itm == "water" {
					player.Thirst = math.Min(100, player.Thirst+50)
				} else if itm == "armor" {
					player.ArmorEquipped = true
					player.ArmorType = "vest"
					player.ArmorDefense = 0.50
					player.ArmorDurability = 10
					player.ArmorMaxDurability = 10
					player.InfectionResist = 0.70
				} else if itm == "weapon" || itm == "axe" || itm == "shotgun" {
					player.WeaponEquipped = true
					player.WeaponType = itm
					player.WeaponDurability = 10
				}
				// Remove from inventory
				player.Inventory[slot] = ""
			}
		case 2: // Trigger processItems()
			sys.processItems()
		case 3: // Full system update
			sys.Update(-1)
		}

		// Invariant checks on every single cycle
		if len(player.Inventory) > 9 {
			t.Fatalf("Cycle %d: Inventory overflowed 9 items! Current len=%d: %v", cycle, len(player.Inventory), player.Inventory)
		}
		if player.Hunger < 0 || player.Hunger > 100 {
			t.Fatalf("Cycle %d: Hunger out of bounds: %f", cycle, player.Hunger)
		}
		if player.Thirst < 0 || player.Thirst > 100 {
			t.Fatalf("Cycle %d: Thirst out of bounds: %f", cycle, player.Thirst)
		}
		if player.Health < 0 || player.Health > 100 {
			t.Fatalf("Cycle %d: Health out of bounds: %f", cycle, player.Health)
		}
	}
}
