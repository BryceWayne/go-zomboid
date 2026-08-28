## 2026-08-28T17:20:07Z
You are a Worker subagent (teamwork_preview_worker_m1_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m1_1
Project root: /home/bryce/code/go-zomboid
Original Request: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan: /home/bryce/code/go-zomboid/PROJECT.md

Scope: Milestone 1 - Procedural Sprite Enhancements & Asset Pipeline Integration

Explorer Handoff References:
- Character sprites: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_1/handoff.md
- Environment tiles: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_2/handoff.md
- Items, weapons, armor & assets.go: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_3/handoff.md

Write Ownership:
You own and may modify:
- `cmd/tools/genassets/main.go`
- `internal/assets/assets.go`
- `internal/assets/images/*.png` (by running `go run ./cmd/tools/genassets`)

Tasks:
1. Implement the upgraded procedural pixel-art generation in `cmd/tools/genassets/main.go` incorporating:
   - Drawing helpers and color manipulation primitives (`setPixel`, `fillRect`, `drawHLine`, `drawVLine`, `drawShadedRect`, `darken`, `lighten`, `blend`, `drawMatrix`, `addSelectiveOutline`).
   - Character entities (16x32): `generatePlayer`, `generateZombie`, `generateRunner`.
   - Floor tiles (64x32): `generateGrass`, `generateDirt`, `generateWoodFloor`, `generateAsphalt`, `generateConcrete`, `generateTileFloor`.
   - Vertical obstacles (64x64): `generateIsoWall`, `generateIsoTree`, `generateIsoFence`, `generateIsoDebris`.
   - Items & equipment (16x16): `generateFood`, `generateWater`, `generateWeapon` (spiked bat), `generateAxe`, `generateShotgun`, `generateAmmo`, `generateArmor` (ballistic Kevlar vest).
2. Execute `go run ./cmd/tools/genassets` to generate all 20 PNG files in `internal/assets/images/`.
3. Update `internal/assets/assets.go` to expose all image handles (`PlayerImage`, `ZombieImage`, `RunnerImage`, `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`, `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `FoodImage`, `WaterImage`) and load them in `Load()`.
4. Run `CC=gcc go test ./...` and `CC=gcc go build -o bin/game ./cmd/game` to verify all builds and existing tests pass.
5. Document your implementation and verification in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m1_1/handoff.md`.
