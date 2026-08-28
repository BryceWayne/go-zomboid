# Handoff Report: Milestone 1 Obstacles (256x256) & Entities (64x128) Scaling Investigation

**Agent**: `teamwork_preview_explorer_m1_2`  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_2`  
**Report Target**: `f7a8f969-fc3f-4f72-a625-45c03a6444ae` (Parent Orchestrator)  
**Detailed Report Path**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_2/m1_obstacles_entities_analysis.md`  

---

## 1. Observation

1. **Current Codebase State**:
   - `cmd/tools/genassets/main.go` currently generates 27 PNG assets into `internal/assets/images/`:
     - 3 Character Entities at $16 \times 32$ (`generatePlayer`, `generateZombie`, `generateRunner`) [lines 204-345]
     - 6 Floor Tiles at $64 \times 32$ [lines 351-698]
     - 10 Vertical Obstacles/Props at $64 \times 64$ (`generateIsoWall`, `generateIsoTree`, `generateIsoFence`, `generateIsoDebris`, `generateIsoTent`, `generateIsoStump`, `generateIsoMushroom`, `generateIsoSign`, `generateElevationBlock`, `generateElevationRamp`) [lines 701-996, 1612-1798]
     - 8 Items/Weapons/Armor at $16 \times 16$ [lines 1002-1609, 1799-1850]
   - `internal/assets/assets_test.go` and `internal/assets/assets_stress_test.go` validate asset dimensions, decodability, non-nil pointers, isometric diamond bounding, character grounding anchors, and contrast.
   - `internal/game/game.go` `DrawSystem` currently maps world coordinates to screen offsets:
     - Floor draw offset: `drawX = isoX - 32 - camX`, `drawY = isoY - 0 - camY`
     - Obstacle draw offset: `drawX = isoX - 32 - camX`, `drawY = isoY - 32 - camY`
     - Entity draw offset: `drawX = isoX - 8 - camX`, `drawY = isoY - 32 - camY`
     - Item draw offset: `drawX = isoX - 8 - camX`, `drawY = isoY - 8 - camY`

2. **Milestone 1 & Milestone 2 Requirements**:
   - Quadruple base tile pixel size from $64 \times 32$ to $256 \times 128$ for floors (`PROJECT.md` Feature 1).
   - Proportionally scale vertical obstacles/props to $256 \times 256$ (`PROJECT.md` Feature 2).
   - Proportionally scale character entities to $64 \times 128$ (`PROJECT.md` Feature 3).
   - Proportionally scale items to $64 \times 64$ (`PROJECT.md` Feature 4).
   - Update `DrawSystem` anchor offsets in Milestone 2 (`PROJECT.md` Feature 7):
     - Floors: `(-128, 0)`
     - Obstacles: `(-128, -128)`
     - Entities: `(-32, -128)`
     - Items: `(-32, -32)`

---

## 2. Logic Chain

1. **Obstacle Geometry ($256 \times 256$)**:
   - In 2:1 isometric projection, a world tile cell of size $128 \times 128$ produces a ground footprint diamond of width $256\text{px}$ and height $128\text{px}$.
   - When drawn with offset `drawX = isoX - 128, drawY = isoY - 128`, the obstacle canvas top-left is positioned at `(isoX - 128, isoY - 128)`.
   - In the $256 \times 256$ canvas:
     - The bottom ground footprint diamond sits at $y \in [128..256]$, centered at $(128, 192)$, with top vertex at $(128, 128)$, left at $(0, 192)$, right at $(256, 192)$, and bottom at $(128, 256)$.
     - The elevated top face (height 128px) sits at $y \in [0..128]$, centered at $(128, 64)$, with top vertex at $(128, 0)$, left at $(0, 64)$, right at $(256, 64)$, and bottom at $(128, 128)$.
     - The West left face connects $x \in [0..128]$ with $\text{topY} = 64 + x/2$ and $\text{botY} = 192 + x/2$.
     - The South right face connects $x \in [128..256]$ with $\text{topY} = 128 - (x-128)/2$ and $\text{botY} = 256 - (x-128)/2$.
   - This exact 4x geometry ensures perfect visual tiling and z-sorting with floor tiles.

2. **Entity Geometry ($64 \times 128$)**:
   - Character entities are anchored at the center of their feet on the ground tile: `drawX = isoX - 32, drawY = isoY - 128`.
   - The grounding drop shadow ellipse is drawn at $(32, 122)$ with radii $r_x = 24.0, r_y = 6.0$.
   - Boots/feet are placed in rows $y \in [116..124]$ to satisfy the grounding invariant tested in `TestCharacterGroundAnchor`.
   - Torso ($y \in [48..82]$) and head ($y \in [10..48], R=18$) are scaled by $4\times$ with anti-aliasing and facial feature highlights (sclera, pupil, catchlight, snarl).

---

## 3. Caveats

- **Test Suite Synchronization**: Updating asset sizes in `cmd/tools/genassets` must be accompanied by updating the expected bounds in `internal/assets/assets_test.go`, `internal/assets/assets_stress_test.go`, and `cmd/tools/genassets/genassets_test.go` during M1 execution.
- **Milestone Coupling**: Asset generation (M1) generates 4x assets to disk; the engine coordinate and offset updates in `internal/game/game.go` and `internal/game/world/map.go` will be executed in Milestone 2.

---

## 4. Conclusion

- The mathematical specifications, coordinate mappings, color palettes, and drop-shadow formulations for all 10 vertical obstacles/props ($256 \times 256$) and all 3 character entities ($64 \times 128$) are fully defined and documented in `m1_obstacles_entities_analysis.md`.
- Complete, drop-in replacement Go implementations for `generatePlayer()`, `generateZombie()`, `generateRunner()`, and obstacle generators have been formulated with alpha compositing (`blendPixel`) and anti-aliased primitives (`drawAAEllipse`).

---

## 5. Verification Method

To independently verify the asset generation and dimensions once implemented:
1. `go run ./cmd/tools/genassets`
2. `CC=gcc go test ./cmd/tools/genassets/...`
3. `CC=gcc go test ./internal/assets/...`
4. Inspect generated PNG bounds in `internal/assets/images/` using `file` or image decoder:
   - Characters (`player.png`, `zombie.png`, `runner.png`): $64 \times 128$
   - Obstacles (`wall.png`, `tree.png`, `fence.png`, `debris.png`, `tent.png`, `stump.png`, `mushroom.png`, `sign.png`, `elevation_block.png`, `elevation_ramp.png`): $256 \times 256$
