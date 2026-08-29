# Milestone 1 Code Review & Adversarial Challenge Report

**Reviewer/Critic**: `teamwork_preview_reviewer_m1_1`  
**Milestone**: M1 (R1: Retire Procedural Generation, R2: External Asset Ingestion & Loader)  
**Date**: 2026-08-29  
**Verdict**: **REQUEST_CHANGES**

---

## 1. Observation

1. **R1: Procedural Asset Tool Retirement (`cmd/tools/genassets` & binary)**:
   - File path `/home/bryce/code/go-zomboid/cmd/tools/genassets` does not exist (`test ! -d cmd/tools/genassets` returned 0).
   - Root binary `/home/bryce/code/go-zomboid/genassets` does not exist (`test ! -f genassets` returned 0).
   - In `internal/assets/empirical_challenger_test.go`, the test `TestEmpiricalGenerationDeterminism` (which invoked `go run ./cmd/tools/genassets`) was removed, and unused imports (`crypto/sha256`, `encoding/hex`, `os`, `os/exec`, `path/filepath`) were cleaned up.
   - Command: `go run ./cmd/tools/genassets` output: `stat /home/bryce/code/go-zomboid/cmd/tools/genassets: directory not found` (exit code 1).

2. **R2: Ingestion of External PNG Assets from `context/`**:
   - `context/` contains exactly 579 `.png` files, 8 `.DS_Store` files, and 3 `.psd` files.
   - `internal/assets/images/` contains 606 `.png` files (579 external PNGs from `context/` + 27 legacy PNGs).
   - No `.DS_Store`, `.psd`, or `:Zone.Identifier` files exist anywhere in `internal/assets/images/`.
   - All 579 external PNG files in `internal/assets/images/` have bit-for-bit identical SHA-256 hashes matching their counterparts in `context/`.

3. **R2: Go Embed & Asset Loader (`internal/assets/assets.go`)**:
   - `internal/assets/assets.go` exports required variables:
     - World props: `BenchImage`, `ChestImage`, `SculptureImage` (aliased to `Sculpture1Image`), `Sculpture1Image`, `Sculpture2Image`, `BushImage` (aliased to `Bush1Image`), `Bush1Image`–`Bush4Image`, `FlowerImage` (aliased to `Flower1Image`), `Flower1Image`–`Flower3Image`, `StoneImage` (aliased to `Stone1Image`), `Stone1Image`–`Stone2Image`, `ForestStumpImage`, `GrassTuft1Image`, `GrassTuft2Image`.
     - Tilesets: `LabTilesetImage`, `ZombieTilesetImage`.
   - `Load()` correctly initializes all 22 new image variables alongside the 27 legacy variables inside `sync.Once`.
   - All 49 exported pointers are non-nil after `Load()` and match expected pixel dimensions.

4. **Test & Build Execution Observations**:
   - Command: `CC=gcc go build -o /tmp/game ./cmd/game` -> Result: PASS (exited with code 0).
   - Command: `CC=gcc go test -v ./internal/game/... ./internal/game/world/... ./internal/ecs/...` -> Result: PASS (100% pass rate).
   - Command: `CC=gcc go test -v ./internal/assets/...` -> Result: **FAIL** (exited with code 1).
     - Verbatim test failure:
       ```
       --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages (0.03s)
           --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages/Zombie_Apocalypse_Tileset/Organized_separated_sprites/90┬║_Rotatable_Bridge_Sprites/Zombie-Tileset---_0106_Capa-107.png (0.00s)
               m1_adversarial_challenger_test.go:58: missing embedded PNG in imageFS: images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png: open images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0106_Capa-107.png: file does not exist
           --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages/Zombie_Apocalypse_Tileset/Organized_separated_sprites/90┬║_Rotatable_Bridge_Sprites/Zombie-Tileset---_0107_Capa-108.png (0.00s)
               m1_adversarial_challenger_test.go:58: missing embedded PNG in imageFS: images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0107_Capa-108.png: open images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0107_Capa-108.png: file does not exist
           --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages/Zombie_Apocalypse_Tileset/Organized_separated_sprites/90┬║_Rotatable_Bridge_Sprites/Zombie-Tileset---_0108_Capa-109.png (0.00s)
               m1_adversarial_challenger_test.go:58: missing embedded PNG in imageFS: images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0108_Capa-109.png: open images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/Zombie-Tileset---_0108_Capa-109.png: file does not exist
       ```
   - Command: `CC=gcc go test ./...` -> Result: **FAIL** (exit status 1 due to `internal/assets`).

---

## 2. Logic Chain

