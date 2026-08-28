package ecs

import (
	"image/color"
)

// Position component for world coordinates.
type Position struct {
	X, Y float64
}

// Velocity component for movement.
type Velocity struct {
	X, Y float64
}

// Sprite component for rendering placeholder objects.
type Sprite struct {
	Color color.RGBA
	W, H  float64
}

// Collider component for AABB collision detection.
type Collider struct {
	Width, Height float64
}

// Player marker component.
type Player struct{
	Infected       bool
	Health         float64
	Dead           bool
	FacingX        float64
	FacingY        float64
	WeaponEquipped bool
	AttackCooldown int
}

// Item marker component
type Item struct{
	Type string
}

// Zombie marker component.
type Zombie struct{
	Speed       float64
	Chasing     bool
	IsRunner    bool
	WanderTimer int
	WanderDirX  float64
	WanderDirY  float64
}
