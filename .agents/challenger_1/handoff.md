# Empirical Stress Testing & Verification Handoff Report — Challenger 1

## 1. Observation

1. **Coordinate Conversion Implementation** (`internal/game/game.go:859-865`, `internal/game/game.go:205-221`):
   - `WorldToIso(wx, wy)` directly returns `(wx, wy)`.
   - `IsoToWorld(isoX, isoY)` directly returns `(isoX, isoY)`.
   - `ScreenToWorld(screenX, screenY, camX, camY)` evaluates `(camX + (screenX - 640.0)/DefaultZoom, camY + (screenY - 360.0)/DefaultZoom)` with `DefaultZoom = 0.5`.
   - `WorldToScreen(wx, wy, camX, camY)` evaluates `((wx - camX)*DefaultZoom + 640.0, (wy - camY)*DefaultZoom + 360.0)`.

2. **Empirical Fuzzing & Extreme Coordinate Stress** (`internal/game/orthogonal_stress_challenger_test.go:17-106`):
   - Tested boundary coordinates at $\pm 10^8$, near-zero sub-micron ($10^{-7}$), origin $(0, 0)$, and 50,000 randomized floating-point tuples in $[-10^8, 10^8]$.
   - Bidirectional round-trip precision verified: $|ScreenToWorld(WorldToScreen(wx, wy)) - wx| < 10^{-5}$ and $|WorldToScreen(ScreenToWorld(sx, sy)) - sx| < 10^{-6}$.
   - Viewport center invariant holds: `WorldToScreen(camX, camY, camX, camY) == (640.0, 360.0)`.
   - Scale factor linearity holds: $(WorldToScreen(wx+\Delta) - WorldToScreen(wx)) = \Delta \cdot DefaultZoom$.

3. **Adjacent Tile Seam & Black Gap Verification Across 10,000 Edges** (`internal/game/orthogonal_stress_challenger_test.go:108-164`):
   - Evaluated 10,000 adjacent tile pairs across $[-10000, 10000]$ tile indices under fractional sub-pixel camera offsets (e.g. $camX, camY \in [-5\cdot 10^5, +5\cdot 10^5]$).
   - Horizontal seam gap $|(s_{0x} + \text{TileSize}\cdot \text{DefaultZoom}) - s_{\text{right\_tile}, x}| = 0.0 \le 10^{-9}$.
   - Vertical seam gap $|(s_{0y} + \text{TileSize}\cdot \text{DefaultZoom}) - s_{\text{bottom\_tile}, y}| = 0.0 \le 10^{-9}$.
   - Diagonal corner coincidence $|(s_{0x} + \text{TileSize}\cdot \text{DefaultZoom}) - s_{\text{diag\_tile}, x}| \le 10^{-9}$ and $|(s_{0y} + \text{TileSize}\cdot \text{DefaultZoom}) - s_{\text{diag\_tile}, y}| \le 10^{-9}$.

4. **Camera Tracking, Exponential Lerp, and Sub-Pixel Snapping** (`internal/game/game.go:163-201`, `internal/game/orthogonal_stress_challenger_test.go:166-235`):
   - Camera initialization: uninitialized camera snaps immediately to target on first `Update` without lerp lag.
   - Sub-pixel bifurcation at $\text{hypot}(\Delta X, \Delta Y) < 0.01$: distances $\ge 0.01$ lerp smoothly by $10\%$ per frame; distances $< 0.01$ snap strictly to target without asymptotic drift or jitter.
   - Extreme coordinate convergence: camera initialized at $(-10^8, -10^8)$ tracking target $(+10^8, +10^8)$ converges monotonically without NaN, Inf, or overshoot, snapping at distance zero.

5. **Top-Down Y-Depth Sorting Monotonicity & Occlusion** (`internal/game/game.go:1251-1258`, `internal/game/orthogonal_stress_challenger_test.go:237-306`):
   - Monotonicity verified across 1,000 randomized arrays of 100 renderables ($10^5$ total elements) in $[-10^8, 10^8]$: `sort.SliceStable` strictly satisfies $\text{Depth}[i] \le \text{Depth}[i+1]$ for all $i$.
   - Stability verified on duplicate depths preserving original insertion order.
   - Occlusion hierarchy verified: Entity north of wall base ($\text{Depth} < \text{wallWorldY} + \text{TileSize}$) is drawn prior to wall (occluded); entity south of wall base ($\text{Depth} > \text{wallWorldY} + \text{TileSize}$) is drawn after wall (occludes).

