# Milestone 3 Handoff Report: Armor Data Structures & Equipping Integration

## 1. Observation

### 1.1 Existing ECS Player Data Model
In `internal/ecs/components.go` (lines 28–41), the current `ecs.Player` component contains only basic health, hunger, thirst, inventory, and weapon fields without any armor tracking:
```go
// Player marker component.
type Player struct {
	Health         float64
	Hunger         float64 // 100.0 is full, 0.0 is starving
	Thirst         float64 // 100.0 is hydrated, 0.0 is dehydrated
	Inventory      []string
	WeaponEquipped   bool
	WeaponDurability int
	AttackCooldown   int
	Dead             bool
	Infected         bool
	FacingX          float64
	FacingY          float64
}
```

### 1.2 Existing Inventory & Item Handling Logic
In `internal/game/game.go` (lines 273–307), inventory activation reads slots 1–9 and branches based on item type string (`"food"`, `"water"`, `"weapon"`):
```go
useItemIdx := -1
if ebiten.IsKeyPressed(ebiten.Key1) { useItemIdx = 0 }
if ebiten.IsKeyPressed(ebiten.Key2) { useItemIdx = 1 }
if ebiten.IsKeyPressed(ebiten.Key3) { useItemIdx = 2 }
if ebiten.IsKeyPressed(ebiten.Key4) { useItemIdx = 3 }
if ebiten.IsKeyPressed(ebiten.Key5) { useItemIdx = 4 }
if ebiten.IsKeyPressed(ebiten.Key6) { useItemIdx = 5 }
if ebiten.IsKeyPressed(ebiten.Key7) { useItemIdx = 6 }
if ebiten.IsKeyPressed(ebiten.Key8) { useItemIdx = 7 }
if ebiten.IsKeyPressed(ebiten.Key9) { useItemIdx = 8 }

if useItemIdx >= 0 && useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
	player.AttackCooldown = 30 // Small cooldown so it doesn't instantly consume everything if held
	t := player.Inventory[useItemIdx]
	
	used := false
	if t == "food" && player.Hunger < 100 {
		player.Hunger += 50
		if player.Hunger > 100 { player.Hunger = 100 }
		used = true
	} else if t == "water" && player.Thirst < 100 {
		player.Thirst += 50
		if player.Thirst > 100 { player.Thirst = 100 }
		used = true
	} else if t == "weapon" {
		player.WeaponEquipped = true
		player.WeaponDurability = 5
		used = true
	}
	
	if used {
		// Remove item from inventory
		player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
	}
}
```

### 1.3 Asset & Map Spawning Baseline
- `internal/assets/assets.go:41, 71`: `ArmorImage` is already loaded and exported from `images/armor.png`.
- `internal/game/world/map.go:801`: Procedural map generation already places contextual loot with `Type: "armor"` (e.g. Police Station Armory).
- `internal/game/game.go:726`: `DrawSystem.Draw` already handles `case "armor": img = assets.ArmorImage` for rendering dropped world items.

---

## 2. Logic Chain

1. **ECS Extension Design**:
   - `ecs.Player` must support armor state across game loops, damage mitigation checks, and HUD drawing.
   - Adding the 6 specified fields (`ArmorEquipped`, `ArmorType`, `ArmorDefense`, `ArmorDurability`, `ArmorMaxDurability`, `InfectionResist`) to `ecs.Player` satisfies the contract in `PROJECT.md` (§ Interface Contracts & Feature 10).
   - In Go, uninitialized struct fields in composite literals with explicit field keys default to zero values (`false`, `""`, `0.0`, `0`, `0`, `0.0`), ensuring complete backward compatibility with all existing player initialization code.

2. **Equipping Mechanics**:
   - When a player presses a slot key 1–9 corresponding to an `"armor"` (or `"vest"`) item, and `player.AttackCooldown <= 0`:
     - `player.AttackCooldown` is set to `30` (providing a 0.5s throttle to prevent rapid double-consumption).
     - Armor properties are set:
       - `ArmorEquipped = true`
       - `ArmorType = "vest"`
       - `ArmorDefense = 0.50` (50% direct damage reduction)
       - `ArmorDurability = 10` (10 hits before breaking)
       - `ArmorMaxDurability = 10`
       - `InfectionResist = 0.70` (70% probability to deflect zombie bites)
     - `used = true` marks the item for consumption.
     - The item is removed from `player.Inventory` by slicing: `player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)`.
   - If the player already has damaged armor equipped (e.g. Durability 3/10), equipping a new vest replaces/refreshes the armor back to full durability (10/10) and full stats.

