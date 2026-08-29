# Milestone 1 Adversarial Challenge Report & Validation Handoff

## 1. Observation

1. **R1 Procedural Pipeline Retirement**:
   - Directory `/home/bryce/code/go-zomboid/cmd/tools/genassets` does not exist on disk (`ls cmd/tools/genassets` -> `No such file or directory`).
   - Root binary `/home/bryce/code/go-zomboid/genassets` does not exist on disk.
   - Command `go run ./cmd/tools/genassets` fails empirically with:
     `stat /home/bryce/code/go-zomboid/cmd/tools/genassets: directory not found` (exit status 1).

2. **R2 External Asset Ingestion & Embedded FS Defect (CRITICAL)**:
   - On disk, `context/` contains 579 PNG files (590 total files including 8 `.DS_Store` and 3 `.psd`).
   - On disk, `internal/assets/images/` contains 606 PNG files (27 legacy + 579 external) and 0 non-PNG files.
   - **However, Go's `//go:embed images/*` filesystem (`imageFS`) only embeds 603 PNG files (27 legacy + 576 external), omitting 3 external PNG files:**
     - `images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png`
     - `images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0107_Capa-108.png`
     - `images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0108_Capa-109.png`
   - Attempting to read these 3 files from `imageFS` fails with:
     `open images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png: file does not exist`
   - Root cause in Go compiler (`cmd/go/internal/load/pkg.go:2229-2245`):
     Go embed's directory walker calls `isBadEmbedName()`, which invokes `module.CheckFilePath()`. Because `90┬║ Rotatable Bridge Sprites` contains Unicode box-drawing characters `┬` (U+252C) and `║` (U+2551) from an un-sanitized CP437/DOS zip extraction of `90° Rotatable Bridge Sprites`, `module.CheckFilePath()` rejects the directory name and Go silently returns `fs.SkipDir`, skipping embedding for all files in that folder.

3. **R2 Pointer Non-Nil Integrity & Dimensions**:
   - `assets.Load()` successfully initializes all 22 new exported `*ebiten.Image` pointers:
     - `BenchImage` (52x37), `ChestImage` (22x21), `Sculpture1Image` (23x31), `Sculpture2Image` (29x32), `SculptureImage` (23x31)
     - `Bush1Image` (24x18), `Bush2Image` (19x15), `Bush3Image` (25x19), `Bush4Image` (28x19), `BushImage` (24x18)
     - `Flower1Image` (26x25), `Flower2Image` (24x22), `Flower3Image` (26x18), `FlowerImage` (26x25)
     - `Stone1Image` (28x19), `Stone2Image` (29x25), `StoneImage` (28x19), `ForestStumpImage` (29x19)
     - `GrassTuft1Image` (25x24), `GrassTuft2Image` (31x15)
     - `LabTilesetImage` (768x768), `ZombieTilesetImage` (764x300)
   - All 27 legacy pointers (`PlayerImage`, `GrassImage`, `WallImage`, `WeaponImage`, etc.) remain non-nil with exact dimensions.

4. **Concurrency & Idempotency**:
   - Concurrency stress test (`TestEmpiricalM1_ConcurrentLoadIdempotencyStress`) executed 100 concurrent goroutines invoking `assets.Load()` simultaneously across 50 iterations with 0 data races, 0 nil pointers, and 0 panics.

5. **Test Suite Execution**:
   - `CC=gcc go test ./internal/assets/...` fails with:
     ```
     --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages (0.03s)
         --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages/Zombie_Apocalypse_Tileset/Organized_separated_sprites/90┬║_Rotatable_Bridge_Sprites/Zombie-Tileset---_0106_Capa-107.png (0.00s)
             m1_adversarial_challenger_test.go:58: missing embedded PNG in imageFS: images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png: open images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png: file does not exist
     --- FAIL: TestEmpiricalM1_EmbeddedFSStructureAndCount (0.00s)
         m1_adversarial_challenger_test.go:127: expected exactly 606 embedded PNGs (27 legacy + 579 external), found 603 (disk has 606)
     ```

## 2. Logic Chain

1. Requirement R2 dictates ingesting all external PNG assets from `context/` into `internal/assets/images/` such that the game can natively load these assets.
2. The runtime asset loader relies on Go's `//go:embed images/*` directive to access all embedded PNG files via `imageFS`.
3. Observation #2 reveals that although 579 files were copied to disk, Go's embed toolchain rejects the directory name `90┬║ Rotatable Bridge Sprites` due to invalid box-drawing characters `┬` and `║` under `module.CheckFilePath()`.
4. As a direct result, 3 PNG files are silently excluded from `imageFS`, causing runtime `ReadFile` failures and leaving only 576 out of 579 external PNGs accessible in the binary.
5. While R1 (retirement of genassets), pointer initialization, and concurrency are fully compliant, the asset embedding pipeline fails Requirement 1 of the milestone ("Verify all 579 PNG files from context/ exist and are valid readable PNGs in internal/assets/images/").
6. Therefore, Milestone 1 must be REJECTED until the worker sanitizes the directory name to a valid path (e.g. `90 Rotatable Bridge Sprites` or `90_Rotatable_Bridge_Sprites` or `90-deg Rotatable Bridge Sprites` in both `context/` and `internal/assets/images/` or in `internal/assets/images/`).

## 3. Caveats

- All 22 new pointers needed for Milestone 2 (`world/map.go`) and Milestone 3 (`game.go`) are fully loaded and operational; the 3 omitted files belong to bridge sprites in the Zombie Apocalypse Tileset pack, which are not currently referenced in `assets.go`'s top-level variables but violate the contract of complete asset ingestion.
- The 27 legacy assets are completely unaffected and 100% operational.

## 4. Conclusion

**Milestone 1 Empirical Confirmation Verdict**: **REJECT**

### Actionable Fix for Worker:
1. Rename the directory `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites` to a valid ASCII/UTF-8 path without box-drawing runes (e.g. `90 Rotatable Bridge Sprites` or `90_Rotatable_Bridge_Sprites` or `90-deg Rotatable Bridge Sprites`).
2. (Optional / recommended) Update `context/Zombie Apocalypse Tileset/Organized separated sprites/` directory similarly to keep 100% path parity with `context/`.
3. Re-run `CC=gcc go test -v ./internal/assets/...` and verify all 606 embedded files pass.

## 5. Verification Method

To independently reproduce the failure and verify after fix:

```bash
# 1. Run the empirical adversarial test suite
CC=gcc go test -v -run "TestEmpiricalM1" ./internal/assets/...

# 2. Inspect the exact missing embedded files
python3 -c "
import subprocess, json
d = json.loads(subprocess.run(['go', 'list', '-json', './internal/assets'], capture_output=True, text=True).stdout)
print('Embedded count:', len(d.get('EmbedFiles', [])))
print('Rotatable files in EmbedFiles:', [f for f in d.get('EmbedFiles', []) if 'Rotatable' in f])
"

# 3. Full package test
CC=gcc go test -v ./internal/assets/...
```
