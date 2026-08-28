package ecs

import (
	"image/color"
	"testing"
)

func TestPositionComponent(t *testing.T) {
	pos := Position{X: 123.45, Y: 678.90}
	if pos.X != 123.45 || pos.Y != 678.90 {
		t.Errorf("Position mismatch: got (%f, %f), want (123.45, 678.90)", pos.X, pos.Y)
	}
}

func TestVelocityComponent(t *testing.T) {
	vel := Velocity{X: -2.5, Y: 3.5}
	if vel.X != -2.5 || vel.Y != 3.5 {
		t.Errorf("Velocity mismatch: got (%f, %f), want (-2.5, 3.5)", vel.X, vel.Y)
	}
}

func TestSpriteAndColliderComponents(t *testing.T) {
	col := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	sprite := Sprite{Color: col, W: 16, H: 32}
	if sprite.W != 16 || sprite.H != 32 || sprite.Color != col {
		t.Errorf("Sprite mismatch: got %+v", sprite)
	}

	collider := Collider{Width: 16, Height: 16}
	if collider.Width != 16 || collider.Height != 16 {
		t.Errorf("Collider mismatch: got %+v", collider)
	}
}

func TestPlayerComponentState(t *testing.T) {
	p := Player{
		Health:             100.0,
		Hunger:             100.0,
		Thirst:             100.0,
		Inventory:          []string{"food", "armor"},
		WeaponEquipped:     true,
		WeaponType:         "axe",
		WeaponDurability:   12,
		ArmorEquipped:      true,
		ArmorType:          "armor",
		ArmorDefense:       0.5,
		ArmorDurability:    10,
		ArmorMaxDurability: 10,
		InfectionResist:    0.8,
		AttackCooldown:     0,
		Dead:               false,
		Infected:           false,
		FacingX:            1.0,
		FacingY:            0.0,
	}

	if p.Health != 100.0 || p.Hunger != 100.0 || p.Thirst != 100.0 {
		t.Errorf("Player stats mismatch: %+v", p)
	}
	if !p.WeaponEquipped || p.WeaponType != "axe" || p.WeaponDurability != 12 {
		t.Errorf("Player weapon state mismatch: %+v", p)
	}
	if !p.ArmorEquipped || p.ArmorType != "armor" || p.ArmorDefense != 0.5 || p.ArmorDurability != 10 || p.InfectionResist != 0.8 {
		t.Errorf("Player armor state mismatch: %+v", p)
	}
}

func TestZombieAndItemComponents(t *testing.T) {
	item := Item{Type: "shotgun"}
	if item.Type != "shotgun" {
		t.Errorf("Item type mismatch: got %s, want shotgun", item.Type)
	}

	zombie := Zombie{
		Speed:       1.5,
		Chasing:     true,
		IsRunner:    false,
		WanderTimer: 60,
		WanderDirX:  0.707,
		WanderDirY:  0.707,
		StunTimer:   15,
	}
	if zombie.Speed != 1.5 || !zombie.Chasing || zombie.IsRunner || zombie.StunTimer != 15 {
		t.Errorf("Zombie state mismatch: %+v", zombie)
	}
}