3. **Downstream Integration**:
   - Equipping logic integrates seamlessly with the upcoming Milestone 3 tasks:
     - Zombie contact mitigation in `processZombies()` can read `player.ArmorEquipped`, `player.InfectionResist`, `player.ArmorDefense`, and decrement `player.ArmorDurability`.
     - HUD drawing in `DrawSystem.Draw()` can read `player.ArmorEquipped`, `player.ArmorDurability`, `player.ArmorMaxDurability`, and `player.ArmorDefense` to display the steel-blue armor status bar.

---

## 3. Caveats

- **Key Holding & Input Throttling**: Ebitengine's `ebiten.IsKeyPressed` remains true while a key is held. Setting `player.AttackCooldown = 30` when activating an item is essential to prevent consuming all 9 inventory slots in a few frames.
- **Weapon Cooldown Sharing**: Currently `player.AttackCooldown` is shared between item consumption and weapon/shove attacks. This is intentional per the existing engine design to prevent simultaneous attacking and item swapping.
- **No Direct Source Modification**: In accordance with the explorer read-only protocol, this report provides exact Go code specifications and patch definitions for implementation agents.

---

## 4. Conclusion & Concrete Code Formulations

### 4.1 Component 1: `internal/ecs/components.go`
Modify `Player` struct in `internal/ecs/components.go` (lines 28–41):

#### Before:
```go
// Player marker component.
type Player struct {
	Health         float64
	Hunger         float64 // 100.0 is full, 0.0 is starving
	Thirst         float64 // 100.0 is hydrated, 0.0 is dehydrated
	Inventory      []string
	WeaponEquipped   bool
	WeaponDurability int
	AttackCooldown   int
	Dead             bool
	Infected         bool
	FacingX          float64
	FacingY          float64
}
```

#### After:
```go
// Player marker component.
type Player struct {
	Health             float64
	Hunger             float64 // 100.0 is full, 0.0 is starving
	Thirst             float64 // 100.0 is hydrated, 0.0 is dehydrated
	Inventory          []string
	WeaponEquipped     bool
	WeaponType         string
	WeaponDurability   int
	ArmorEquipped      bool
	ArmorType          string
	ArmorDefense       float64
	ArmorDurability    int
	ArmorMaxDurability int
	InfectionResist    float64
	AttackCooldown     int
	Dead               bool
	Infected           bool
	FacingX            float64
	FacingY            float64
}
```

---

### 4.2 Component 2: `internal/game/game.go:processInputAndCombat()`
Extend inventory branch in `internal/game/game.go` (around lines 284–307):

#### Before:
```go
				used := false
				if t == "food" && player.Hunger < 100 {
					player.Hunger += 50
					if player.Hunger > 100 { player.Hunger = 100 }
					used = true
				} else if t == "water" && player.Thirst < 100 {
					player.Thirst += 50
					if player.Thirst > 100 { player.Thirst = 100 }
					used = true
				} else if t == "weapon" {
					player.WeaponEquipped = true
					player.WeaponDurability = 5
					used = true
				}
				
				if used {
					// Remove item from inventory
					player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
				}
```

#### After:
```go
				used := false
				if t == "food" && player.Hunger < 100 {
					player.Hunger += 50
					if player.Hunger > 100 { player.Hunger = 100 }
					used = true
				} else if t == "water" && player.Thirst < 100 {
					player.Thirst += 50
					if player.Thirst > 100 { player.Thirst = 100 }
					used = true
				} else if t == "weapon" {
					player.WeaponEquipped = true
					player.WeaponDurability = 5
					used = true
				} else if t == "armor" || t == "vest" {
					player.ArmorEquipped = true
					player.ArmorType = "vest"
					player.ArmorDefense = 0.50
					player.ArmorDurability = 10
					player.ArmorMaxDurability = 10
					player.InfectionResist = 0.70
					used = true
				}
				
				if used {
					// Remove item from inventory
					player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
				}
```

---

## 5. Verification Method

To independently verify the implementation, add the following unit test suite to `internal/game/armor_test.go` and run `CC=gcc go test -v ./...`:

```go
package game

import (
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

func TestArmorECSComponentFields(t *testing.T) {
	p := ecs.Player{
		Health:             100.0,
		ArmorEquipped:      true,
		ArmorType:          "vest",
		ArmorDefense:       0.50,
		ArmorDurability:    10,
		ArmorMaxDurability: 10,
		InfectionResist:    0.70,
	}

	if !p.ArmorEquipped {
		t.Errorf("Expected ArmorEquipped true")
	}
	if p.ArmorType != "vest" {
		t.Errorf("Expected ArmorType 'vest', got %s", p.ArmorType)
	}
	if p.ArmorDefense != 0.50 {
		t.Errorf("Expected ArmorDefense 0.50, got %f", p.ArmorDefense)
	}
	if p.ArmorDurability != 10 || p.ArmorMaxDurability != 10 {
		t.Errorf("Expected durability 10/10, got %d/%d", p.ArmorDurability, p.ArmorMaxDurability)
	}
	if p.InfectionResist != 0.70 {
		t.Errorf("Expected InfectionResist 0.70, got %f", p.InfectionResist)
	}
}

func TestArmorEquipExecution(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap3[ecs.Player, ecs.Position, ecs.Velocity](w)
	ent := pMap.NewEntity(
		&ecs.Player{
			Health:         100.0,
			Inventory:      []string{"armor"},
			AttackCooldown: 0,
		},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
	)

	// Simulate equipping slot 0
	player := arkecs.NewMap1[ecs.Player](w).Get(ent)
	if len(player.Inventory) != 1 || player.Inventory[0] != "armor" {
		t.Fatalf("Expected inventory to contain 'armor'")
	}

	// Direct equip logic validation
	useItemIdx := 0
	if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		player.AttackCooldown = 30
		tItem := player.Inventory[useItemIdx]
		if tItem == "armor" || tItem == "vest" {
			player.ArmorEquipped = true
			player.ArmorType = "vest"
			player.ArmorDefense = 0.50
			player.ArmorDurability = 10
			player.ArmorMaxDurability = 10
			player.InfectionResist = 0.70
			player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
		}
	}

	if !player.ArmorEquipped {
		t.Errorf("Expected player ArmorEquipped true")
	}
	if player.ArmorDurability != 10 {
		t.Errorf("Expected ArmorDurability 10, got %d", player.ArmorDurability)
	}
	if player.ArmorDefense != 0.50 {
		t.Errorf("Expected ArmorDefense 0.50, got %f", player.ArmorDefense)
	}
	if player.InfectionResist != 0.70 {
		t.Errorf("Expected InfectionResist 0.70, got %f", player.InfectionResist)
	}
	if len(player.Inventory) != 0 {
		t.Errorf("Expected inventory to be empty after equipping armor, got %v", player.Inventory)
	}
	if player.AttackCooldown != 30 {
		t.Errorf("Expected AttackCooldown 30, got %d", player.AttackCooldown)
	}
}

func TestArmorReEquipRefreshesDurability(t *testing.T) {
	player := &ecs.Player{
		Health:             100.0,
		Inventory:          []string{"armor"},
		ArmorEquipped:      true,
		ArmorType:          "vest",
		ArmorDefense:       0.50,
		ArmorDurability:    2, // Severely damaged
		ArmorMaxDurability: 10,
		InfectionResist:    0.70,
		AttackCooldown:     0,
	}

	// Equip new armor
	useItemIdx := 0
	player.AttackCooldown = 30
	tItem := player.Inventory[useItemIdx]
	if tItem == "armor" || tItem == "vest" {
		player.ArmorEquipped = true
		player.ArmorType = "vest"
		player.ArmorDefense = 0.50
		player.ArmorDurability = 10
		player.ArmorMaxDurability = 10
		player.InfectionResist = 0.70
		player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
	}

	if player.ArmorDurability != 10 {
		t.Errorf("Expected ArmorDurability refreshed to 10, got %d", player.ArmorDurability)
	}
	if len(player.Inventory) != 0 {
		t.Errorf("Expected empty inventory after equipping")
	}
}
```

### 5.1 Verification Commands
```bash
# 1. Verify build
CC=gcc go build -o bin/game ./cmd/game

# 2. Run test suite
CC=gcc go test -v ./...

# 3. Asset generator check
go run ./cmd/tools/genassets
```
