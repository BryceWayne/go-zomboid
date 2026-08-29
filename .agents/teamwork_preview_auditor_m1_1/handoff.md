# Forensic Audit Report & Handoff: Milestone 1 (R1 & R2)

**Work Product**: Milestone 1 Asset Ingestion & Retirement of genassets (`cmd/tools/genassets`, `internal/assets`)  
**Integrity Mode**: Demo (evaluated against Development, Demo, and Benchmark strictness)  
**Forensic Verdict**: **CLEAN (Integrity)** / **FUNCTIONAL DEFECT NOTED (Embedding)**

---

## 1. Observation

### A. Procedural Generation Retirement (R1)
- **Directory Deletion**: `/home/bryce/code/go-zomboid/cmd/tools/genassets` does not exist on disk (`test ! -d cmd/tools/genassets` exits with code 0).
- **Binary Deletion**: Root binary `/home/bryce/code/go-zomboid/genassets` does not exist on disk (`test ! -f genassets` exits with code 0).
- **Phantom Script Search**: `find . -type f \( -name "*.sh" -o -name "*.py" -o -name "*.go" \)` confirmed no rogue scripts or hidden generation tools exist.
- **Direct Invocation Test**: Executing `go run ./cmd/tools/genassets` returns exit code 1 (`directory not found`).
- **Determinism Test Decoupling**: In `internal/assets/empirical_challenger_test.go`, the test `TestEmpiricalGenerationDeterminism` (which executed `go run ./cmd/tools/genassets`) was cleanly excised.

### B. External Asset Ingestion Authenticity (R2)
- **Asset Count & Hash Comparison**:
  - `context/`: Contains 590 total files (579 PNGs, 8 `.DS_Store`, 3 `.psd`).
  - `internal/assets/images/`: Contains 606 total files (579 external PNGs + 27 legacy PNGs, 0 non-PNG files).
  - **Bit-for-Bit Hash Integrity**: SHA-256 hash comparison across all 579 external PNG files showed a **100.0% exact match** (0 missing, 0 hash mismatches, 0 corrupted files).
  - **No Dummy/Mock Assets**: Dimensional inspection via image decoder confirmed all assets have authentic non-zero dimensions (e.g., `Bench.png` is 52x37 RGBA, `Inside_C.png` is 768x768 RGBA, `Zombie Apocalypse Tileset Reference.png` is 764x300 RGBA, `Sculpture-1.png` is 23x31 RGBA). None are 1x1 dummy pixels.
  - **Purity**: Zero non-PNG files (`.DS_Store`, `.psd`, `:Zone.Identifier`, `.gitkeep`) were copied into `internal/assets/images/`.

### C. Native Asset Loader Implementation (`internal/assets/assets.go`)
- **Genuine Image Loading**: `loadEbitenImage` (lines 147–159) reads raw bytes from `imageFS embed.FS`, decodes standard PNG headers via `image.Decode(bytes.NewReader(data))`, and calls `ebiten.NewImageFromImage(img)`. No dummy mock structs, hardcoded colors, or `return nil` facades exist.
- **Exported Pointers**: Declares and initializes all required `*ebiten.Image` pointers in `Load()`:
  - World Props: `BenchImage`, `ChestImage`, `Sculpture1Image`, `Sculpture2Image`, `SculptureImage`, `Bush1Image`, `Bush2Image`, `Bush3Image`, `Bush4Image`, `BushImage`, `Flower1Image`, `Flower2Image`, `Flower3Image`, `FlowerImage`, `Stone1Image`, `Stone2Image`, `StoneImage`, `ForestStumpImage`, `GrassTuft1Image`, `GrassTuft2Image`.
  - Master Tilesets: `LabTilesetImage`, `ZombieTilesetImage`.
  - Preserves all 27 legacy entity, floor, obstacle, and item image pointers.

### D. Anti-Cheating & Forensic Check Matrix
| # | Forensic Check | Result | Evidence |
|---|----------------|--------|----------|
| 1 | Hardcoded test results | **PASS** | No test results, bypasses, or canned boolean checks embedded in code. |
| 2 | Facade implementations | **PASS** | `loadEbitenImage()` and `Load()` perform genuine decoding and pointer assignment. |
| 3 | Fabricated verification outputs | **PASS** | All test results generated through live execution. |
| 4 | Self-certifying tests | **PASS** | Tests decode embedded assets dynamically and compute bounds/alpha fill ratios. |
| 5 | Execution delegation | **PASS** | No external CLI or third-party wrappers performing runtime work. |

