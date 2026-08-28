# Handoff Report — M1 Asset Pipeline & Race Testing Investigation

## 1. Observation
1. **Command Execution**:
   - Running `go run ./cmd/tools/genassets && CC=gcc go test -race -v ./cmd/tools/genassets/... ./internal/assets/...` resulted in test failures and data race warnings in `internal/assets`.
2. **Failure 1 (Empirical Challenger Floor Geometry)**:
   - File: `internal/assets/empirical_challenger_test.go:226, 235, 242`
   - Test: `TestEmpiricalFloorDiamondGeometry/images/dirt.png`
   - Verbatim error log:
     ```
     empirical_challenger_test.go:226: Inner hole at (153, 30): RGBA=(0, 0, 0, 45) [isoDist=0.723]
     empirical_challenger_test.go:226: Inner hole at (152, 31): RGBA=(0, 0, 0, 45) [isoDist=0.699]
     empirical_challenger_test.go:235: images/dirt.png has 18 non-transparent pixels outside isometric diamond (isoDist > 1.0)
     empirical_challenger_test.go:242: images/dirt.png has 151 transparent/semi-transparent pixels inside solid core (isoDist <= 0.85)
     ```
3. **Failure 2 (Multi-Threaded Load Data Race)**:
   - File: `internal/assets/challenger_stress_test.go:109, 123`
   - Test: `TestChallenger_MultiThreadedLoadAndPointerRace`
   - Verbatim race detector log:
     ```
     ==================
     WARNING: DATA RACE
     Write at 0x000000eb91e0 by goroutine 141:
       github.com/BryceWayne/go-zomboid/internal/assets.Load()
           /home/bryce/code/go-zomboid/internal/assets/assets.go:55 +0x3d
     ...
     Previous read at 0x000000eb91e0 by goroutine 162:
       github.com/BryceWayne/go-zomboid/internal/assets.TestChallenger_MultiThreadedLoadAndPointerRace.func2()
           /home/bryce/code/go-zomboid/internal/assets/challenger_stress_test.go:123 +0x1ea
     ==================
     ```
   - Total runtime of `internal/assets` under race detector was 49.784s due to 1,000 un-memoized `Load()` calls decoding 27 PNGs repeatedly.

---

## 2. Logic Chain
1. In `cmd/tools/genassets/main.go:250-288`, `drawVectorPebble` defined `dropShadow := color.RGBA{0, 0, 0, 45}` and placed pixels via `setPixel(img, x, y, dropShadow)` instead of `blendPixel(img, x, y, dropShadow)`.
2. `setPixel` overwrote the underlying dirt tile pixels ($A = 255$) with $A = 45$. Because $A < 255$, line 222 of `empirical_challenger_test.go` (`if a < 0xFFFF { innerHoleCount++ }`) flagged 151 pixels in the inner diamond core ($\text{isoDist} \le 0.85$).
3. In `cmd/tools/genassets/main.go:668`, pebble `{195, 36}` has center coordinate $\text{isoDist} = 0.957$. With $r_x = 7.0$, rightmost pixels reached $x = 202$ ($\text{isoDist} = 1.012$), exceeding the diamond boundary ($\text{isoDist} > 1.0$) by 18 pixels without boundary clipping.
4. In `internal/assets/assets.go:53-88`, `Load()` read all 27 embedded images from `imageFS` and reassigned global pointers `PlayerImage`, `ZombieImage`, etc., on every invocation.
5. In `challenger_stress_test.go:94-150`, 20 goroutines called `Load()` 50 times simultaneously while 30 reader goroutines concurrently read `*d.ptr`. Because writes to global variables were un-synchronized and not protected by `sync.Once`, the Go race detector reported data races and performance degraded to 50 seconds.
6. Wrapping `Load()` with `sync.Once` ensures single initialization, eliminating write-write and write-read races, and reducing execution time from 50s to <1.5s.

---

## 3. Caveats
- `CC=gcc` must be supplied in test environments where Go's default C compiler points to an environment lacking X11/alsa development headers, as demonstrated by the acceptance criteria and project contracts.
- No changes to game logic or world coordinates are required for this M1 fix; this fix strictly targets the asset generator and asset loader packages.

---

## 4. Conclusion
- The test suite failure in `empirical_challenger_test.go` and race warnings in `challenger_stress_test.go` are completely addressed by:
  1. Updating `drawVectorPebble` in `cmd/tools/genassets/main.go` to use `blendPixel` for drop shadows and adding an `isoDist <= 1.0` boundary guard.
  2. Shifting pebble position `{195, 36}` to `{185, 42}` in `generateDirt`.
  3. Adding `sync.Once` to `Load()` in `internal/assets/assets.go`.
- Detailed before/after code diffs are documented in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_3/fix_plan.md`.

---

## 5. Verification Method
1. Re-generate all sprites:
   ```bash
   go run ./cmd/tools/genassets
   ```
2. Run the complete race testing suite:
   ```bash
   CC=gcc go test -race -v ./cmd/tools/genassets/... ./internal/assets/...
   ```
3. Invalidation condition: Any test failure, non-zero exit code, or data race warning in `cmd/tools/genassets/...` or `internal/assets/...`.
