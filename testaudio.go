package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()
	fmt.Println("Creating context...")
	_ = audio.NewContext(44100)
	fmt.Println("Done")
}
