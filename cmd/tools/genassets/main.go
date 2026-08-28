package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

const outDir = "internal/assets/images"

func main() {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	// 1. Character Entities (64x128)
	generatePlayer("player.png")
	generateZombie("zombie.png")
	generateRunner("runner.png")

	// 2. Floor Tiles (256x128)
	generateGrass("grass.png")
	generateDirt("dirt.png")
	generateWoodFloor("wood.png")
	generateAsphalt("asphalt.png")
	generateConcrete("concrete.png")
	generateTileFloor("tile_floor.png")

	// 3. Vertical Obstacles & Props (256x256)
	generateIsoWall("wall.png")
	generateIsoTree("tree.png")
	generateIsoFence("fence.png")
	generateIsoDebris("debris.png")
	generateIsoTent("tent.png")
	generateIsoStump("stump.png")
	generateIsoMushroom("mushroom.png")
	generateIsoSign("sign.png")
	generateElevationBlock("elevation_block.png")
	generateElevationRamp("elevation_ramp.png")

	// 4. Items & Equipment (64x64)
	generateFood("food.png")
	generateWater("water.png")
	generateWeapon("weapon.png")
	generateAxe("axe.png")
	generateShotgun("shotgun.png")
	generateAmmo("ammo.png")
	generateArmor("armor.png")
	generateAntidote("antidote.png")

	log.Println("Asset generation completed successfully.")
}

// -------------------------------------------------------------
// DRAWING HELPERS & VECTOR PRIMITIVES
// -------------------------------------------------------------

// Bounds-checked pixel setter
func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	bounds := img.Bounds()
	if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
		img.SetRGBA(x, y, c)
	}
}

// Alpha-blended pixel setter (Porter-Duff Over)
func blendPixel(img *image.RGBA, x, y int, c color.RGBA) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	if c.A == 255 {
		img.SetRGBA(x, y, c)
		return
	}
	if c.A == 0 {
		return
	}
	dst := img.RGBAAt(x, y)
	if dst.A == 0 {
		img.SetRGBA(x, y, c)
		return
	}
	srcA := float64(c.A) / 255.0
	dstA := float64(dst.A) / 255.0
	outA := srcA + dstA*(1.0-srcA)
	if outA <= 0 {
		return
	}
	outR := (float64(c.R)*srcA + float64(dst.R)*dstA*(1.0-srcA)) / outA
	outG := (float64(c.G)*srcA + float64(dst.G)*dstA*(1.0-srcA)) / outA
	outB := (float64(c.B)*srcA + float64(dst.B)*dstA*(1.0-srcA)) / outA
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(math.Round(outR)),
		G: uint8(math.Round(outG)),
		B: uint8(math.Round(outB)),
		A: uint8(math.Round(outA * 255.0)),
	})
}

// Fill solid rectangle
func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			setPixel(img, x+dx, y+dy, c)
		}
	}
}

// Horizontal line
func drawHLine(img *image.RGBA, x1, x2, y int, c color.RGBA) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		setPixel(img, x, y, c)
	}
}

// Vertical line
func drawVLine(img *image.RGBA, x, y1, y2 int, c color.RGBA) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		setPixel(img, x, y, c)
	}
}

// Shaded rectangle with top-left highlight and bottom-right shadow
func drawShadedRect(img *image.RGBA, x, y, w, h int, base, highlight, shadow color.RGBA) {
	fillRect(img, x, y, w, h, base)
	drawHLine(img, x, x+w-1, y, highlight)
	drawVLine(img, x, y, y+h-1, highlight)
	drawHLine(img, x, x+w-1, y+h-1, shadow)
	drawVLine(img, x+w-1, y, y+h-1, shadow)
}

// Darken color by factor (e.g. 0.75 for 25% darker)
func darken(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Max(0, math.Min(255, float64(c.R)*factor))),
		G: uint8(math.Max(0, math.Min(255, float64(c.G)*factor))),
		B: uint8(math.Max(0, math.Min(255, float64(c.B)*factor))),
		A: c.A,
	}
}

// Lighten color by factor (e.g. 1.25 for 25% brighter)
func lighten(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Max(0, math.Min(255, float64(c.R)*factor))),
		G: uint8(math.Max(0, math.Min(255, float64(c.G)*factor))),
		B: uint8(math.Max(0, math.Min(255, float64(c.B)*factor))),
		A: c.A,
	}
}

// Color interpolation
func blend(c1, c2 color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	return color.RGBA{
		R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
		G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
		B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
		A: uint8(float64(c1.A)*(1-t) + float64(c2.A)*t),
	}
}

// Draw filled circle with bounds checking
func drawFilledCircle(img *image.RGBA, cx, cy int, r float64, c color.RGBA) {
	r2 := r * r
	minX := int(math.Floor(float64(cx) - r))
	maxX := int(math.Ceil(float64(cx) + r))
	minY := int(math.Floor(float64(cy) - r))
	maxY := int(math.Ceil(float64(cy) + r))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			if dx*dx+dy*dy <= r2 {
				setPixel(img, x, y, c)
			}
		}
	}
}

// Anti-aliased filled ellipse
func drawAAEllipse(img *image.RGBA, cx, cy, rx, ry float64, c color.RGBA) {
	minX := int(math.Floor(cx - rx - 1))
	maxX := int(math.Ceil(cx + rx + 1))
	minY := int(math.Floor(cy - ry - 1))
	maxY := int(math.Ceil(cy + ry + 1))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(y) + 0.5 - cy) / ry
			d := dx*dx + dy*dy
			if d <= 1.0 {
				edgeDist := 1.0 - math.Sqrt(d)
				alphaFactor := math.Min(1.0, edgeDist*math.Min(rx, ry))
				if alphaFactor > 0.05 {
					col := c
					col.A = uint8(float64(c.A) * alphaFactor)
					blendPixel(img, x, y, col)
				}
			}
		}
	}
}

// Draw vector grass blade cluster (3 blades + root)
func drawVectorChevron(img *image.RGBA, cx, cy int, col color.RGBA) {
	// Center root
	fillRect(img, cx-1, cy, 3, 2, col)
	// Center blade (vertical)
	for i := 0; i < 8; i++ {
		setPixel(img, cx, cy-i, col)
		setPixel(img, cx+1, cy-i, col)
	}
	// Left blade (diagonal up-left)
	for i := 1; i <= 7; i++ {
		setPixel(img, cx-i, cy-i, col)
		setPixel(img, cx-i, cy-i+1, col)
	}
	// Right blade (diagonal up-right)
	for i := 1; i <= 7; i++ {
		setPixel(img, cx+i+1, cy-i, col)
		setPixel(img, cx+i+1, cy-i+1, col)
	}
}

// Draw 5-petal wildflower with yellow center
func drawVectorFlower(img *image.RGBA, cx, cy int, petalColor, centerColor color.RGBA) {
	for k := 0; k < 5; k++ {
		angle := float64(k)*(2.0*math.Pi/5.0) - math.Pi/2.0
		px := int(math.Round(float64(cx) + 4.5*math.Cos(angle)))
		py := int(math.Round(float64(cy) + 4.5*math.Sin(angle)))
		drawFilledCircle(img, px, py, 2.5, petalColor)
	}
	drawFilledCircle(img, cx, cy, 2.5, centerColor)
}

// Draw rounded vector pebble with highlight and drop shadow
func drawVectorPebble(img *image.RGBA, cx, cy int, rx, ry float64, base, light, shadow color.RGBA) {
	dropShadow := color.RGBA{0, 0, 0, 45}
	// Drop shadow
	minY := int(math.Floor(float64(cy+2) - ry))
	maxY := int(math.Ceil(float64(cy+2) + ry))
	minX := int(math.Floor(float64(cx+2) - rx))
	maxX := int(math.Ceil(float64(cx+2) + rx))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x-(cx+2)) / rx
			dy := float64(y-(cy+2)) / ry
			if dx*dx+dy*dy <= 1.0 {
				isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0
				if isoDist <= 1.0 {
					blendPixel(img, x, y, dropShadow)
				}
			}
		}
	}

	// Pebble body
	minY = int(math.Floor(float64(cy) - ry))
	maxY = int(math.Ceil(float64(cy) + ry))
	minX = int(math.Floor(float64(cx) - rx))
	maxX = int(math.Ceil(float64(cx) + rx))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			normDist := (dx*dx)/(rx*rx) + (dy*dy)/(ry*ry)
			if normDist <= 1.0 {
				isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0
				if isoDist <= 1.0 {
					c := base
					if dx+dy < -2.0 {
						c = light
					} else if dx+dy > 2.5 {
						c = shadow
					}
					setPixel(img, x, y, c)
				}
			}
		}
	}
}

// Add 1px dark contour around entity silhouette
func addSelectiveOutline(img *image.RGBA, outlineColor color.RGBA) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	temp := image.NewRGBA(bounds)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			temp.SetRGBA(x, y, img.RGBAAt(x, y))
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if temp.RGBAAt(x, y).A == 0 {
				hasSolidNeighbor := false
				if (x > 0 && temp.RGBAAt(x-1, y).A > 0) ||
					(x < w-1 && temp.RGBAAt(x+1, y).A > 0) ||
					(y > 0 && temp.RGBAAt(x, y-1).A > 0) ||
					(y < h-1 && temp.RGBAAt(x, y+1).A > 0) {
					hasSolidNeighbor = true
				}
				if hasSolidNeighbor {
					img.SetRGBA(x, y, outlineColor)
				}
			}
		}
	}
}

