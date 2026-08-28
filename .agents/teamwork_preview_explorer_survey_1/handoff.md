# Survey & Analysis of Procedural Asset Generation and Engine Integration

## 1. Observation

### 1.1 Existing Sprite Inventory & `cmd/tools/genassets` Implementation
The `cmd/tools/genassets` utility generates all visual assets for the game without external dependencies. The current codebase produces **11 sprite files** saved in `internal/assets/images/`:

| Sprite File | Dimensions | Generator Function | Visual Description & Current Implementation |
|---|---|---|---|
| `player.png` | 16x32 | `generateEntity("player.png", RGBA{0, 255, 0, 255})` (`main.go:18, 43-57`) | Solid lime-green rectangular block with 4 black pixels at `(10,4)`, `(10,5)`, `(10,10)`, `(10,11)`. |
| `zombie.png` | 16x32 | `generateEntity("zombie.png", RGBA{0, 100, 0, 255})` (`main.go:19, 43-57`) | Solid dark-green rectangular block with identical 4 black eye pixels. |
| `runner.png` | 16x32 | `generateEntity("runner.png", RGBA{150, 0, 0, 255})` (`main.go:20, 43-57`) | Solid red rectangular block with identical 4 black eye pixels. |
| `weapon.png` | 16x16 | `generateWeapon("weapon.png")` (`main.go:22, 208-217`) | 1-pixel thick diagonal brown line from `(2, 13)` to `(13, 2)`. |
| `food.png` | 16x16 | `generateItem("food.png", RGBA{255, 165, 0, 255})` (`main.go:23, 33-41`) | Solid 8x8 orange square centered between `(4,4)` and `(11,11)`. |
| `water.png` | 16x16 | `generateItem("water.png", RGBA{0, 191, 255, 255})` (`main.go:24, 33-41`) | Solid 8x8 cyan square centered between `(4,4)` and `(11,11)`. |
| `grass.png` | 64x32 | `generateIsoFloor("grass.png", RGBA{34, 139, 34, 255}, true)` (`main.go:26, 59-88`) | 64x32 isometric diamond (`|x-32|/32 + |y-16|/16 <= 1`), uniform green with random color subtraction noise (`variance := rand.Intn(20)`) and darkened edge when distance ratio > 0.9. |
| `dirt.png` | 64x32 | `generateIsoFloor("dirt.png", RGBA{139, 69, 19, 255}, true)` (`main.go:27, 59-88`) | Same 64x32 isometric diamond formula using brown color with random noise. |
| `wood.png` | 64x32 | `generateIsoFloor("wood.png", RGBA{205, 133, 63, 255}, true)` (`main.go:28, 59-88`) | Same 64x32 isometric diamond formula using wood-tan color with random noise. |
| `wall.png` | 64x64 | `generateIsoWall("wall.png", RGBA{105, 105, 105, 255})` (`main.go:29, 90-159`) | 64x64 3D block: 64x32 diamond top face (`topColor`), left vertical face (`80%` brightness), right vertical face (`60%` brightness). |
| `tree.png` | 64x64 | `generateIsoTree("tree.png")` (`main.go:30, 175-206`) | Brown trunk rectangle (`x: 28-36`, `y: 48-60`), green triangle canopy from tip `(32, 4)` to base at `y=50` with green noise. |

### 1.2 Asset Loading & Storage Pipeline
- **Embedding (`internal/assets/assets.go:13-14`)**:
  ```go
  //go:embed images/*
  var imageFS embed.FS
  ```
- **Exported Global Handles (`internal/assets/assets.go:16-28`)**:
  `PlayerImage`, `ZombieImage`, `RunnerImage`, `GrassImage`, `WallImage`, `DirtImage`, `WoodImage`, `TreeImage`, `WeaponImage`, `FoodImage`, `WaterImage` are package-level `*ebiten.Image` variables.
- **Decoding Mechanism (`internal/assets/assets.go:30-56`)**:
  `Load()` calls `loadEbitenImage(path)` which executes `imageFS.ReadFile(path)`, `image.Decode(bytes.NewReader(data))`, and `ebiten.NewImageFromImage(img)`.

### 1.3 Game Engine Coordinate Systems & Drawing Pipeline
- **Isometric Projection (`internal/game/game.go:546-550`)**:
  $$isoX = wx - wy$$
  $$isoY = \frac{wx + wy}{2}$$
- **Camera Offset (`internal/game/game.go:581-584`)**:
  $$camX = playerIsoX - 400$$
  $$camY = playerIsoY - 300$$
