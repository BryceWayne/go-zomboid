package game

import (
	"testing"
)

func TestWorldToIso(t *testing.T) {
	tests := []struct {
		name       string
		wx, wy     float64
		wantX, wantY float64
	}{
		{"Origin", 0, 0, 0, 0},
		{"X Axis", 32, 0, 32, 16},
		{"Y Axis", 0, 32, -32, 16},
		{"Diagonal", 32, 32, 0, 32},
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
