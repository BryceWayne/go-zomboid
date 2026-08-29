# Independent Review and Adversarial Challenge Report: Milestone 1 (R1 & R2)

**Reviewer**: `teamwork_preview_reviewer_m1_2` (Roles: Reviewer, Critic)  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_reviewer_m1_2`  
**Verdict**: **REQUEST_CHANGES**  
**Overall Risk Assessment**: **MEDIUM**

---

## 1. Review Summary

- **Verdict**: **REQUEST_CHANGES**
- **Rationale**: While procedural generation retirement (R1) and core external asset loader declarations (R2) are implemented cleanly, the repository test suite `CC=gcc go test ./...` fails. Specifically, 3 out of 579 external PNG files (`Zombie-Tileset---_0106_Capa-107.png`, `Zombie-Tileset---_0107_Capa-108.png`, `Zombie-Tileset---_0108_Capa-109.png`) located in `images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/` are excluded by Go's `//go:embed images/*` directive due to non-ASCII box-drawing characters (`┬║` / `\xe2\x94\xac\xe2\x95\x91`) failing Go's `module.CheckFilePath()` validation. This causes `TestEmpiricalM1_All579ContextPNGsMatchImages`, `TestEmpiricalM1_EmbeddedFSStructureAndCount`, and `TestChallenger_All606EmbeddedPNGIntegrity` to fail.

---

## 2. Findings

### [Major] Finding 1: Non-ASCII Directory Name Bypasses Go `//go:embed` Directive

- **What**: 3 external PNG files are missing from Go's embedded filesystem (`imageFS`), resulting in only 603 files embedded instead of the expected 606 (27 legacy + 579 external).
- **Where**: `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/`
- **Why**: The directory name contains non-ASCII box-drawing characters `┬║` (`\xe2\x94\xac\xe2\x95\x91`, CP437 degree symbol artifact). Go's compiler (`cmd/go/internal/load/pkg.go:isBadEmbedName()`) delegates to `module.CheckFilePath()`, which rejects non-ASCII and special punctuation directory names. When `isBadEmbedName()` returns true for a directory during `//go:embed images/*` traversal, the compiler invokes `fs.SkipDir`, skipping embedding for that entire subdirectory without failing compilation. Consequently, `imageFS.ReadFile()` fails for these 3 files at runtime.
- **Suggestion**:
  1. Rename directory `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90┬║ Rotatable Bridge Sprites/` to `internal/assets/images/Zombie Apocalypse Tileset/Organized separated sprites/90 Rotatable Bridge Sprites/` (or `90 Degree Rotatable Bridge Sprites/`).
  2. Update any path references in test suites to reflect the sanitized ASCII directory name.

---

## 3. Verified Claims

| Claim | Verification Method | Result | Notes |
|---|---|---|---|
| `cmd/tools/genassets` permanently retired (R1) | `test ! -d cmd/tools/genassets && test ! -f genassets` | **PASS** | Directory and root binary deleted; determinism test calling it removed from `empirical_challenger_test.go`. |
| Legacy asset compatibility | `CC=gcc go test -v ./internal/assets -run TestAssetsLoadAllPointersNonNil` | **PASS** | All 27 legacy pointers (`PlayerImage`, `GrassImage`, etc.) load and match dimensions. |
| External asset pointer initialization (R2) | `CC=gcc go test -v ./internal/assets -run TestExternalAssetsLoadAllPointersNonNil` | **PASS** | All 22 new asset pointers (`BenchImage`, `ChestImage`, `Sculpture1Image`, etc.) non-nil and valid. |
| Thread safety & idempotency | `CC=gcc go test -race -v ./internal/assets -run TestAssetsLoadIdempotency` & `TestEmpiricalM1_ConcurrentLoadIdempotencyStress` | **PASS** | `sync.Once` ensures safe concurrent `Load()` across 100 goroutines without data races. |
| No Integrity Violations | Source code inspection of `internal/assets/assets.go` | **PASS** | Real `ebiten.NewImageFromImage` loading from `imageFS`; no hardcoded mock data or fake passes. |
| Repository test suite execution | `CC=gcc go test ./...` | **FAIL** | Failed in `internal/assets` due to 3 missing embedded files from non-ASCII path. |

---

## 4. Adversarial Review & Stress Testing

### Challenge Summary
- **Overall Risk**: Medium
- **Primary Stress Mode**: Embed completeness and character set validation on external assets.

### Challenges Evaluated:
1. **Concurrent Initialization Race**:
   - **Hypothesis**: Rapid concurrent calls to `assets.Load()` during game startup or parallel tests could cause torn reads or multiple loads.
   - **Result**: **PASS**. Protected by `loadOnce sync.Once`. Concurrency stress tests with 100 parallel goroutines and `-race` pass without race warnings.
