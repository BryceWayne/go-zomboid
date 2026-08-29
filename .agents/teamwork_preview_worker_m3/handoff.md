# Milestone 3 Handoff Report: DrawSystem Rendering & Depth-Sorting

## 1. Observation
- `internal/game/game.go` defines `DrawSystem.Draw(screen *ebiten.Image, timeOfDay float64)`.
- Ground Pass (Pass 1): Handled flat terrain rendering (lines 879-927). Previously only mapped legacy terrain/obstacles to base ground diamonds. Missing cases for new props (`TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`), which left transparent/unrendered holes on the tile diamond.
- Depth-Sorted Sprite Pass (Pass 2): Handled obstacle and entity collection into `sprites []Renderable` (lines 936-1003) with stable sort by `Depth = worldX + worldY`. Previously only checked legacy obstacle tile types.
- Prop sprite dimensions: `BenchImage` (52x37), `ChestImage` (22x21), `SculptureImage` (23x31), `BushImage` (24x18), `FlowerImage` (26x25), `StoneImage` (28x19), compared to legacy 256x256 tiles.
- Applying unified geometric anchor translation `op.GeoM.Translate(-float64(imgW)/2.0, 128.0 - float64(imgH))` centers any sprite of width `imgW` on the tile in isometric X ($isoX$) and aligns the bottom of the sprite with the tile base ($isoY + 128.0$), while reducing identically to `-128, -128` for 256x256 obstacles.
- `CC=gcc go test -v ./internal/game/...` passes 100%.
- `CC=gcc go test ./...` passes 100%.
- `CC=gcc go build ./cmd/game` builds cleanly.

## 2. Logic Chain
1. **Ground Pass Alignment**: Prop tiles occupy a tile cell on the map. Without a base ground diamond rendered during Pass 1, black/transparent gaps would appear under the prop sprite. Adding `TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone` to the `assets.GrassImage` case guarantees a seamless terrain base beneath all props.
2. **Sprite Collection & Binding**: In Pass 2, when iterating through the map tiles, checking for all 6 new prop tile types and binding them to their respective `*ebiten.Image` handles (`assets.BenchImage`, `assets.ChestImage`, `assets.SculptureImage`, `assets.BushImage`, `assets.FlowerImage`, `assets.StoneImage`) brings them into the rendering pipeline.
3. **Dynamic Geometric Anchoring**: For any sprite with dimensions $(W, H)$, translating horizontally by $-W/2$ places its center at $isoX$. Translating vertically by $128.0 - H$ places its foot/bottom at $isoY + 128.0$ (the isometric diamond's bottom vertex). For 256x256 sprites, $-256/2 = -128$ and $128 - 256 = -128$, preserving identical legacy rendering behavior while positioning newly ingested variable-sized props naturally on the ground.
4. **Depth Sorting Invariants**: Setting `Depth = worldX + worldY` for props ensures they are sorted stably alongside walls, trees, items, character entities (player, zombies), and indicators via `sort.SliceStable`.
5. **FOV Visibility & Memory Tinting**: Non-visible but explored cells receive the memory tint `op.ColorScale.Scale(0.2, 0.2, 0.3, 1)`, and unvisited cells are omitted from rendering.

## 3. Caveats
- `internal/assets/assets.go` currently binds `SculptureImage` to `Sculpture1Image`, `BushImage` to `Bush1Image`, `FlowerImage` to `Flower1Image`, and `StoneImage` to `Stone1Image`. If randomized visual variations (e.g. `Sculpture2Image`, `Bush2Image`) are desired in future milestones, `DrawSystem.Draw` can select variant images based on tile coordinates $(x, y)$.

## 4. Conclusion
- Milestone 3 is complete.
- `DrawSystem.Draw` now renders base ground diamonds in Pass 1 and depth-sorts prop sprites with unified geometric anchoring in Pass 2.
- Added comprehensive unit and stress tests in `internal/game/draw_depth_test.go` and updated `internal/game/game_stress_test.go`.
- All tests pass and game binary compiles cleanly.

## 5. Verification Method
- `CC=gcc go test -v ./internal/game/...`
- `CC=gcc go test ./...`
- `CC=gcc go build ./cmd/game`
