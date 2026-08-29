# Comprehensive Technical Survey: Asset Pipeline & External PNG Ingestion

**Author**: `teamwork_preview_explorer_survey_1`  
**Date**: 2026-08-29  
**Milestone**: Asset Pipeline Survey & Migration Plan (R1 & R2)  
**Status**: Complete  

---

## 1. Executive Summary

This survey provides an in-depth technical analysis and migration plan for transitioning the `go-zomboid` project from its legacy procedural asset generator (`cmd/tools/genassets`) to external pixel-art PNG assets located in `context/`.

### Key Findings
1. **Context Asset Catalog**: The `context/` directory contains **590 total files**, including **579 PNG image files**, **3 Photoshop (.psd) master files**, and **8 macOS `.DS_Store` metadata files** (along with Windows `:Zone.Identifier` stream files).
   - `Lab/`: 1 PNG master indoor tileset sheet (`Inside_C.png`, 768x768).
   - `Small Forest/`: 45 PNG files comprising modular world props (benches, chests, sculptures, bushes, fences, flowers, grass, ground tilesets, stones, trees).
   - `Zombie Apocalypse Tileset/`: 533 PNG files (1 reference sheet + 532 separated sprites across 46 animation, vehicle, weapon, prop, building, and UI categories) and 3 PSD master files.
2. **Procedural System (`genassets`) Footprint**:
   - `cmd/tools/genassets` consists of `main.go` (2,413 lines) and `genassets_test.go` (119 lines) generating 27 legacy PNG files into `internal/assets/images/`.
   - A compiled binary `/home/bryce/code/go-zomboid/genassets` resides at the repository root.
   - Codebase references to `genassets` exist in `README.md`, `PROJECT.md`, `TEST_READY.md`, `TEST_INFRA.md`, `ART_STYLE_GUIDE.md`, and one test (`internal/assets/empirical_challenger_test.go:302-357`).
3. **Asset Loading Architecture**:
   - `internal/assets/assets.go` uses Go standard `//go:embed images/*` via `embed.FS` and decodes PNGs into `*ebiten.Image` pointers guarded by `sync.Once`.
   - 27 existing `*ebiten.Image` variables are currently exported and consumed by `internal/game/game.go`, `internal/game/world/map.go`, and test suites.
   - `//go:embed images/*` recursively supports nested subdirectories out of the box in Go 1.16+.
4. **Migration Strategy**:
   - **R1**: Completely delete `cmd/tools/genassets/` and `./genassets`. Update `internal/assets/empirical_challenger_test.go` to remove/retire `TestEmpiricalGenerationDeterminism`. Clean up documentation references.
   - **R2**: Ingest all 579 PNG files from `context/` into `internal/assets/images/`, preserving clean directory structures while maintaining existing 27 legacy asset files for backwards compatibility. Update `assets.go` to declare and load exported `*ebiten.Image` pointers for new assets (Benches, Chests, Sculptures, Bushes, Flowers, Stones, Trees, Tilesets, Items).

---

## 2. Complete Context Directory Inventory & Categorization

The `context/` directory contains rich pixel-art sprites across three major packages. All files have been inspected, decoded, and categorized.

### 2.1 Summary by Top-Level Directory

| Directory | PNG Files | PSD Files | .DS_Store | Total Files | Primary Contents / Themes |
|---|---|---|---|---|---|
| `Lab/` | 1 | 0 | 0 | 1 | High-resolution indoor research facility tileset sheet (768x768) |
| `Small Forest/` | 45 | 0 | 0 | 45 | Isometric/top-down outdoor props: benches, chests, sculptures, foliage, fences, ground tiles |
| `Zombie Apocalypse Tileset/` | 533 | 3 | 8 | 544 | Reference tileset (764x300), 532 separated sprites: characters, zombies, weapons, buildings, vehicles, FX, UI |
| **Total** | **579** | **3** | **8** | **590** | |

*(Note: Windows `:Zone.Identifier` stream files created during download are omitted from the above counts and must be excluded during ingestion).*

---

### 2.2 Detailed Inventory of `Lab/` (1 file)

| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Lab/Inside_C.png` | PNG | 768x768 | RGBA | 100,232 B | Laboratory interior tileset sheet containing floor tiles, scientific machinery, cryogenic pods, monitors, consoles, and lab tables. |

---

### 2.3 Detailed Inventory of `Small Forest/` (45 files)

`Small Forest` contains outdoor environmental props and ground tiles.

#### A. Bench and Chest (2 files)
| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Small Forest/Bench and chest/Bench.png` | PNG | 52x37 | RGBA | 1,464 B | Wooden park bench (`TileBench` / `assets.BenchImage`) |
| `Small Forest/Bench and chest/Chest.png` | PNG | 22x21 | RGBA | 1,286 B | Wooden treasure/loot container (`TileChest` / `assets.ChestImage`) |

#### B. Bushes and Stumps (5 files)
| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Small Forest/Bushes/Bush-1.png` | PNG | 24x18 | RGBA | 1,260 B | Small green bush (`TileBush` / `assets.Bush1Image`) |
| `Small Forest/Bushes/Bush-2.png` | PNG | 19x15 | RGBA | 1,280 B | Dense round bush (`assets.Bush2Image`) |
| `Small Forest/Bushes/Bush-3.png` | PNG | 25x19 | RGBA | 1,296 B | Wide foliage bush (`assets.Bush3Image`) |
| `Small Forest/Bushes/Bush-4.png` | PNG | 28x19 | RGBA | 1,361 B | Large bushy shrub (`assets.Bush4Image`) |
| `Small Forest/Bushes/Stump.png` | PNG | 29x19 | RGBA | 1,520 B | Cut tree stump (`TileStump` / `assets.ForestStumpImage`) |

#### C. Fences (12 files)
| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Small Forest/Fences/Big wooden fence/Big-wooden-fence-1.png` | PNG | 54x23 | RGBA | 1,636 B | Long horizontal tall wooden fence |
| `Small Forest/Fences/Big wooden fence/Big-wooden-fence-2.png` | PNG | 64x23 | RGBA | 1,664 B | Extra long tall wooden fence |
| `Small Forest/Fences/Big wooden fence/Big-wooden-fence-3.png` | PNG | 54x23 | RGBA | 1,634 B | Tall wooden fence segment with post |
| `Small Forest/Fences/Big wooden fence/Big-wooden-fence-4.png` | PNG | 14x44 | RGBA | 1,383 B | Vertical tall wooden fence segment |
| `Small Forest/Fences/Stone fence/Stone-fence-1.png` | PNG | 32x22 | RGBA | 1,396 B | Horizontal low stone wall section 1 |
| `Small Forest/Fences/Stone fence/Stone-fence-2.png` | PNG | 32x22 | RGBA | 1,383 B | Horizontal low stone wall section 2 |
| `Small Forest/Fences/Stone fence/Stone-fence-3.png` | PNG | 32x22 | RGBA | 1,409 B | Horizontal low stone wall section 3 |
| `Small Forest/Fences/Stone fence/Stone-fence-4.png` | PNG | 13x32 | RGB | 1,208 B | Vertical low stone wall section |
| `Small Forest/Fences/Wooden fence/Wooden-fence-1.png` | PNG | 29x17 | RGBA | 1,344 B | Standard horizontal wooden fence |
| `Small Forest/Fences/Wooden fence/Wooden-fence-2.png` | PNG | 32x17 | RGBA | 1,390 B | Extended wooden fence segment |
| `Small Forest/Fences/Wooden fence/Wooden-fence-3.png` | PNG | 29x17 | RGBA | 1,375 B | Wooden fence segment variant |
| `Small Forest/Fences/Wooden fence/Wooden-fence-4.png` | PNG | 15x36 | RGBA | 1,407 B | Vertical wooden fence segment |