- **Layer 1: Ground Tiles (`internal/game/game.go:588-630`)**:
  - Direct drawing to `screen` (no Y-sorting required for flat ground).
  - Screen placement: `drawX = isoX - 32 - camX`, `drawY = isoY - 0 - camY`.
  - Memory fog: `op.ColorScale.Scale(0.2, 0.2, 0.3, 1)` if explored but outside active player vision radius (250px).
- **Layer 2: Y-Depth Sorted Renderables (`internal/game/game.go:632-828`)**:
  - `Renderable` struct contains `Image *ebiten.Image`, `Depth float64`, `Op *ebiten.DrawImageOptions`.
  - Walls & Trees (64x64): `drawX = isoX - 32 - camX`, `drawY = isoY - 32 - camY`, `Depth = worldX + worldY`.
  - Items (16x16): `drawX = isoX - 8 - camX`, `drawY = isoY - 8 - camY`, `Depth = iPos.X + iPos.Y`.
  - Entities (16x32): `drawX = isoX - 8 - camX`, `drawY = isoY - 32 - camY`, `Depth = pos.X + pos.Y`. (Feet grounded at `(isoX, isoY)`).
  - Facing indicator: `drawX = isoX - 4 - camX`, `drawY = isoY - 4 - camY`, `Depth = targetX + targetY`.
  - Sorting: `sort.SliceStable(sprites, func(i, j int) bool { return sprites[i].Depth < sprites[j].Depth })`.
- **Layer 3: Lighting & HUD (`internal/game/game.go:830-885`)**:
  - Day/Night cycle overlay: `alpha := 0.45 + 0.45*math.Cos((timeOfDay/24.0)*math.Pi*2)`, fills screen with `RGBA{0, 0, 15, alpha*255}`.
  - HUD draws Health bar, Hunger bar, Thirst bar, Weapon info, and 9 inventory slots.

### 1.4 Entities, Inventory, and Combat Systems
- **ECS Definition (`internal/ecs/components.go`)**:
  - `Player`: `Health`, `Hunger`, `Thirst`, `Inventory []string`, `WeaponEquipped bool`, `WeaponDurability int`, `AttackCooldown int`, `Dead bool`, `Infected bool`, `FacingX`, `FacingY`.
  - `Item`: `Type string`.
  - `Zombie`: `Speed float64`, `Chasing bool`, `IsRunner bool`, `WanderTimer int`, `WanderDirX`, `WanderDirY`, `StunTimer int`.
- **Combat Mechanics (`internal/game/game.go:345-391`)**:
  - When attacking without a weapon: shoves zombies back (`StunTimer = 45`, velocity pushback).
  - When attacking with a weapon equipped: removes zombie entity on hit, decrements `WeaponDurability` (breaks when durability reaches 0).
  - When zombies touch the player within distance < 14: `playerComp.Infected = true` (`internal/game/game.go:463-468`), which continuously drains health (`player.Health -= 0.05` per frame).

---

## 2. Logic Chain

### 2.1 Limitations of Existing Procedural Generation
1. **Lack of Visual Anatomy & Contrast**: Player, Zombie, and Runner are currently monolithic rectangles of solid color differing only by RGB base fill. They lack anatomical features (head, torso, legs, boots, arms, hands, facial expressions, clothes, decaying flesh).
2. **Abstract Item Icons**: Items (`food.png`, `water.png`, `weapon.png`) are primitive geometric squares or single-pixel diagonal lines, making them difficult to visually recognize in the world.
3. **Monolithic Isometric Surfaces**: Floor tiles (`grass`, `dirt`, `wood`) and 3D blocks (`wall`, `tree`) use uniform random RGB noise subtraction without structured patterns (e.g. grass blades, stone mortar, brick courses, wood plank seams, foliage clusters, or directional lighting bevels).

### 2.2 Requirements for Fulfilling R1 & R2 from `ORIGINAL_REQUEST.md`
1. **R1 (Procedural Sprite Enhancements)**:
   - Enhance all existing sprites using pure algorithmic generation in `cmd/tools/genassets`.
   - Implement pixel-art drawing primitives (anti-aliased/crisp lines, circles, beveled rects, polygons).
   - Use cohesive palettes and shading models (isometric directional lighting from top-left, highlights, drop shadows, ambient occlusion).
   - Add structural textures: brick patterns, wood grain, foliage cluster noise, grass tufts.
