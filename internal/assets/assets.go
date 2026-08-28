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
	PlayerImage *ebiten.Image
	ZombieImage *ebiten.Image
	RunnerImage *ebiten.Image
	GrassImage  *ebiten.Image
	WallImage   *ebiten.Image
	DirtImage   *ebiten.Image
	WoodImage   *ebiten.Image
	TreeImage   *ebiten.Image
	WeaponImage *ebiten.Image
	FoodImage   *ebiten.Image
	WaterImage  *ebiten.Image
)

func Load() {
	PlayerImage = loadEbitenImage("images/player.png")
	ZombieImage = loadEbitenImage("images/zombie.png")
	RunnerImage = loadEbitenImage("images/runner.png")
	GrassImage = loadEbitenImage("images/grass.png")
	WallImage = loadEbitenImage("images/wall.png")
	DirtImage = loadEbitenImage("images/dirt.png")
	WoodImage = loadEbitenImage("images/wood.png")
	TreeImage = loadEbitenImage("images/tree.png")
	WeaponImage = loadEbitenImage("images/weapon.png")
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
