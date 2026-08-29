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
		// Entities (64x128)
		PlayerImage = loadEbitenImage("images/player.png")
		ZombieImage = loadEbitenImage("images/zombie.png")
		RunnerImage = loadEbitenImage("images/runner.png")

		// Floor Tiles (256x128)
		GrassImage = loadEbitenImage("images/grass.png")
		DirtImage = loadEbitenImage("images/dirt.png")
		WoodImage = loadEbitenImage("images/wood.png")
		AsphaltImage = loadEbitenImage("images/asphalt.png")
		ConcreteImage = loadEbitenImage("images/concrete.png")
		TileFloorImage = loadEbitenImage("images/tile_floor.png")

		// Vertical Obstacles & Props (256x256)
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

		// Items / Weapons / Armor (64x64)
		WeaponImage = loadEbitenImage("images/weapon.png")
		AxeImage = loadEbitenImage("images/axe.png")
		ShotgunImage = loadEbitenImage("images/shotgun.png")
		AmmoImage = loadEbitenImage("images/ammo.png")
		ArmorImage = loadEbitenImage("images/armor.png")
		AntidoteImage = loadEbitenImage("images/antidote.png")
		FoodImage = loadEbitenImage("images/food.png")
		WaterImage = loadEbitenImage("images/water.png")

		// External World Props & Foliage
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

		// External Tilesets
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
