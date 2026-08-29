# Technical Survey Handoff Report

**Agent**: `teamwork_preview_explorer_survey_1`  
**Date**: 2026-08-29  
**Task**: In-depth Technical Survey of Asset Pipeline (Context assets, Genassets references, Internal/Assets architecture, and R1/R2 refactoring plan)  

---

## 1. Observation

1. **Context Directory Assets**:
   - Total files found in `/home/bryce/code/go-zomboid/context/`: 590 non-stream files.
   - Breakdown by format: 579 PNG files, 3 Photoshop PSD files (`Zombie Apocalypse Tileset.psd`, `INTERFACE WITH NEW INVENTORY.psd`, `SLOT AND ITEMS.psd`), 8 macOS `.DS_Store` files, and Windows `:Zone.Identifier` stream files.
   - `Lab/`: 1 PNG tileset sheet (`Inside_C.png`, 768x768, 100,232 bytes).
   - `Small Forest/`: 45 PNG files across 9 subdirectories (Bench and chest: `Bench.png` 52x37, `Chest.png` 22x21; Bushes: 5 files 19x15 to 29x19; Fences: 12 files 13x32 to 64x23; Flowers: 3 files 24x22 to 26x25; Grass: 2 files; Ground tileset: 5 files 37x182 to 365x331; Sculptures: 2 files 23x31 and 29x32; Stones: 2 files 28x19 and 29x25; Trees: 12 files 15x18 to 55x67 across 3 species and 4 stages).
   - `Zombie Apocalypse Tileset/`: 533 PNG files (1 reference sheet `Zombie Apocalypse Tileset Reference.png` 764x300 + 532 separated PNG sprites across 46 subdirectories), 3 PSD source files, and 8 `.DS_Store` files.

2. **Procedural Pipeline (`cmd/tools/genassets`) & Repo Footprint**:
   - `cmd/tools/genassets/main.go` (2,413 lines) contains procedural algorithms generating 27 PNG assets into `internal/assets/images/`.
   - `cmd/tools/genassets/genassets_test.go` (119 lines) tests determinism and PNG dimensions.
   - Binary at repository root: `/home/bryce/code/go-zomboid/genassets` (ELF executable, 2.37 MB).
   - References across the codebase:
     - `README.md:10, 40`
     - `PROJECT.md:10, 34, 41, 67, 68`
     - `TEST_READY.md:40`
     - `TEST_INFRA.md:30`
     - `ART_STYLE_GUIDE.md:4, 12`
     - `internal/assets/empirical_challenger_test.go:302-357`: `TestEmpiricalGenerationDeterminism` invokes `exec.Command("go", "run", "./cmd/tools/genassets")`.

3. **Current Asset Loading Architecture (`internal/assets/`)**:
   - `internal/assets/assets.go` uses `//go:embed images/*` to embed all image assets into `imageFS embed.FS`.
   - `Load()` initializes 27 exported `*ebiten.Image` pointers (`PlayerImage`, `ZombieImage`, `RunnerImage`, 6 floor tile pointers, 10 vertical obstacle pointers, 8 item/weapon pointers) using `loadEbitenImage()`.
   - Go's `//go:embed images/*` recursively includes subdirectories and nested files without syntax modifications.
   - All tests in `internal/assets/` (`assets_test.go`, `assets_stress_test.go`, `challenger_stress_test.go`, `empirical_challenger_test.go`) pass with `CC=gcc go test -v ./internal/assets/...`.

---

## 2. Logic Chain

1. **Retiring Procedural Generation (R1)**:
   - *Premise*: Requirement R1 requires complete deletion of `cmd/tools/genassets`.
   - *Inference*: Removing `cmd/tools/genassets` will break `TestEmpiricalGenerationDeterminism` in `internal/assets/empirical_challenger_test.go` because that test directly executes `go run ./cmd/tools/genassets`.
   - *Deduction*: `TestEmpiricalGenerationDeterminism` must be removed or retired as part of R1 so that `CC=gcc go test ./...` passes cleanly.
   - *Cleanup*: The compiled root binary `./genassets` and documentation references should be cleaned up simultaneously.

2. **Ingesting External PNG Assets (R2)**:
   - *Premise*: Requirement R2 requires copying external PNG files from `context/` into `internal/assets/images/` and updating `internal/assets/assets.go` to load them into `ebiten.Image` variables.
   - *Inference*: `context/` contains 579 PNG files, 3 PSD files, and 8 `.DS_Store` files.
   - *Deduction*: Ingestion must copy PNG files while ignoring `.DS_Store`, `*.psd`, and `:Zone.Identifier`.
   - *Inference*: Existing game systems and test suites rely on the 27 legacy pointers (`PlayerImage`, `GrassImage`, `WallImage`, etc.).
   - *Deduction*: Retaining existing 27 PNG assets and adding the new PNG assets from `context/` ensures 100% backwards compatibility with existing unit tests while exposing new `*ebiten.Image` variables (`BenchImage`, `ChestImage`, `SculptureImage`, `BushImage`, `FlowerImage`, `StoneImage`, `LabTilesetImage`, `ZombieTilesetImage`).

---

## 3. Caveats

- **No Caveats**: All 590 files in `context/` were fully enumerated, decoded, dimension-checked, and cataloged. Go `//go:embed` recursion was empirically verified via standard toolchain execution.

---

## 4. Conclusion

The asset pipeline survey is complete. The roadmap for R1 and R2 is clearly defined:
1. **R1**: Delete `cmd/tools/genassets/`, delete root `./genassets`, retire `TestEmpiricalGenerationDeterminism` in `empirical_challenger_test.go`, and update documentation.
2. **R2**: Copy all 579 PNG files from `context/` to `internal/assets/images/` (omitting PSDs, `.DS_Store`, and `:Zone.Identifier`), declare new `*ebiten.Image` variables in `internal/assets/assets.go`, and load them inside `Load()`.

Full technical details, catalogs, dimension matrices, and code plans are documented in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1/survey.md`.

---

## 5. Verification Method

To independently verify the findings in this report:

1. **Verify Context Directory Inventory**:
   ```bash
   python3 -c "
   import os
   cnt = sum(1 for root, _, files in os.walk('/home/bryce/code/go-zomboid/context') for f in files if not f.endswith(':Zone.Identifier'))
   pngs = sum(1 for root, _, files in os.walk('/home/bryce/code/go-zomboid/context') for f in files if f.endswith('.png') and not f.endswith(':Zone.Identifier'))
   print(f'Total: {cnt}, PNGs: {pngs}')
   "
   # Output: Total: 590, PNGs: 579
   ```

2. **Verify Codebase References to `genassets`**:
   ```bash
   grep -rn --exclude-dir=".agents" --exclude-dir=".git" "genassets" /home/bryce/code/go-zomboid
   ```

3. **Verify Existing Asset Tests**:
   ```bash
   CC=gcc go test -v ./internal/assets/...
   ```
