# Milestone 1 Empirical Challenge & Verification Report

## 1. Observation

1. **Procedural Generation Pipeline Retirement (R1)**:
   - Verified that `cmd/tools/genassets` directory does not exist on disk (`stat cmd/tools/genassets` -> `No such file or directory`).
   - Verified that root executable `genassets` does not exist on disk (`stat genassets` -> `No such file or directory`).
   - Attempting to execute `go run ./cmd/tools/genassets` fails empirically with:
     ```
     stat /home/bryce/code/go-zomboid/cmd/tools/genassets: directory not found
     ```

2. **Unwanted File Ingestion Scan**:
   - Executed scan across `internal/assets/images/`:
     `find internal/assets/images/ -name "*DS_Store*" -o -name "*.psd*" -o -name "*Zone.Identifier*" -o -name "*Thumbs.db*" -o -name ".*"`
   - Result: Exactly 0 unwanted files found. All 8 `.DS_Store` files, 3 `.psd` files, and zone identifier streams from `context/` were completely excluded.

3. **External Image Dimensions & Alpha Channel Integrity**:
   - Inspected all 579 PNG files in `context/` and `internal/assets/images/`:
     - 0 corrupted files.
     - 0 zero-byte files.
     - 100% of files have valid PNG magic header bytes `\x89PNG\r\n\x1a\n`.
     - 100% of files have positive dimensions ($W > 0, H > 0$).
     - 100% of files contain non-zero alpha pixels (0 files are 100% transparent).

4. **Pointer Accessibility, Dimensions & Concurrency**:
   - `assets.Load()` successfully initializes all 27 legacy pointers (`PlayerImage`, `GrassImage`, `WallImage`, `WeaponImage`, etc.) with exact legacy dimensions (Entities: 64x128, Floors: 256x128, Obstacles: 256x256, Items: 64x64).
   - `assets.Load()` successfully initializes all 22 new pointers:
     - `BenchImage` (52x37)
     - `ChestImage` (22x21)
     - `Sculpture1Image` (23x31), `Sculpture2Image` (29x32), `SculptureImage` (23x31)
     - `Bush1Image` (24x18), `Bush2Image` (19x15), `Bush3Image` (25x19), `Bush4Image` (28x19), `BushImage` (24x18)
     - `Flower1Image` (26x25), `Flower2Image` (24x22), `Flower3Image` (26x18), `FlowerImage` (26x25)
     - `Stone1Image` (28x19), `Stone2Image` (29x25), `StoneImage` (28x19), `ForestStumpImage` (29x19)
     - `GrassTuft1Image` (25x24), `GrassTuft2Image` (31x15)
     - `LabTilesetImage` (768x768), `ZombieTilesetImage` (764x300)
   - Concurrency stress test (`TestChallenger_ConcurrentLoadStress` and `TestEmpiricalM1_ConcurrentLoadIdempotencyStress`) verified that 100 concurrent goroutines executing 50 calls each to `Load()` suffer 0 data races, 0 nil pointers, and 0 panics.

5. **Defect: Go embed Toolchain Rejects Non-ASCII Folder Name (CRITICAL)**:
   - On disk in `internal/assets/images/`: 606 total PNG files exist (27 legacy + 579 external).
   - In Go `embed.FS` (`imageFS`): Only 603 total PNG files are embedded (27 legacy + 576 external).
   - The Go compiler (`//go:embed images/*`) silently skipped the entire subdirectory:
     `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/`
   - Files omitted from the binary `imageFS`:
     1. `images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png`
     2. `images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0107_Capa-108.png`
     3. `images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0108_Capa-109.png`
   - Attempting to read these files from `imageFS` produces:
     `open images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png: file does not exist`
   - Test suite execution:
     `CC=gcc go test -v ./internal/assets/...` fails on `TestEmpiricalM1_All579ContextPNGsMatchImages` and `TestEmpiricalM1_EmbeddedFSStructureAndCount` due to missing embedded files.

## 2. Logic Chain

1. Requirement R2 mandates copying all external PNG assets from `context/` into `internal/assets/images/` and ensuring they can be loaded by Go.
2. The runtime uses Go's standard `//go:embed images/*` to embed assets into `imageFS`.
3. In Go's embed package implementation (`cmd/go/internal/load/pkg.go`), directory paths are validated against `module.CheckFilePath()`. Directory paths with non-ASCII or invalid path characters (such as the CP437/DOS box-drawing runes `┬` (U+252C) and `║` (U+2551) in `90┬║ Rotatable Bridge Sprites`) are rejected as invalid embed paths, causing Go to skip the directory during compilation.
4. Because the directory was skipped by Go, 3 external PNG files are physically present on disk but absent from the compiled Go binary's embedded filesystem.
5. As observed in Observation #5, any test or game logic attempting to load all external assets fails with `file does not exist`.
6. Therefore, the external asset ingestion acceptance criterion is not fully satisfied until the directory path is sanitized to standard ASCII/UTF-8.

## 3. Caveats

- All 22 new pointers needed for Milestone 2 (`internal/game/world`) and Milestone 3 (`internal/game`) are properly loaded and functional; the 3 omitted files belong to bridge sprites in the Zombie Apocalypse Tileset pack, which are not currently bound to top-level variables.
- Legacy asset pointers (27 total) remain 100% backwards compatible and pass all regression and stress tests.

## 4. Conclusion

**Verdict**: **REJECT**

Milestone 1 cannot be approved in its current state because 3 external PNG assets fail to embed into the binary due to an invalid directory name, causing package test failure.

### Action Plan for Worker:
1. Rename the directory `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites` to a valid ASCII name (e.g. `90 Rotatable Bridge Sprites` or `90_Rotatable_Bridge_Sprites` or `90-deg Rotatable Bridge Sprites`).
2. (Optional/Recommended) Update `context/Zombie Apocalypse Tileset/Organized separated sprites/` similarly to maintain exact relative path parity.
3. Verify that `CC=gcc go test -v ./internal/assets/...` and `CC=gcc go test -v ./...` pass 100% with all 606 embedded files.

## 5. Verification Method

To independently verify the failure and test the fix:

```bash
# 1. Run the empirical asset test suite
CC=gcc go test -v ./internal/assets/...

# 2. Check the embedded file count vs disk count
python3 -c "
import os, json, subprocess
d = json.loads(subprocess.run(['go', 'list', '-json', './internal/assets'], capture_output=True, text=True).stdout)
embed_count = len(d.get('EmbedFiles', []))
print(f'Embedded files: {embed_count} / 606')
assert embed_count == 606, f'Expected 606 embedded files, got {embed_count}'
"

# 3. Verify all tests across repo
CC=gcc go test ./...
```
