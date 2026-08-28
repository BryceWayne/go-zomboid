# Handoff Report: Milestone 1 Empirical Asset Challenge

**Agent**: `m1_challenger_2`  
**Verdict**: **FAIL**  
**Role**: Empirical Challenger (critic, specialist)  
**Date**: 2026-08-28T18:59:35Z  

---

## 1. Observation

1. **Test Execution**:
   - Running `CC=gcc go test -v ./internal/assets -run TestEmpiricalFloorDiamondGeometry/images/dirt.png` produces verbatim failure:
     ```
     --- FAIL: TestEmpiricalFloorDiamondGeometry (0.00s)
         --- FAIL: TestEmpiricalFloorDiamondGeometry/images/dirt.png (0.00s)
             empirical_challenger_test.go:226: Inner hole at (153, 30): RGBA=(0, 0, 0, 45) [isoDist=0.723]
             empirical_challenger_test.go:235: images/dirt.png has 18 non-transparent pixels outside isometric diamond (isoDist > 1.0)
             empirical_challenger_test.go:242: images/dirt.png has 151 transparent/semi-transparent pixels inside solid core (isoDist <= 0.85)
     ```
   - Running `CC=gcc go test -v -race ./internal/assets -run TestChallenger_MultiThreadedLoadAndPointerRace` produces verbatim race detector failure:
     ```
     ==================
     WARNING: DATA RACE
     Write at 0x000000eb92a8 by goroutine 40:
       github.com/BryceWayne/go-zomboid/internal/assets.Load()
           /home/bryce/code/go-zomboid/internal/assets/assets.go:86 +0x84a
     Previous write at 0x000000eb92a8 by goroutine 42:
       github.com/BryceWayne/go-zomboid/internal/assets.Load()
           /home/bryce/code/go-zomboid/internal/assets/assets.go:86 +0x84a
     ==================
     ```

2. **Code Inspection**:
   - In `cmd/tools/genassets/main.go:250-265`, `drawVectorPebble()` draws pebble drop shadows using `setPixel(img, x, y, dropShadow)` with `dropShadow := color.RGBA{0, 0, 0, 45}` without checking isometric diamond bounds `isoDist <= 1.0`.
   - In `internal/assets/assets.go:53-88`, `Load()` performs direct unsynchronized writes to global variables `PlayerImage`, `GrassImage`, `WallImage`, etc., without `sync.Once`.

3. **Exported Pointer & Bounds Inspection**:
   - Checked all 27 exported pointers via `internal/assets/challenger_stress_test.go`:
     - Character entities (3): `PlayerImage`, `ZombieImage`, `RunnerImage` -> 64x128.
     - Floor tiles (6): `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage` -> 256x128.
     - Vertical obstacles/props (10): `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`, `TentImage`, `StumpImage`, `MushroomImage`, `SignImage`, `ElevationBlockImage`, `ElevationRampImage` -> 256x256.
     - Item/equipment (8): `FoodImage`, `WaterImage`, `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `AntidoteImage` -> 64x64.
   - All 27 pointers are non-nil, correctly decoded, and match specified bounds.

4. **Visual & Statistical Analysis**:
   - Luminance dynamic range $\ge 90.0$ and RMS contrast $\ge 25.0$ on all 27 assets.
   - All character entities have drop shadows anchored in rows 112..127 and full posture in rows 0..40.
   - All 8 item icons have centered bounding centroids within $[20..44]$.

---

## 2. Logic Chain

1. From Observation 1 and Observation 2, `drawVectorPebble()` directly replaces opaque earthen pixels with `RGBA{0, 0, 0, 45}`.
2. Because `setPixel` is used rather than `blendPixel`, the alpha channel is overwritten to 45 rather than preserving opacity 255.
3. Therefore, 151 core pixels in `images/dirt.png` become transparent holes, causing background bleeding during game rendering.
4. Additionally, because pebble shadow center `(cx+2, cy+2)` is drawn without checking $d_{iso} \le 1.0$, 18 shadow pixels extend outside the 2:1 isometric diamond, causing seams.
5. From Observation 1 and Observation 2, `internal/assets.Load()` lacks synchronization primitives (`sync.Once`).
6. When `Load()` is invoked concurrently from multiple goroutines or read during loading, unsynchronized writes occur on the 27 package variables, creating a data race detected by `-race`.
7. Therefore, Milestone 1 cannot be approved in its current state.

---

## 3. Caveats

- Single-threaded sequential execution of `assets.Load()` behaves correctly and decodes all 27 assets into non-nil `*ebiten.Image` handles.
- Asset pixel artwork quality, vector styling, dynamic range, and determinism across the other 26 assets (`grass.png`, `wood.png`, `player.png`, etc.) meet and exceed requirements.

---

## 4. Conclusion

**Verdict: FAIL**  
Milestone 1 asset pipeline has 2 reproducible defects that must be fixed:
1. `cmd/tools/genassets/main.go:250-265`: Fix `drawVectorPebble()` to use `blendPixel()` and constrain shadow drawing to `isoDist <= 1.0`.
2. `internal/assets/assets.go:53-88`: Wrap `Load()` with `sync.Once` to make asset loading thread-safe and race-free.

---

## 5. Verification Method

To independently verify these findings:
1. Run `CC=gcc go test -v ./internal/assets -run TestEmpiricalFloorDiamondGeometry/images/dirt.png` to reproduce the dirt tile alpha hole and boundary bleed failure.
2. Run `CC=gcc go test -v -race ./internal/assets -run TestChallenger_MultiThreadedLoadAndPointerRace` to reproduce the concurrent data race on `assets.Load()`.
3. Inspect `cmd/tools/genassets/main.go` lines 251–265 and `internal/assets/assets.go` lines 53–88.
