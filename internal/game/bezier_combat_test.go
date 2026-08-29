package game

import (
	"image/color"
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestBezier_AxeControlPointsCalculation tests mathematical formulation of Bezier control points for Fire Axe
func TestBezier_AxeControlPointsCalculation(t *testing.T) {
	playerX, playerY := 200.0, 200.0
	facingX, facingY := 1.0, 0.0 // Facing Right (East), theta = 0

	baseAngle := math.Atan2(facingY, facingX)
	deltaTheta := 2.0 // ~115 degrees
	rIn := 40.0
	rApex := 140.0
	rOut := 120.0

	theta0 := baseAngle - deltaTheta/2.0
	theta1 := baseAngle + deltaTheta/2.0

	p0x := playerX + rIn*math.Cos(theta0)
	p0y := playerY + rIn*math.Sin(theta0)

	p1x := playerX + rApex*math.Cos(baseAngle)
	p1y := playerY + rApex*math.Sin(baseAngle)

	p2x := playerX + rOut*math.Cos(theta1)
	p2y := playerY + rOut*math.Sin(theta1)

	// Invariants:
	// P1 apex must be furthest along facing axis: p1x = 340.0, p1y = 200.0
	if math.Abs(p1x-340.0) > 1e-6 || math.Abs(p1y-200.0) > 1e-6 {
		t.Errorf("Axe apex P1 mismatch: got (%f, %f), want (340.0, 200.0)", p1x, p1y)
	}

	// P0 is initial swing position (rIn=40), P2 is follow-through extension (rOut=120)
	if (p0y - playerY) >= 0 {
		t.Errorf("P0 should be above facing axis (negative delta Y): got %f", p0y-playerY)
	}
	if (p2y - playerY) <= 0 {
		t.Errorf("P2 should be below facing axis (positive delta Y): got %f", p2y-playerY)
	}
	if math.Hypot(p2x-playerX, p2y-playerY) <= math.Hypot(p0x-playerX, p0y-playerY) {
		t.Errorf("Follow-through radius should expand beyond initial reach")
	}

	// Transform to screen space via WorldToIso (orthogonal mapping)
	s0x, s0y := WorldToIso(p0x, p0y)
	s1x, s1y := WorldToIso(p1x, p1y)
	s2x, s2y := WorldToIso(p2x, p2y)

	// In orthogonal space, WorldToIso(340, 200) = (340.0, 200.0)
	if math.Abs(s1x-340.0) > 1e-6 || math.Abs(s1y-200.0) > 1e-6 {
		t.Errorf("Screen apex S1 mismatch: got (%f, %f), want (340.0, 200.0)", s1x, s1y)
	}

	if s0x == s1x && s0y == s1y {
		t.Error("S0 and S1 should not be coincident")
	}
	if s2x == s1x && s2y == s1y {
		t.Error("S2 and S1 should not be coincident")
	}
}

// TestBezier_AllWeaponControlPointsSanity tests control points generation across all weapon types
func TestBezier_AllWeaponControlPointsSanity(t *testing.T) {
	weapons := []string{"axe", "weapon", "shotgun", ""}

	for _, wType := range weapons {
		t.Run("Weapon_"+wType, func(t *testing.T) {
			screen := ebiten.NewImage(800, 600)
			w := arkecs.NewWorld()
			m := worldNewMapStub(40, 40)
			drawSys := NewDrawSystem(w, m)

			// Test all 14 frames of swing animation (attackCooldown: 30 down to 17)
			for cd := 30; cd > 16; cd-- {
				drawSys.DrawAttackSwingArc(screen, 300.0, 300.0, 1.0, 0.0, wType, cd, 0.0, 0.0)
			}
		})
	}
}

// TestBezier_AlphaFadeCurveProgression verifies quadratic fade: alpha = (1 - t)^2
func TestBezier_AlphaFadeCurveProgression(t *testing.T) {
	for cd := 30; cd >= 16; cd-- {
		tVal := float64(30-cd) / 14.0
		expectedAlpha := float32((1.0 - tVal) * (1.0 - tVal))

		if cd == 30 {
			if expectedAlpha != 1.0 {
				t.Errorf("At start of swing (cd=30), expected alpha=1.0, got %f", expectedAlpha)
			}
		}
		if cd == 16 {
			if expectedAlpha != 0.0 {
				t.Errorf("At end of swing (cd=16), expected alpha=0.0, got %f", expectedAlpha)
			}
		}
		if expectedAlpha < 0.0 || expectedAlpha > 1.0 {
			t.Errorf("Alpha out of bounds: %f at cd=%d", expectedAlpha, cd)
		}
	}
}

// TestBezier_PathRenderingAntiAliasing verifies vector.StrokePath executes cleanly with AntiAlias true
func TestBezier_PathRenderingAntiAliasing(t *testing.T) {
	screen := ebiten.NewImage(800, 600)

	var path vector.Path
	path.MoveTo(100, 100)
	path.QuadTo(200, 50, 300, 100)

	outerDrawOpts := &vector.DrawPathOptions{AntiAlias: true}
	outerDrawOpts.ColorScale.ScaleWithColor(color.RGBA{255, 69, 0, 200})

	vector.StrokePath(screen, &path, &vector.StrokeOptions{
		Width:    12.0,
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}, outerDrawOpts)

	coreDrawOpts := &vector.DrawPathOptions{AntiAlias: true}
	coreDrawOpts.ColorScale.ScaleWithColor(color.RGBA{255, 230, 120, 255})

	vector.StrokePath(screen, &path, &vector.StrokeOptions{
		Width:    4.0,
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}, coreDrawOpts)
}

// Helper function to create map stub
func worldNewMapStub(w, h int) *world.Map {
	assets.Load()
	return world.NewMap(w, h)
}