6. **Bezier Combat Arc Affine Projection & Invariants** (`internal/game/game.go:1373-1554`, `internal/game/orthogonal_stress_challenger_test.go:308-384`):
   - Affine linearity verified: $WorldToScreen(B_{\text{world}}(t)) \equiv B_{\text{screen}}(t)$ with error $< 10^{-9}$ across all $t \in [0, 1]$.
   - Alpha fade curve $\alpha(t) = (1-t)^2$ is strictly monotonic decreasing across frames $30 \to 17$.
   - Safe zero-length facing vector fallback to $(1, 0)$ without panic or NaN.
   - Headless rendering across 360-degree rotation sweeps for all weapon types ("axe", "weapon", "shotgun", fists/shove) executed cleanly.

7. **Project Test Execution**:
   - Command: `CC=gcc go test -v -run "TestOrthogonal|TestCamera|TestChallenger" ./internal/game` -> **PASS** (all subtests passed).
   - Command: `CC=gcc go test -count=1 ./...` -> **PASS** across all packages (`internal/assets`, `internal/ecs`, `internal/game`, `internal/game/world`).

## 2. Logic Chain

1. Observations (1) and (2) demonstrate that `WorldToIso`, `IsoToWorld`, `ScreenToWorld`, and `WorldToScreen` are strict bijective affine Cartesian transformations that preserve distance ratios, screen centering, and numerical stability up to extreme bounds of $\pm 10^8$.
2. Observation (3) empirically proves that adjacent tile edges and corner vertices calculated via the orthogonal rendering pipeline meet with zero sub-pixel seams ($< 10^{-9}$) across 10,000 pseudo-random tile pairs and floating-point camera positions, guaranteeing the elimination of black gaps or diamond voids.
3. Observation (4) establishes that the camera system exhibits robust initialization, exponential convergence, and a precise $< 0.01$ sub-pixel snapping threshold that prevents micro-jitter and asymptotic drift.
4. Observation (5) confirms that Y-depth sorting is mathematically monotonic, stable under duplicate keys, and correctly implements top-down vertical occlusion for all standing props, items, and entities.
5. Observation (6) proves that Bezier attack swing arcs project without isometric distortion or shearing, and behave predictably across all orientations and weapon categories.
6. Observation (7) verifies that all existing and newly added stress harnesses compile and pass 100% cleanly.

## 3. Caveats

- Tests were run in headless Ebitengine execution mode without hardware GPU acceleration; however, all vector rasterization and image transformation pathways were fully exercised.
- Extreme coordinate fuzzing was conducted up to $\pm 10^8$, well exceeding the playable 100x100 tile world boundaries ($100 \times 128 = 12,800$ units).
- No caveats affecting production gameplay stability were identified.

## 4. Conclusion

**Verdict: APPROVE**

The 2D Orthogonal Engine Overhaul meets all mathematical, rendering, camera tracking, and depth sorting requirements. Coordinate transformations are bijective and robust, tile adjacencies are seamless with 0 black gaps across 10,000 edge tests, camera sub-pixel snapping is stable, Y-depth sorting is strictly monotonic, and Bezier combat arcs project accurately.

## 5. Verification Method

To independently verify this report, execute the following commands in `/home/bryce/code/go-zomboid`:

```bash
# 1. Run targeted orthogonal, camera, and challenger stress test suite
CC=gcc go test -v -run "TestOrthogonal|TestCamera|TestChallenger" ./internal/game

# 2. Run all repository tests uncached
CC=gcc go test -count=1 ./...
```

Inspect files:
- `/home/bryce/code/go-zomboid/internal/game/orthogonal_stress_challenger_test.go`
- `/home/bryce/code/go-zomboid/internal/game/game.go`
