package main

import (
	"log"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Go Zomboid")

	assets.Load()

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