2. **R2 (Environment, Weapons & Armor)**:
   - **Armor System**:
     - Generate procedural armor sprites: `armor.png` (tactical Kevlar body armor / reinforced vest with plates, straps, buckles), and optionally `helmet.png`.
     - Register `ArmorImage` in `internal/assets/assets.go`.
     - Add armor fields to `ecs.Player` (e.g. `ArmorEquipped bool`, `ArmorDurability int`, `ArmorAbsorption float64`).
     - Update inventory consumption to equip armor when selected.
     - Update zombie attack logic so armor absorbs damage or prevents/reduces infection chance.
   - **Expanded Weapons**:
     - Generate new procedural weapon sprites: e.g. `axe.png` (fire axe with polished wood handle and red steel head with sharpened blade glint), `knife.png` (tactical knife), `crowbar.png`, or `shotgun.png`.
     - Register new weapon images in `internal/assets/assets.go`.
     - Update item spawning in `internal/game/game.go` and `internal/game/world/map.go`.
     - Update combat logic with distinct weapon stats (durability, attack reach, damage).

### 2.3 Algorithmic Design for Upgraded Procedural Sprites in Pure Go

```
+---------------------------------------------------------------------------------------+
|                              PROCEDURAL SPRITE PIPELINE                               |
+---------------------------------------------------------------------------------------+
|  1. Color Palette Engine    |  2. Geometry & Pixel Shader |  3. Texture & Noise Engine |
|  - Curated 32-color palette |  - Bresenham line drawer    |  - Perlin / Octave noise   |
|  - Warm highlight shifting  |  - Midpoint circle / disc   |  - Bayer matrix dithering  |
|  - Cool shadow shifting     |  - Scanline polygon filler  |  - Wood grain sine mod     |
|  - Alpha blending & borders |  - Isometric face projection|  - Brick/mortar course grid|
+---------------------------------------------------------------------------------------+
                                           |
    +--------------------------------------+--------------------------------------+
    |                                      |                                      |
    v                                      v                                      v
[Characters 16x32]                 [Items & Equipment 16x16]              [Environment 64x32 & 64x64]
- Player: Skin, Hair, Shirt,       - Food: Can with label & pull-tab      - Grass: Blade tufts, border bevel
  Jeans, Belt, Combat Boots        - Water: Contoured bottle & cap        - Dirt: Soil ruts, pebble clusters
- Zombie: Decaying flesh, torn     - Weapon: Spiked bat / grip tape       - Wood: Diagonal plank seams, nails
  ragged shirt, rib wounds, claws  - Axe: Fire axe, steel glint           - Wall: Masonry brick courses & cap
- Runner: Hunched predator pose,   - Armor: Kevlar vest with buckles      - Tree: 3-tier foliage & bark roots
  crimson sinew, glowing eyes      - Medkit: White box with red cross
```

#### Detailed Sprite Generator Specifications:
1. **Humanoid Entities (16x32)**:
   - `generateHumanoid(name, skinColor, hairColor, shirtColor, pantsColor, bootColor, isZombie, isRunner)`
   - Layer 0: 1px Dark Silhouette Outline.
   - Layer 1: Head (`y: 2..9`, `x: 5..10`): Skin base, hair volume (`y: 2..5`), eyes (`y: 6`), mouth/beard (`y: 8`).
   - Layer 2: Torso & Arms (`y: 10..19`): Shirt with shading, collar (`y: 10`), chest highlight, arms (`x: 2..4, 11..13`), hands (`y: 17..19`).
   - Layer 3: Waist & Belt (`y: 20`): Dark leather belt with metallic buckle at `x: 8`.
   - Layer 4: Legs (`y: 21..28`): Denim/khaki pants with center inseam split (`x: 7..8`) and knee crease shadows.
   - Layer 5: Boots (`y: 29..31`): Heavy boots with dark soles at `y: 31`.
   - Zombie Variations: Asymmetrical torn clothes, green/grey rotting skin, exposed red flesh wounds, vacant eyes.
   - Runner Variations: Aggressive hunched silhouette, glowing red eyes, deep crimson muscle sinews.

2. **Items & Weapons (16x16)**:
   - `generateItemCannedFood()`: Cylindrical tin can (`x: 4..11, y: 3..13`) with top lid oval, red/yellow label, silver rim highlights.
   - `generateItemWaterBottle()`: Translucent cyan bottle body, narrow neck (`x: 7..8, y: 3..5`), white plastic cap (`y: 2`), water meniscus reflection.
   - `generateItemBaseballBat()`: Diagonal wooden bat with tapering grip, white tape wraps (`y: 11..14`), wood grain, iron studs/nails at the head.
   - `generateItemFireAxe()`: Curved wooden handle, double-beveled red axe head (`x: 7..14, y: 2..7`), sharpened silver cutting edge.
   - `generateItemArmorVest()`: Tactical ballistic vest silhouette (`x: 3..12, y: 2..13`), shoulder straps, modular webbing (MOLLE), center zipper, front armor plate shading.
   - `generateItemMedkit()`: White emergency kit box (`x: 3..12, y: 4..13`) with bright red medical cross at center and dark carry handle.

