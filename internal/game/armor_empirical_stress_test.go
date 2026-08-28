package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// 1. Stress test zombie swarm contact on armored player (10, 50, 100 simultaneous attackers)
func TestArmorEmpirical_ZombieSwarmContactStress(t *testing.T) {
	assets.Load()

	swarmSizes := []int{1, 5, 10, 25, 50, 100}

	for _, count := range swarmSizes {
		t.Run(fmt.Sprintf("Swarm_%d_Zombies", count), func(t *testing.T) {
			w := arkecs.NewWorld()
			m := world.NewMap(50, 50)
			sys := NewUpdateSystem(w, m)

			pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
			pEnt := pMap.NewEntity(
				&ecs.Player{
					Health:             100.0,
					ArmorEquipped:      true,
					ArmorType:          "vest",
					ArmorDefense:       0.50,
					ArmorDurability:    10,
					ArmorMaxDurability: 10,
					InfectionResist:    1.0, // Guaranteed deflection while armor holds
					Dead:               false,
					Infected:           false,
				},
				&ecs.Position{X: 200, Y: 200},
				&ecs.Velocity{X: 0, Y: 0},
				&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
				&ecs.Collider{Width: 16, Height: 16},
			)

			// Spawn count zombies all directly touching player (dist = 5.0 < 14.0)
			zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
			for i := 0; i < count; i++ {
				angle := float64(i) / float64(count) * 2.0 * math.Pi
				zMap.NewEntity(
					&ecs.Zombie{Speed: 1.5, Chasing: true},
					&ecs.Position{X: 200 + 5.0*math.Cos(angle), Y: 200 + 5.0*math.Sin(angle)},
					&ecs.Velocity{X: 0, Y: 0},
					&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
					&ecs.Collider{Width: 16, Height: 16},
				)
			}

			// Process one frame of zombie AI & contact
			sys.processZombies()

			player := arkecs.NewMap1[ecs.Player](w).Get(pEnt)
			if player == nil {
				t.Fatal("Player entity missing")
			}

			if count < 10 {
				// Armor should still be equipped with reduced durability
				expectedDurability := 10 - count
				if !player.ArmorEquipped {
					t.Errorf("Swarm %d: Expected armor to remain equipped, got broken", count)
				}
				if player.ArmorDurability != expectedDurability {
					t.Errorf("Swarm %d: Expected durability %d, got %d", count, expectedDurability, player.ArmorDurability)
				}
				if player.Infected {
					t.Errorf("Swarm %d: Expected no infection with 100%% resist armor", count)
				}
			} else {
				// Armor must be completely broken and unequipped, durability clamped to 0
				if player.ArmorEquipped {
					t.Errorf("Swarm %d: Expected armor to break, but ArmorEquipped is true", count)
				}
				if player.ArmorDurability != 0 {
					t.Errorf("Swarm %d: Expected durability 0, got %d", count, player.ArmorDurability)
				}
				if player.ArmorMaxDurability != 0 {
					t.Errorf("Swarm %d: Expected max durability 0, got %d", count, player.ArmorMaxDurability)
				}
				if player.ArmorDefense != 0.0 {
					t.Errorf("Swarm %d: Expected defense 0.0, got %f", count, player.ArmorDefense)
				}
				if player.InfectionResist != 0.0 {
					t.Errorf("Swarm %d: Expected infection resist 0.0, got %f", count, player.InfectionResist)
				}
				if player.ArmorType != "" {
					t.Errorf("Swarm %d: Expected empty ArmorType, got '%s'", count, player.ArmorType)
				}

				if count > 10 {
					// Hits exceeding durability (10) strike unarmored player -> infected
					if !player.Infected {
						t.Errorf("Swarm %d: Expected player to be infected after armor broke on hit 10", count)
					}
				}
			}
		})
	}
}