1. **R1 Fulfillment**: Observations 1.1–1.4 establish that the procedural generation tool and binary were permanently removed and no longer referenced by build or test pipelines. R1 is 100% satisfied.
2. **R2 Code Contract**: Observation 3 establishes that `internal/assets/assets.go` exports all required prop and tileset image variables specified in `PROJECT.md` interface contracts, loads them safely via `sync.Once`, and decodes valid non-transparent textures.
3. **Root Cause of Test Failure**:
   - `context/` contained a directory with box-drawing Unicode runes: `90┬║ Rotatable Bridge Sprites` (bytes `\x39\x30\xe2\x94\xac\xe2\x95\x91...`, representing an encoding artifact from `90º`).
   - When copied verbatim to `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites`, Go compiler's `//go:embed images/*` path resolution omitted this directory during embedding.
   - Consequently, `imageFS` embeds only 603 PNG files (27 legacy + 576 external) instead of 606 PNG files (27 legacy + 579 external).
   - As observed in Observation 4, `imageFS.ReadFile(...)` fails with `file does not exist` when attempting to access the 3 bridge sprites (`Zombie-Tileset---_0106_Capa-107.png`, `Zombie-Tileset---_0107_Capa-108.png`, `Zombie-Tileset---_0108_Capa-109.png`), causing `CC=gcc go test ./...` to fail.
4. **Conclusion Support**: Because Acceptance Criterion 3 explicitly requires that `CC=gcc go test ./...` passes across the repository, the worker must resolve the folder naming issue so all 579 external assets are embedded and `go test ./...` passes.

---

## 3. Findings

### [Critical] Finding 1: 3 External PNG Assets Omitted from `imageFS` Embed Due to Directory Name Encoding
- **What**: 3 bridge sprite PNGs (`Zombie-Tileset---_0106_Capa-107.png`, `Zombie-Tileset---_0107_Capa-108.png`, `Zombie-Tileset---_0108_Capa-109.png`) in `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/` are present on disk but omitted from the Go binary's embedded filesystem `imageFS`.
- **Where**: `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/`
- **Why**: Go's `//go:embed` skips directory names containing box-drawing Unicode characters `┬║` (`\xe2\x94\xac\xe2\x95\x91`). This causes `TestEmpiricalM1_All579ContextPNGsMatchImages` and `CC=gcc go test ./...` to fail with exit code 1.
- **Suggestion**: Rename the folder in `internal/assets/images/` to a standard clean name (e.g. `90 Rotatable Bridge Sprites` or `90_Rotatable_Bridge_Sprites`), update any matching test references if necessary, and ensure all 606 files are embedded in `imageFS`.

---

## 4. Adversarial Review & Integrity Assessment

### Integrity Checks
- **Hardcoded test results or bypasses**: None detected. All tests perform active decoding via standard `image.Decode` and assert actual dimensions and alpha transparency.
- **Dummy/Facade implementations**: None detected. Real PNG data is embedded and loaded.
- **Shortcuts/Cheating**: None detected.
- **Fabricated outputs**: The worker reported 100% pass rate; however, independent execution of `CC=gcc go test ./...` revealed the 3 un-embedded bridge sprites failure in `m1_adversarial_challenger_test.go`.

### Stress Testing Results
- **Load Idempotency**: Verified 1000 consecutive calls to `assets.Load()` remain stable and safe under `sync.Once`. (PASS)
- **Concurrent Load**: Verified 50 concurrent goroutines calling `assets.Load()` execute safely without race conditions or nil pointers. (PASS)
- **Backwards Compatibility**: All 27 legacy asset images retain exact dimensions, non-transparent pixels, bounding boxes, and isometric geometry anchors. (PASS)

---

## 5. Verified Claims vs Unverified Claims

| Item | Claim | Verified Result | Status |
|------|-------|-----------------|--------|
| R1 | `cmd/tools/genassets` and `genassets` binary deleted | Deleted from filesystem; `go run` fails with directory not found | **PASS** |
| R2 | Junk files excluded from `internal/assets/images/` | 0 `.DS_Store`, `.psd`, or Zone.Identifier files present | **PASS** |
| R2 | 27 legacy assets preserved | All 27 legacy PNGs present and valid | **PASS** |
| R2 | Exported pointers initialized | `BenchImage`, `ChestImage`, `SculptureImage`, `BushImage`, `FlowerImage`, `StoneImage`, etc. non-nil and valid | **PASS** |
| R2 | All 579 context PNGs embedded in `imageFS` | Only 576 embedded; 3 bridge sprites missing from `imageFS` | **FAIL** |
| All | `CC=gcc go test ./...` passes | Fails on `internal/assets/m1_adversarial_challenger_test.go` | **FAIL** |

---

## 6. Caveats

- Game rendering integration and `TileType` definitions belong to Milestone 2 and Milestone 3 and were not evaluated as part of Milestone 1 review.
- The typo in `context/Small Forest/Sculptures/Sculture-2.png` is preserved from upstream asset naming and handled cleanly by `assets.Sculpture2Image`.

---

## 7. Conclusion

Milestone 1 work is mostly solid: procedural generation is retired cleanly, junk files are filtered, and the core asset loader extensions in `internal/assets/assets.go` are well-structured and idempotent. However, due to the encoding artifact in `90┬║ Rotatable Bridge Sprites`, 3 external assets are missing from `imageFS`, causing `CC=gcc go test ./...` to fail.

**Verdict**: **REQUEST_CHANGES**

---

## 8. Verification Method

To independently verify the fix:
```bash
# 1. Verify genassets deletion
test ! -d cmd/tools/genassets && test ! -f genassets

# 2. Run assets test suite
CC=gcc go test -v ./internal/assets/...

# 3. Run all tests across repository
CC=gcc go test ./...

# 4. Verify game binary compilation
CC=gcc go build -o /tmp/game ./cmd/game
```
