# Comprehensive Remediation Analysis: Assets & DrawSystem Depth Sorting

## 1. Executive Summary & Root Cause Analysis

### 1.1 Summary of Findings
During Milestone 1 and Milestone 4 iterations, an erroneous modification was made to `internal/assets/assets.go` (commit `7e05822`), where 19 legacy image pointer variables (`PlayerImage`, `ZombieImage`, `RunnerImage`, `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`, `TentImage`, `StumpImage`, `MushroomImage`, `SignImage`, `ElevationBlockImage`, `ElevationRampImage`) were repointed from their canonical 256x128/256x256/64x128 image files to small external PNG files from `context/` (e.g. 14x15 player sprite, 25x24 grass sprite, 32x17 fence sprite).

This created two critical regressions:
1. **Legacy Asset Contract Breach**: Existing tests in `internal/assets` that assert on expected dimensions (e.g., `TestEmbeddedAssetDimensionsAndValidity`, `TestAssetsLoadAllPointersNonNil`, `TestChallenger_All27ExportedPointersAndExactBounds`) failed because the loaded image dimensions did not match the legacy specifications.
2. **Geometric Anchor Distortion in DrawSystem**: In `internal/game/game.go`, vertical obstacle and prop sprites are positioned using the universal anchor transformation `op.GeoM.Translate(-imgW/2.0, 128.0 - imgH)`. When `WallImage` and `TreeImage` were repointed to small 32x17 and 15x19 sprites, this formula evaluated to `(-16.0, 111.0)` and `(-7.5, 109.0)` respectively, failing the geometric anchor test `TestDrawSystem_SpriteGeometricAnchors` which expects `(-128.0, -128.0)` for 256x256 obstacle tiles.

### 1.2 Remediation Goal
1. Restore all **27 legacy pointers** in `internal/assets/assets.go` to load from `images/<name>.png`.
2. Maintain all **22 new external asset pointers** in `internal/assets/assets.go` loading from their ingested paths in `images/Small Forest/...`, `images/Lab/...`, and `images/Zombie Apocalypse Tileset/...`.
3. Provide full test suites in `internal/assets/assets_test.go`, `internal/assets/challenger_stress_test.go`, and `internal/game/draw_depth_test.go`.
4. Ensure 100% test pass rate (`CC=gcc go test ./...`) and clean build (`CC=gcc go build ./cmd/game`).

---

## 2. Complete Asset Inventory

### Part A: 27 Legacy `*ebiten.Image` Pointers

| # | Pointer Variable | Category | Embedded Path | Width | Height | Role in Game Engine |
|---|------------------|----------|---------------|-------|--------|---------------------|
| 1 | `PlayerImage` | Entity | `images/player.png` | 64 | 128 | Player character sprite |
| 2 | `ZombieImage` | Entity | `images/zombie.png` | 64 | 128 | Standard walker zombie sprite |
| 3 | `RunnerImage` | Entity | `images/runner.png` | 64 | 128 | Fast runner zombie sprite |
| 4 | `GrassImage` | Floor Tile | `images/grass.png` | 256 | 128 | Base terrain & ground diamond under props |
| 5 | `DirtImage` | Floor Tile | `images/dirt.png` | 256 | 128 | Dirt terrain ground diamond |
| 6 | `WoodImage` | Floor Tile | `images/wood.png` | 256 | 128 | Indoor wooden floor ground diamond |
| 7 | `AsphaltImage` | Floor Tile | `images/asphalt.png` | 256 | 128 | Road / asphalt ground diamond |
| 8 | `ConcreteImage` | Floor Tile | `images/concrete.png` | 256 | 128 | Sidewalk / concrete ground diamond |
| 9 | `TileFloorImage` | Floor Tile | `images/tile_floor.png` | 256 | 128 | Indoor tiled floor ground diamond |
| 10 | `WallImage` | Obstacle/Prop | `images/wall.png` | 256 | 256 | Solid vision-blocking wall obstacle |
| 11 | `TreeImage` | Obstacle/Prop | `images/tree.png` | 256 | 256 | Solid foliage obstacle |
| 12 | `FenceImage` | Obstacle/Prop | `images/fence.png` | 256 | 256 | Solid perimeter fence obstacle |
| 13 | `DebrisImage` | Obstacle/Prop | `images/debris.png` | 256 | 256 | Solid rubble/debris obstacle |
| 14 | `TentImage` | Obstacle/Prop | `images/tent.png` | 256 | 256 | Solid campsite tent obstacle |
| 15 | `StumpImage` | Obstacle/Prop | `images/stump.png` | 256 | 256 | Solid tree stump obstacle |
| 16 | `MushroomImage` | Obstacle/Prop | `images/mushroom.png` | 256 | 256 | Environmental mushroom obstacle |
| 17 | `SignImage` | Obstacle/Prop | `images/sign.png` | 256 | 256 | Solid wooden signpost obstacle |
| 18 | `ElevationBlockImage` | Obstacle/Prop | `images/elevation_block.png` | 256 | 256 | Solid raised terrain block obstacle |
| 19 | `ElevationRampImage` | Obstacle/Prop | `images/elevation_ramp.png` | 256 | 256 | Terrain elevation ramp obstacle |
| 20 | `FoodImage` | Item / Pickup | `images/food.png` | 64 | 64 | Food consumable item icon |
| 21 | `WaterImage` | Item / Pickup | `images/water.png` | 64 | 64 | Water consumable item icon |
| 22 | `WeaponImage` | Item / Pickup | `images/weapon.png` | 64 | 64 | Melee spiked club/bat item icon |
| 23 | `AxeImage` | Item / Pickup | `images/axe.png` | 64 | 64 | Fire axe item icon |
| 24 | `ShotgunImage` | Item / Pickup | `images/shotgun.png` | 64 | 64 | Pump shotgun item icon |
| 25 | `AmmoImage` | Item / Pickup | `images/ammo.png` | 64 | 64 | Shotgun ammunition box icon |
| 26 | `ArmorImage` | Item / Pickup | `images/armor.png` | 64 | 64 | Tactical armor vest item icon |
| 27 | `AntidoteImage` | Item / Pickup | `images/antidote.png` | 64 | 64 | Zombie infection cure antidote icon |