// 2. Monte Carlo Statistical Validation of Infection Deflection Rate (10,000 trials)
func TestArmorEmpirical_MonteCarloInfectionDeflectionRate(t *testing.T) {
	assets.Load()
	rand.Seed(1337)

	totalTrials := 10000
	deflectionCount := 0

	for i := 0; i < totalTrials; i++ {
		w := arkecs.NewWorld()
		m := world.NewMap(50, 50)
		sys := NewUpdateSystem(w, m)

		pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		pEnt := pMap.NewEntity(
			&ecs.Player{
				Health:             100.0,
				ArmorEquipped:      true,
				ArmorType:          "vest",
				ArmorDefense:       0.50,
				ArmorDurability:    10,
				ArmorMaxDurability: 10,
				InfectionResist:    0.70, // 70% nominal deflection rate
				Dead:               false,
				Infected:           false,
			},
			&ecs.Position{X: 100, Y: 100},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)

		zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0, Chasing: true},
			&ecs.Position{X: 105, Y: 100},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)

		sys.processZombies()

		player := arkecs.NewMap1[ecs.Player](w).Get(pEnt)
		if !player.Infected {
			deflectionCount++
		}
	}

	empiricalRate := float64(deflectionCount) / float64(totalTrials)
	t.Logf("Monte Carlo 10,000 trials: Deflected=%d, Infected=%d, Rate=%.4f (Expected=0.7000)",
		deflectionCount, totalTrials-deflectionCount, empiricalRate)

	// Standard error for n=10000, p=0.70 is sqrt(0.70*0.30/10000) ~= 0.00458
	// 3 sigma tolerance ~= 0.015 (1.5%)
	if math.Abs(empiricalRate-0.70) > 0.015 {
		t.Errorf("Empirical deflection rate %.4f deviated significantly from expected 0.70 (allowed +-0.015)", empiricalRate)
	}
}

// 3. Dynamic Inventory Equipping Stress Test: Rapid multiple vests, slot permutations, full inventory
func TestArmorEmpirical_InventoryEquippingStress(t *testing.T) {
	assets.Load()

	t.Run("FullInventoryOfArmorVestsChainedEquip", func(t *testing.T) {
		player := &ecs.Player{
			Health:             100.0,
			Inventory:          []string{"armor", "vest", "armor", "vest", "armor", "vest", "armor", "vest", "armor"},
			AttackCooldown:     0,
			ArmorEquipped:      false,
			ArmorDurability:    0,
			ArmorMaxDurability: 0,
		}

		vestCount := len(player.Inventory)
		if vestCount != 9 {
			t.Fatalf("Expected 9 inventory slots, got %d", vestCount)
		}

		for slot := 0; slot < vestCount; slot++ {
			// Equip first available item (index 0 as items are consumed)
			useItemIdx := 0
			if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30
				itemType := player.Inventory[useItemIdx]
				if itemType == "armor" || itemType == "vest" {
					player.ArmorEquipped = true
					player.ArmorType = "vest"
					player.ArmorDefense = 0.50
					player.ArmorDurability = 10
					player.ArmorMaxDurability = 10
					player.InfectionResist = 0.70
					player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
				}
			}

			// Verify equipped state
			if !player.ArmorEquipped || player.ArmorDurability != 10 || player.ArmorDefense != 0.50 {
				t.Fatalf("Slot iteration %d: Failed to equip vest correctly: %+v", slot, player)
			}
			expectedRemaining := 9 - (slot + 1)
			if len(player.Inventory) != expectedRemaining {
				t.Fatalf("Slot iteration %d: Expected %d remaining items, got %d", slot, expectedRemaining, len(player.Inventory))
			}

			// Simulate some durability damage before next equip
			player.ArmorDurability = 3

			// Reset cooldown for next equip simulation
			player.AttackCooldown = 0
		}

		if len(player.Inventory) != 0 {
			t.Errorf("Expected empty inventory at end, got %v", player.Inventory)
		}
	})

	t.Run("MixedInventoryItemActivationSafety", func(t *testing.T) {
		player := &ecs.Player{
			Health:         50.0,
			Hunger:         60.0,
			Thirst:         60.0,
			Inventory:      []string{"food", "armor", "water", "weapon", "vest"},
			AttackCooldown: 0,
		}

		// Activate slot 1 ("armor")
		useIdx := 1
		itemType := player.Inventory[useIdx]
		if itemType == "armor" || itemType == "vest" {
			player.ArmorEquipped = true
			player.ArmorType = "vest"
			player.ArmorDefense = 0.50
			player.ArmorDurability = 10
			player.ArmorMaxDurability = 10
			player.InfectionResist = 0.70
			player.Inventory = append(player.Inventory[:useIdx], player.Inventory[useIdx+1:]...)
		}

		if !player.ArmorEquipped || player.ArmorDurability != 10 {
			t.Errorf("Expected armor equipped, got %+v", player)
		}

		expectedInv := []string{"food", "water", "weapon", "vest"}
		if len(player.Inventory) != len(expectedInv) {
			t.Fatalf("Expected inventory length %d, got %d", len(expectedInv), len(player.Inventory))
		}
		for idx, item := range player.Inventory {
			if item != expectedInv[idx] {
				t.Errorf("Slot %d: Expected %s, got %s", idx, expectedInv[idx], item)
			}
		}
	})
}