#### D. Flowers and Grass Sprites (5 files)
| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Small Forest/Flowers/Flower-1.png` | PNG | 26x25 | RGBA | 1,389 B | Pink flower cluster (`TileFlower` / `assets.Flower1Image`) |
| `Small Forest/Flowers/Flower-2.png` | PNG | 24x22 | RGBA | 1,326 B | Red/orange flower cluster (`assets.Flower2Image`) |
| `Small Forest/Flowers/Flower-3.png` | PNG | 26x18 | RGBA | 1,322 B | Yellow/blue flower cluster (`assets.Flower3Image`) |
| `Small Forest/Grass/Grass-1.png` | PNG | 25x24 | RGBA | 1,263 B | Tuft of wild grass 1 (`assets.GrassTuft1Image`) |
| `Small Forest/Grass/Grass-2.png` | PNG | 31x15 | RGBA | 1,170 B | Tuft of wild grass 2 (`assets.GrassTuft2Image`) |

#### E. Ground Tilesets (5 files)
| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Small Forest/Ground tileset/Bright-grass-tileset.png` | PNG | 365x331 | RGBA | 7,823 B | Complete autotile sheet for bright green grass terrain |
| `Small Forest/Ground tileset/Dark-grass-tileset.png` | PNG | 365x331 | RGBA | 7,901 B | Complete autotile sheet for dark green forest grass |
| `Small Forest/Ground tileset/Earth-tileset.png` | PNG | 206x92 | RGBA | 1,764 B | Earth/dirt autotile transition sheet |
| `Small Forest/Ground tileset/Stone-path-tileset-horizontal.png` | PNG | 182x37 | RGBA | 1,497 B | Horizontal cobbled stone pathway tileset |
| `Small Forest/Ground tileset/Stone-path-tileset-vertical.png` | PNG | 37x182 | RGBA | 1,559 B | Vertical cobbled stone pathway tileset |