---

### Part B: 22 New External Asset Pointers

| # | Pointer Variable | Ingested PNG Path | Width | Height | Game World Role |
|---|------------------|-------------------|-------|--------|-----------------|
| 1 | `BenchImage` | `images/Small Forest/Bench and chest/Bench.png` | 52 | 37 | `TileBench` (16) solid world prop |
| 2 | `ChestImage` | `images/Small Forest/Bench and chest/Chest.png` | 22 | 21 | `TileChest` (17) solid loot container prop |
| 3 | `Sculpture1Image` | `images/Small Forest/Sculptures/Sculpture-1.png` | 23 | 31 | `TileSculpture` (18) variant 1 solid prop |
| 4 | `Sculpture2Image` | `images/Small Forest/Sculptures/Sculture-2.png` | 29 | 32 | `TileSculpture` variant 2 solid prop |
| 5 | `SculptureImage` | (alias to `Sculpture1Image`) | 23 | 31 | Primary `TileSculpture` pointer |
| 6 | `Bush1Image` | `images/Small Forest/Bushes/Bush-1.png` | 24 | 18 | `TileBush` (19) variant 1 walkable foliage |
| 7 | `Bush2Image` | `images/Small Forest/Bushes/Bush-2.png` | 19 | 15 | `TileBush` variant 2 walkable foliage |
| 8 | `Bush3Image` | `images/Small Forest/Bushes/Bush-3.png` | 25 | 19 | `TileBush` variant 3 walkable foliage |
| 9 | `Bush4Image` | `images/Small Forest/Bushes/Bush-4.png` | 28 | 19 | `TileBush` variant 4 walkable foliage |
| 10 | `BushImage` | (alias to `Bush1Image`) | 24 | 18 | Primary `TileBush` pointer |
| 11 | `Flower1Image` | `images/Small Forest/Flowers/Flower-1.png` | 26 | 25 | `TileFlower` (20) variant 1 walkable foliage |
| 12 | `Flower2Image` | `images/Small Forest/Flowers/Flower-2.png` | 24 | 22 | `TileFlower` variant 2 walkable foliage |
| 13 | `Flower3Image` | `images/Small Forest/Flowers/Flower-3.png` | 26 | 18 | `TileFlower` variant 3 walkable foliage |
| 14 | `FlowerImage` | (alias to `Flower1Image`) | 26 | 25 | Primary `TileFlower` pointer |
| 15 | `Stone1Image` | `images/Small Forest/Stones/Stone-1.png` | 28 | 19 | `TileStone` (21) variant 1 solid obstacle |
| 16 | `Stone2Image` | `images/Small Forest/Stones/Stone-2.png` | 29 | 25 | `TileStone` variant 2 solid obstacle |
| 17 | `StoneImage` | (alias to `Stone1Image`) | 28 | 19 | Primary `TileStone` pointer |
| 18 | `ForestStumpImage` | `images/Small Forest/Bushes/Stump.png` | 29 | 19 | External forest stump prop |
| 19 | `GrassTuft1Image` | `images/Small Forest/Grass/Grass-1.png` | 25 | 24 | External decorative grass tuft 1 |
| 20 | `GrassTuft2Image` | `images/Small Forest/Grass/Grass-2.png` | 31 | 15 | External decorative grass tuft 2 |
| 21 | `LabTilesetImage` | `images/Lab/Inside_C.png` | 768 | 768 | External laboratory tileset sheet |
| 22 | `ZombieTilesetImage` | `images/Zombie Apocalypse Tileset/Zombie Apocalypse Tileset Reference.png` | 764 | 300 | External zombie apocalypse reference sheet |