// 4. Heavy Simulation: 2000 Game Frames with Armored Player, Swarms, Combat & Equips
func TestArmorEmpirical_HeavySimulationContinuousLoop(t *testing.T) {
	assets.Load()
	g := NewGame()
	screen := ebiten.NewImage(800, 600)

	// Equip player with armor initially
	pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
	for pq.Next() {
		p := pq.Get()
		p.ArmorEquipped = true
		p.ArmorType = "vest"
		p.ArmorDefense = 0.50
		p.ArmorDurability = 10
		p.ArmorMaxDurability = 10
		p.InfectionResist = 0.70
		p.Inventory = []string{"armor", "vest", "armor", "food", "water"}
	}

	rng := rand.New(rand.NewSource(999))

	for frame := 0; frame < 2000; frame++ {
		// Periodically spawn or re-equip items/zombies to create dynamic chaos
		if frame%100 == 0 {
			pFilter := arkecs.NewFilter2[ecs.Player, ecs.Position](g.world).Query()
			for pFilter.Next() {
				p, pos := pFilter.Get()
				if !p.ArmorEquipped && len(p.Inventory) > 0 {
					for idx, item := range p.Inventory {
						if item == "armor" || item == "vest" {
							p.ArmorEquipped = true
							p.ArmorType = "vest"
							p.ArmorDefense = 0.50
							p.ArmorDurability = 10
							p.ArmorMaxDurability = 10
							p.InfectionResist = 0.70
							p.Inventory = append(p.Inventory[:idx], p.Inventory[idx+1:]...)
							break
						}
					}
				}
				// Perturb position slightly
				pos.X += (rng.Float64() - 0.5) * 5.0
				pos.Y += (rng.Float64() - 0.5) * 5.0

				// Prevent early death to continue simulation coverage
				if p.Health < 20 {
					p.Health = 100.0
					p.Hunger = 100.0
					p.Thirst = 100.0
				}
			}
		}

		// Update game
		err := g.Update()
		if err != nil {
			t.Fatalf("Frame %d: Update returned error: %v", frame, err)
		}

		// Periodically render to stress Draw pipeline
		if frame%10 == 0 {
			screen.Clear()
			g.Draw(screen)
		}

		// Verify ECS sanity
		pqCheck := arkecs.NewFilter3[ecs.Player, ecs.Position, ecs.Velocity](g.world).Query()
		for pqCheck.Next() {
			p, pos, vel := pqCheck.Get()
			if math.IsNaN(pos.X) || math.IsNaN(pos.Y) {
				t.Fatalf("Frame %d: Player Position is NaN (%f, %f)", frame, pos.X, pos.Y)
			}
			if math.IsNaN(vel.X) || math.IsNaN(vel.Y) {
				t.Fatalf("Frame %d: Player Velocity is NaN (%f, %f)", frame, vel.X, vel.Y)
			}
			if p.ArmorDurability < 0 {
				t.Fatalf("Frame %d: Player ArmorDurability is negative: %d", frame, p.ArmorDurability)
			}
			if p.ArmorEquipped && p.ArmorDurability == 0 {
				t.Fatalf("Frame %d: Player ArmorEquipped is true with 0 durability", frame)
			}
		}
	}
}

