package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed images/*
var imageFS embed.FS

var (
	// Entity Sprites (16x32)
	PlayerImage *ebiten.Image
	ZombieImage *ebiten.Image
	RunnerImage *ebiten.Image

	// Floor Tiles (64x32)
	GrassImage     *ebiten.Image
	DirtImage      *ebiten.Image
	WoodImage      *ebiten.Image
	AsphaltImage   *ebiten.Image
	ConcreteImage  *ebiten.Image
	TileFloorImage *ebiten.Image

	// Vertical Obstacles / Props (64x64)
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

	// Item / Weapon / Armor Sprites (16x16)
	WeaponImage  *ebiten.Image
	AxeImage     *ebiten.Image
	ShotgunImage *ebiten.Image
	AmmoImage    *ebiten.Image
	ArmorImage   *ebiten.Image
	FoodImage    *ebiten.Image
	WaterImage   *ebiten.Image
)

func Load() {
	// Entities
	PlayerImage = loadEbitenImage("images/player.png")
	ZombieImage = loadEbitenImage("images/zombie.png")
	RunnerImage = loadEbitenImage("images/runner.png")

	// Floor Tiles
	GrassImage = loadEbitenImage("images/grass.png")
	DirtImage = loadEbitenImage("images/dirt.png")
	WoodImage = loadEbitenImage("images/wood.png")
	AsphaltImage = loadEbitenImage("images/asphalt.png")
	ConcreteImage = loadEbitenImage("images/concrete.png")
	TileFloorImage = loadEbitenImage("images/tile_floor.png")

	// Vertical Obstacles
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

	// Items / Weapons / Armor
	WeaponImage = loadEbitenImage("images/weapon.png")
	AxeImage = loadEbitenImage("images/axe.png")
	ShotgunImage = loadEbitenImage("images/shotgun.png")
	AmmoImage = loadEbitenImage("images/ammo.png")
	ArmorImage = loadEbitenImage("images/armor.png")
	FoodImage = loadEbitenImage("images/food.png")
	WaterImage = loadEbitenImage("images/water.png")
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