### E. Behavioral Build & Test Execution
1. **Compilation (`CC=gcc go build ./cmd/game`)**:
   - Exit code: 0 (compiles cleanly).
2. **Package Tests (`CC=gcc go test ./...`)**:
   - `internal/ecs`: PASS
   - `internal/game`: PASS
   - `internal/game/world`: PASS
   - `internal/assets`: FAIL on `TestEmpiricalM1_All579ContextPNGsMatchImages` and `TestEmpiricalM1_EmbeddedFSStructureAndCount` (see Functional Defect below).

### F. Functional Defect Surfaced
- **Go Embed Directory Character Restriction**:
  - Directory path: `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites`
  - Cause: The directory name contains box-drawing Unicode characters `0xe2 0x94 0xac 0xe2 0x95 0x91` (`┬║`, artifact of legacy codepage extraction).
  - Effect: Go compiler's `embed` package rejects this directory (`cannot embed file ... in invalid directory 90┬║ Rotatable Bridge Sprites`), silently omitting the 3 PNG files inside (`Zombie-Tileset---_0106_Capa-107.png`, `Zombie-Tileset---_0107_Capa-108.png`, `Zombie-Tileset---_0108_Capa-109.png`) from `imageFS embed.FS`.
  - Disk has 606 PNGs, but `imageFS` embeds 603 PNGs.
  - Remediation: Renaming this subdirectory to standard ASCII (e.g. `90 Rotatable Bridge Sprites` or `90_Rotatable_Bridge_Sprites`) in `internal/assets/images/` will allow `embed.FS` to embed all 606 PNG files without error.

---

## 2. Logic Chain

1. **R1 Compliance**: The user requirement (ORIGINAL_REQUEST §R1) mandates completely deleting `cmd/tools/genassets` and retiring procedural asset generation. Verification confirms `cmd/tools/genassets` directory, `genassets` root binary, and all related direct invocations are completely removed from disk and cannot be executed.
2. **R2 Compliance & Authenticity**: Ingestion of PNGs from `context/` was verified by comparing SHA-256 checksums across all 579 files, proving bit-for-bit authenticity with 0 mocked or dummy assets. All 27 legacy assets remain intact in `internal/assets/images/`.
3. **Loader Integrity**: Inspection of `internal/assets/assets.go` proves that `loadEbitenImage` genuinely reads from `embed.FS` and decodes images using the standard Go `image` and `ebiten` APIs.
4. **Integrity vs Functional Classification**: The failure in `CC=gcc go test ./internal/assets/...` is not an integrity violation or facade; it is an encoding defect in the asset subdirectory naming that prevents Go's `//go:embed` directive from including 3 bridge sprites.

---

## 3. Caveats

- The subdirectory `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites` must be sanitized/renamed so that Go's `embed.FS` can embed its 3 PNG files, enabling `CC=gcc go test ./...` to achieve 100% pass rate.
- Milestone 2 and Milestone 3 logic (tile mappings, world generation, and depth-sorting draw systems) are outside Milestone 1 scope and will be audited in subsequent milestones.

---

## 4. Conclusion

**Final Forensic Verdict**: **CLEAN (Integrity)**.
There are **ZERO integrity violations**, **ZERO mock/dummy facades**, **ZERO hardcoded test cheats**, and **ZERO procedural generation leftovers**. 
The work product authentically satisfies R1 and R2 integrity standards. The directory naming defect noted above is documented for remediation in the next development step.

---

## 5. Verification Method

To independently verify all findings:

```bash
# 1. Verify R1 genassets deletion
test ! -d cmd/tools/genassets && test ! -f genassets && echo "R1 PASS: genassets deleted"

# 2. Verify SHA256 integrity of all 579 context PNGs vs internal/assets/images
python3 -c "
import os, hashlib
ctx, img = 'context', 'internal/assets/images'
for root, _, files in os.walk(ctx):
    for f in files:
        if f.lower().endswith('.png'):
            rel = os.path.relpath(os.path.join(root, f), ctx)
            h1 = hashlib.sha256(open(os.path.join(ctx, rel), 'rb').read()).hexdigest()
            h2 = hashlib.sha256(open(os.path.join(img, rel), 'rb').read()).hexdigest()
            assert h1 == h2, f'Mismatch in {rel}'
print('All 579 PNGs verified bit-for-bit identical!')
"

# 3. Verify game compiles cleanly
CC=gcc go build ./cmd/game
```