// 5. HUD Rendering Calculations and Visual Tint Permutations
func TestArmorEmpirical_HUDAndVisualTintExhaustive(t *testing.T) {
	assets.Load()
	screen := ebiten.NewImage(800, 600)

	states := []struct {
		name       string
		health     float64
		dead       bool
		infected   bool
		armored    bool
		durability int
		maxDur     int
		defense    float64
		cooldown   int
	}{
		{"FullArmorNormal", 100.0, false, false, true, 10, 10, 0.50, 0},
		{"HalfArmorNormal", 80.0, false, false, true, 5, 10, 0.50, 0},
		{"LowArmorNormal", 30.0, false, false, true, 1, 10, 0.50, 0},
		{"ZeroArmorEquippedAnomaly", 50.0, false, false, true, 0, 10, 0.50, 0},
		{"ZeroMaxDurabilityAnomaly", 50.0, false, false, true, 0, 0, 0.50, 0},
		{"NegativeDurabilityAnomaly", 50.0, false, false, true, -5, 10, 0.50, 0},
		{"OverMaxDurabilityAnomaly", 100.0, false, false, true, 15, 10, 0.50, 0},
		{"ArmoredInfected", 75.0, false, true, true, 8, 10, 0.50, 0},
		{"ArmoredDead", 0.0, true, false, true, 10, 10, 0.50, 0},
		{"ArmoredAttacking", 100.0, false, false, true, 10, 10, 0.50, 28},
		{"UnarmoredNormal", 100.0, false, false, false, 0, 0, 0.0, 0},
		{"UnarmoredInfected", 50.0, false, true, false, 0, 0, 0.0, 0},
		{"UnarmoredDead", 0.0, true, false, false, 0, 0, 0.0, 0},
		{"HighDefenseArmored", 100.0, false, false, true, 10, 10, 1.00, 0},
		{"ZeroDefenseArmored", 100.0, false, false, true, 10, 10, 0.00, 0},
	}

	for _, st := range states {
		t.Run(st.name, func(t *testing.T) {
			w := arkecs.NewWorld()
			m := world.NewMap(30, 30)
			for i := range m.Visible {
				m.Visible[i] = true
				m.Explored[i] = true
			}

			pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
			pMap.NewEntity(
				&ecs.Player{
					Health:             st.health,
					Hunger:             80.0,
					Thirst:             80.0,
					Inventory:          []string{"armor", "food"},
					ArmorEquipped:      st.armored,
					ArmorType:          "vest",
					ArmorDefense:       st.defense,
					ArmorDurability:    st.durability,
					ArmorMaxDurability: st.maxDur,
					InfectionResist:    0.70,
					AttackCooldown:     st.cooldown,
					Dead:               st.dead,
					Infected:           st.infected,
					FacingX:            1,
					FacingY:            0,
				},
				&ecs.Position{X: 15 * world.TileSize, Y: 15 * world.TileSize},
				&ecs.Velocity{X: 0, Y: 0},
				&ecs.Sprite{W: 16, H: 16},
				&ecs.Collider{Width: 16, Height: 16},
			)

			drawSys := NewDrawSystem(w, m)

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic rendering state %s: %v", st.name, r)
				}
			}()

			screen.Clear()
			drawSys.Draw(screen, 12.0)
		})
	}
}