---

## 3. Geometric Anchor & Depth-Sorting System Analysis

### 3.1 Isometric Projection Math
In `internal/game/game.go`, coordinates are transformed via 2:1 isometric projection:
$$\text{isoX} = \text{worldX} - \text{worldY}$$
$$\text{isoY} = \frac{\text{worldX} + \text{worldY}}{2}$$

The base diamond of a tile with top vertex at $(\text{isoX}, \text{isoY})$ has:
- Left vertex: $(\text{isoX} - 128, \text{isoY} + 64)$
- Right vertex: $(\text{isoX} + 128, \text{isoY} + 64)$
- Bottom vertex: $(\text{isoX}, \text{isoY} + 128)$

### 3.2 Dynamic Geometric Anchor Formula
For any sprite of dimensions $(W, H)$, the drawing pipeline in `game.go` applies:
$$\text{transX} = -\frac{W}{2.0}$$
$$\text{transY} = 128.0 - H$$

#### Mathematical Behavior:
1. **Horizontal Centering**: Translating by $-W/2$ aligns the horizontal center of the sprite with $\text{isoX}$ (the vertical centerline of the isometric tile diamond).
2. **Vertical Ground Anchoring**: Translating by $128.0 - H$ aligns the bottom edge ($y = H$) of the sprite with $\text{isoY} + 128.0$ (the bottom vertex of the diamond).
3. **Legacy 256x256 Reduction**:
   $$\text{transX} = -\frac{256}{2.0} = -128.0$$
   $$\text{transY} = 128.0 - 256.0 = -128.0$$
   This matches the exact legacy coordinate offset $(-128.0, -128.0)$!
4. **New Prop Reduction**:
   - Bench ($52 \times 37$): $\text{transX} = -26.0$, $\text{transY} = 91.0$
   - Chest ($22 \times 21$): $\text{transX} = -11.0$, $\text{transY} = 107.0$
   - Sculpture ($23 \times 31$): $\text{transX} = -11.5$, $\text{transY} = 97.0$
   - Bush ($24 \times 18$): $\text{transX} = -12.0$, $\text{transY} = 110.0$
   - Flower ($26 \times 25$): $\text{transX} = -13.0$, $\text{transY} = 103.0$
   - Stone ($28 \times 19$): $\text{transX} = -14.0$, $\text{transY} = 109.0$

### 3.3 Two-Pass Rendering Architecture in `DrawSystem.Draw`
1. **Pass 1 (Base Ground Pass)**:
   - Renders flat ground diamond tiles (`assets.GrassImage`, `assets.DirtImage`, etc.).
   - All obstacle and prop tiles (`world.TileWall`, `world.TileTree`, `world.TileBench`, `world.TileChest`, `world.TileSculpture`, `world.TileBush`, `world.TileFlower`, `world.TileStone`) render `assets.GrassImage` in Pass 1 to prevent transparent holes beneath vertical props.
2. **Pass 2 (Depth-Sorted Vertical Sprites)**:
   - Collects all vertical obstacles, props, items, and entities into `sprites []Renderable`.
   - Assigns depth key: $\text{Depth} = \text{worldX} + \text{worldY}$.
   - Executes `sort.SliceStable` to guarantee back-to-front rendering order without depth flicker.

---

## 4. Concrete Code Remediations

### 4.1 Target File 1: `internal/assets/assets.go`
Restore all 27 legacy pointers to their canonical `images/<name>.png` paths and load all 22 external pointers.

