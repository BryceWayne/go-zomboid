## 2026-08-29T15:38:39Z

Scope & Task:
1. Update `internal/assets/assets.go`:
   - In `Load()`, ensure all 27 legacy pointers are loaded from their canonical legacy paths:
     - `PlayerImage`: `images/player.png`
     - `ZombieImage`: `images/zombie.png`
     - `RunnerImage`: `images/runner.png`
     - `GrassImage`: `images/grass.png`
     - `DirtImage`: `images/dirt.png`
     - `WoodImage`: `images/wood.png`
     - `AsphaltImage`: `images/asphalt.png`
     - `ConcreteImage`: `images/concrete.png`
     - `TileFloorImage`: `images/tile_floor.png`
     - `WallImage`: `images/wall.png`
     - `TreeImage`: `images/tree.png`
     - `FenceImage`: `images/fence.png`
     - `DebrisImage`: `images/debris.png`
     - `TentImage`: `images/tent.png`
     - `StumpImage`: `images/stump.png`
     - `MushroomImage`: `images/mushroom.png`
     - `SignImage`: `images/sign.png`
     - `ElevationBlockImage`: `images/elevation_block.png`
     - `ElevationRampImage`: `images/elevation_ramp.png`
     - `FoodImage`: `images/food.png`
     - `WaterImage`: `images/water.png`
     - `WeaponImage`: `images/weapon.png`
     - `AxeImage`: `images/axe.png`
     - `ShotgunImage`: `images/shotgun.png`
     - `AmmoImage`: `images/ammo.png`
     - `ArmorImage`: `images/armor.png`
     - `AntidoteImage`: `images/antidote.png`
   - In `Load()`, ensure all 22 new external asset pointers are loaded from their respective external paths under `images/Small Forest/...`, `images/Lab/...`, `images/Zombie Apocalypse Tileset/...`:
     - `BenchImage`, `ChestImage`, `Sculpture1Image`, `Sculpture2Image`, `SculptureImage`, `Bush1Image`, `Bush2Image`, `Bush3Image`, `Bush4Image`, `BushImage`, `Flower1Image`, `Flower2Image`, `Flower3Image`, `FlowerImage`, `Stone1Image`, `Stone2Image`, `StoneImage`, `ForestStumpImage`, `GrassTuft1Image`, `GrassTuft2Image`, `LabTilesetImage`, `ZombieTilesetImage`.
2. Update tests in `internal/assets/` (`assets_test.go`, `challenger_stress_test.go`) and `internal/game/` (`draw_depth_test.go`) to verify both the 27 legacy pointers and the new external prop pointers, and their geometric anchors.
3. Verify:
   - Run `CC=gcc go test -v -count=1 ./...`
   - Run `CC=gcc go test -race ./...`
   - Run `CC=gcc go build ./cmd/game`
   Ensure all tests in all packages pass with exit code 0.

Write your handoff report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_remediation_1/handoff.md`. Send a message when complete.
