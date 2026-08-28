package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
)

func main() {
	generateEntity("player.png", color.RGBA{0, 255, 0, 255})
	generateEntity("zombie.png", color.RGBA{255, 0, 0, 255})
	generateEntity("runner.png", color.RGBA{255, 140, 0, 255}) // Dark orange
	generateIsoFloor("grass.png", color.RGBA{34, 139, 34, 255}, true)
	generateIsoFloor("dirt.png", color.RGBA{139, 69, 19, 255}, true)
	generateIsoFloor("wood.png", color.RGBA{205, 133, 63, 255}, true)
	generateIsoWall("wall.png", color.RGBA{105, 105, 105, 255})
	generateIsoTree("tree.png")
	generateWeapon("weapon.png")
}

func generateEntity(name string, baseColor color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 32))
	for x := 0; x < 16; x++ {
		for y := 0; y < 32; y++ {
			img.Set(x, y, baseColor)
		}
	}
	black := color.RGBA{0, 0, 0, 255}
	img.Set(10, 4, black)
	img.Set(10, 5, black)
	img.Set(10, 10, black)
	img.Set(10, 11, black)

	saveImg(name, img)
}

func generateIsoFloor(name string, baseColor color.RGBA, addNoise bool) {
	// Isometric floor tile: 64x32
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Diamond equation
			dx := float64(x - w/2)
			dy := float64(y - h/2)
			if (math.Abs(dx)/(float64(w)/2) + math.Abs(dy)/(float64(h)/2)) <= 1.0 {
				c := baseColor
				if addNoise {
					variance := uint8(rand.Intn(20))
					if c.R > variance { c.R -= variance }
					if c.G > variance { c.G -= variance }
					if c.B > variance { c.B -= variance }
				}
				// Simple edge highlight/shadow
				if (math.Abs(dx)/(float64(w)/2) + math.Abs(dy)/(float64(h)/2)) > 0.9 {
					c.R = uint8(float64(c.R) * 0.8)
					c.G = uint8(float64(c.G) * 0.8)
					c.B = uint8(float64(c.B) * 0.8)
				}
				img.Set(x, y, c)
			}
		}
	}
	saveImg(name, img)
}

func generateIsoWall(name string, baseColor color.RGBA) {
	// Isometric wall block: 64x64
	// Base is at bottom (y from 32 to 64)
	// Top face is at top (y from 0 to 32)
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	
	topColor := baseColor
	leftColor := color.RGBA{uint8(float64(baseColor.R)*0.8), uint8(float64(baseColor.G)*0.8), uint8(float64(baseColor.B)*0.8), 255}
	rightColor := color.RGBA{uint8(float64(baseColor.R)*0.6), uint8(float64(baseColor.G)*0.6), uint8(float64(baseColor.B)*0.6), 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x - w/2)
			
			// Top face (center at 32, 16)
			dyTop := float64(y - 16)
			if (math.Abs(dx)/32.0 + math.Abs(dyTop)/16.0) <= 1.0 {
				img.Set(x, y, topColor)
				continue
			}

			// Left face
			if x < w/2 && y >= 16 && y < 16+32 {
				// Check if within bounds
				// Top edge is the top face's bottom-left edge. 
				// Bottom edge is y < 16 + 32 + (x / 2) -> wait, math is easier:
				// just check if it's below the top face and above the bottom face.
			}
			
			// Let's do a simpler raycast/heightmap approach for a block
			// For a given pixel (x,y), we check if it falls in the 3D bounds
			// A point (X, Y, Z) projects to x = X - Y, y = (X + Y)/2 - Z
		}
	}
	
	// Simpler wall drawing:
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x - w/2)
			dyTop := float64(y - 16)

			isTop := (math.Abs(dx)/32.0 + math.Abs(dyTop)/16.0) <= 1.0
			
			if isTop {
				img.Set(x, y, topColor)
			} else if y > 16 && x < w/2 && (float64(y) <= 16 + float64(x)/2.0 + 32.0) && (float64(y) >= 16 - float64(x)/2.0) {
				// Left face approx
				// Top edge of left face: y = 16 - x/2
				// Bottom edge of left face: y = 48 + x/2
				// Left edge of left face: x = 0
				// Right edge of left face: x = 32
				// Wait, x goes from 0 to 32.
				// Top edge goes from (0, 16) to (32, 32). So y = 16 + x/2.
				if float64(y) >= 16.0 + float64(x)/2.0 && float64(y) <= 48.0 + float64(x)/2.0 {
					img.Set(x, y, leftColor)
				}
			} else if y > 16 && x >= w/2 {
				// Right face approx
				// x goes from 32 to 64
				// Top edge goes from (32, 32) to (64, 16). So y = 32 - (x-32)/2 = 48 - x/2
				if float64(y) >= 48.0 - float64(x)/2.0 && float64(y) <= 80.0 - float64(x)/2.0 {
					img.Set(x, y, rightColor)
				}
			}
		}
	}

	saveImg(name, img)
}

func saveImg(name string, img *image.RGBA) {
	path := "internal/assets/images/" + name
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully generated", path)
}

func generateIsoTree(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	
	trunkColor := color.RGBA{101, 67, 33, 255} // Dark brown
	leafColor := color.RGBA{0, 60, 0, 255}     // Dark green fir

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Trunk (x: 28-36, y: 48-60)
			if x >= 28 && x <= 36 && y >= 48 && y <= 60 {
				img.Set(x, y, trunkColor)
			}
			
			// Pine tree canopy (triangle)
			// Base of triangle at y=50, tip at y=4
			if y >= 4 && y <= 50 {
				// Width of triangle increases as y increases
				// At y=4, width=0. At y=50, width=40 (x from 12 to 52)
				maxWidth := float64(y-4) / 46.0 * 20.0
				dx := float64(x - 32)
				if math.Abs(dx) <= maxWidth {
					c := leafColor
					variance := uint8(rand.Intn(20))
					if c.G > variance { c.G -= variance }
					img.Set(x, y, c)
				}
			}
		}
	}
	saveImg(name, img)
}

func generateWeapon(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	brown := color.RGBA{139, 69, 19, 255}
	for i := 2; i < 14; i++ {
		img.Set(i, 15-i, brown)
		img.Set(i+1, 15-i, brown)
		img.Set(i, 15-i+1, brown)
	}
	saveImg(name, img)
}