```go
package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed images/*
var imageFS embed.FS

var loadOnce sync.Once

var (
	// Entity Sprites (64x128)
	PlayerImage *ebiten.Image
	ZombieImage *ebiten.Image
	RunnerImage *ebiten.Image

	// Floor Tiles (256x128)
	GrassImage     *ebiten.Image
	DirtImage      *ebiten.Image
	WoodImage      *ebiten.Image
	AsphaltImage   *ebiten.Image
	ConcreteImage  *ebiten.Image
	TileFloorImage *ebiten.Image

	// Vertical Obstacles / Props (256x256)
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

	// Item / Weapon / Armor Sprites (64x64)
	WeaponImage   *ebiten.Image
	AxeImage      *ebiten.Image
	ShotgunImage  *ebiten.Image
	AmmoImage     *ebiten.Image
	ArmorImage    *ebiten.Image
	AntidoteImage *ebiten.Image
	FoodImage     *ebiten.Image
	WaterImage    *ebiten.Image

	// External World Props & Foliage (from context/)
	BenchImage       *ebiten.Image
	ChestImage       *ebiten.Image
	Sculpture1Image  *ebiten.Image
	Sculpture2Image  *ebiten.Image
	SculptureImage   *ebiten.Image
	Bush1Image       *ebiten.Image
	Bush2Image       *ebiten.Image
	Bush3Image       *ebiten.Image
	Bush4Image       *ebiten.Image
	BushImage        *ebiten.Image
	Flower1Image     *ebiten.Image
	Flower2Image     *ebiten.Image
	Flower3Image     *ebiten.Image
	FlowerImage      *ebiten.Image
	Stone1Image      *ebiten.Image
	Stone2Image      *ebiten.Image
	StoneImage       *ebiten.Image
	ForestStumpImage *ebiten.Image
	GrassTuft1Image  *ebiten.Image
	GrassTuft2Image  *ebiten.Image

	// External Tilesets (from context/)
	LabTilesetImage    *ebiten.Image
	ZombieTilesetImage *ebiten.Image
)

func Load() {
	loadOnce.Do(func() {
		// 1. Entities (64x128)
		PlayerImage = loadEbitenImage("images/player.png")
		ZombieImage = loadEbitenImage("images/zombie.png")
		RunnerImage = loadEbitenImage("images/runner.png")

		// 2. Floor Tiles (256x128)
		GrassImage = loadEbitenImage("images/grass.png")
		DirtImage = loadEbitenImage("images/dirt.png")
		WoodImage = loadEbitenImage("images/wood.png")
		AsphaltImage = loadEbitenImage("images/asphalt.png")
		ConcreteImage = loadEbitenImage("images/concrete.png")
		TileFloorImage = loadEbitenImage("images/tile_floor.png")

		// 3. Vertical Obstacles & Props (256x256)
		WallImage = loadEbitenImage("images/wall.png")
		TreeImage = loadEbitenImage("images/tree.png")
		FenceImage = loadEbitenImage("images/fence.png")
		DebrisImage = loadEbitenImage("images/debris.png")
		TentImage = loadEbitenImage("images/tent.png")
		StumpImage = loadEbitenImage("images/stump.png")
		MushroomImage = loadEbitenImage("images/mushroom.png")
		SignImage = loadEbitenImage("images/sign.png")
		ElevationBlockImage = loadEbitenImage("images/elevation_block.png")
		ElevationRampImage = loadEbitenImage("images/elevation_ramp.png")

		// 4. Items / Weapons / Armor (64x64)
		WeaponImage = loadEbitenImage("images/weapon.png")
		AxeImage = loadEbitenImage("images/axe.png")
		ShotgunImage = loadEbitenImage("images/shotgun.png")
		AmmoImage = loadEbitenImage("images/ammo.png")
		ArmorImage = loadEbitenImage("images/armor.png")
		AntidoteImage = loadEbitenImage("images/antidote.png")
		FoodImage = loadEbitenImage("images/food.png")
		WaterImage = loadEbitenImage("images/water.png")

		// 5. External World Props & Foliage (from context/)
		BenchImage = loadEbitenImage("images/Small Forest/Bench and chest/Bench.png")
		ChestImage = loadEbitenImage("images/Small Forest/Bench and chest/Chest.png")
		Sculpture1Image = loadEbitenImage("images/Small Forest/Sculptures/Sculpture-1.png")
		Sculpture2Image = loadEbitenImage("images/Small Forest/Sculptures/Sculture-2.png")
		SculptureImage = Sculpture1Image
		Bush1Image = loadEbitenImage("images/Small Forest/Bushes/Bush-1.png")
		Bush2Image = loadEbitenImage("images/Small Forest/Bushes/Bush-2.png")
		Bush3Image = loadEbitenImage("images/Small Forest/Bushes/Bush-3.png")
		Bush4Image = loadEbitenImage("images/Small Forest/Bushes/Bush-4.png")
		BushImage = Bush1Image
		Flower1Image = loadEbitenImage("images/Small Forest/Flowers/Flower-1.png")
		Flower2Image = loadEbitenImage("images/Small Forest/Flowers/Flower-2.png")
		Flower3Image = loadEbitenImage("images/Small Forest/Flowers/Flower-3.png")
		FlowerImage = Flower1Image
		Stone1Image = loadEbitenImage("images/Small Forest/Stones/Stone-1.png")
		Stone2Image = loadEbitenImage("images/Small Forest/Stones/Stone-2.png")
		StoneImage = Stone1Image
		ForestStumpImage = loadEbitenImage("images/Small Forest/Bushes/Stump.png")
		GrassTuft1Image = loadEbitenImage("images/Small Forest/Grass/Grass-1.png")
		GrassTuft2Image = loadEbitenImage("images/Small Forest/Grass/Grass-2.png")

		// 6. External Tilesets (from context/)
		LabTilesetImage = loadEbitenImage("images/Lab/Inside_C.png")
		ZombieTilesetImage = loadEbitenImage("images/Zombie Apocalypse Tileset/Zombie Apocalypse Tileset Reference.png")
	})
}

func loadEbitenImage(path string) *ebiten.Image {
	data, err := imageFS.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read embedded image %s: %v", path, err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to decode image %s: %v", path, err)
	}

	return ebiten.NewImageFromImage(img)
}
```

