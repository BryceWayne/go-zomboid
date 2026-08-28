# Handoff Report: Milestone 1 Items / Weapons / Equipment (64x64) & Asset Test Suite Analysis

**Agent**: `m1_explorer_3`
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_3`
**Milestone**: Milestone 1 (High-Fidelity Asset Generation 4x Scaling)
**Target Files Analyzed**:
- `cmd/tools/genassets/main.go`
- `cmd/tools/genassets/genassets_test.go`
- `internal/assets/assets.go`
- `internal/assets/assets_test.go`
- `internal/assets/assets_stress_test.go`

---

## 1. Observation

Direct observations from codebase inspection:
1. **Item Generators in `cmd/tools/genassets/main.go`**:
   - `generateFood` (lines 1002–1083), `generateWater` (lines 1085–1167), `generateWeapon` (lines 1169–1257), `generateAxe` (lines 1259–1365), `generateShotgun` (lines 1367–1446), `generateAmmo` (lines 1448–1514), `generateArmor` (lines 1516–1610), and `generateAntidote` (lines 1799–1850) currently initialize canvases with `image.NewRGBA(image.Rect(0, 0, 16, 16))`.
   - Drawing routines place coordinates in the range $[0..15]$, lacking the high-resolution vector details (e.g. concentric lid rings, pull-tab ring cutouts, contoured ergonomic bottle waist ribs, woodgrain sweeps, multi-spike clusters, gun receiver milling, brass bullet rows, laser-cut MOLLE webbing, and fluid meniscus/air bubbles) required by the 4x vector art style.
2. **Asset Registration in `internal/assets/assets.go`**:
   - Lines 17–51 declare global image handles for 27 assets:
     - 3 Character Entities: `PlayerImage`, `ZombieImage`, `RunnerImage`
     - 6 Floor Tiles: `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage`
     - 10 Vertical Obstacles/Props: `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`, `TentImage`, `StumpImage`, `MushroomImage`, `SignImage`, `ElevationBlockImage`, `ElevationRampImage`
     - 8 Items/Equipment: `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `AntidoteImage`, `FoodImage`, `WaterImage`
   - `Load()` (lines 53–88) loads all 27 images into these pointer variables from `images/*.png`.
3. **Unit Tests in `internal/assets/assets_test.go`**:
   - `TestEmbeddedAssetDimensionsAndValidity` (lines 12–93) only tests 20 assets, asserting dimensions: 16x32 for entities, 64x32 for floors, 64x64 for obstacles, and 16x16 for 7 items (`antidote.png` and 6 props omitted). Contains assertion `if len(expectedAssets) != 20`.
   - `TestAssetsLoadAllPointersNonNil` (lines 95–150) tests only 20 handles, asserting dimensions 16x32, 64x32, 64x64, 16x16. Contains assertion `if len(handles) != 20`.
4. **Stress Tests in `internal/assets/assets_stress_test.go`**:
   - `TestFloorTileIsometricBounds` (lines 13–66) tests 6 floor tiles against 64x32 bounding box, diamond center $(31.5, 15.5)$, and radii $(32.5, 16.5)$.
   - `TestCharacterGroundAnchor` (lines 68–105) asserts grounding pixels in rows $y \in [28..31]$ on a 16x32 canvas.
   - `TestItemOutlineContrast` (lines 107–159) loops over $16 \times 16$ bounds for 7 items (`antidote.png` omitted) and asserts `solidCount >= 20` and `darkContourCount > 0`.
   - `TestAssetsLoadIdempotency` (lines 161–176) only checks 20 asset pointers for non-nil.
5. **Generator Determinism in `cmd/tools/genassets/genassets_test.go`**:
   - `expectedAssetFiles` (lines 15–48) defines 20 assets with 16x32, 64x32, 64x64, and 16x16 dimensions.
   - `TestAssetDimensionsAndIntegrity` (lines 113–167) checks `len(expectedAssetFiles) != 20` and checks 5% fill density.

---

## 2. Logic Chain

1. **4x Resolution Requirement**:
   - Per `ORIGINAL_REQUEST.md` §R1 and `PROJECT.md` Feature 4 & Interface Contracts, floor tiles scale from 64x32 to 256x128, vertical obstacles/props scale to 256x256, character entities scale from 16x32 to 64x128, and items scale from 16x16 to 64x64.