#### F. Sculptures and Stones (4 files)
| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Small Forest/Sculptures/Sculpture-1.png` | PNG | 23x31 | RGBA | 15,596 B | Carved stone garden statue 1 (`TileSculpture` / `assets.Sculpture1Image`) |
| `Small Forest/Sculptures/Sculture-2.png` | PNG | 29x32 | RGBA | 15,739 B | Carved stone garden statue 2 (`assets.Sculpture2Image`) |
| `Small Forest/Stones/Stone-1.png` | PNG | 28x19 | RGBA | 1,584 B | Small forest boulder (`TileStone` / `assets.Stone1Image`) |
| `Small Forest/Stones/Stone-2.png` | PNG | 29x25 | RGBA | 1,471 B | Medium mossy forest boulder (`assets.Stone2Image`) |

#### G. Trees (12 files across 3 species and 4 growth stages)
| Relative Path | Format | Dimensions | Mode | Size | Description / Suggested Mapping |
|---|---|---|---|---|---|
| `Small Forest/Trees/Tree-1/Tree-1-1.png` | PNG | 15x19 | RGBA | 1,503 B | Species 1 Sapling |
| `Small Forest/Trees/Tree-1/Tree-1-2.png` | PNG | 23x29 | RGBA | 1,805 B | Species 1 Young tree |
| `Small Forest/Trees/Tree-1/Tree-1-3.png` | PNG | 32x41 | RGBA | 2,149 B | Species 1 Mature tree |
| `Small Forest/Trees/Tree-1/Tree-1-4.png` | PNG | 37x50 | RGBA | 2,167 B | Species 1 Full-grown canopy tree |
| `Small Forest/Trees/Tree-2/Tree-2-1.png` | PNG | 15x18 | RGBA | 1,431 B | Species 2 Sapling |
| `Small Forest/Trees/Tree-2/Tree-2-2.png` | PNG | 23x29 | RGBA | 1,823 B | Species 2 Young tree |
| `Small Forest/Trees/Tree-2/Tree-2-3.png` | PNG | 32x39 | RGBA | 2,123 B | Species 2 Mature tree |
| `Small Forest/Trees/Tree-2/Tree-2-4.png` | PNG | 36x50 | RGBA | 2,099 B | Species 2 Full-grown pine/conifer tree |
| `Small Forest/Trees/Tree-3/Tree-3-1.png` | PNG | 15x19 | RGBA | 1,396 B | Species 3 Sapling |
| `Small Forest/Trees/Tree-3/Tree-3-2.png` | PNG | 23x30 | RGBA | 1,678 B | Species 3 Young tree |
| `Small Forest/Trees/Tree-3/Tree-3-3.png` | PNG | 28x46 | RGBA | 1,965 B | Species 3 Mature tree |
| `Small Forest/Trees/Tree-3/Tree-3-4.png` | PNG | 55x67 | RGBA | 2,500 B | Species 3 Giant oak tree |

---

### 2.4 Detailed Inventory of `Zombie Apocalypse Tileset/` (544 files)

The Zombie Apocalypse package contains the master reference sheet, 3 Photoshop source documents, and 532 separated PNG sprites in 46 thematic directories.

#### A. Master Reference & Source Files (4 files)
| Relative Path | Format | Dimensions | Size | Purpose |
|---|---|---|---|---|
| `Zombie Apocalypse Tileset/Zombie Apocalypse Tileset Reference.png` | PNG | 764x300 | 48,155 B | Master atlas with all sprites in grid alignment |
| `Zombie Apocalypse Tileset/Zombie Apocalypse Tileset.psd` | PSD | N/A | 613,137 B | Photoshop layered source file (excluded from build) |
| `Zombie Apocalypse Tileset/Organized separated sprites/Inventory interface/INTERFACE WITH NEW INVENTORY.psd` | PSD | N/A | 58,149 B | UI master source (excluded from build) |
| `Zombie Apocalypse Tileset/Organized separated sprites/Inventory interface/SLOT AND ITEMS.psd` | PSD | N/A | 33,095 B | UI master source (excluded from build) |

#### B. Character & Zombie Animation Sprites (83 PNG files)
| Subfolder | Count | Dimensions | Description |
|---|---|---|---|
| `Player Character Walking Animation Frames` | 9 | 11x15 to 14x16 | 8-directional walking cycle frames for Player |
| `Damaged Player Animation Frames` | 9 | 13x15 to 14x14 | Hit/damage reaction frames for Player |
| `Skinny Walking Zombie Animation` | 9 | 9x15 to 11x15 | Walking frames for standard skinny zombie |
| `Damaged Skinny Zombie Animation Frames` | 9 | 10x14 to 10x15 | Hit reaction frames for skinny zombie |
| `Big Zombie Walking Animation Frames` | 9 | 16x15 to 16x16 | Heavy zombie / brute walking animation |
| `Damaged Big Zombie Animation Frames` | 9 | 16x15 to 16x16 | Heavy zombie hit reaction animation |
| `Kid Zombie Animation Frames` | 9 | 8x10 to 12x11 | Fast runner / kid zombie walking cycle |
| `Damaged Kid Zombie Animation Frames` | 9 | 10x10 to 12x10 | Runner zombie hit reaction frames |
| `Turret Zombie Animation Frames` | 12 | 12x13 to 12x15 | Stationary / acid spitter zombie idle/aim |
| `Damaged Turret Zombie Animation Frames` | 9 | 12x15 | Turret zombie hit reaction frames |
| `Sitting Zombie` | 2 | 9x12 to 11x13 | Slumped / dormant ambusher zombie |
| `Dead Corpses With Flies Animation Frames` | 6 | 12x14 to 15x13 | Decaying zombie corpse with animated flies |

#### C. Weapons, Combat & VFX Sprites (38 PNG files)
| Subfolder | Count | Dimensions | Description |
|---|---|---|---|
| `Pistol Shooting Animation Frames` | 5 | 5x6 to 10x3 | Gunshot muzzle flash and recoil frames |
| `Shotgun Shooting Animation Frames` | 6 | 6x8 to 10x12 | Shotgun blast flash and smoke frames |
| `Knife Attack Animation Frames` | 4 | 11x10 to 16x11 | Melee knife slash arc animation |
| `Turret Zombie Vomit Shooting Animation Frames` | 7 | 5x10 to 14x8 | Acid spit projectile travel & splash |
| `Blood Animation Frames` | 5 | 12x3 to 15x9 | Blood splatter impact burst animation |
| `Random Blood Stains` | 5 | 7x9 to 10x9 | Floor decals for dried/fresh blood |
| `Explosion Animation Frames` | 6 | 10x10 to 16x16 | Multi-frame fiery explosive burst |
| `Exploding Barrel Animation Frames` | 4 | 12x16 | Red hazardous explosive barrel detonation |
| `Smoke Animation Frames` | 6 | 13x12 to 16x16 | Drifting smoke plume / puff animation |

#### D. Modular Buildings & Architectural Sprites (138 PNG files)
| Subfolder | Count | Dimensions | Description |
|---|---|---|---|
| `Modular Barns` | 58 | 13x9 to 16x16 | Red barn walls, corrugated roofs, barn doors, lofts, silos |
| `Modular Big Building` | 29 | 16x7 to 16x16 | Brick exterior walls, commercial windows, storefronts, signage |
| `Modular Small Building` | 15 | 16x7 to 16x16 | Residential siding, wood shingle roofs, residential doors |
| `Modular Fences` | 31 | 3x16 to 16x16 | Barbed wire, chain link, wooden pickets, metal gates |
| `Modular Road` | 26 | 16x16 to 18x20 | Asphalt intersections, lane markers, crosswalks, sidewalk curbs |
| `Modular Terrain Path` | 12 | 16x16 | Dirt paths, gravel trails, edge blending corners |

#### E. Vehicles & Industrial Props (30 PNG files)
| Subfolder | Count | Dimensions | Description |
|---|---|---|---|
| `Drivable Car with 8 Directions` | 8 | 18x33 to 33x18 | 8-directional drivable sedan vehicle |
| `Broken Cars and Tires` | 15 | 7x8 to 33x18 | Wrecked station wagons, burned chassis, loose tires |
| `Gas Station` | 3 | 17x14 to 99x90 | Gas pump island, fuel pumps, large canopy |
| `Store Truck with Smoking Guy Animation Frames` | 3 | 40x23 to 41x23 | Armored delivery truck with merchant NPC |
| `Tractor` | 4 | 13x11 to 16x16 | Farm tractor body, oversized rear wheels, engine block |

#### F. Items, Loot & Equipment (20 PNG files)
| Subfolder | Count | Dimensions | Description |
|---|---|---|---|
| `Pickable Items and Weapons` | 20 | 3x8 to 16x5 | Pistols, shotguns, knives, ammo boxes, medkits, canned food, water bottles, keys, batteries, gold coins |

#### G. World Environment, Foliage & Wildlife (56 PNG files)
| Subfolder | Count | Dimensions | Description |
|---|---|---|---|
| `Trees` | 9 | 7x16 to 16x16 | Dead leafless trees, pine trees, autumnal trees |
| `Modular Bushes` | 10 | 8x9 to 16x16 | Brambles, hedge rows, overgrown shrubs |
| `Modular Stacked Straw` | 6 | 8x5 to 16x16 | Hay bales, stacked straw piles |
| `Different Crops Lengths` | 8 | 16x9 to 16x16 | Corn stalks, wheat, potato plants at various growth heights |
| `Scarecrow` | 2 | 9x10 to 16x15 | Farm scarecrow on wooden stake |
| `Tombstone` | 1 | 14x11 | Graveyard headstone |
| `Terrain Variations` | 4 | 16x16 | Cracked mud, grassy patches, rock outcrop |
| `Terrain wall` | 1 | 16x16 | Cliff edge / elevation boundary |
| `90º Rotatable Bridge Sprites` | 3 | 16x16 | Wooden plank bridge segments |
| `Water animation frames` | 3 | 16x16 | Animated rippling river/pond water |
| `Under Bridge Water animation frames` | 6 | 16x16 | Shaded water under bridge piers |
| `Windmill with Fan Animation Frames` | 7 | 5x5 to 16x16 | Functional windmill structure with rotating blades |
| `Black Bird Flying and Ground Eating Animation Frames` | 12 | 3x6 to 13x11 | Crows / ravens in flight and pecking on ground |
| `White Bird Flying, Ground Eating + Being Shot Blood Animation Frames` | 12 | 3x6 to 13x11 | Doves / seagulls flying, feeding, and exploding on impact |

#### H. UI & Interactive Elements (44 PNG files)
| Subfolder | Count | Dimensions | Description |
|---|---|---|---|
| `Inventory interface` | 15 | 32x22 to 480x270 | Full HUD layout, equipment slots, hotbar, inventory grid |
| `UI Elements` | 23 | 1x3 to 40x6 | Health bars, stamina gauges, ammo counters, selection brackets |
| `Spawning Item Box Animation Frames + Broken Box Pieces` | 6 | 10x27 to 15x13 | Wooden crate drop, impact burst, wood debris |
| `Spawning Money Animation Frames` | 5 | 7x4 to 12x11 | Floating dollar signs / coin sparkles |
| `Shootable Coke Can Animation Frames` | 5 | 6x10 to 10x12 | Soda can bouncing / getting shot |
| `Music Notes Animation Frames` | 3 | 9x6 to 13x8 | Floating audio notes for radios/jukeboxes |
| `Zombie Poster` | 1 | 11x14 | Wanted / quarantine warning poster |

---

## 3. Procedural Pipeline (`cmd/tools/genassets`) Forensic Survey

### 3.1 Directory Structure & Line Counts
```
cmd/tools/genassets/
├── main.go            # 2,413 lines (60,094 bytes) — 27 procedural math generation routines
└── genassets_test.go  #   119 lines  (4,686 bytes) — Determinism & PNG dimension verification tests
```

### 3.2 Analysis of Generated Assets (27 files)
`cmd/tools/genassets/main.go` procedurally computed 27 PNG files in `internal/assets/images/`:
- **Character Entities (64x128)**: `player.png`, `zombie.png`, `runner.png`.
- **Floor Tiles (256x128)**: `grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png`.
- **Vertical Obstacles / Props (256x256)**: `wall.png`, `tree.png`, `fence.png`, `debris.png`, `tent.png`, `stump.png`, `mushroom.png`, `sign.png`, `elevation_block.png`, `elevation_ramp.png`.
- **Items & Equipment (64x64)**: `weapon.png`, `axe.png`, `shotgun.png`, `ammo.png`, `armor.png`, `antidote.png`, `food.png`, `water.png`.

### 3.3 Codebase References to `genassets`
Grep search identified references across the repository:

1. **Repository Root Binary**:
   - `/home/bryce/code/go-zomboid/genassets` (ELF 64-bit executable, 2,367,008 bytes).
2. **Go Test Suites**:
   - `cmd/tools/genassets/genassets_test.go`: Executes `go run ./cmd/tools/genassets`.
   - `internal/assets/empirical_challenger_test.go` (Lines 302–357):
     ```go
     // TestEmpiricalGenerationDeterminism checks SHA-256 hashes across repeated executions of genassets.
     func TestEmpiricalGenerationDeterminism(t *testing.T) {
         ...
         cmd := exec.Command("go", "run", "./cmd/tools/genassets")
         ...
     }
     ```
3. **Documentation & Project Specifications**:
   - `README.md` (Lines 10, 40): References `genassets` in project overview and build instructions.
   - `PROJECT.md` (Lines 10, 34, 41, 67, 68): Lists `genassets` in architecture and milestone tables.
   - `TEST_READY.md` (Line 40): References `go test -v ./cmd/tools/genassets`.
   - `TEST_INFRA.md` (Line 30): References `go test -v ./cmd/tools/genassets`.
   - `ART_STYLE_GUIDE.md` (Lines 4, 12): References procedural generation in `cmd/tools/genassets/main.go`.

### 3.4 Impact of Retirement (R1)
- When `cmd/tools/genassets` is deleted, `TestEmpiricalGenerationDeterminism` in `internal/assets/empirical_challenger_test.go` will fail unless it is removed or retired.
- `cmd/tools/` will be empty once `genassets` is deleted and can be pruned if desired.
- The root binary `./genassets` should be removed to ensure no stale artifacts remain.

---

## 4. Current Asset Loading & Embedding Architecture (`internal/assets/`)

### 4.1 Embedding Mechanism
In `internal/assets/assets.go`:
```go
//go:embed images/*
var imageFS embed.FS
```
- Standard library `embed.FS` embeds all files in `images/`.
- In Go 1.16+, when a directory is matched by `images/*`, all files and subdirectories within that directory are recursively embedded in the binary.
- Loading helper `loadEbitenImage(path string)` opens the embedded byte slice, passes it to `image.Decode()`, and converts it to GPU/memory format via `ebiten.NewImageFromImage(img)`.

### 4.2 Current Exported `ebiten.Image` Variables
`internal/assets/assets.go` currently exports 27 pointers:

```go
var (
	// Entity Sprites
	PlayerImage *ebiten.Image
	ZombieImage *ebiten.Image
	RunnerImage *ebiten.Image

	// Floor Tiles
	GrassImage     *ebiten.Image
	DirtImage      *ebiten.Image
	WoodImage      *ebiten.Image
	AsphaltImage   *ebiten.Image
	ConcreteImage  *ebiten.Image
	TileFloorImage *ebiten.Image

	// Vertical Obstacles / Props
	WallImage           *ebiten.Image
	TreeImage           *ebiten.Image
	FenceImage          *ebiten.Image
	DebrisImage         *ebiten.Image
	TentImage           *ebiten.Image
	StumpImage          *ebiten.Image
	MushroomImage       *ebiten.Image
	SignImage           *ebiten.Image
	ElevationBlockImage *ebiten.Image
	ElevationRampImage  *ebiten.Image

	// Item / Weapon / Armor Sprites
	WeaponImage   *ebiten.Image
	AxeImage      *ebiten.Image
	ShotgunImage  *ebiten.Image
	AmmoImage     *ebiten.Image
	ArmorImage    *ebiten.Image
	AntidoteImage *ebiten.Image
	FoodImage     *ebiten.Image
	WaterImage    *ebiten.Image
)
```

### 4.3 Test Suite Constraints in `internal/assets/`
The package `internal/assets` has four comprehensive test files:
1. `assets_test.go`: Validates that all 27 pointers are non-nil after `Load()`, and verifies exact dimensions and non-empty pixel alpha.
2. `assets_stress_test.go`: Validates 2:1 isometric diamond geometry on floor tiles, grounding anchors on characters, item outlines, and `Load()` idempotency.
3. `challenger_stress_test.go`: Validates multi-threaded concurrent `Load()` calls, ITU-R BT.601 color saturation, RMS contrast, and dynamic range.
4. `empirical_challenger_test.go`: Validates fill ratios, diamond boundaries, character ground contact rows 112..127, item centroids, and `TestEmpiricalGenerationDeterminism` (the sole `genassets` dependency).

**Critical Insight**: By keeping the existing 27 baseline PNGs in `internal/assets/images/` and appending the new assets from `context/`, all existing tests for legacy textures will pass 100% with zero regressions.

---

## 5. File Copying and Code Refactoring Plan for R1 & R2

### 5.1 Plan for Requirement R1: Retire Procedural Generation

1. **Delete Procedural Directory**:
   ```bash
   rm -rf /home/bryce/code/go-zomboid/cmd/tools/genassets
   ```
2. **Remove Root Binary**:
   ```bash
   rm -f /home/bryce/code/go-zomboid/genassets
   ```
3. **Update Test File (`internal/assets/empirical_challenger_test.go`)**:
   - Remove `TestEmpiricalGenerationDeterminism` (lines 302–357) which executes `go run ./cmd/tools/genassets`.
4. **Update Documentation**:
   - Update `README.md` to remove `go run ./cmd/tools/genassets` instructions.
   - Update `PROJECT.md`, `TEST_READY.md`, `TEST_INFRA.md`, `ART_STYLE_GUIDE.md` to note that procedural generation has been retired in favor of direct embedded external PNG assets.

---

### 5.2 Plan for Requirement R2: External Asset Ingestion

#### A. File Ingestion & Organization in `internal/assets/images/`
1. **Copy all PNG assets from `context/`**:
   - Copy `context/Lab/` -> `internal/assets/images/Lab/`
   - Copy `context/Small Forest/` -> `internal/assets/images/Small Forest/`
   - Copy `context/Zombie Apocalypse Tileset/` -> `internal/assets/images/Zombie Apocalypse Tileset/`
2. **Filter out unwanted files**:
   - Omit all `.DS_Store` files.
   - Omit all `.psd` files (`Zombie Apocalypse Tileset.psd`, `INTERFACE WITH NEW INVENTORY.psd`, `SLOT AND ITEMS.psd`).
   - Omit all `:Zone.Identifier` metadata files.
3. **Add Direct Prop Aliases**:
   - In addition to preserving the hierarchical subfolders, create canonical top-level copies/names in `internal/assets/images/` for direct access:
     - `bench.png` (from `Small Forest/Bench and chest/Bench.png`)
     - `chest.png` (from `Small Forest/Bench and chest/Chest.png`)
     - `sculpture_1.png` (from `Small Forest/Sculptures/Sculpture-1.png`)
     - `sculpture_2.png` (from `Small Forest/Sculptures/Sculture-2.png`)
     - `bush_1.png` (from `Small Forest/Bushes/Bush-1.png`)
     - `flower_1.png` (from `Small Forest/Flowers/Flower-1.png`)
     - `stone_1.png` (from `Small Forest/Stones/Stone-1.png`)

#### B. Code Refactoring in `internal/assets/assets.go`
1. **Declare New Exported Variables**:
   ```go
   var (
       // New World Props from context/
       BenchImage       *ebiten.Image
       ChestImage       *ebiten.Image
       Sculpture1Image  *ebiten.Image
       Sculpture2Image  *ebiten.Image
       SculptureImage   *ebiten.Image // Alias to Sculpture1Image
       Bush1Image       *ebiten.Image
       BushImage        *ebiten.Image // Alias to Bush1Image
       Flower1Image     *ebiten.Image
       FlowerImage      *ebiten.Image // Alias to Flower1Image
       Stone1Image      *ebiten.Image
       StoneImage       *ebiten.Image // Alias to Stone1Image

       // Tileset Sheets
       LabTilesetImage    *ebiten.Image
       ZombieTilesetImage *ebiten.Image
   )
   ```
2. **Update `Load()` in `assets.go`**:
   ```go
   // New World Props
   BenchImage = loadEbitenImage("images/Small Forest/Bench and chest/Bench.png")
   ChestImage = loadEbitenImage("images/Small Forest/Bench and chest/Chest.png")
   Sculpture1Image = loadEbitenImage("images/Small Forest/Sculptures/Sculpture-1.png")
   Sculpture2Image = loadEbitenImage("images/Small Forest/Sculptures/Sculture-2.png")
   SculptureImage = Sculpture1Image
   Bush1Image = loadEbitenImage("images/Small Forest/Bushes/Bush-1.png")
   BushImage = Bush1Image
   Flower1Image = loadEbitenImage("images/Small Forest/Flowers/Flower-1.png")
   FlowerImage = Flower1Image
   Stone1Image = loadEbitenImage("images/Small Forest/Stones/Stone-1.png")
   StoneImage = Stone1Image

   // Sheets
   LabTilesetImage = loadEbitenImage("images/Lab/Inside_C.png")
   ZombieTilesetImage = loadEbitenImage("images/Zombie Apocalypse Tileset/Zombie Apocalypse Tileset Reference.png")
   ```
3. **Maintain Existing 27 Pointers**:
   Keep all existing 27 pointer variables and their loading calls intact so existing systems (`game.go`, `combat.go`, `camera.go`) remain 100% operational.

---

## 6. Downstream Inferences & Integration Plan (R3 Roadmap)

For the subsequent R3 task, the following mappings are inferred:

### 6.1 Map `TileType` Constants in `internal/game/world/map.go`
Add new tile types for the imported environmental objects:
```go
const (
    TileGrass TileType = iota
    TileWall
    TileDirt
    TileWoodFloor
    TileTree
    TileAsphalt
    TileConcrete
    TileTileFloor
    TileFence
    TileDebris
    TileTent
    TileElevationBlock
    TileRamp
    TileStump
    TileMushroom
    TileSign
    // New R3 Tile Types:
    TileBench
    TileChest
    TileSculpture
    TileBush
    TileFlower
    TileStone
)
```

Update helper functions in `map.go`:
- `IsSolid()`: Return `true` for `TileBench`, `TileChest`, `TileSculpture`, `TileStone`. Return `false` for `TileBush`, `TileFlower`.
- `BlocksVision()`: Return `false` for low props (`TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`).
- `IsFloor()`: Return `false` for vertical props (`TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`).
- `String()`: Add string cases for debugging.

### 6.2 Map Generation & Placement in `internal/game/world/map.go`
- In `NewMap()` procedural generation, place benches in residential gardens and town parks, chests inside residential/police/warehouse rooms, and sculptures/bushes/flowers/stones in outdoor plazas and park areas.
- Ensure that existing tile counts (Grass, Wall, Dirt, WoodFloor, Tree, Asphalt, Concrete, TileFloor, Fence, Debris) tested by `TestEmpirical_All10TileTypesGenerated` remain positive.

### 6.3 DrawSystem & Depth Sorting in `internal/game/game.go`
- Update `DrawSystem`:
  - In the vertical obstacle rendering loop, include `TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`.
  - Map each new tile type to its corresponding `*ebiten.Image` (`assets.BenchImage`, `assets.ChestImage`, `assets.SculptureImage`, `assets.BushImage`, `assets.FlowerImage`, `assets.StoneImage`).
  - Calculate `Depth: worldX + worldY` so new props are correctly depth-sorted behind and in front of the player and zombies.

---

## 7. Verification & Acceptance Checklist

| Item | Verification Command / Method | Expected Result |
|---|---|---|
| R1: `genassets` Deletion | `test ! -d cmd/tools/genassets && test ! -f genassets` | Directory and binary no longer exist |
| R2: PNG Asset Ingestion | `ls internal/assets/images/Lab internal/assets/images/"Small Forest"` | All PNG files present and non-empty |
| R2: Asset Loading | `CC=gcc go test -v ./internal/assets/...` | All asset loading and dimension tests pass |
| Full Test Suite | `CC=gcc go test ./...` | 100% tests pass across all packages |
| Game Launch & Render | `CC=gcc go run ./cmd/game` | Game starts without panic, renders world |