---

### 4.2 Target File 2: `internal/assets/assets_test.go`
Create unit test suite validating embedded assets and non-nil pointer dimensions.

```go
package assets

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestEmbeddedAssetDimensionsAndValidity(t *testing.T) {
	expectedAssets := []struct {
		path   string
		width  int
		height int
	}{
		// Character Entities (64x128)
		{"images/player.png", 64, 128},
		{"images/zombie.png", 64, 128},
		{"images/runner.png", 64, 128},

		// Floor Tiles (256x128)
		{"images/grass.png", 256, 128},
		{"images/dirt.png", 256, 128},
		{"images/wood.png", 256, 128},
		{"images/asphalt.png", 256, 128},
		{"images/concrete.png", 256, 128},
		{"images/tile_floor.png", 256, 128},

		// Vertical Obstacles & Props (256x256)
		{"images/wall.png", 256, 256},
		{"images/tree.png", 256, 256},
		{"images/fence.png", 256, 256},
		{"images/debris.png", 256, 256},
		{"images/tent.png", 256, 256},
		{"images/stump.png", 256, 256},
		{"images/mushroom.png", 256, 256},
		{"images/sign.png", 256, 256},
		{"images/elevation_block.png", 256, 256},
		{"images/elevation_ramp.png", 256, 256},

		// Items, Weapons & Equipment (64x64)
		{"images/food.png", 64, 64},
		{"images/water.png", 64, 64},
		{"images/weapon.png", 64, 64},
		{"images/axe.png", 64, 64},
		{"images/shotgun.png", 64, 64},
		{"images/ammo.png", 64, 64},
		{"images/armor.png", 64, 64},
		{"images/antidote.png", 64, 64},
	}

	if len(expectedAssets) != 27 {
		t.Fatalf("expected 27 assets to be tested, found %d", len(expectedAssets))
	}

	for _, tc := range expectedAssets {
		t.Run(tc.path, func(t *testing.T) {
			data, err := imageFS.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("failed to read embedded file %s: %v", tc.path, err)
			}

			if len(data) == 0 {
				t.Fatalf("embedded file %s is empty", tc.path)
			}

			img, format, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode image %s: %v", tc.path, err)
			}

			if format != "png" {
				t.Errorf("image %s format = %s, want png", tc.path, format)
			}

			bounds := img.Bounds()
			if bounds.Dx() != tc.width || bounds.Dy() != tc.height {
				t.Errorf("image %s dimensions = %dx%d, want %dx%d",
					tc.path, bounds.Dx(), bounds.Dy(), tc.width, tc.height)
			}

			nonTransparentCount := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						nonTransparentCount++
					}
				}
			}

			if nonTransparentCount == 0 {
				t.Errorf("image %s has no non-transparent pixels", tc.path)
			}
		})
	}
}

func TestAssetsLoadAllPointersNonNil(t *testing.T) {
	Load()

	handles := []struct {
		name  string
		img   *ebiten.Image
		wantW int
		wantH int
	}{
		// Entities (64x128)
		{"PlayerImage", PlayerImage, 64, 128},
		{"ZombieImage", ZombieImage, 64, 128},
		{"RunnerImage", RunnerImage, 64, 128},

		// Floors (256x128)
		{"GrassImage", GrassImage, 256, 128},
		{"DirtImage", DirtImage, 256, 128},
		{"WoodImage", WoodImage, 256, 128},
		{"AsphaltImage", AsphaltImage, 256, 128},
		{"ConcreteImage", ConcreteImage, 256, 128},
		{"TileFloorImage", TileFloorImage, 256, 128},

		// Obstacles & Props (256x256)
		{"WallImage", WallImage, 256, 256},
		{"TreeImage", TreeImage, 256, 256},
		{"FenceImage", FenceImage, 256, 256},
		{"DebrisImage", DebrisImage, 256, 256},
		{"TentImage", TentImage, 256, 256},
		{"StumpImage", StumpImage, 256, 256},
		{"MushroomImage", MushroomImage, 256, 256},
		{"SignImage", SignImage, 256, 256},
		{"ElevationBlockImage", ElevationBlockImage, 256, 256},
		{"ElevationRampImage", ElevationRampImage, 256, 256},

		// Items & Equipment (64x64)
		{"FoodImage", FoodImage, 64, 64},
		{"WaterImage", WaterImage, 64, 64},
		{"WeaponImage", WeaponImage, 64, 64},
		{"AxeImage", AxeImage, 64, 64},
		{"ShotgunImage", ShotgunImage, 64, 64},
		{"AmmoImage", AmmoImage, 64, 64},
		{"ArmorImage", ArmorImage, 64, 64},
		{"AntidoteImage", AntidoteImage, 64, 64},
	}

	if len(handles) != 27 {
		t.Fatalf("expected 27 asset pointers to be checked, found %d", len(handles))
	}

	for _, h := range handles {
		t.Run(h.name, func(t *testing.T) {
			if h.img == nil {
				t.Fatalf("asset pointer %s is nil after Load()", h.name)
			}
			bounds := h.img.Bounds()
			if bounds.Dx() != h.wantW || bounds.Dy() != h.wantH {
				t.Errorf("asset %s dimensions = %dx%d, want %dx%d",
					h.name, bounds.Dx(), bounds.Dy(), h.wantW, h.wantH)
			}
		})
	}
}
```