2. **Item Procedural Generation Enhancement**:
   - With a 64x64 canvas, item area expands 16-fold ($16 \times 16 = 256 \to 64 \times 64 = 4096$).
   - Each item generator must be rewritten to generate rich, anti-aliased, multi-tone pixel/vector art:
     - `food.png`: Soup can with top lid ellipse, pull tab ring, red/gold label with cylindrical lighting, tomato emblem, metallic rims.
     - `water.png`: Contoured sports bottle with threaded cap, meniscus fill line, translucent cyan/blue volume gradient, grip ribs, air bubbles, base feet.
     - `weapon.png`: Diagonal wooden bat with athletic tape spiral wraps, ash woodgrain barrel, 6 steel spikes with collars, blood splatters.
     - `axe.png`: Curved hickory handle, rubberized base grip, steel socket/wedge, rear breaching pick, red enamel blade, mirror-polished cutting edge.
     - `shotgun.png`: Diagonal pump shotgun with walnut stock, rubber buttpad, blued steel receiver with ejection port, ribbed forend, dual barrel/mag tubes, brass bead sight.
     - `ammo.png`: Olive drab ammo can with yellow stencil text, stamped side panel, 4 corner rivets, 6 brass cartridges with copper tips protruding from top.
     - `armor.png`: Ballistic Kevlar vest with shoulder straps, buckles, scooped neckline, Velcro ID patch, 3 MOLLE webbing rows, 3 magazine pouches with pull tabs.
     - `antidote.png`: Laboratory glass vial with cork stopper, flanged lip, narrow neck, glowing neon-green fluid with meniscus, rising bubbles, measurement graduations, glass highlights.
   - All items must apply `addSelectiveOutline(img, darkBorder)` for crisp contrast against any floor type.
3. **Synchronized Test Suite Updates**:
   - As assets scale to their 4x dimensions, running `go test ./...` will fail unless test assertions are updated.
   - `internal/assets/assets_test.go`:
     - Update all 27 assets in `expectedAssets` to their new dimensions (256x128 floors, 256x256 props, 64x128 entities, 64x64 items). Update count check to 27.
     - Update all 27 handles in `handles` to target dimensions. Update count check to 27.
   - `internal/assets/assets_stress_test.go`:
     - Update `TestFloorTileIsometricBounds` for 256x128 ($W=256, H=128$, center $(127.5, 63.5)$, radii $(128.5, 64.5)$).
     - Update `TestCharacterGroundAnchor` to check bottom 16 rows ($y \in [112..127]$) on 64x128 entities.
     - Update `TestItemOutlineContrast` to loop $64 \times 64$, include all 8 items (adding `antidote.png`), and update solid pixel threshold to $\ge 320$ pixels.
     - Update `TestAssetsLoadIdempotency` to check all 27 pointers.
   - `cmd/tools/genassets/genassets_test.go`:
     - Update `expectedAssetFiles` to all 27 assets with target dimensions. Update count check to 27.

---

## 3. Caveats

1. **Test-First vs Generation Sequence**:
   - If tests are updated before running `go run ./cmd/tools/genassets`, tests will fail on existing 16x16/64x32 PNG assets.
   - If `main.go` is executed first to generate 4x PNGs, `internal/assets/assets_test.go` will fail until test assertions are updated.
   - Therefore, the implementation in worker tasks should update generator functions in `main.go`, regenerate assets with `go run ./cmd/tools/genassets`, and update test files in tandem.
2. **Game Loop Offsets**:
   - In `internal/game/game.go`, item rendering offset is `drawX = isoX - 8 - camX`, `drawY = isoY - 8 - camY`. For 64x64 items, Milestone 2 will update this to `drawX = isoX - 32 - camX`, `drawY = isoY - 32 - camY`. Milestone 1 focuses on asset generation and asset test validation.

---

## 4. Conclusion

- A complete mathematical and procedural art specification has been constructed for all 8 items (`food.png`, `water.png`, `weapon.png`, `axe.png`, `shotgun.png`, `ammo.png`, `armor.png`, `antidote.png`) on 64x64 canvases in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_3/m1_items_tests_analysis.md`.
- All necessary test suite modifications across `internal/assets/assets_test.go`, `internal/assets/assets_stress_test.go`, and `cmd/tools/genassets/genassets_test.go` have been mapped out for all 27 assets.

---

## 5. Verification Method

Once implemented by worker agents:
1. `go run ./cmd/tools/genassets`
   - Must execute without errors and output "Asset generation completed successfully."
2. `CC=gcc go test -v ./cmd/tools/genassets/...`
   - Must pass all SHA-256 determinism and dimension tests for 27 assets.
3. `CC=gcc go test -v ./internal/assets/...`
   - Must pass `TestEmbeddedAssetDimensionsAndValidity`, `TestAssetsLoadAllPointersNonNil`, `TestFloorTileIsometricBounds`, `TestCharacterGroundAnchor`, `TestItemOutlineContrast`, and `TestAssetsLoadIdempotency`.
4. `CC=gcc go test ./...`
   - All unit and stress tests across the entire repository must pass cleanly.
