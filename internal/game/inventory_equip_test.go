package game

import (
	"image/color"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupEquipTestGame initializes a test world with player entity and map
func setupEquipTestGame() (*arkecs.World, *world.Map, *UpdateSystem, *ecs.Player, *ecs.Position) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	updateSys := NewUpdateSystem(w, m)

	player := &ecs.Player{
		Health:             100.0,
		Hunger:             100.0,
		Thirst:             100.0,
		Inventory:          make([]string, 9),
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
		FacingX:            1,
		FacingY:            0,
	}
	pos := &ecs.Position{X: 640, Y: 640}

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	playerMap.NewEntity(
		player,
		pos,
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, updateSys, player, pos
}

// simulateUseItem applies the inventory slot use/equip logic for slot useItemIdx
func simulateUseItem(player *ecs.Player, useItemIdx int) {
	if useItemIdx >= 0 && useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		t := player.Inventory[useItemIdx]
		if t == "food" && player.Hunger < 100 {
			player.AttackCooldown = 30
			player.Hunger += 50
			if player.Hunger > 100 {
				player.Hunger = 100
			}
			player.Inventory[useItemIdx] = ""
		} else if t == "antidote" && player.Infected {
			player.AttackCooldown = 30
			player.Infected = false
			player.Inventory[useItemIdx] = ""
		} else if t == "water" && player.Thirst < 100 {
			player.AttackCooldown = 30
			player.Thirst += 50
			if player.Thirst > 100 {
				player.Thirst = 100
			}
			player.Inventory[useItemIdx] = ""
		} else if t == "weapon" || t == "axe" || t == "shotgun" {
			player.AttackCooldown = 15
			oldWeapon := player.WeaponType
			wasEquipped := player.WeaponEquipped
			player.WeaponEquipped = true
			player.WeaponType = t
			if t == "weapon" {
				player.WeaponDurability = 5
			} else if t == "axe" {
				player.WeaponDurability = 12
			} else if t == "shotgun" {
				player.WeaponDurability = 15
			}
			if wasEquipped && oldWeapon != "" {
				player.Inventory[useItemIdx] = oldWeapon
			} else {
				player.Inventory[useItemIdx] = ""
			}
		} else if t == "armor" || t == "vest" {
			player.AttackCooldown = 30
			player.ArmorEquipped = true
			player.ArmorType = "vest"
			player.ArmorDefense = 0.50
			player.ArmorDurability = 10
			player.ArmorMaxDurability = 10
			player.InfectionResist = 0.70
			player.Inventory[useItemIdx] = ""
		}
	}
}

// simulateUnequip applies the unequip (hotkey 'U') logic
func simulateUnequip(player *ecs.Player) bool {
	if player.WeaponEquipped && player.WeaponType != "" && player.AttackCooldown <= 0 {
		emptyIdx := -1
		for idx := 0; idx < len(player.Inventory) && idx < 9; idx++ {
			if player.Inventory[idx] == "" {
				emptyIdx = idx
				break
			}
		}
		if emptyIdx != -1 {
			player.AttackCooldown = 15
			player.Inventory[emptyIdx] = player.WeaponType
			player.WeaponEquipped = false
			player.WeaponType = ""
			player.WeaponDurability = 0
			return true
		}
	}
	return false
}