---

### 4.3 Target File 3: `internal/assets/challenger_stress_test.go`
Create stress and concurrency test suite for `internal/assets`.

```go
package assets

import (
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type assetDescriptor struct {
	name  string
	ptr   **ebiten.Image
	path  string
	wantW int
	wantH int
	cat   string
}

func getAssetDescriptors() []assetDescriptor {
	return []assetDescriptor{
		// Character Entities (3) - 64x128
		{"PlayerImage", &PlayerImage, "images/player.png", 64, 128, "Entity"},
		{"ZombieImage", &ZombieImage, "images/zombie.png", 64, 128, "Entity"},
		{"RunnerImage", &RunnerImage, "images/runner.png", 64, 128, "Entity"},

		// Floor Tiles (6) - 256x128
		{"GrassImage", &GrassImage, "images/grass.png", 256, 128, "Floor"},
		{"DirtImage", &DirtImage, "images/dirt.png", 256, 128, "Floor"},
		{"WoodImage", &WoodImage, "images/wood.png", 256, 128, "Floor"},
		{"AsphaltImage", &AsphaltImage, "images/asphalt.png", 256, 128, "Floor"},
		{"ConcreteImage", &ConcreteImage, "images/concrete.png", 256, 128, "Floor"},
		{"TileFloorImage", &TileFloorImage, "images/tile_floor.png", 256, 128, "Floor"},

		// Vertical Obstacles / Props (10) - 256x256
		{"WallImage", &WallImage, "images/wall.png", 256, 256, "Obstacle/Prop"},
		{"TreeImage", &TreeImage, "images/tree.png", 256, 256, "Obstacle/Prop"},
		{"FenceImage", &FenceImage, "images/fence.png", 256, 256, "Obstacle/Prop"},
		{"DebrisImage", &DebrisImage, "images/debris.png", 256, 256, "Obstacle/Prop"},
		{"TentImage", &TentImage, "images/tent.png", 256, 256, "Obstacle/Prop"},
		{"StumpImage", &StumpImage, "images/stump.png", 256, 256, "Obstacle/Prop"},
		{"MushroomImage", &MushroomImage, "images/mushroom.png", 256, 256, "Obstacle/Prop"},
		{"SignImage", &SignImage, "images/sign.png", 256, 256, "Obstacle/Prop"},
		{"ElevationBlockImage", &ElevationBlockImage, "images/elevation_block.png", 256, 256, "Obstacle/Prop"},
		{"ElevationRampImage", &ElevationRampImage, "images/elevation_ramp.png", 256, 256, "Obstacle/Prop"},

		// Item / Weapon / Equipment (8) - 64x64
		{"FoodImage", &FoodImage, "images/food.png", 64, 64, "Item"},
		{"WaterImage", &WaterImage, "images/water.png", 64, 64, "Item"},
		{"WeaponImage", &WeaponImage, "images/weapon.png", 64, 64, "Item"},
		{"AxeImage", &AxeImage, "images/axe.png", 64, 64, "Item"},
		{"ShotgunImage", &ShotgunImage, "images/shotgun.png", 64, 64, "Item"},
		{"AmmoImage", &AmmoImage, "images/ammo.png", 64, 64, "Item"},
		{"ArmorImage", &ArmorImage, "images/armor.png", 64, 64, "Item"},
		{"AntidoteImage", &AntidoteImage, "images/antidote.png", 64, 64, "Item"},
	}
}

func TestChallenger_All27ExportedPointersAndExactBounds(t *testing.T) {
	Load()

	descriptors := getAssetDescriptors()
	if len(descriptors) != 27 {
		t.Fatalf("expected exactly 27 exported image descriptors, got %d", len(descriptors))
	}

	for _, d := range descriptors {
		t.Run(d.name, func(t *testing.T) {
			img := *d.ptr
			if img == nil {
				t.Fatalf("Pointer %s is nil after Load()", d.name)
			}

			bounds := img.Bounds()
			if bounds.Min.X != 0 || bounds.Min.Y != 0 {
				t.Errorf("Pointer %s Bounds().Min = (%d, %d), want (0, 0)", d.name, bounds.Min.X, bounds.Min.Y)
			}
			if bounds.Dx() != d.wantW || bounds.Dy() != d.wantH {
				t.Errorf("Pointer %s Bounds() = %dx%d, want %dx%d (%s category)",
					d.name, bounds.Dx(), bounds.Dy(), d.wantW, d.wantH, d.cat)
			}
		})
	}
}

func TestChallenger_MultiThreadedLoadAndPointerRace(t *testing.T) {
	const numLoaders = 20
	const numReaders = 30
	const iterations = 50

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for i := 0; i < numLoaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal
			for j := 0; j < iterations; j++ {
				Load()
			}
		}()
	}

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			<-startSignal
			for j := 0; j < iterations; j++ {
				descriptors := getAssetDescriptors()
				d := descriptors[readerID%len(descriptors)]
				img := *d.ptr
				if img != nil {
					b := img.Bounds()
					if b.Dx() != d.wantW || b.Dy() != d.wantH {
						t.Errorf("Reader %d detected bounds mismatch: %dx%d want %dx%d",
							readerID, b.Dx(), b.Dy(), d.wantW, d.wantH)
					}
				}
			}
		}(i)
	}

	close(startSignal)
	wg.Wait()
}
```

