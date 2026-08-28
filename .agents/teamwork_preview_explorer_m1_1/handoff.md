# Handoff Report: Milestone 1 Floor Tiles (256x128 Scaling)

**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_1`  
**Target File Investigated**: `cmd/tools/genassets/main.go`  
**Detailed Report**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_1/m1_floor_analysis.md`  
**Date**: 2026-08-28  

---

## 1. Observation

Direct examination of the codebase revealed the following:
1. **Floor Tile Sizes in Asset Generator** (`cmd/tools/genassets/main.go:25-32, 351-695`):
   - `generateGrass`: `w, h := 64, 32` (line 352). Diamond equation: `dx := float64(x) - 31.5; dy := float64(y) - 15.5; isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0` (lines 365-367).
   - `generateDirt`: `w, h := 64, 32` (line 423). Pebble overlays: 4 hardcoded 3x2 flat rectangles (lines 454-458).
   - `generateWoodFloor`: `w, h := 64, 32` (line 465). UV projection: `u := dx/64.0 + dy/32.0 + 0.5; v := dy/32.0 - dx/64.0 + 0.5` (lines 483-484).
   - `generateAsphalt`: `w, h := 64, 32` (line 541). Yellow marking: `v >= 0.43 && v <= 0.57 && (u <= 0.38 || u >= 0.62)` (line 558).
   - `generateConcrete`: `w, h := 64, 32` (line 580). Quadrant joints at `distU < 0.025 || distV < 0.025` (line 612).
   - `generateTileFloor`: `w, h := 64, 32` (line 634). Grout lines at `subU < 0.05 || subV < 0.05` (line 671).
2. **Asset Dimensions Testing Contracts** (`internal/assets/assets_test.go:23-30, 110-117`):
   - `{"images/grass.png", 64, 32}`, `{"images/dirt.png", 64, 32}`, `{"images/wood.png", 64, 32}`, `{"images/asphalt.png", 64, 32}`, `{"images/concrete.png", 64, 32}`, `{"images/tile_floor.png", 64, 32}`.
   - `TestAssetsLoadAllPointersNonNil` asserts non-nil `ebiten.Image` pointers of dimensions `64x32`.
3. **Style & Feature Requirements** (`PROJECT.md:16, 20` and `ART_STYLE_GUIDE.md:7-30`):
   - Feature 1: "Quadruple base floor tiles (`grass`, `dirt`, `wood`, `asphalt`, `concrete`, `tile_floor`) from 64x32 to 256x128 preserving 2:1 dimetric ratio."
   - Feature 5: "Scale and anti-alias chevron grass tufts, wildflower clusters, pebbles, UV plank lanes & nails, asphalt yellow dashes, concrete joints, and ceramic tile grout."

---

## 2. Logic Chain

1. **Resolution Quadrupling & Dimension Invariants**:
   Scaling base dimensions $W_{old} = 64, H_{old} = 32$ by $4\times$ yields $W_{new} = 256, H_{new} = 128$.
   Center points shift from $(31.5, 15.5)$ to $(127.5, 63.5)$.
   Semi-axes scale from $(32.0, 16.0)$ to $(128.0, 64.0)$.

2. **Isometric Diamond Boundary**:
   The normalized $L_1$ metric becomes $\text{isoDist}(x, y) = \frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0} \le 1.0$.
   The ratio $\frac{64.0}{128.0} = 0.5$ preserves the 2:1 dimetric projection exactly.

3. **Bi-directional UV Orthogonal Surface Coordinates**:
   Using the four diamond corner vertices $(127.5, -0.5) \to (0, 0)$, $(255.5, 63.5) \to (1, 0)$, $(127.5, 127.5) \to (1, 1)$, and $(-0.5, 63.5) \to (0, 1)$:
   - Forward: $dx = (u - v) \cdot 128.0$, $dy = (u + v - 1.0) \cdot 64.0$.
   - Inverse: $u = \frac{dx}{256.0} + \frac{dy}{128.0} + 0.5$, $v = \frac{dy}{128.0} - \frac{dx}{256.0} + 0.5$.
   - Inside diamond: $u \in [0, 1] \land v \in [0, 1] \iff \text{isoDist} \le 1.0$.

4. **Procedural Vector Overlays Scaling**:
   - *Grass*: Chevrons become 3-blade vector clusters (vertical stem + 2 diagonal arms) across 8 distributed anchors; wildflowers become 5-lobed petal clusters around a yellow pistil disk ($r=2.5$).
   - *Dirt*: Pebbles scale from 3x2 rectangular stamps to smooth $14\times 8$ px ellipses ($r_x=7.0, r_y=4.0$) with drop shadows, specular highlights, and shadow creases.
   - *Wood*: 4 longitudinal lanes with 3px dark seams ($v_{\text{local}} < 0.04 \lor v_{\text{local}} > 0.96$), staggered end joints, and $5\times 5$ px iron nailheads with specular highlights.
   - *Asphalt*: Dashed centerline markings in UV with 10% lane ribbon ($v \in [0.45, 0.55]$) and distinct dash intervals ($u \in [0.08, 0.40] \cup [0.60, 0.92]$).
   - *Concrete*: 2x2 quadrants with 3px expansion joints ($|u-0.5| < 0.010 \lor |v-0.5| < 0.010$) and top-edge chamfered bevel highlights.
   - *Tile Floor*: 4x4 checkerboard with 2.5px mortar grout lines ($subU < 0.045 \lor subV < 0.045$) and dual bevels (top-left specular highlight, bottom-right shadow).

---

## 3. Caveats

1. **Scope Boundary**: This investigation focuses specifically on the 6 floor generators (`grass`, `dirt`, `wood`, `asphalt`, `concrete`, `tile_floor`). Obstacles/props (256x256), character entities (64x128), and item sprites (64x64) are analyzed in parallel explorer tracks.
2. **Coordinated Testing Update**: Updating floor generators to 256x128 will cause `internal/assets/assets_test.go` to fail until `expectedAssets` dimensions in the test files are updated to match 256x128 during implementation.

---

## 4. Conclusion

The floor generator scaling architecture is fully defined and ready for direct implementation. All mathematical transformations, coordinate equations, geometric overlay formulas, and complete drop-in Go implementations have been documented in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_1/m1_floor_analysis.md`.

---

## 5. Verification Method

1. **Generator Execution**:
   ```bash
   go run ./cmd/tools/genassets
   ```
2. **Image Dimension Verification**:
   Inspect the generated floor PNGs in `internal/assets/images/`:
   ```bash
   file internal/assets/images/grass.png internal/assets/images/dirt.png internal/assets/images/wood.png internal/assets/images/asphalt.png internal/assets/images/concrete.png internal/assets/images/tile_floor.png
   ```
   *Expected output*: `256 x 128, 8-bit/color RGBA, non-interlaced` for all 6 floor tiles.
3. **Unit Test Suite**:
   ```bash
   CC=gcc go test -v ./internal/assets/...
   ```