3. **Isometric Environment (64x32 Floor Tiles & 64x64 Blocks)**:
   - `generateIsoGrass()`: Diamond mask, multi-frequency green noise, scattered 2-3px grass blade tufts, yellow flower flecks, 1px isometric bevel.
   - `generateIsoDirt()`: Diamond mask, rich brown soil palette, pebble clusters, subtle tire track grooves.
   - `generateIsoWoodFloor()`: Diamond mask, diagonal plank divider lines along the isometric axes, staggered plank ends, nail head dots, wood grain variation.
   - `generateIsoWall()`:
     - Top Face (64x32 diamond): Smooth beveled concrete coping stone.
     - Left Face (32x32 isometric plane): Running bond brick pattern (alternating courses every 4 pixels), mortar lines, cast shadow from top coping.
     - Right Face (32x32 isometric plane): Shadowed brick courses with darker palette and corner quoin stones.
   - `generateIsoTree()`:
     - Trunk: Cylindrical textured trunk (`x: 28..36, y: 46..60`) with root flares grounding into tile center at `(32, 60)`.
     - Canopy: 3 distinct overlapping isometric canopy cones/domes (Top `y: 4..22`, Mid `y: 18..36`, Bottom `y: 30..48`), top-left directional lighting, dark under-canopy cast shadows.

---

## 3. Caveats

1. **Strict Engine Alignment Dimensions**:
   - Floor tiles MUST remain 64x32.
   - Wall/Tree blocks MUST remain 64x64 with base grounded at `(32, 32)` to `(32, 64)`.
   - Items MUST remain 16x16 to match item entity positioning offset `(isoX - 8, isoY - 8)`.
   - Character entities MUST remain 16x32 to match entity drawing offset `(isoX - 8, isoY - 32)`.
2. **Ebitengine Color Scale Interactions**:
   - The engine applies dynamic color scales in `internal/game/game.go` (e.g. `op.ColorScale.Scale(0.2, 0.2, 0.3, 1)` for memory fog, `op.ColorScale.Scale(2, 2, 2, 1)` for attack flash, `op.ColorScale.Scale(1.5, 1.5, 2.5, 1)` for zombie stun). Base sprites should have balanced midtones so multiplicative color scales do not wash out or over-saturate.
3. **Purely Procedural Generation Constraint**:
   - No external image files, third-party asset packs, or network downloads can be introduced. All assets must be produced purely through Go code in `cmd/tools/genassets`.

---

## 4. Conclusion

- The codebase is clean, modular, and directly amenable to procedural sprite expansion.
- `cmd/tools/genassets` can be upgraded with rich pixel-art procedural generation algorithms (palette management, anatomical layering, geometric line/circle rasterization, isometric face shading, wood/brick/foliage procedural texturing).
- Upgrading existing sprites (`player`, `zombie`, `runner`, `weapon`, `food`, `water`, `grass`, `dirt`, `wood`, `wall`, `tree`) and adding new sprites (`armor`, `axe`, `medkit`, etc.) directly supports both R1 and R2.
- The pipeline requires:
  1. Updating `cmd/tools/genassets/main.go` to generate the new/upgraded PNGs.
  2. Updating `internal/assets/assets.go` to expose the new image handles.
  3. Connecting the new armor mechanics and weapon types to `internal/ecs/components.go` and `internal/game/game.go`.

---

## 5. Verification Method

To independently verify asset generation, compilation, tests, and runtime integrity:

1. **Verify Asset Generation Tool**:
   ```bash
   go run ./cmd/tools/genassets
   ```
   *Expected result*: Executes with exit code 0 and logs successful generation of all PNG files in `internal/assets/images/`.

2. **Verify All Unit Tests**:
   ```bash
   CC=gcc go test ./...
   ```
   *Expected result*: All package tests pass cleanly (`internal/game`, `internal/game/world`, etc.).

3. **Verify Build & Compilation**:
   ```bash
   go build -o /tmp/go-zomboid-test ./cmd/game
   ```
   *Expected result*: Compiles without error.

4. **Verify Game Launch**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   *Expected result*: Game window opens with the isometric world, rendered sprites, and gameplay systems functioning properly.