---

### 4.4 Target File 4: `internal/game/draw_depth_test.go`
Create unit test suite validating DrawSystem anchor calculations, depth-sorting, and two-pass rendering for props and obstacles.

```go
package game

import (
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

func TestDrawSystem_SpriteGeometricAnchors(t *testing.T) {
	assets.Load()
	tests := []struct {
		name       string
		img        *ebiten.Image
		wantTransX float64
		wantTransY float64
	}{
		// Legacy Obstacles (256x256)
		{"Wall", assets.WallImage, -128.0, -128.0},
		{"Tree", assets.TreeImage, -128.0, -128.0},
		{"Fence", assets.FenceImage, -128.0, -128.0},
		{"Debris", assets.DebrisImage, -128.0, -128.0},
		{"Tent", assets.TentImage, -128.0, -128.0},
		{"Stump", assets.StumpImage, -128.0, -128.0},
		{"Mushroom", assets.MushroomImage, -128.0, -128.0},
		{"Sign", assets.SignImage, -128.0, -128.0},
		{"ElevationBlock", assets.ElevationBlockImage, -128.0, -128.0},
		{"ElevationRamp", assets.ElevationRampImage, -128.0, -128.0},

		// New Environmental Props (Variable Dimensions)
		{"Bench", assets.BenchImage, -26.0, 91.0},        // 52x37: -52/2 = -26, 128 - 37 = 91
		{"Chest", assets.ChestImage, -11.0, 107.0},      // 22x21: -22/2 = -11, 128 - 21 = 107
		{"Sculpture", assets.SculptureImage, -11.5, 97.0}, // 23x31: -23/2 = -11.5, 128 - 31 = 97
		{"Bush", assets.BushImage, -12.0, 110.0},         // 24x18: -24/2 = -12, 128 - 18 = 110
		{"Flower", assets.FlowerImage, -13.0, 103.0},     // 26x25: -26/2 = -13, 128 - 25 = 103
		{"Stone", assets.StoneImage, -14.0, 109.0},       // 28x19: -28/2 = -14, 128 - 19 = 109
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.img == nil {
				t.Fatalf("image %s is nil", tt.name)
			}
			b := tt.img.Bounds()
			imgW := float64(b.Dx())
			imgH := float64(b.Dy())

			transX := -imgW / 2.0
			transY := 128.0 - imgH

			if transX != tt.wantTransX {
				t.Errorf("%s transX = %f, want %f", tt.name, transX, tt.wantTransX)
			}
			if transY != tt.wantTransY {
				t.Errorf("%s transY = %f, want %f", tt.name, transY, tt.wantTransY)
			}
		})
	}
}

func TestDrawSystem_NewPropTilesLoadedAndDrawn(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(20, 20)

	props := []world.TileType{
		world.TileBench,
		world.TileChest,
		world.TileSculpture,
		world.TileBush,
		world.TileFlower,
		world.TileStone,
	}

	for idx, p := range props {
		m.SetTile(idx+1, idx+1, p)
		m.Visible[(idx+1)*20+(idx+1)] = true
		m.Explored[(idx+1)*20+(idx+1)] = true
	}

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(800, 600)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DrawSystem panicked while drawing prop tiles: %v", r)
		}
	}()

	drawSys.Draw(screen, 12.0)
}

func TestDrawSystem_GroundPassUnderNewProps(t *testing.T) {
	assets.Load()
	if assets.GrassImage == nil {
		t.Fatal("GrassImage is nil")
	}

	b := assets.GrassImage.Bounds()
	if b.Dx() != 256 || b.Dy() != 128 {
		t.Errorf("GrassImage dimensions = %dx%d, want 256x128", b.Dx(), b.Dy())
	}
}

func TestDrawSystem_DepthSortingOrdering(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(10, 10)
	drawSys := NewDrawSystem(w, m)

	if drawSys == nil {
		t.Fatal("NewDrawSystem returned nil")
	}

	// Verify WorldToIso isometric transform monotonicity
	// When worldX + worldY increases, depth increases
	isoX1, isoY1 := WorldToIso(100.0, 100.0)
	isoX2, isoY2 := WorldToIso(200.0, 200.0)

	if (isoX1 + isoY1) >= (isoX2 + isoY2) {
		t.Errorf("Expected depth ordering monotonicity, got %f >= %f", isoX1+isoY1, isoX2+isoY2)
	}
}
```

---

## 5. Verification Plan

The implementer must execute and confirm the following:

```bash
# 1. Test canonical test suite across all packages
CC=gcc go test ./...

# 2. Test assets package specifically
CC=gcc go test -v ./internal/assets -run "TestAssetsLoadAllPointersNonNil|TestEmbeddedAssetDimensionsAndValidity|TestChallenger"

# 3. Test game draw and depth-sorting package specifically
CC=gcc go test -v ./internal/game -run "TestDrawSystem_SpriteGeometricAnchors|TestDrawSystem_NewPropTilesLoadedAndDrawn"

# 4. Stress and race detection verification
CC=gcc go test -race -count=1 ./...

# 5. Build game executable
CC=gcc go build ./cmd/game
```