2. **Embedded Asset Completeness**:
   - **Hypothesis**: All 579 external assets in `context/` can be read and decoded from `imageFS`.
   - **Result**: **FAIL**. 3 files inside `90┬║ Rotatable Bridge Sprites/` are omitted from `imageFS` due to Go embed module restrictions.
3. **Legacy Regression**:
   - **Hypothesis**: Modifications to `assets.go` break downstream packages (`internal/game`, `internal/game/world`, `internal/ecs`).
   - **Result**: **PASS**. All tests in `internal/ecs`, `internal/game`, and `internal/game/world` pass cleanly, and `CC=gcc go build ./cmd/game` succeeds.

---

## 5. Handoff Protocol (5 Components)

### 1. Observation
1. **Repository Test Execution**:
   Command: `CC=gcc go test ./...`
   Output:
   ```
   --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages (0.03s)
       --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages/.../90┬║_Rotatable_Bridge_Sprites/Zombie-Tileset---_0106_Capa-107.png
           missing embedded PNG in imageFS: ...: open ...: file does not exist
       --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages/.../90┬║_Rotatable_Bridge_Sprites/Zombie-Tileset---_0107_Capa-108.png
           missing embedded PNG in imageFS: ...: open ...: file does not exist
       --- FAIL: TestEmpiricalM1_All579ContextPNGsMatchImages/.../90┬║_Rotatable_Bridge_Sprites/Zombie-Tileset---_0108_Capa-109.png
           missing embedded PNG in imageFS: ...: open ...: file does not exist
   FAIL internal/assets
   ```
2. **Go Toolchain `go list` Embed File Count**:
   Command: `go list -json ./internal/assets`
   Result: `EmbedFiles` count is `603`. Disk contains `606` PNG files (`find internal/assets/images -type f -name "*.png" | wc -l`).
3. **Go Embed Specification & Implementation**:
   In Go's `cmd/go/internal/load/pkg.go:isBadEmbedName()`, `module.CheckFilePath("90┬║ Rotatable Bridge Sprites")` fails on non-ASCII characters, triggering `fs.SkipDir` during `resolveEmbed()`.

### 2. Logic Chain
1. Milestone 1 requirement R2 requires copying external PNG files from `context/` into `internal/assets/images/` and embedding them via `//go:embed images/*`.
2. One directory in `context/` is named `90┬║ Rotatable Bridge Sprites` (containing 3 PNG files: `Zombie-Tileset---_0106_Capa-107.png`, `Zombie-Tileset---_0107_Capa-108.png`, `Zombie-Tileset---_0108_Capa-109.png`).
3. Go's `embed` toolchain ignores directories with names that fail `module.CheckFilePath()`. Non-ASCII box drawing characters `┬║` cause Go to execute `fs.SkipDir` and omit all 3 files from `imageFS`.
4. As observed in Observation 1, challenger test suites verifying embed completeness fail because `imageFS.ReadFile()` returns `file does not exist`.
5. Therefore, `CC=gcc go test ./...` fails, requiring a **REQUEST_CHANGES** verdict to sanitize the directory name.

### 3. Caveats
- The failure is isolated to asset embedding of 3 rotatable bridge sprites in `Zombie Apocalypse Tileset`.
- All game-critical props required for Milestone 2 (`BenchImage`, `ChestImage`, `SculptureImage`, `BushImage`, `FlowerImage`, `StoneImage`, `LabTilesetImage`) are in standard ASCII paths and load with 100% success.
- Procedural generation retirement (R1) is completely sound and verified.

### 4. Conclusion
Milestone 1 satisfies architectural requirements R1 and R2 for prop/foliage/tileset pointer loading, but fails the acceptance criteria of `CC=gcc go test ./...` passing across the repository due to the non-ASCII directory name in `internal/assets/images/`.
**Verdict**: **REQUEST_CHANGES**  
**Required Action**: Sanitize the directory name to `90 Rotatable Bridge Sprites` (or `90 Degree Rotatable Bridge Sprites`) in `internal/assets/images/` and update test assertions to match.

### 5. Verification Method
To independently verify the failure and subsequent fix:
```bash
# 1. Reproduce embed failure
CC=gcc go test -v ./internal/assets -run "TestEmpiricalM1_All579ContextPNGsMatchImages|TestChallenger_All606EmbeddedPNGIntegrity"

# 2. Verify all tests across repository
CC=gcc go test ./...

# 3. Verify game compilation
CC=gcc go build ./cmd/game
```
Invalidation condition for this REQUEST_CHANGES verdict: When all 606 PNG files are embedded in `imageFS` and `CC=gcc go test ./...` exits with code 0 across all packages.