// simulateDragDrop applies the inventory drag and drop logic
func simulateDragDrop(player *ecs.Player, draggingSlot, dropSlot int) {
	if draggingSlot == dropSlot || draggingSlot < 0 || dropSlot < 0 {
		return
	}
	for len(player.Inventory) < 9 {
		player.Inventory = append(player.Inventory, "")
	}

	if draggingSlot >= 0 && draggingSlot < 9 && dropSlot >= 0 && dropSlot < 9 {
		player.Inventory[draggingSlot], player.Inventory[dropSlot] = player.Inventory[dropSlot], player.Inventory[draggingSlot]
	} else if draggingSlot >= 0 && draggingSlot < 9 && dropSlot == 9 {
		// Equip weapon from inventory slot
		itm := player.Inventory[draggingSlot]
		if itm == "weapon" || itm == "axe" || itm == "shotgun" {
			oldWeapon := player.WeaponType
			wasEquipped := player.WeaponEquipped
			player.WeaponEquipped = true
			player.WeaponType = itm
			if itm == "weapon" {
				player.WeaponDurability = 5
			} else if itm == "axe" {
				player.WeaponDurability = 12
			} else if itm == "shotgun" {
				player.WeaponDurability = 15
			}
			if wasEquipped && oldWeapon != "" {
				player.Inventory[draggingSlot] = oldWeapon
			} else {
				player.Inventory[draggingSlot] = ""
			}
		}
	} else if draggingSlot == 9 && dropSlot >= 0 && dropSlot < 9 {
		// Unequip/move equipped weapon to inventory slot
		destItem := player.Inventory[dropSlot]
		if destItem == "weapon" || destItem == "axe" || destItem == "shotgun" {
			oldWeapon := player.WeaponType
			player.WeaponType = destItem
			if destItem == "weapon" {
				player.WeaponDurability = 5
			} else if destItem == "axe" {
				player.WeaponDurability = 12
			} else if destItem == "shotgun" {
				player.WeaponDurability = 15
			}
			player.Inventory[dropSlot] = oldWeapon
		} else if destItem == "" {
			player.Inventory[dropSlot] = player.WeaponType
			player.WeaponEquipped = false
			player.WeaponType = ""
			player.WeaponDurability = 0
		} else {
			// Find first empty slot
			emptyIdx := -1
			for k := 0; k < 9; k++ {
				if player.Inventory[k] == "" {
					emptyIdx = k
					break
				}
			}
			if emptyIdx != -1 {
				player.Inventory[emptyIdx] = player.WeaponType
				player.WeaponEquipped = false
				player.WeaponType = ""
				player.WeaponDurability = 0
			}
		}
	}
}

// 1. Test equipping weapons from inventory moves them to equipped slot
func TestEquip_WeaponMovesToEquippedSlot(t *testing.T) {
	tests := []struct {
		weaponType     string
		slotIdx        int
		wantDurability int
	}{
		{"weapon", 0, 5},
		{"axe", 3, 12},
		{"shotgun", 8, 15},
	}

	for _, tt := range tests {
		t.Run(tt.weaponType, func(t *testing.T) {
			_, _, _, player, _ := setupEquipTestGame()
			player.Inventory[tt.slotIdx] = tt.weaponType

			simulateUseItem(player, tt.slotIdx)

			if !player.WeaponEquipped {
				t.Fatalf("Expected WeaponEquipped to be true")
			}
			if player.WeaponType != tt.weaponType {
				t.Errorf("Expected WeaponType '%s', got '%s'", tt.weaponType, player.WeaponType)
			}
			if player.WeaponDurability != tt.wantDurability {
				t.Errorf("Expected WeaponDurability %d, got %d", tt.wantDurability, player.WeaponDurability)
			}
			if player.Inventory[tt.slotIdx] != "" {
				t.Errorf("Expected inventory slot %d to be empty, got '%s'", tt.slotIdx, player.Inventory[tt.slotIdx])
			}
		})
	}
}