// Save RGBA image to internal/assets/images/<name>
func saveImg(name string, img *image.RGBA) {
	path := outDir + "/" + name
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("failed to create image file %s: %v", path, err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatalf("failed to encode PNG %s: %v", path, err)
	}
	log.Println("Generated", path)
}

// -------------------------------------------------------------
// 1. CHARACTER ENTITIES (64x128)
// -------------------------------------------------------------

func generatePlayer(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 128))

	skin := color.RGBA{255, 204, 153, 255}
	skinShadow := color.RGBA{230, 175, 125, 255}
	hair := color.RGBA{92, 58, 34, 255}
	hairHi := color.RGBA{135, 88, 54, 255}
	shirt := color.RGBA{70, 172, 230, 255}
	shirtHi := color.RGBA{115, 205, 255, 255}
	shirtSh := color.RGBA{42, 130, 185, 255}
	pants := color.RGBA{75, 95, 145, 255}
	pantsHi := color.RGBA{98, 122, 175, 255}
	pantsSh := color.RGBA{48, 62, 102, 255}
	belt := color.RGBA{52, 40, 32, 255}
	boots := color.RGBA{42, 34, 28, 255}
	shadow := color.RGBA{0, 0, 0, 60}

	// 1. Ground drop shadow
	drawAAEllipse(img, 32, 122, 24, 6, shadow)

	// 2. Boots (rows 116..124)
	fillRect(img, 18, 116, 11, 8, boots)
	fillRect(img, 36, 116, 11, 8, boots)
	drawHLine(img, 18, 28, 123, darken(boots, 0.6))
	drawHLine(img, 36, 46, 123, darken(boots, 0.6))

	// 3. Pants
	fillRect(img, 20, 80, 24, 36, pants)
	fillRect(img, 20, 80, 11, 36, pantsHi)
	fillRect(img, 33, 80, 11, 36, pantsSh)
	// Inseam split
	fillRect(img, 31, 92, 2, 24, pantsSh)
	// Belt
	fillRect(img, 20, 80, 24, 4, belt)
	setPixel(img, 31, 81, color.RGBA{220, 220, 220, 255})
	setPixel(img, 32, 81, color.RGBA{220, 220, 220, 255})

	// 4. Shirt (Torso)
	fillRect(img, 18, 48, 28, 32, shirt)
	fillRect(img, 18, 48, 14, 32, shirtHi)
	fillRect(img, 32, 48, 14, 32, shirtSh)
	// V-Neck collar
	for y := 48; y <= 56; y++ {
		dy := y - 48
		drawHLine(img, 32-dy/2, 32+dy/2, y, skin)
	}

	// 5. Sleeves & Arms
	fillRect(img, 10, 48, 8, 20, shirtHi)
	fillRect(img, 10, 68, 8, 16, skin)
	fillRect(img, 46, 48, 8, 20, shirtSh)
	fillRect(img, 46, 68, 8, 16, skinShadow)

	// 6. Head
	for y := 12; y <= 48; y++ {
		for x := 14; x <= 50; x++ {
			dx := float64(x - 32)
			dy := float64(y - 30)
			if dx*dx+dy*dy <= 324 {
				c := skin
				if dx > 4 {
					c = skinShadow
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 7. Hair
	for y := 12; y <= 26; y++ {
		for x := 14; x <= 50; x++ {
			dx := float64(x - 32)
			dy := float64(y - 24)
			if dx*dx+dy*dy*2.0 <= 260 {
				c := hair
				if y < 20 && dx < 0 {
					c = hairHi
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 8. Eyes & Face Details
	fillRect(img, 24, 27, 4, 4, color.RGBA{255, 255, 255, 255})
	fillRect(img, 25, 28, 3, 3, color.RGBA{20, 20, 25, 255})
	setPixel(img, 25, 28, color.RGBA{255, 255, 255, 255})

	fillRect(img, 37, 27, 4, 4, color.RGBA{255, 255, 255, 255})
	fillRect(img, 37, 28, 3, 3, color.RGBA{20, 20, 25, 255})
	setPixel(img, 37, 28, color.RGBA{255, 255, 255, 255})

	drawHLine(img, 23, 28, 25, hair)
	drawHLine(img, 36, 41, 25, hair)
	drawHLine(img, 30, 34, 38, color.RGBA{180, 110, 85, 255})

	saveImg(name, img)
}

func generateZombie(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 128))

	skin := color.RGBA{145, 195, 145, 255}
	skinSh := color.RGBA{105, 155, 105, 255}
	shirt := color.RGBA{145, 95, 95, 255}
	shirtSh := color.RGBA{105, 65, 65, 255}
	pants := color.RGBA{95, 95, 95, 255}
	shadow := color.RGBA{0, 0, 0, 60}

	// 1. Ground drop shadow
	drawAAEllipse(img, 32, 122, 24, 6, shadow)

	// 2. Ragged feet (rows 116..124)
	fillRect(img, 18, 116, 11, 8, skinSh)
	fillRect(img, 36, 116, 11, 8, skinSh)

	// 3. Tattered Pants
	fillRect(img, 20, 80, 24, 34, pants)
	for x := 20; x <= 44; x++ {
		if x%4 == 0 {
			setPixel(img, x, 114, skinSh)
			setPixel(img, x, 115, skinSh)
		}
	}

	// 4. Decayed Shirt
	fillRect(img, 18, 48, 28, 32, shirt)
	fillRect(img, 32, 48, 14, 32, shirtSh)
	for y := 56; y <= 66; y++ {
		drawHLine(img, 26, 34, y, skinSh)
	}

	// 5. Reaching Outstretched Arms
	fillRect(img, 4, 54, 16, 8, skin)
	fillRect(img, 44, 54, 16, 8, skinSh)
	fillRect(img, 4, 52, 4, 4, skinSh)
	fillRect(img, 56, 52, 4, 4, skinSh)

	// 6. Head
	for y := 12; y <= 48; y++ {
		for x := 14; x <= 50; x++ {
			dx := float64(x - 32)
			dy := float64(y - 30)
			if dx*dx+dy*dy <= 324 {
				c := skin
				if dx > 4 {
					c = skinSh
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 7. Glowing Blood-Red Eyes
	fillRect(img, 24, 27, 4, 4, color.RGBA{255, 40, 40, 255})
	setPixel(img, 25, 28, color.RGBA{255, 180, 80, 255})
	fillRect(img, 37, 27, 4, 4, color.RGBA{255, 40, 40, 255})
	setPixel(img, 38, 28, color.RGBA{255, 180, 80, 255})

	// Snarl
	fillRect(img, 28, 38, 8, 4, color.RGBA{30, 20, 20, 255})
	setPixel(img, 30, 38, color.RGBA{230, 230, 210, 255})
	setPixel(img, 33, 38, color.RGBA{230, 230, 210, 255})

	saveImg(name, img)
}

func generateRunner(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 128))

	skin := color.RGBA{235, 55, 55, 255}
	skinHi := color.RGBA{255, 120, 120, 255}
	skinSh := color.RGBA{160, 25, 25, 255}
	shadow := color.RGBA{0, 0, 0, 60}

	// 1. Ground drop shadow
	drawAAEllipse(img, 32, 122, 26, 6, shadow)

	// 2. Sprinting Legs (rows 118..124 grounded)
	fillRect(img, 14, 84, 12, 38, skinSh)
	fillRect(img, 38, 80, 14, 42, skin)
	fillRect(img, 12, 118, 6, 4, skinSh)
	fillRect(img, 46, 118, 8, 4, skin)

	// 3. Forward Leaning Torso
	for y := 48; y <= 92; y++ {
		for x := 12; x <= 52; x++ {
			dx := float64(x - 32)
			dy := float64(y - 70)
			if (dx*dx)/400.0+(dy*dy)/784.0 <= 1.0 {
				c := skin
				if dx < -4 {
					c = skinHi
				} else if dx > 4 {
					c = skinSh
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 4. Predatory Arms
	fillRect(img, 6, 64, 16, 10, skinHi)
	fillRect(img, 44, 64, 16, 10, skinSh)

	// 5. Head
	for y := 18; y <= 54; y++ {
		for x := 16; x <= 48; x++ {
			dx := float64(x - 32)
			dy := float64(y - 36)
			if dx*dx+dy*dy <= 256 {
				c := skin
				if dx < -3 {
					c = skinHi
				} else if dx > 3 {
					c = skinSh
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 6. Glowing Yellow Eyes
	fillRect(img, 26, 33, 4, 4, color.RGBA{255, 240, 50, 255})
	setPixel(img, 27, 34, color.RGBA{255, 255, 210, 255})
	fillRect(img, 37, 33, 4, 4, color.RGBA{255, 240, 50, 255})
	setPixel(img, 38, 34, color.RGBA{255, 255, 210, 255})

	// Fang Maw
	fillRect(img, 28, 44, 10, 5, color.RGBA{40, 10, 10, 255})
	setPixel(img, 30, 44, color.RGBA{255, 255, 255, 255})
	setPixel(img, 34, 44, color.RGBA{255, 255, 255, 255})

	saveImg(name, img)
}

// -------------------------------------------------------------
// 2. FLOOR TILES (256x128)
// -------------------------------------------------------------

func generateGrass(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	topColor := color.RGBA{113, 225, 157, 255}
	leftColor := color.RGBA{74, 184, 131, 255}
	rightColor := color.RGBA{54, 150, 103, 255}
	flowerWhite := color.RGBA{255, 255, 255, 255}
	flowerYellow := color.RGBA{255, 220, 100, 255}
	chevronColor := color.RGBA{154, 238, 181, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				c := topColor

				rimThickness := 0.08 + 0.03*math.Sin(float64(x)*0.08)
				if isoDist > 1.0-rimThickness {
					if x < 128 {
						c = leftColor
					} else {
						c = rightColor
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// 4x Scaled vector grass blades
	chevrons := [][2]int{
		{64, 48}, {160, 32}, {96, 80}, {192, 64},
		{128, 56}, {48, 68}, {176, 92}, {140, 24},
	}
	for _, pos := range chevrons {
		drawVectorChevron(img, pos[0], pos[1], chevronColor)
	}

	// 4x Scaled wildflower clusters
	flowers := [][2]int{
		{96, 32}, {160, 80}, {52, 70}, {180, 44}, {120, 96},
	}
	for _, pos := range flowers {
		drawVectorFlower(img, pos[0], pos[1], flowerWhite, flowerYellow)
	}

	saveImg(name, img)
}

func generateDirt(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	topColor := color.RGBA{151, 103, 81, 255}
	leftColor := color.RGBA{122, 79, 59, 255}
	rightColor := color.RGBA{94, 60, 44, 255}
	pebbleBase := color.RGBA{175, 140, 120, 255}
	pebbleLight := color.RGBA{215, 185, 165, 255}
	pebbleShadow := color.RGBA{110, 75, 60, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				c := topColor

				rimThickness := 0.08 + 0.03*math.Sin(float64(x)*0.08)
				if isoDist > 1.0-rimThickness {
					if x < 128 {
						c = leftColor
					} else {
						c = rightColor
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// 4x Scaled rounded vector pebbles (~14x8px)
	pebbles := [][2]int{
		{80, 40}, {180, 56}, {120, 88}, {60, 80}, {185, 42}, {145, 30},
	}
	for _, pos := range pebbles {
		drawVectorPebble(img, pos[0], pos[1], 7.0, 4.0, pebbleBase, pebbleLight, pebbleShadow)
	}

	saveImg(name, img)
}

func generateWoodFloor(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	plankColors := []color.RGBA{
		{142, 92, 54, 255},
		{128, 80, 46, 255},
		{156, 104, 62, 255},
		{136, 88, 50, 255},
	}
	seamDark := color.RGBA{45, 26, 14, 255}
	nailColor := color.RGBA{30, 22, 18, 255}
	nailHighlight := color.RGBA{100, 80, 70, 255}
	endJoints := []float64{0.60, 0.30, 0.75, 0.45}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				laneFloat := v * 4.0
				lane := int(math.Floor(laneFloat))
				if lane < 0 {
					lane = 0
				}
				if lane > 3 {
					lane = 3
				}
				vInLane := laneFloat - float64(lane)

				c := plankColors[lane%len(plankColors)]

				// 3px Seams along longitudinal lanes
				if vInLane < 0.04 || vInLane > 0.96 {
					c = seamDark
				}

				// Transverse end joints
				endU := endJoints[lane]
				if math.Abs(u-endU) < 0.012 {
					c = seamDark
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = darken(c, 0.85)
					} else {
						c = darken(c, 0.70)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// 4x Vector nailheads
	for lane := 0; lane < 4; lane++ {
		endU := endJoints[lane]
		for _, offset := range []float64{-0.035, 0.035} {
			nu := endU + offset
			nv := (float64(lane) + 0.5) / 4.0
			nx := int(math.Round(127.5 + (nu-nv)*128.0))
			ny := int(math.Round(63.5 + (nu+nv-1.0)*64.0))
			if nx >= 0 && nx < w && ny >= 0 && ny < h {
				dx := float64(nx) - 127.5
				dy := float64(ny) - 63.5
				if math.Abs(dx)/128.0+math.Abs(dy)/64.0 <= 0.88 {
					drawFilledCircle(img, nx, ny, 2.5, nailColor)
					setPixel(img, nx-1, ny-1, nailHighlight)
				}
			}
		}
	}

	saveImg(name, img)
}

func generateAsphalt(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	baseDark := color.RGBA{38, 40, 44, 255}
	baseMid := color.RGBA{48, 50, 55, 255}
	yellowMarking := color.RGBA{240, 195, 45, 255}
	yellowShadow := color.RGBA{180, 140, 30, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				c := baseMid

				// Crisp highway yellow dashed stripe in UV
				if v >= 0.45 && v <= 0.55 {
					if (u >= 0.08 && u <= 0.40) || (u >= 0.60 && u <= 0.92) {
						if v >= 0.535 {
							c = yellowShadow
						} else {
							c = yellowMarking
						}
					}
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = baseDark
					} else {
						c = darken(baseDark, 0.75)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}

func generateConcrete(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	slabBase := color.RGBA{145, 145, 142, 255}
	slabLight := color.RGBA{168, 168, 165, 255}
	jointDark := color.RGBA{45, 45, 45, 255}
	jointBevelLight := color.RGBA{195, 195, 190, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				quadX := 0
				if u >= 0.5 {
					quadX = 1
				}
				quadY := 0
				if v >= 0.5 {
					quadY = 1
				}

				var c color.RGBA
				if (quadX+quadY)%2 == 0 {
					c = slabBase
				} else {
					c = slabLight
				}

				distU := math.Abs(u - 0.5)
				distV := math.Abs(v - 0.5)

				// Expansion joints (3px deep grooving)
				if distU < 0.010 || distV < 0.010 {
					c = jointDark
				} else if (distU < 0.020 && u > 0.5) || (distV < 0.020 && v > 0.5) {
					c = jointBevelLight
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = darken(slabBase, 0.8)
					} else {
						c = darken(slabBase, 0.6)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}

func generateTileFloor(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	tileABase := color.RGBA{210, 210, 205, 255}
	tileBBase := color.RGBA{65, 75, 85, 255}
	groutDark := color.RGBA{32, 34, 38, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				gridU := u * 4.0
				gridV := v * 4.0
				tileU := int(math.Floor(gridU))
				tileV := int(math.Floor(gridV))
				if tileU < 0 {
					tileU = 0
				}
				if tileU > 3 {
					tileU = 3
				}
				if tileV < 0 {
					tileV = 0
				}
				if tileV > 3 {
					tileV = 3
				}

				subU := gridU - float64(tileU)
				subV := gridV - float64(tileV)

				isTileA := (tileU+tileV)%2 == 0
				var baseCol color.RGBA
				if isTileA {
					baseCol = tileABase
				} else {
					baseCol = tileBBase
				}

				var c color.RGBA
				if subU < 0.045 || subV < 0.045 {
					c = groutDark
				} else if subU < 0.09 || subV < 0.09 {
					c = lighten(baseCol, 1.15)
				} else if subU > 0.94 || subV > 0.94 {
					c = darken(baseCol, 0.82)
				} else {
					c = baseCol
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = darken(c, 0.85)
					} else {
						c = darken(c, 0.70)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}

// -------------------------------------------------------------
// 3. VERTICAL OBSTACLES & PROPS (256x256)
// -------------------------------------------------------------

func generateIsoWall(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	copingTop := color.RGBA{228, 224, 218, 255}
	brickLeftBase := color.RGBA{154, 62, 48, 255}
	brickLeftMortar := color.RGBA{185, 95, 78, 255}
	brickRightBase := color.RGBA{118, 44, 32, 255}
	brickRightMortar := color.RGBA{88, 30, 20, 255}

	// 1. Top Coping Face (Diamond at center (128, 56), rx=128, ry=56)
	for y := 0; y < 112; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 55.5
			if math.Abs(dx)/128.0+math.Abs(dy)/56.0 <= 1.0 {
				c := copingTop
				if math.Abs(dx)/128.0+math.Abs(dy)/56.0 > 0.94 {
					if x < 128 {
						c = color.RGBA{245, 242, 238, 255}
					} else {
						c = color.RGBA{200, 195, 188, 255}
					}
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 2. Left Face (West Wall) for x in [0..127]
	for x := 0; x < 128; x++ {
		topY := 56 + x/2
		botY := 184 + x/2
		for y := topY; y <= botY && y < h; y++ {
			c := brickLeftBase
			relY := y - (56 + x/2)
			// Horizontal mortar joints every 16px
			if relY%16 == 0 {
				c = brickLeftMortar
			} else {
				// Vertical mortar joints staggered every 32px
				row := relY / 16
				offset := (row % 2) * 16
				if (x+offset)%32 == 0 {
					c = brickLeftMortar
				}
			}
			setPixel(img, x, y, c)
		}
	}

	// 3. Right Face (South Wall) for x in [128..255]
	for x := 128; x < 256; x++ {
		topY := 120 - (x-128)/2
		botY := 248 - (x-128)/2
		for y := topY; y <= botY && y < h; y++ {
			c := brickRightBase
			relY := y - (120 - (x-128)/2)
			// Horizontal mortar joints every 16px
			if relY%16 == 0 {
				c = brickRightMortar
			} else {
				// Vertical mortar joints staggered every 32px
				row := relY / 16
				offset := (row % 2) * 16
				if (x+offset)%32 == 0 {
					c = brickRightMortar
				}
			}
			setPixel(img, x, y, c)
		}
	}

	// Highlight ridge along top edge of wall faces
	for x := 0; x < 128; x++ {
		topY := 56 + x/2
		setPixel(img, x, topY, color.RGBA{255, 250, 245, 255})
		setPixel(img, x, topY+1, color.RGBA{255, 250, 245, 255})
	}
	for x := 128; x < 256; x++ {
		topY := 120 - (x-128)/2
		setPixel(img, x, topY, color.RGBA{210, 205, 200, 255})
		setPixel(img, x, topY+1, color.RGBA{210, 205, 200, 255})
	}

	saveImg(name, img)
}

func generateIsoTree(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	trunkHighlight := color.RGBA{135, 100, 75, 255}
	trunkBase := color.RGBA{101, 74, 57, 255}
	trunkShadow := color.RGBA{68, 48, 36, 255}
	leafHighlight := color.RGBA{110, 218, 158, 255}
	leafMid := color.RGBA{74, 184, 131, 255}
	leafShadow := color.RGBA{44, 142, 96, 255}
	leafDeepShadow := color.RGBA{28, 98, 64, 255}
	groundShadow := color.RGBA{0, 0, 0, 50}

	// 1. Ground shadow
	drawAAEllipse(img, 128, 220, 64, 20, groundShadow)

	// 2. Trunk Cylinder
	for y := 148; y <= 222; y++ {
		flare := 0
		if y > 214 {
			flare = (y - 214) * 2
		}
		for x := 112 - flare; x <= 143 + flare; x++ {
			c := trunkBase
			if x < 120-flare/2 {
				c = trunkHighlight
			} else if x > 135+flare/2 {
				c = trunkShadow
			}
			setPixel(img, x, y, c)
		}
	}

	// 3. Multi-tier Foliage Canopy
	canopySpheres := []struct {
		cx, cy float64
		r      float64
	}{
		{128, 100, 80},
		{128, 60, 56},
		{84, 108, 54},
		{172, 108, 54},
	}

	for y := 0; y < 200; y++ {
		for x := 0; x < w; x++ {
			minDistRatio := 2.0
			bestSphereIdx := -1
			var bestDX, bestDY float64

			for i, s := range canopySpheres {
				dx := float64(x) - s.cx
				dy := float64(y) - s.cy
				dist := math.Sqrt(dx*dx + dy*dy)
				ratio := dist / s.r
				if ratio < minDistRatio {
					minDistRatio = ratio
					bestSphereIdx = i
					bestDX = dx
					bestDY = dy
				}
			}

			if minDistRatio <= 1.0 && bestSphereIdx >= 0 {
				s := canopySpheres[bestSphereIdx]
				c := leafMid
				if bestDX < -0.15*s.r && bestDY < -0.15*s.r {
					c = leafHighlight
				} else if bestDX > 0.40*s.r && bestDY > 0.40*s.r {
					c = leafDeepShadow
				} else if bestDX > 0.20*s.r && bestDY > 0.20*s.r && minDistRatio > 0.45 {
					c = leafShadow
				}
				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}

func generateIsoFence(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	woodLight := color.RGBA{178, 158, 135, 255}
	woodMid := color.RGBA{142, 120, 98, 255}
	woodDark := color.RGBA{88, 72, 56, 255}
	nailColor := color.RGBA{35, 30, 25, 255}
	nailHighlight := color.RGBA{180, 180, 190, 255}

	// 1. Horizontal Rails (x in [8..128])
	for x := 8; x <= 128; x++ {
		tRailY := 112 + x/2
		bRailY := 160 + x/2
		for dy := 0; dy < 8; dy++ {
			setPixel(img, x, tRailY+dy, woodMid)
			if dy >= 5 {
				setPixel(img, x, tRailY+dy, woodDark)
			}
			setPixel(img, x, bRailY+dy, woodDark)
			if dy < 3 {
				setPixel(img, x, bRailY+dy, woodMid)
			}
		}
	}

	// 2. Vertical Pickets
	pickets := []int{20, 36, 52, 68, 84, 100, 116}
	for _, px := range pickets {
		baseY := 184 + px/2
		topY := baseY - 96

		// Pointed peak triangle (8px high)
		for dy := 0; dy < 8; dy++ {
			pw := dy + 1
			startX := px + 5 - pw
			endX := px + 5 + pw
			for x := startX; x <= endX; x++ {
				c := woodMid
				if x < px+4 {
					c = woodLight
				} else if x > px+6 {
					c = woodDark
				}
				setPixel(img, x, topY-8+dy, c)
			}
		}

		// Picket body (width 10)
		for y := topY; y <= baseY; y++ {
			for x := px; x < px+10; x++ {
				c := woodMid
				if x < px+3 {
					c = woodLight
				} else if x >= px+7 {
					c = woodDark
				}
				setPixel(img, x, y, c)
			}
		}

		// Fastening nails
		tRailY := 112 + (px+5)/2 + 4
		bRailY := 160 + (px+5)/2 + 4
		fillRect(img, px+4, tRailY, 2, 2, nailColor)
		setPixel(img, px+4, tRailY, nailHighlight)
		fillRect(img, px+4, bRailY, 2, 2, nailColor)
		setPixel(img, px+4, bRailY, nailHighlight)
	}

	// 3. Corner Post (x in [120..135], y in [112..240])
	for y := 112; y <= 240; y++ {
		for x := 120; x <= 135; x++ {
			c := woodMid
			if x < 125 {
				c = woodLight
			} else if x > 130 {
				c = woodDark
			}
			setPixel(img, x, y, c)
		}
	}
	// Pyramid cap for corner post (y in [96..111])
	for dy := 0; dy < 16; dy++ {
		halfW := dy / 2
		for x := 127 - halfW; x <= 128 + halfW; x++ {
			c := woodMid
			if x < 127 {
				c = woodLight
			} else if x > 128 {
				c = woodDark
			}
			setPixel(img, x, 96+dy, c)
		}
	}

	// 4. Left Post (x in [8..23], y in [56..184])
	for y := 56; y <= 184; y++ {
		for x := 8; x <= 23; x++ {
			c := woodMid
			if x < 13 {
				c = woodLight
			} else if x > 18 {
				c = woodDark
			}
			setPixel(img, x, y, c)
		}
	}
	// Pyramid cap for left post (y in [40..55])
	for dy := 0; dy < 16; dy++ {
		halfW := dy / 2
		for x := 15 - halfW; x <= 16 + halfW; x++ {
			c := woodMid
			if x < 15 {
				c = woodLight
			} else if x > 16 {
				c = woodDark
			}
			setPixel(img, x, 40+dy, c)
		}
	}

	saveImg(name, img)
}

func generateIsoDebris(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	crateTop := color.RGBA{190, 142, 92, 255}
	crateLeft := color.RGBA{156, 112, 70, 255}
	crateRight := color.RGBA{102, 70, 42, 255}
	frameWood := color.RGBA{124, 84, 52, 255}
	ironBracket := color.RGBA{78, 82, 88, 255}
	ironRivet := color.RGBA{190, 195, 205, 255}

	concreteMid := color.RGBA{130, 130, 125, 255}
	concreteLight := color.RGBA{170, 170, 165, 255}
	concreteDark := color.RGBA{75, 75, 72, 255}
	brickRed := color.RGBA{145, 58, 42, 255}
	groundShadow := color.RGBA{20, 20, 20, 120}

	// 1. Ground drop shadow
	drawAAEllipse(img, 128, 212, 80, 28, groundShadow)

	// 2. Crate Top Face (Diamond at (128, 104), rx=56, ry=28)
	for y := 76; y <= 132; y++ {
		for x := 72; x <= 184; x++ {
			dx := float64(x - 128)
			dy := float64(y - 104)
			metric := math.Abs(dx)/56.0 + math.Abs(dy)/28.0
			if metric <= 1.0 {
				c := crateTop
				if metric > 0.76 {
					c = frameWood
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 3. Crate Left Face (x in [72..128])
	for x := 72; x <= 128; x++ {
		topY := 104 + (x-72)/2
		botY := 168 + (x-72)/2
		for y := topY; y <= botY; y++ {
			c := crateLeft
			relX := x - 72
			relY := y - topY
			if relX < 6 || relX > 50 || relY < 6 || relY > 58 {
				c = frameWood
			} else {
				normX := float64(relX-6) / 44.0
				normY := float64(relY-6) / 52.0
				if math.Abs(normX-normY) < 0.10 || math.Abs(normX-(1.0-normY)) < 0.10 {
					c = frameWood
				}
			}
			setPixel(img, x, y, c)
		}
	}

	// 4. Crate Right Face (x in [128..184])
	for x := 128; x <= 184; x++ {
		topY := 132 - (x-128)/2
		botY := 196 - (x-128)/2
		for y := topY; y <= botY; y++ {
			c := crateRight
			relX := x - 128
			relY := y - topY
			if relX < 6 || relX > 50 || relY < 6 || relY > 58 {
				c = darken(frameWood, 0.75)
			} else {
				normX := float64(relX-6) / 44.0
				normY := float64(relY-6) / 52.0
				if math.Abs(normX-normY) < 0.10 || math.Abs(normX-(1.0-normY)) < 0.10 {
					c = darken(frameWood, 0.75)
				}
			}
			setPixel(img, x, y, c)
		}
	}

	// 5. Iron Corner Brackets & Rivets
	cornerPts := [][2]int{
		{72, 104}, {128, 132}, {184, 104}, {128, 76},
		{72, 168}, {128, 196}, {184, 168},
	}
	for _, pt := range cornerPts {
		fillRect(img, pt[0]-4, pt[1]-4, 9, 9, ironBracket)
		fillRect(img, pt[0]-1, pt[1]-1, 3, 3, ironRivet)
	}

	// 6. Concrete & Brick Rubble
	drawDebrisChunk := func(rx, ry, rw, rh int, base, light, dark color.RGBA) {
		for y := ry; y < ry+rh; y++ {
			for x := rx; x < rx+rw; x++ {
				dx := float64(x - rx - rw/2)
				dy := float64(y - ry - rh/2)
				if (dx*dx)/float64(rw*rw/4)+(dy*dy)/float64(rh*rh/4) <= 1.0 {
					c := base
					if dy < -float64(rh)*0.2 {
						c = light
					} else if dy > float64(rh)*0.2 {
						c = dark
					}
					setPixel(img, x, y, c)
				}
			}
		}
	}

	drawDebrisChunk(32, 192, 32, 24, concreteMid, concreteLight, concreteDark)
	drawDebrisChunk(184, 204, 36, 28, concreteMid, concreteLight, concreteDark)
	drawDebrisChunk(56, 220, 24, 16, brickRed, lighten(brickRed, 1.2), darken(brickRed, 0.7))
	drawDebrisChunk(152, 216, 20, 14, brickRed, lighten(brickRed, 1.2), darken(brickRed, 0.7))

	saveImg(name, img)
}

func generateIsoTent(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	canvasLight := color.RGBA{88, 132, 68, 255}
	canvasShadow := color.RGBA{48, 78, 36, 255}
	canvasInside := color.RGBA{22, 36, 18, 255}
	poleSilver := color.RGBA{200, 205, 210, 255}
	ropeBeige := color.RGBA{195, 175, 140, 255}
	groundShadow := color.RGBA{0, 0, 0, 50}

	// 1. Ground shadow
	drawAAEllipse(img, 128, 208, 88, 32, groundShadow)

	// 2. Left Slope face
	for y := 64; y <= 180; y++ {
		for x := 0; x <= 160; x++ {
			topBound := 64.0 + float64(x-96)*0.5
			if x < 96 {
				topBound = 64.0 - float64(96-x)*0.83
			}
			botBound := 144.0 + float64(x)*0.5
			if float64(y) >= topBound && float64(y) <= botBound && float64(y) <= 64.0+float64(x)*1.17 {
				setPixel(img, x, y, canvasLight)
			}
		}
	}

	// 3. Right Slope / Front Opening
	for y := 96; y <= 224; y++ {
		for x := 64; x <= 224; x++ {
			// Right canvas face
			if x >= 160 && x <= 224 {
				topY := 96.0 + float64(x-160)*2.0
				botY := 176.0 + float64(x-64)*0.38
				if float64(y) >= topY && float64(y) <= botY {
					setPixel(img, x, y, canvasShadow)
				}
			}
			// Front triangle opening
			if x >= 64 && x <= 192 {
				leftSlope := 96.0 + float64(160-x)*0.83
				rightSlope := 96.0 + float64(x-160)*4.0
				botY := 176.0 + float64(x-64)*0.38
				if float64(y) >= leftSlope && float64(y) >= rightSlope && float64(y) <= botY {
					setPixel(img, x, y, canvasInside)
				}
			}
		}
	}

	// 4. Ridge Pole
	for i := 0; i <= 64; i++ {
		px := 96 + i
		py := 64 + i/2
		setPixel(img, px, py, poleSilver)
		setPixel(img, px, py+1, poleSilver)
	}

	// 5. Guy lines and stakes
	drawVLine(img, 160, 96, 210, ropeBeige)
	drawHLine(img, 156, 164, 210, color.RGBA{220, 180, 40, 255})

	saveImg(name, img)
}

func generateIsoStump(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	barkLight := color.RGBA{115, 78, 48, 255}
	barkMid := color.RGBA{88, 56, 32, 255}
	barkDark := color.RGBA{54, 34, 18, 255}
	woodTop := color.RGBA{205, 155, 105, 255}
	ringColor := color.RGBA{168, 118, 74, 255}
	mossGreen := color.RGBA{88, 148, 62, 255}
	groundShadow := color.RGBA{0, 0, 0, 50}

	// 1. Ground shadow
	drawAAEllipse(img, 128, 208, 60, 22, groundShadow)

	// 2. Trunk Body (x in [80..176], y in [136..208])
	for y := 136; y <= 208; y++ {
		progress := float64(y-136) / 72.0
		halfW := 48.0 + progress*16.0
		minX := int(128.0 - halfW)
		maxX := int(128.0 + halfW)
		for x := minX; x <= maxX; x++ {
			c := barkMid
			if x < 128-int(halfW*0.4) {
				c = barkLight
			} else if x > 128+int(halfW*0.4) {
				c = barkDark
			}
			if y > 185 && (x+y)%5 < 2 {
				c = mossGreen
			}
			setPixel(img, x, y, c)
		}
	}

	// 3. Top Cut Surface Ellipse at (128, 136), rx=48, ry=24
	for y := 112; y <= 160; y++ {
		for x := 80; x <= 176; x++ {
			dx := float64(x - 128)
			dy := float64(y - 136)
			dist := (dx*dx)/(48.0*48.0) + (dy*dy)/(24.0*24.0)
			if dist <= 1.0 {
				c := woodTop
				if dist > 0.82 {
					c = barkMid
				} else {
					d := math.Sqrt(dist)
					if math.Abs(d-0.28) < 0.035 || math.Abs(d-0.52) < 0.035 || math.Abs(d-0.72) < 0.035 {
						c = ringColor
					}
					angle := math.Atan2(dy*2.0, dx)
					if math.Abs(angle-0.35) < 0.04 && d > 0.15 {
						c = barkDark
					}
				}
				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}

func generateIsoMushroom(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	capBase := color.RGBA{220, 42, 42, 255}
	capGloss := color.RGBA{255, 110, 110, 255}
	capShadow := color.RGBA{140, 24, 24, 255}
	dotWhite := color.RGBA{250, 250, 250, 255}
	stemBase := color.RGBA{235, 228, 212, 255}
	stemShadow := color.RGBA{180, 170, 150, 255}
	groundShadow := color.RGBA{0, 0, 0, 50}

	// 1. Ground shadow
	drawAAEllipse(img, 128, 216, 56, 20, groundShadow)

	// 2. Hero Stem (Stipe)
	for y := 136; y <= 212; y++ {
		for x := 112; x <= 144; x++ {
			c := stemBase
			if x > 132 {
				c = stemShadow
			}
			setPixel(img, x, y, c)
		}
	}
	drawHLine(img, 108, 148, 156, color.RGBA{250, 245, 235, 255})
	drawHLine(img, 108, 148, 157, color.RGBA{220, 210, 195, 255})

	// 3. Hero Cap Dome (Ellipse at (128, 104), rx=72, ry=48)
	for y := 56; y <= 144; y++ {
		for x := 56; x <= 200; x++ {
			dx := float64(x - 128)
			dy := float64(y - 104)
			dist := (dx*dx)/(72.0*72.0) + (dy*dy)/(48.0*48.0)
			if dist <= 1.0 {
				c := capBase
				if dx < -15 && dy < -10 {
					c = capGloss
				} else if dx > 20 && dy > 10 {
					c = capShadow
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 4. White Polka Dots
	dots := []struct {
		cx, cy int
		r      float64
	}{
		{104, 88, 8.0},
		{144, 82, 9.0},
		{124, 68, 7.5},
		{88, 112, 7.0},
		{160, 108, 8.5},
		{128, 108, 9.0},
		{172, 88, 6.5},
	}
	for _, dot := range dots {
		drawFilledCircle(img, dot.cx, dot.cy, dot.r, dotWhite)
	}

	// 5. Companion Sprout Mushroom
	drawAAEllipse(img, 76, 196, 16, 6, groundShadow)
	for y := 178; y <= 196; y++ {
		for x := 72; x <= 80; x++ {
			c := stemBase
			if x > 76 {
				c = stemShadow
			}
			setPixel(img, x, y, c)
		}
	}
	drawAAEllipse(img, 76, 178, 24, 16, capBase)
	drawFilledCircle(img, 70, 174, 3.5, dotWhite)
	drawFilledCircle(img, 82, 176, 3.0, dotWhite)

	saveImg(name, img)
}

func generateIsoSign(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	postWood := color.RGBA{110, 76, 46, 255}
	postShadow := color.RGBA{75, 50, 30, 255}
	boardLight := color.RGBA{165, 122, 80, 255}
	boardMid := color.RGBA{135, 95, 60, 255}
	hazardYellow := color.RGBA{245, 195, 35, 255}
	hazardBlack := color.RGBA{35, 30, 25, 255}
	boltColor := color.RGBA{180, 185, 190, 255}
	groundShadow := color.RGBA{0, 0, 0, 50}

	// 1. Ground shadow
	drawAAEllipse(img, 128, 220, 48, 16, groundShadow)

	// 2. Vertical Post (x in [120..135], y in [96..224])
	for y := 96; y <= 224; y++ {
		for x := 120; x <= 135; x++ {
			c := postWood
			if x > 128 {
				c = postShadow
			}
			setPixel(img, x, y, c)
		}
	}
	// Post pyramid top
	for dy := 0; dy < 12; dy++ {
		hw := dy / 2
		for x := 127 - hw; x <= 128 + hw; x++ {
			c := postWood
			if x > 128 {
				c = postShadow
			}
			setPixel(img, x, 84+dy, c)
		}
	}

	// 3. Directional Arrow 1 (NW pointing, x in [56..168], y in [64..104])
	for y := 64; y <= 104; y++ {
		for x := 56; x <= 168; x++ {
			tipOffset := math.Abs(float64(y-84)) * 0.75
			if float64(x-56) >= tipOffset {
				c := boardMid
				if y < 72 {
					c = boardLight
				}
				if (x+y)%24 < 12 {
					c = hazardYellow
				} else {
					c = hazardBlack
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 4. Directional Arrow 2 (NE pointing, x in [112..216], y in [116..156])
	for y := 116; y <= 156; y++ {
		for x := 112; x <= 216; x++ {
			tipOffset := math.Abs(float64(y-136)) * 0.75
			if float64(216-x) >= tipOffset {
				c := boardMid
				if y < 124 {
					c = boardLight
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 5. Bolts
	drawFilledCircle(img, 128, 84, 3.0, boltColor)
	drawFilledCircle(img, 128, 136, 3.0, boltColor)

	saveImg(name, img)
}

func generateElevationBlock(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	grassTop := color.RGBA{106, 186, 70, 255}
	cliffWest := color.RGBA{120, 95, 70, 255}
	cliffSouth := color.RGBA{85, 65, 45, 255}
	ridgeHighlight := color.RGBA{180, 230, 140, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// 1. Top Face (Grass Diamond, centered at (128, 64), rx=128, ry=64)
			if y <= 128 {
				dx := float64(x) - 127.5
				dy := float64(y) - 63.5
				if math.Abs(dx)/128.0+math.Abs(dy)/64.0 <= 1.0 {
					setPixel(img, x, y, grassTop)
				}
			}
			// 2. Left Cliff Face (West) for x in [0..127]
			if x < 128 {
				topY := 64 + x/2
				botY := 192 + x/2
				if y > topY && y <= botY && y < h {
					setPixel(img, x, y, cliffWest)
				}
			}
			// 3. Right Cliff Face (South) for x in [128..255]
			if x >= 128 {
				topY := 128 - (x-128)/2
				botY := 256 - (x-128)/2
				if y > topY && y <= botY && y < h {
					setPixel(img, x, y, cliffSouth)
				}
			}
		}
	}

	// Ridge Bevel Highlight
	for x := 0; x < 128; x++ {
		topY := 64 + x/2
		setPixel(img, x, topY, ridgeHighlight)
		setPixel(img, x, topY+1, ridgeHighlight)
	}
	for x := 128; x < 256; x++ {
		topY := 128 - (x-128)/2
		setPixel(img, x, topY, ridgeHighlight)
		setPixel(img, x, topY+1, ridgeHighlight)
	}

	saveImg(name, img)
}

func generateElevationRamp(name string) {
	w, h := 256, 256
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	rampGrass := color.RGBA{125, 195, 80, 255}
	rampDirtSide := color.RGBA{85, 65, 45, 255}
	stonePaver := color.RGBA{160, 155, 150, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			topY := x / 2
			botY := 128 + x/2
			if y >= topY && y <= botY {
				c := rampGrass
				if (x+y)%32 < 4 {
					c = stonePaver
				}
				setPixel(img, x, y, c)
			}
			if x >= 128 && y > 128+(x-128)/2 && y < h {
				setPixel(img, x, y, rampDirtSide)
			}
		}
	}

	saveImg(name, img)
}

// -------------------------------------------------------------
// 4. ITEMS & EQUIPMENT (64x64)
// -------------------------------------------------------------

func generateFood(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{35, 38, 42, 255}
	tinLid := color.RGBA{185, 190, 198, 255}
	tinLidHi := color.RGBA{240, 245, 252, 255}
	tinLidSh := color.RGBA{130, 135, 142, 255}
	tinRim := color.RGBA{215, 220, 228, 255}
	metalBodyHi := color.RGBA{220, 225, 235, 255}
	metalBodySh := color.RGBA{110, 115, 125, 255}
	labelRed := color.RGBA{205, 35, 28, 255}
	labelRedHi := color.RGBA{245, 75, 65, 255}
	labelRedSh := color.RGBA{125, 18, 14, 255}
	labelGold := color.RGBA{245, 195, 45, 255}
	labelGoldHi := color.RGBA{255, 230, 120, 255}
	emblemGreen := color.RGBA{45, 145, 40, 255}

	// 1. Bottom Base Ellipse (center at (31.5, 53.5), rx=17.5, ry=5.5)
	for y := 49; y <= 58; y++ {
		for x := 14; x <= 49; x++ {
			dx := (float64(x) - 31.5) / 17.5
			dy := (float64(y) - 53.5) / 5.5
			if dx*dx+dy*dy <= 1.0 {
				setPixel(img, x, y, tinRim)
			}
		}
	}

	// 2. Cylinder Body (y in [15..52], x in [14..49])
	for y := 15; y <= 52; y++ {
		for x := 14; x <= 49; x++ {
			t := math.Cos((float64(x-22) / 35.0) * (math.Pi / 2.0))
			t = math.Max(0, math.Min(1, t))

			var c color.RGBA
			if y <= 18 || y >= 49 {
				c = blend(metalBodySh, metalBodyHi, t)
			} else {
				c = labelRed
				if x >= 20 && x <= 26 {
					c = labelRedHi
				} else if x >= 42 {
					c = labelRedSh
				}
				if (y >= 26 && y <= 28) || (y >= 38 && y <= 40) {
					c = blend(labelGold, labelGoldHi, t)
				}
				dx := float64(x) - 31.5
				dy := float64(y) - 33.5
				if dx*dx+dy*dy <= 25.0 {
					if dx*dx+dy*dy <= 16.0 {
						c = emblemGreen
					} else {
						c = labelGoldHi
					}
				}
			}
			setPixel(img, x, y, c)
		}
	}

	// 3. Top Lid Ellipse (center at (31.5, 13.5), rx=17.5, ry=5.5)
	for y := 8; y <= 18; y++ {
		for x := 14; x <= 49; x++ {
			dx := (float64(x) - 31.5) / 17.5
			dy := (float64(y) - 13.5) / 5.5
			dist := dx*dx + dy*dy
			if dist <= 1.0 {
				c := tinLid
				if dist > 0.85 {
					c = tinRim
				} else if dx < -0.2 {
					c = tinLidHi
				} else if dx > 0.3 {
					c = tinLidSh
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 4. Pull-Tab Ring at (31.5, 12)
	fillRect(img, 28, 9, 8, 4, color.RGBA{240, 245, 252, 255})
	fillRect(img, 30, 10, 4, 2, color.RGBA{50, 55, 62, 255})
	setPixel(img, 31, 12, color.RGBA{180, 185, 192, 255})

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}

func generateWater(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{25, 45, 75, 255}
	capWhite := color.RGBA{245, 248, 255, 255}
	capBlue := color.RGBA{170, 200, 235, 255}
	capSh := color.RGBA{115, 145, 185, 255}
	highlight := color.RGBA{245, 252, 255, 255}
	waterDeep := color.RGBA{18, 65, 175, 255}
	waterLight := color.RGBA{145, 210, 255, 255}
	bubbleColor := color.RGBA{210, 240, 255, 255}

	getProfileHalfW := func(y int) float64 {
		if y < 4 || y > 59 {
			return 0
		}
		if y <= 11 {
			return 6.0
		}
		if y <= 14 {
			return 7.0
		}
		if y <= 19 {
			return 5.0
		}
		if y <= 27 {
			t := float64(y-19) / 8.0
			return 5.0 + t*11.0
		}
		if y <= 35 {
			return 16.0
		}
		if y <= 45 {
			t := float64(y-35) / 10.0
			return 16.0 - 3.0*math.Sin(t*math.Pi)
		}
		if y <= 54 {
			return 16.0
		}
		return 15.0
	}

	for y := 4; y <= 59; y++ {
		hw := getProfileHalfW(y)
		minX := int(math.Round(31.5 - hw))
		maxX := int(math.Round(31.5 + hw))

		for x := minX; x <= maxX; x++ {
			normX := (float64(x) - 31.5) / hw

			if y <= 14 {
				c := capBlue
				if (x-minX)%3 == 0 {
					c = capWhite
				} else if normX > 0.4 {
					c = capSh
				}
				setPixel(img, x, y, c)
			} else if y < 24 {
				c := color.RGBA{180, 220, 250, 200}
				if normX < -0.6 {
					c = highlight
				}
				setPixel(img, x, y, c)
			} else {
				t := (normX + 1.0) / 2.0
				c := blend(waterLight, waterDeep, t)
				if normX < -0.5 && normX > -0.8 {
					c = highlight
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 3 Horizontal Grip Ridges on waist
	for _, ry := range []int{37, 40, 43} {
		hw := getProfileHalfW(ry)
		minX := int(31.5 - hw + 2)
		maxX := int(31.5 + hw - 2)
		drawHLine(img, minX, maxX, ry, highlight)
		drawHLine(img, minX, maxX, ry+1, waterDeep)
	}

	// Air bubbles
	drawFilledCircle(img, 26, 33, 1.5, bubbleColor)
	drawFilledCircle(img, 38, 42, 2.0, bubbleColor)
	drawFilledCircle(img, 29, 48, 1.2, bubbleColor)

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}

func generateWeapon(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{45, 28, 15, 255}
	woodSh := color.RGBA{115, 65, 25, 255}
	woodMid := color.RGBA{185, 128, 65, 255}
	woodHi := color.RGBA{238, 188, 122, 255}
	tapeBase := color.RGBA{230, 226, 215, 255}
	tapeSh := color.RGBA{145, 138, 122, 255}
	steelSpike := color.RGBA{245, 250, 255, 255}
	steelBase := color.RGBA{90, 100, 115, 255}
	blood := color.RGBA{168, 22, 22, 255}

	p0x, p0y := 8.0, 56.0
	p1x, p1y := 54.0, 9.0
	batLen := math.Hypot(p1x-p0x, p1y-p0y)
	dirX := (p1x - p0x) / batLen
	dirY := (p1y - p0y) / batLen
	normX := -dirY
	normY := dirX

	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			vx := float64(x) - p0x
			vy := float64(y) - p0y
			s := (vx*dirX + vy*dirY) / batLen
			dPerp := vx*normX + vy*normY

			if s >= -0.05 && s <= 1.05 {
				var r float64
				if s < 0.05 {
					r = 4.5
				} else if s < 0.30 {
					r = 3.5
				} else if s < 0.50 {
					t := (s - 0.30) / 0.20
					r = 3.5 + t*1.5
				} else if s < 0.95 {
					t := (s - 0.50) / 0.45
					r = 5.0 + t*1.8
				} else {
					t := (s - 0.95) / 0.05
					r = 6.8 * math.Sqrt(math.Max(0, 1.0-t*t))
				}

				if math.Abs(dPerp) <= r {
					var c color.RGBA
					if s >= 0.05 && s < 0.30 {
						c = tapeBase
						relTape := int(s*batLen) % 4
						if relTape == 0 || dPerp > 1.0 {
							c = tapeSh
						}
					} else {
						tLight := (dPerp + r) / (2.0 * r)
						if tLight < 0.35 {
							c = woodHi
						} else if tLight < 0.70 {
							c = woodMid
						} else {
							c = woodSh
						}
					}
					setPixel(img, x, y, c)
				}
			}
		}
	}

	// 6 Steel Spikes
	spikes := []struct {
		bx, by, tx, ty int
	}{
		{40, 22, 34, 16},
		{44, 18, 48, 12},
		{47, 15, 43, 21},
		{51, 11, 46, 5},
		{37, 25, 43, 29},
		{49, 13, 56, 17},
	}
	for _, sp := range spikes {
		steps := 8
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			sx := int(math.Round(float64(sp.bx)*(1.0-t) + float64(sp.tx)*t))
			sy := int(math.Round(float64(sp.by)*(1.0-t) + float64(sp.ty)*t))
			c := steelBase
			if t > 0.6 {
				c = steelSpike
			}
			setPixel(img, sx, sy, c)
		}
	}

	drawFilledCircle(img, 46, 5, 2.0, blood)
	drawFilledCircle(img, 56, 17, 2.2, blood)
	drawFilledCircle(img, 50, 8, 2.5, blood)

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}

func generateAxe(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{32, 30, 32, 255}
	gripRubber := color.RGBA{38, 40, 45, 255}
	gripHi := color.RGBA{75, 80, 90, 255}
	woodMid := color.RGBA{220, 155, 70, 255}
	woodHi := color.RGBA{250, 195, 125, 255}
	woodSh := color.RGBA{135, 80, 25, 255}
	axeRed := color.RGBA{220, 32, 32, 255}
	axeRedHi := color.RGBA{255, 88, 88, 255}
	axeRedSh := color.RGBA{135, 15, 15, 255}
	steelEye := color.RGBA{85, 92, 102, 255}
	steelBevel := color.RGBA{180, 195, 210, 255}
	steelEdge := color.RGBA{250, 253, 255, 255}

	// 1. Curved Hickory Handle
	steps := 100
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		hx := (1-t)*(1-t)*10.0 + 2*(1-t)*t*20.0 + t*t*34.0
		hy := (1-t)*(1-t)*58.0 + 2*(1-t)*t*42.0 + t*t*18.0

		for r := -3; r <= 3; r++ {
			px := int(math.Round(hx + float64(r)*0.7))
			py := int(math.Round(hy - float64(r)*0.7))

			if hy >= 48 {
				c := gripRubber
				if r < -1 {
					c = gripHi
				}
				setPixel(img, px, py, c)
			} else {
				c := woodMid
				if r < -1 {
					c = woodHi
				} else if r > 1 {
					c = woodSh
				}
				setPixel(img, px, py, c)
			}
		}
	}

	// 2. Steel Eye Collar
	fillRect(img, 28, 12, 11, 13, steelEye)
	fillRect(img, 28, 12, 4, 13, color.RGBA{115, 125, 138, 255})
	fillRect(img, 36, 12, 3, 13, color.RGBA{55, 60, 68, 255})

	// 3. Rear Breaching Pick
	for x := 15; x <= 28; x++ {
		t := float64(x-15) / 13.0
		halfH := int(math.Round(t * 3.5))
		for y := 16 - halfH; y <= 16 + halfH; y++ {
			c := steelBevel
			if y == 16-halfH {
				c = steelEdge
			}
			setPixel(img, x, y, c)
		}
	}

	// 4. Axe Blade Body
	for x := 36; x <= 52; x++ {
		t := float64(x-36) / 16.0
		halfH := int(math.Round(5.0 + t*6.0))
		for y := 18 - halfH; y <= 18 + halfH; y++ {
			c := axeRed
			if y <= 18-halfH+2 {
				c = axeRedHi
			} else if y >= 18+halfH-2 {
				c = axeRedSh
			}
			setPixel(img, x, y, c)
		}
	}

	// 5. Beveled Cutting Edge
	for y := 6; y <= 29; y++ {
		dy := float64(y - 17)
		apexX := int(math.Round(57.0 - (dy*dy)/36.0))
		for x := apexX - 4; x <= apexX; x++ {
			c := steelBevel
			if x >= apexX-1 {
				c = steelEdge
			}
			setPixel(img, x, y, c)
		}
	}

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}

func generateShotgun(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{24, 26, 30, 255}
	buttRubber := color.RGBA{38, 38, 42, 255}
	woodStock := color.RGBA{148, 90, 44, 255}
	woodHi := color.RGBA{200, 132, 75, 255}
	woodSh := color.RGBA{95, 52, 22, 255}
	steelRec := color.RGBA{68, 74, 85, 255}
	steelRecHi := color.RGBA{118, 128, 145, 255}
	ejectPort := color.RGBA{22, 24, 28, 255}
	pumpWood := color.RGBA{160, 102, 50, 255}
	pumpRib := color.RGBA{75, 42, 18, 255}
	barrelHi := color.RGBA{190, 200, 215, 255}
	barrelMid := color.RGBA{85, 92, 105, 255}
	barrelSh := color.RGBA{42, 48, 56, 255}
	beadSight := color.RGBA{250, 215, 60, 255}

	// 1. Buttpad
	fillRect(img, 6, 49, 4, 7, buttRubber)
	setPixel(img, 7, 51, color.RGBA{140, 145, 155, 255})
	setPixel(img, 7, 53, color.RGBA{140, 145, 155, 255})

	// 2. Walnut Stock
	for x := 9; x <= 22; x++ {
		t := float64(x-9) / 13.0
		topY := int(math.Round(49.0 - t*11.0))
		botY := int(math.Round(55.0 - t*14.0))
		for y := topY; y <= botY; y++ {
			c := woodStock
			if y <= topY+1 {
				c = woodHi
			} else if y >= botY-1 {
				c = woodSh
			}
			setPixel(img, x, y, c)
		}
	}

	// 3. Trigger Guard & Trigger
	fillRect(img, 22, 36, 5, 6, darkBorder)
	fillRect(img, 23, 37, 3, 4, color.RGBA{0, 0, 0, 0})
	setPixel(img, 24, 38, color.RGBA{170, 175, 185, 255})

	// 4. Milled Steel Receiver
	for x := 22; x <= 34; x++ {
		for y := 26; y <= 37; y++ {
			c := steelRec
			if y <= 28 {
				c = steelRecHi
			} else if y >= 35 {
				c = barrelSh
			}
			setPixel(img, x, y, c)
		}
	}
	fillRect(img, 26, 28, 6, 4, ejectPort)
	setPixel(img, 27, 29, color.RGBA{235, 190, 55, 255})

	// 5. Pump Forend Slide
	for x := 34; x <= 44; x++ {
		for y := 20; y <= 29; y++ {
			c := pumpWood
			if (x-34)%2 == 0 {
				c = pumpRib
			}
			setPixel(img, x, y, c)
		}
	}

	// 6. Top Barrel & Mag Tube
	for x := 34; x <= 58; x++ {
		t := float64(x-34) / 24.0
		barrelY := int(math.Round(25.0 - t*12.0))

		setPixel(img, x, barrelY-1, barrelHi)
		setPixel(img, x, barrelY, barrelMid)
		setPixel(img, x, barrelY+1, barrelSh)

		if x <= 52 {
			setPixel(img, x, barrelY+3, barrelMid)
			setPixel(img, x, barrelY+4, barrelSh)
		}
	}

	// 7. Bead Sight
	setPixel(img, 56, 12, beadSight)
	setPixel(img, 57, 12, beadSight)
	fillRect(img, 58, 13, 2, 2, darkBorder)

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}

func generateAmmo(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{22, 30, 16, 255}
	boxGreen := color.RGBA{68, 90, 45, 255}
	boxGreenHi := color.RGBA{100, 132, 65, 255}
	boxGreenSh := color.RGBA{40, 55, 25, 255}
	stencilYellow := color.RGBA{240, 190, 40, 255}
	stencilText := color.RGBA{25, 32, 20, 255}
	brassHi := color.RGBA{255, 238, 135, 255}
	brassMid := color.RGBA{235, 190, 55, 255}
	brassSh := color.RGBA{155, 115, 25, 255}
	copperTip := color.RGBA{215, 105, 45, 255}
	copperTipHi := color.RGBA{255, 165, 98, 255}
	rivetColor := color.RGBA{170, 180, 175, 255}

	// 1. Standing Cartridges (6 cartridges)
	cartridgesX := []int{15, 21, 27, 33, 39, 45}
	for _, cx := range cartridgesX {
		fillRect(img, cx, 8, 3, 6, copperTip)
		setPixel(img, cx, 8, copperTipHi)
		setPixel(img, cx, 9, copperTipHi)

		fillRect(img, cx-1, 14, 5, 8, brassMid)
		drawVLine(img, cx-1, 14, 21, brassHi)
		drawVLine(img, cx+3, 14, 21, brassSh)
	}

	// 2. Ammo Box Top Rim
	fillRect(img, 10, 21, 44, 4, boxGreenHi)

	// 3. Main Box Front Wall
	for y := 24; y <= 54; y++ {
		drawHLine(img, 10, 53, y, boxGreen)
		setPixel(img, 10, y, boxGreenHi)
		setPixel(img, 11, y, boxGreenHi)
		setPixel(img, 52, y, boxGreenSh)
		setPixel(img, 53, y, boxGreenSh)
	}

	drawShadedRect(img, 14, 26, 36, 27, boxGreen, boxGreenSh, boxGreenHi)

	// Stencil Band
	for y := 32; y <= 40; y++ {
		drawHLine(img, 16, 47, y, stencilYellow)
	}
	for x := 18; x <= 45; x += 3 {
		drawVLine(img, x, 34, 38, stencilText)
	}

	// Corner Rivets
	rivets := [][2]int{{13, 25}, {50, 25}, {13, 51}, {50, 51}}
	for _, pt := range rivets {
		fillRect(img, pt[0], pt[1], 2, 2, rivetColor)
	}

	// Bottom Base Rim
	fillRect(img, 10, 54, 44, 5, boxGreenSh)

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}

func generateArmor(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{18, 20, 25, 255}
	vestKevlar := color.RGBA{48, 54, 65, 255}
	vestHi := color.RGBA{80, 88, 105, 255}
	vestSh := color.RGBA{28, 32, 40, 255}
	strapHi := color.RGBA{92, 100, 118, 255}
	buckleSteel := color.RGBA{145, 155, 170, 255}
	idPatch := color.RGBA{95, 90, 75, 255}
	idPatchHi := color.RGBA{125, 120, 102, 255}
	molleWeb := color.RGBA{26, 30, 36, 255}
	pouchFlap := color.RGBA{66, 75, 90, 255}
	pouchBody := color.RGBA{38, 44, 54, 255}
	pullTab := color.RGBA{100, 110, 128, 255}

	// 1. Shoulder Straps
	fillRect(img, 12, 6, 10, 13, vestKevlar)
	drawVLine(img, 12, 6, 18, strapHi)
	fillRect(img, 14, 11, 6, 3, buckleSteel)

	fillRect(img, 42, 6, 10, 13, vestKevlar)
	drawVLine(img, 51, 6, 18, vestSh)
	fillRect(img, 44, 11, 6, 3, buckleSteel)

	// 2. Scooped Neckline Area
	for y := 6; y <= 18; y++ {
		for x := 22; x <= 41; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 6.0
			if (dx*dx)/100.0+(dy*dy)/144.0 > 1.0 {
				setPixel(img, x, y, vestKevlar)
			}
		}
	}

	// 3. Chest Plate Carrier Body
	for y := 18; y <= 54; y++ {
		drawHLine(img, 14, 49, y, vestKevlar)
		drawVLine(img, 14, 18, 54, vestHi)
		drawVLine(img, 49, 18, 54, vestSh)
	}

	// 4. Velcro ID Patch
	fillRect(img, 22, 19, 20, 7, idPatch)
	drawHLine(img, 22, 41, 19, idPatchHi)

	// 5. MOLLE Webbing Rows
	for _, my := range []int{27, 33, 39} {
		fillRect(img, 15, my, 34, 2, molleWeb)
		for sx := 15; sx <= 49; sx += 6 {
			drawVLine(img, sx, my, my+1, color.RGBA{60, 68, 80, 255})
		}
	}

	// 6. 3 Magazine Utility Pouches
	pouches := [][2]int{{15, 24}, {27, 36}, {39, 48}}
	for _, p := range pouches {
		fillRect(img, p[0], 41, p[1]-p[0]+1, 13, pouchBody)
		fillRect(img, p[0], 41, p[1]-p[0]+1, 3, pouchFlap)
		midX := (p[0] + p[1]) / 2
		fillRect(img, midX-1, 44, 3, 3, pullTab)
	}

	// 7. Bottom Hem
	fillRect(img, 12, 54, 40, 5, vestSh)

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}

func generateAntidote(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	darkBorder := color.RGBA{12, 18, 15, 255}
	corkMid := color.RGBA{150, 110, 65, 255}
	corkHi := color.RGBA{185, 145, 95, 255}
	corkSh := color.RGBA{105, 72, 38, 255}
	glassHi := color.RGBA{235, 245, 255, 255}
	liquidCore := color.RGBA{65, 250, 85, 255}
	liquidGlow := color.RGBA{130, 255, 150, 255}
	liquidMid := color.RGBA{32, 195, 55, 255}
	liquidDark := color.RGBA{18, 135, 35, 255}
	bubbleGlow := color.RGBA{180, 255, 195, 255}
	gradLine := color.RGBA{220, 245, 230, 255}

	// 1. Cork Stopper
	fillRect(img, 26, 6, 12, 10, corkMid)
	drawHLine(img, 26, 37, 6, corkHi)
	drawVLine(img, 26, 6, 15, corkHi)
	drawVLine(img, 37, 6, 15, corkSh)

	// 2. Glass Flanged Lip & Neck
	fillRect(img, 24, 15, 16, 4, glassHi)
	fillRect(img, 26, 19, 12, 7, color.RGBA{200, 230, 245, 180})

	// 3. Ampoule Body & Antidote Liquid
	for y := 26; y <= 58; y++ {
		var hw float64
		if y <= 31 {
			t := float64(y-25) / 6.0
			hw = 6.0 + t*11.5
		} else if y <= 52 {
			hw = 17.5
		} else {
			t := float64(y-52) / 6.0
			hw = 17.5 * math.Sqrt(math.Max(0, 1.0-t*t))
		}

		minX := int(math.Round(31.5 - hw))
		maxX := int(math.Round(31.5 + hw))

		for x := minX; x <= maxX; x++ {
			normX := (float64(x) - 31.5) / hw

			if y < 29 {
				c := color.RGBA{190, 225, 245, 200}
				if normX < -0.6 {
					c = glassHi
				}
				setPixel(img, x, y, c)
			} else {
				c := liquidMid
				distFromCenter := math.Abs(normX)
				if distFromCenter < 0.35 {
					c = liquidCore
				} else if distFromCenter > 0.70 {
					c = liquidDark
				}
				if y == 29 || y == 30 {
					c = liquidGlow
				}
				if normX >= -0.75 && normX <= -0.55 {
					c = glassHi
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 4. Etched Measurement Graduation Tick Lines
	for _, gy := range []int{35, 41, 47} {
		drawHLine(img, 42, 46, gy, gradLine)
	}

	// 5. Rising micro-bubbles
	drawFilledCircle(img, 24, 46, 1.5, bubbleGlow)
	drawFilledCircle(img, 37, 39, 2.0, bubbleGlow)
	drawFilledCircle(img, 30, 43, 1.2, bubbleGlow)

	addSelectiveOutline(img, darkBorder)
	saveImg(name, img)
}
