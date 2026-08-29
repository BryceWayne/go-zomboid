package game

import (
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	arkecs "github.com/mlange-42/ark/ecs"
)

func TestWorldToIso(t *testing.T) {
	tests := []struct {
		name         string
		wx, wy       float64
		wantX, wantY float64
	}{
		{"Origin", 0, 0, 0, 0},
		{"X Axis", 32, 0, 32, 0},
		{"Y Axis", 0, 32, 0, 32},
		{"Diagonal", 32, 32, 32, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := WorldToIso(tt.wx, tt.wy)
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Errorf("WorldToIso(%f, %f) = (%f, %f), want (%f, %f)",
					tt.wx, tt.wy, gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestNewGameInitialization(t *testing.T) {
	assets.Load()
	g := NewGame()
	if g == nil {
		t.Fatal("NewGame() returned nil")
	}
	if g.gameMap == nil {
		t.Fatal("NewGame() gameMap is nil")
	}
	if g.world == nil {
		t.Fatal("NewGame() arkecs world is nil")
	}
}

func TestGameResetContextualSpawns(t *testing.T) {
	assets.Load()
	g := NewGame()
	if g == nil {
		t.Fatal("NewGame() is nil")
	}

	// Verify player is initialized at safe spawn
	var playerCount int
	pq := arkecs.NewFilter2[ecs.Player, ecs.Position](g.world).Query()
	for pq.Next() {
		playerCount++
		p, pos := pq.Get()
		if p.Health != 100.0 {
			t.Errorf("Expected initial player health 100, got %f", p.Health)
		}
		if pos.X != g.gameMap.PlayerSpawn.X || pos.Y != g.gameMap.PlayerSpawn.Y {
			t.Errorf("Player not at spawn (%f, %f), got (%f, %f)",
				g.gameMap.PlayerSpawn.X, g.gameMap.PlayerSpawn.Y, pos.X, pos.Y)
		}
	}
	if playerCount != 1 {
		t.Errorf("Expected 1 player, got %d", playerCount)
	}

	// Verify items spawned from map loot spawns
	var itemCount int
	iq := arkecs.NewFilter2[ecs.Item, ecs.Position](g.world).Query()
	for iq.Next() {
		itemCount++
	}
	if itemCount < len(g.gameMap.LootSpawns) {
		t.Errorf("Expected %d items spawned, got %d", len(g.gameMap.LootSpawns), itemCount)
	}

	// Verify zombies spawned from map zombie spawns
	var zombieCount int
	zq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](g.world).Query()
	for zq.Next() {
		zombieCount++
	}
	if zombieCount != len(g.gameMap.ZombieSpawns) {
		t.Errorf("Expected %d zombies spawned, got %d", len(g.gameMap.ZombieSpawns), zombieCount)
	}
}