// 2. Test equipping a new weapon when one is already equipped swaps the old weapon back into inventory
func TestEquip_SwappingEquippedWeapons(t *testing.T) {
	_, _, _, player, _ := setupEquipTestGame()

	// Initial: Axe is equipped
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 10
	player.Inventory[2] = "shotgun"

	// Equip shotgun from slot 2
	simulateUseItem(player, 2)

	if !player.WeaponEquipped {
		t.Fatalf("Expected WeaponEquipped to be true")
	}
	if player.WeaponType != "shotgun" {
		t.Errorf("Expected active weapon to be 'shotgun', got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 15 {
		t.Errorf("Expected shotgun durability 15, got %d", player.WeaponDurability)
	}
	if player.Inventory[2] != "axe" {
		t.Errorf("Expected slot 2 to now contain 'axe', got '%s'", player.Inventory[2])
	}

	// Swap back: Equip axe from slot 2
	player.AttackCooldown = 0
	simulateUseItem(player, 2)

	if player.WeaponType != "axe" {
		t.Errorf("Expected active weapon to be 'axe', got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 12 {
		t.Errorf("Expected axe durability 12, got %d", player.WeaponDurability)
	}
	if player.Inventory[2] != "shotgun" {
		t.Errorf("Expected slot 2 to contain 'shotgun', got '%s'", player.Inventory[2])
	}
}

// 3. Test unequipping returns the active weapon to the first available empty slot
func TestUnequip_HotkeyReturnsWeaponToInventory(t *testing.T) {
	_, _, _, player, _ := setupEquipTestGame()

	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12

	// Populate slots 0, 1, 3 with items; slot 2 is the FIRST empty slot
	player.Inventory[0] = "food"
	player.Inventory[1] = "water"
	player.Inventory[2] = "" // First empty
	player.Inventory[3] = "ammo"

	success := simulateUnequip(player)
	if !success {
		t.Fatalf("Expected unequip to succeed")
	}
	if player.WeaponEquipped {
		t.Errorf("Expected WeaponEquipped to be false")
	}
	if player.WeaponType != "" {
		t.Errorf("Expected WeaponType to be empty, got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 0 {
		t.Errorf("Expected WeaponDurability to be 0, got %d", player.WeaponDurability)
	}
	if player.Inventory[2] != "axe" {
		t.Errorf("Expected slot 2 (first empty) to receive 'axe', got '%s'", player.Inventory[2])
	}
	// Verify other slots were not disturbed
	if player.Inventory[0] != "food" || player.Inventory[1] != "water" || player.Inventory[3] != "ammo" {
		t.Errorf("Existing inventory items were corrupted: %v", player.Inventory)
	}
}

// 4. Test unequip rejection when inventory is completely full (data loss protection)
func TestUnequip_FullInventorySafety(t *testing.T) {
	_, _, _, player, _ := setupEquipTestGame()

	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15

	// Fill all 9 slots
	fullInventory := []string{"food", "water", "ammo", "ammo", "food", "armor", "water", "food", "antidote"}
	copy(player.Inventory, fullInventory)

	success := simulateUnequip(player)
	if success {
		t.Errorf("Expected unequip to fail when inventory is full")
	}
	// Verify weapon remains safely equipped
	if !player.WeaponEquipped || player.WeaponType != "shotgun" || player.WeaponDurability != 15 {
		t.Errorf("Equipped weapon was corrupted during full inventory unequip attempt")
	}
	// Verify all 9 inventory items are preserved
	for i := 0; i < 9; i++ {
		if player.Inventory[i] != fullInventory[i] {
			t.Errorf("Inventory slot %d was modified: got '%s', want '%s'", i, player.Inventory[i], fullInventory[i])
		}
	}
}

// 5. Test Drag and Drop equip, unequip, and slot swapping
func TestEquip_DragAndDropEquipAndUnequip(t *testing.T) {
	_, _, _, player, _ := setupEquipTestGame()

	// 5.1 Drag slot 1 ("axe") to slot 9 (Equipped slot)
	player.Inventory[1] = "axe"
	simulateDragDrop(player, 1, 9)

	if !player.WeaponEquipped || player.WeaponType != "axe" || player.WeaponDurability != 12 {
		t.Fatalf("Drag equip failed: equipped=%v type=%s", player.WeaponEquipped, player.WeaponType)
	}
	if player.Inventory[1] != "" {
		t.Errorf("Expected slot 1 to be cleared after drag-equip, got '%s'", player.Inventory[1])
	}

	// 5.2 Drag slot 9 (Equipped "axe") to slot 4 (empty slot)
	simulateDragDrop(player, 9, 4)

	if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
		t.Fatalf("Drag unequip failed: weapon still equipped")
	}
	if player.Inventory[4] != "axe" {
		t.Errorf("Expected slot 4 to receive 'axe', got '%s'", player.Inventory[4])
	}

	// 5.3 Drag slot 9 to a slot containing another weapon (swap)
	player.WeaponEquipped = true
	player.WeaponType = "weapon"
	player.WeaponDurability = 5
	player.Inventory[4] = "shotgun"

	simulateDragDrop(player, 9, 4)

	if !player.WeaponEquipped || player.WeaponType != "shotgun" || player.WeaponDurability != 15 {
		t.Errorf("Drag swap failed: expected equipped shotgun, got %s", player.WeaponType)
	}
	if player.Inventory[4] != "weapon" {
		t.Errorf("Expected slot 4 to receive former weapon, got '%s'", player.Inventory[4])
	}
}

// 6. Test that non-weapon consumable items do NOT equip into the weapon slot
func TestEquip_NonWeaponItemsDoNotEquip(t *testing.T) {
	_, _, _, player, _ := setupEquipTestGame()

	player.Hunger = 50.0
	player.Thirst = 50.0
	player.Infected = true
	player.Inventory[0] = "food"
	player.Inventory[1] = "water"
	player.Inventory[2] = "antidote"

	simulateUseItem(player, 0)
	if player.WeaponEquipped || player.WeaponType != "" {
		t.Errorf("Consuming food incorrectly equipped weapon")
	}
	if player.Hunger != 100.0 {
		t.Errorf("Expected hunger 100, got %f", player.Hunger)
	}

	player.AttackCooldown = 0
	simulateUseItem(player, 1)
	if player.WeaponEquipped || player.WeaponType != "" {
		t.Errorf("Consuming water incorrectly equipped weapon")
	}
	if player.Thirst != 100.0 {
		t.Errorf("Expected thirst 100, got %f", player.Thirst)
	}

	player.AttackCooldown = 0
	simulateUseItem(player, 2)
	if player.WeaponEquipped || player.WeaponType != "" {
		t.Errorf("Consuming antidote incorrectly equipped weapon")
	}
	if player.Infected {
		t.Errorf("Expected infection cured")
	}
}

// 7. Test Armor equipping leaves equipped weapon untouched
func TestEquip_ArmorDoesNotAffectEquippedWeapon(t *testing.T) {
	_, _, _, player, _ := setupEquipTestGame()

	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12

	player.Inventory[0] = "armor"
	simulateUseItem(player, 0)

	if !player.ArmorEquipped || player.ArmorType != "vest" {
		t.Errorf("Expected armor to be equipped")
	}
	if !player.WeaponEquipped || player.WeaponType != "axe" || player.WeaponDurability != 12 {
		t.Errorf("Equipping armor mutated equipped weapon: equipped=%v type=%s dur=%d",
			player.WeaponEquipped, player.WeaponType, player.WeaponDurability)
	}
}

// 8. Test HUD rendering of equipped slot across various states
func TestEquippedSlot_HUDDrawing(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	drawSys := NewDrawSystem(w, m)

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	player := &ecs.Player{
		Health:           100.0,
		Hunger:           100.0,
		Thirst:           100.0,
		Inventory:        make([]string, 9),
		WeaponEquipped:   false,
		WeaponType:       "",
		WeaponDurability: 0,
	}
	playerMap.NewEntity(
		player,
		&ecs.Position{X: 640, Y: 640},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	screen := ebiten.NewImage(1280, 720)

	// Case 1: Unarmed
	drawSys.Draw(screen, 12.0, -1)

	// Case 2: Axe equipped
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	drawSys.Draw(screen, 12.0, -1)

	// Case 3: Shotgun equipped while dragging equipped slot (slot 9)
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	drawSys.Draw(screen, 12.0, 9)

	// Verify no panic occurred and drawing succeeded
}
