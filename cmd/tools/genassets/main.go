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

const outDir = "internal/assets/images"

func main() {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	// 1. Character Entities (16x32)
	generatePlayer("player.png")
	generateZombie("zombie.png")
	generateRunner("runner.png")

	// 2. Floor Tiles (64x32)
	generateGrass("grass.png")
	generateDirt("dirt.png")
	generateWoodFloor("wood.png")
	generateAsphalt("asphalt.png")
	generateConcrete("concrete.png")
	generateTileFloor("tile_floor.png")

	// 3. Vertical Obstacles (64x64)
	generateIsoWall("wall.png")
	generateIsoTree("tree.png")
	generateIsoFence("fence.png")
	generateIsoDebris("debris.png")

	// 4. Items & Equipment (16x16)
	generateFood("food.png")
	generateWater("water.png")
	generateWeapon("weapon.png")
	generateAxe("axe.png")
	generateShotgun("shotgun.png")
	generateAmmo("ammo.png")
	generateArmor("armor.png")

	// 5. New Style Guide Assets (64x64)
	generateIsoTent("tent.png")
	generateIsoStump("stump.png")
	generateIsoMushroom("mushroom.png")
	generateIsoSign("sign.png")
	generateElevationBlock("elevation_block.png")
	generateElevationRamp("elevation_ramp.png")

	log.Println("Asset generation completed successfully.")
}

// -------------------------------------------------------------
// DRAWING HELPERS & COLOR PRIMITIVES
// -------------------------------------------------------------

// Bounds-checked pixel setter
func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	bounds := img.Bounds()
	if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
		img.SetRGBA(x, y, c)
	}
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

// Matrix stamp renderer
func drawMatrix(img *image.RGBA, startX, startY int, rows []string, palette map[rune]color.RGBA) {
	for dy, row := range rows {
		for dx, char := range row {
			if char == '.' || char == ' ' {
				continue
			}
			if c, ok := palette[char]; ok {
				setPixel(img, startX+dx, startY+dy, c)
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
// 1. CHARACTER ENTITIES (16x32)
// -------------------------------------------------------------

func generatePlayer(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 32))

	skinLight := color.RGBA{245, 199, 161, 255}
	skinBase := color.RGBA{224, 169, 122, 255}
	skinShadow := color.RGBA{184, 123, 82, 255}

	hairLight := color.RGBA{110, 71, 42, 255}
	hairBase := color.RGBA{74, 46, 24, 255}
	hairDark := color.RGBA{43, 23, 10, 255}

	shirtLight := color.RGBA{66, 146, 184, 255}
	shirtBase := color.RGBA{42, 111, 143, 255}
	shirtDark := color.RGBA{24, 78, 104, 255}
	shirtSeam := color.RGBA{20, 56, 74, 255}

	beltLeather := color.RGBA{61, 35, 20, 255}
	beltBuckle := color.RGBA{229, 184, 52, 255}

	pantsLight := color.RGBA{75, 101, 132, 255}
	pantsBase := color.RGBA{52, 73, 94, 255}
	pantsDark := color.RGBA{36, 51, 66, 255}

	bootLight := color.RGBA{92, 62, 38, 255}
	bootBase := color.RGBA{59, 40, 24, 255}
	bootSole := color.RGBA{21, 21, 21, 255}

	eyeWhite := color.RGBA{240, 240, 240, 255}
	eyePupil := color.RGBA{30, 63, 102, 255}

	shadowGround := color.RGBA{0, 0, 0, 80}

	palette := map[rune]color.RGBA{
		'h': hairLight,
		'H': hairBase,
		'd': hairDark,
		'k': skinLight,
		'K': skinBase,
		's': skinShadow,
		'e': eyeWhite,
		'E': eyePupil,
		't': shirtLight,
		'T': shirtBase,
		'r': shirtDark,
		'c': shirtSeam,
		'B': beltLeather,
		'G': beltBuckle,
		'p': pantsLight,
		'P': pantsBase,
		'q': pantsDark,
		'f': bootLight,
		'F': bootBase,
		'O': bootSole,
		'S': shadowGround,
	}

	rows := []string{
		"................",
		".....hhhh.......",
		"....hhHHHd......",
		"....hhHHHd......",
		"....kkKKKs......",
		"....kkKKKs......",
		".....kKKs.......",
		"......Ks........",
		"...ttTTTTrr.....",
		"..ttTTTTTTrr....",
		"..ttTTTTTTrr....",
		"..ttTTTTTTrr....",
		"..ttTTTTTTrr....",
		"..kkTTTTTTrs....",
		"..kkTTTTTTrs....",
		"....BBGBB.......",
		"....ppPPPq......",
		"....ppPPPq......",
		"....pPq.pPq.....",
		"....pPq.pPq.....",
		"....pPq.pPq.....",
		"....pPq.pPq.....",
		"....pPq.pPq.....",
		"....pPq.pPq.....",
		"....pPq.pPq.....",
		"....pPq.pPq.....",
		"....fFq.fFq.....",
		"....fFq.fFq.....",
		"...ffFq.ffFq....",
		"...OOOq.OOOq....",
		"................",
		"....SSSSSSSS....",
	}

	drawMatrix(img, 0, 0, rows, palette)
	saveImg(name, img)
}

func generateZombie(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 32))

	fleshLight := color.RGBA{124, 156, 116, 255}
	fleshBase := color.RGBA{94, 122, 88, 255}
	fleshDark := color.RGBA{62, 86, 58, 255}

	bloodLight := color.RGBA{160, 20, 20, 255}
	bloodBase := color.RGBA{120, 20, 20, 255}
	bloodDark := color.RGBA{60, 8, 8, 255}

	boneWhite := color.RGBA{221, 216, 196, 255}

	eyeYellow := color.RGBA{226, 232, 116, 255}
	mouthDark := color.RGBA{26, 10, 10, 255}

	hairDark := color.RGBA{37, 37, 37, 255}

	shirtRagLight := color.RGBA{125, 117, 101, 255}
	shirtRagBase := color.RGBA{85, 78, 65, 255}
	shirtRagDark := color.RGBA{45, 40, 32, 255}

	pantsRagBase := color.RGBA{61, 66, 73, 255}
	pantsRagDark := color.RGBA{34, 37, 42, 255}

	bootBase := color.RGBA{43, 35, 29, 255}
	shadowGround := color.RGBA{0, 0, 0, 80}

	palette := map[rune]color.RGBA{
		'h': hairDark,
		'z': fleshLight,
		'Z': fleshBase,
		'x': fleshDark,
		'b': bloodLight,
		'B': bloodBase,
		'd': bloodDark,
		'W': boneWhite,
		'Y': eyeYellow,
		'M': mouthDark,
		't': shirtRagLight,
		'T': shirtRagBase,
		'r': shirtRagDark,
		'P': pantsRagBase,
		'q': pantsRagDark,
		'F': bootBase,
		'S': shadowGround,
	}

	rows := []string{
		"................",
		"......hhh.......",
		".....zzZZx......",
		".....zzZZx......",
		".....zzZZx......",
		".....zzZZx......",
		"......zZx.......",
		".......Zx.......",
		"....ttTTTrr.....",
		"..zZttTTTrrZ....",
		"..zZttTTTrrZ....",
		"..xZttTTTrrZ....",
		"..xZttTTTrrZ....",
		"....TT.TTrr.....",
		".......TTr......",
		"....PPPPPqq.....",
		"....PPPPPqq.....",
		"....PPq.Pqq.....",
		"....PPq.Pqq.....",
		"....PPq.Pqq.....",
		"....PPq.Bqq.....",
		"....PPq.zBz.....",
		"....PPq.ZZx.....",
		"....PPq.ZZx.....",
		"....PPq.ZZx.....",
		"....PPq.ZZx.....",
		"....FFq.ZZx.....",
		"....FFq.ZZx.....",
		"...FFFq.ZZZx....",
		"...FFFq.BdBB....",
		"................",
		"....SSSSSSSS....",
	}

	drawMatrix(img, 0, 0, rows, palette)
	saveImg(name, img)
}

func generateRunner(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 32))

	boneWhite := color.RGBA{216, 208, 192, 255}
	muscleHighlight := color.RGBA{255, 77, 77, 255}
	muscleArterial := color.RGBA{200, 40, 40, 255}
	muscleBase := color.RGBA{150, 28, 28, 255}
	muscleDark := color.RGBA{82, 10, 10, 255}
	muscleCrevice := color.RGBA{43, 5, 5, 255}

	eyeCore := color.RGBA{255, 158, 27, 255}
	eyeGlow := color.RGBA{255, 21, 21, 255}

	clawBlack := color.RGBA{26, 5, 5, 255}
	clawTip := color.RGBA{140, 32, 32, 255}

	shadowGround := color.RGBA{0, 0, 0, 80}

	palette := map[rune]color.RGBA{
		'W': boneWhite,
		'm': muscleHighlight,
		'a': muscleArterial,
		'M': muscleBase,
		'd': muscleDark,
		'D': muscleCrevice,
		'Y': eyeCore,
		'E': eyeGlow,
		'C': clawBlack,
		'c': clawTip,
		'S': shadowGround,
	}

	rows := []string{
		"................",
		"................",
		"......WWWW......",
		".....WWWWWW.....",
		".....mmMMMd.....",
		".....mmMMMd.....",
		".....mMMMMd.....",
		"....mmMMMMMd....",
		"...mmmMMMMMdd...",
		"..mmmmMMMMMMd...",
		"..mmmmMMMMMMd...",
		"cMmmmmMMMMMMdMc",
		"CMmmmmMMMMMMdMC",
		".CMmmmMMMMMMdC.",
		"..DmmmMMMMMMdD.",
		"..DDmm...mmDD...",
		"..Dmm.....mmD...",
		".Dmmm.....mmmD..",
		".Dmmm.....mmmD..",
		".Dmmm.....mmmD..",
		"..Dmm.....mmD...",
		"..Dmm.....mmD...",
		"..Dmm.....mmD...",
		"..Dmm.....mmD...",
		"..Dmm.....mmD...",
		"..Dmm.....mmD...",
		"..Dmm.....mmD...",
		"..cCC.....CCd...",
		"..CCC.....CCC...",
		"..CCC.....CCC...",
		"................",
		"...SSSSSSSSSS...",
	}

	drawMatrix(img, 0, 0, rows, palette)
	saveImg(name, img)
}

// -------------------------------------------------------------
// 2. FLOOR TILES (64x32)
// -------------------------------------------------------------

func generateGrass(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(42))

	baseDark := color.RGBA{38, 77, 36, 255}
	baseMid := color.RGBA{48, 98, 45, 255}
	baseLight := color.RGBA{68, 128, 55, 255}
	soilDark := color.RGBA{24, 50, 22, 255}
	flowerColor := color.RGBA{225, 210, 110, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				c := baseMid

				// 3D Extrusion - Stepped toon-shading
				if isoDist > 0.90 {
					if x < 32 {
						// Left face - lighter dirt
						c = baseDark
					} else {
						// Right face - darker dirt
						c = soilDark
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// Grass Blades - Distinct geometric V-shapes (chevrons)
	for i := 0; i < 20; i++ {
		bx := 8 + rng.Intn(48)
		by := 6 + rng.Intn(20)
		dx := float64(bx) - 31.5
		dy := float64(by) - 15.5
		if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.85 {
			setPixel(img, bx, by, baseLight)
			setPixel(img, bx-1, by-1, baseLight)
			setPixel(img, bx+1, by-1, baseLight)
		}
	}

	// Wildflower Accents - Distinct Plus Shapes
	for i := 0; i < 4; i++ {
		fx := 10 + rng.Intn(44)
		fy := 6 + rng.Intn(20)
		dx := float64(fx) - 31.5
		dy := float64(fy) - 15.5
		if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.75 {
			setPixel(img, fx, fy, flowerColor)
			setPixel(img, fx-1, fy, flowerColor)
			setPixel(img, fx+1, fy, flowerColor)
			setPixel(img, fx, fy-1, flowerColor)
			setPixel(img, fx, fy+1, flowerColor)
		}
	}

	saveImg(name, img)
}

func generateDirt(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(101))

	baseDark := color.RGBA{68, 46, 32, 255}
	baseMid := color.RGBA{92, 64, 45, 255}
	pebbleHigh := color.RGBA{165, 160, 150, 255}
	pebbleShadow := color.RGBA{40, 28, 20, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				c := baseMid

				if isoDist > 0.90 {
					if x < 32 {
						// Left face - lighter dirt
						c = baseDark
					} else {
						// Right face - darker dirt
						c = darken(baseDark, 0.75)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// Pebbles
	for i := 0; i < 10; i++ {
		px := 8 + rng.Intn(48)
		py := 4 + rng.Intn(24)
		dx := float64(px) - 31.5
		dy := float64(py) - 15.5
		if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.80 {
			// Clean flat rectangles
			setPixel(img, px, py, pebbleHigh)
			setPixel(img, px+1, py, pebbleHigh)
			setPixel(img, px, py+1, pebbleShadow)
			setPixel(img, px+1, py+1, pebbleShadow)
		}
	}

	saveImg(name, img)
}

func generateWoodFloor(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	plankColors := []color.RGBA{
		{142, 92, 54, 255},
		{128, 80, 46, 255},
		{156, 104, 62, 255},
		{136, 88, 50, 255},
	}
	seamDark := color.RGBA{45, 26, 14, 255}
	nailColor := color.RGBA{30, 22, 18, 255}
	endJoints := []float64{0.60, 0.30, 0.75, 0.45}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

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

				if vInLane < 0.05 || vInLane > 0.95 {
					c = seamDark
				}

				endU := endJoints[lane]
				if math.Abs(u-endU) < 0.025 {
					c = seamDark
				}

				if isoDist > 0.92 {
					if x < 32 {
						c = darken(c, 0.85)
					} else {
						c = darken(c, 0.70)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	for lane := 0; lane < 4; lane++ {
		endU := endJoints[lane]
		for _, offset := range []float64{-0.05, 0.05} {
			nu := endU + offset
			nv := (float64(lane) + 0.5) / 4.0
			nx := int(math.Round(31.5 + (nu-nv)*32.0))
			ny := int(math.Round(15.5 + (nu+nv-1.0)*16.0))
			if nx >= 0 && nx < w && ny >= 0 && ny < h {
				dx := float64(nx) - 31.5
				dy := float64(ny) - 15.5
				if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.85 {
					setPixel(img, nx, ny, nailColor)
				}
			}
		}
	}

	saveImg(name, img)
}

func generateAsphalt(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	baseDark := color.RGBA{38, 40, 44, 255}
	baseMid := color.RGBA{48, 50, 55, 255}
	yellowMarking := color.RGBA{220, 180, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

				c := baseMid

				// Solid clean yellow markings
				if v >= 0.43 && v <= 0.57 && (u <= 0.38 || u >= 0.62) {
					c = yellowMarking
				}

				// Flat stepped extrusions
				if isoDist > 0.92 {
					if x < 32 {
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
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	slabBase := color.RGBA{145, 145, 142, 255}
	slabLight := color.RGBA{168, 168, 165, 255}
	jointDark := color.RGBA{50, 50, 50, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

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
				if distU < 0.025 || distV < 0.025 {
					c = jointDark
				}

				// Flat stepped extrusions
				if isoDist > 0.92 {
					if x < 32 {
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
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	tileABase := color.RGBA{200, 200, 195, 255}
	tileBBase := color.RGBA{65, 75, 85, 255}
	groutDark := color.RGBA{35, 38, 42, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

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
				var c color.RGBA

				if subU < 0.05 || subV < 0.05 {
					c = groutDark
				} else {
					if isTileA {
						c = tileABase
					} else {
						c = tileBBase
					}
				}

				if isoDist > 0.92 {
					if x < 32 {
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
// 3. VERTICAL OBSTACLE BLOCKS (64x64)
// -------------------------------------------------------------

func generateIsoWall(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copingTop := color.RGBA{185, 180, 175, 255}
	copingHigh := color.RGBA{220, 215, 210, 255}
	copingShadow := color.RGBA{115, 110, 105, 255}

	brickBaseRed := color.RGBA{148, 56, 42, 255}
	brickDarkRed := color.RGBA{115, 40, 28, 255}
	brickLightRed := color.RGBA{175, 75, 55, 255}

	mortarLeft := color.RGBA{190, 185, 175, 255}
	mortarRight := color.RGBA{110, 105, 100, 255}

	// 1. Top Coping Stone
	for y := 0; y < 28; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 13.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/14.0
			if isoDist <= 1.0 {
				c := copingTop
				if isoDist > 0.88 {
					if y < 14 {
						c = copingHigh
					} else {
						c = copingShadow
					}
				}
				setPixel(img, x, y, c)
			}
		}
	}

	for x := 0; x < 32; x++ {
		edgeY := int(math.Round(14.0 + float64(x)*0.5))
		setPixel(img, x, edgeY+1, copingShadow)
	}
	for x := 32; x < 64; x++ {
		edgeY := int(math.Round(30.0 - float64(x-32)*0.5))
		setPixel(img, x, edgeY+1, darken(copingShadow, 0.8))
	}

	// 2. Left Face (West Wall)
	for x := 0; x < 32; x++ {
		topY := 15 + x/2
		botY := 47 + x/2
		for y := topY + 1; y <= botY && y < h; y++ {
			hRel := y - topY
			course := hRel / 4
			courseY := hRel % 4

			if courseY == 0 {
				setPixel(img, x, y, mortarLeft)
			} else {
				jointOffset := 0
				if course%2 == 1 {
					jointOffset = 4
				}
				if (x+jointOffset)%8 == 0 {
					setPixel(img, x, y, mortarLeft)
				} else {
					brickSeed := (course*13 + (x+jointOffset)/8*7) % 5
					c := brickBaseRed
					if brickSeed == 1 || brickSeed == 3 {
						c = brickLightRed
					} else if brickSeed == 2 {
						c = brickDarkRed
					}
					c = darken(c, 0.90)
					setPixel(img, x, y, c)
				}
			}
		}
	}

	// 3. Right Face (South Wall)
	for x := 32; x < 64; x++ {
		topY := 31 - (x-32)/2
		botY := 63 - (x-32)/2
		for y := topY + 1; y <= botY && y < h; y++ {
			hRel := y - topY
			course := hRel / 4
			courseY := hRel % 4

			if courseY == 0 {
				setPixel(img, x, y, mortarRight)
			} else {
				jointOffset := 0
				if course%2 == 1 {
					jointOffset = 4
				}
				if ((x-32)+jointOffset)%8 == 0 {
					setPixel(img, x, y, mortarRight)
				} else {
					brickSeed := (course*17 + ((x-32)+jointOffset)/8*11) % 5
					c := brickBaseRed
					if brickSeed == 1 || brickSeed == 3 {
						c = brickLightRed
					} else if brickSeed == 2 {
						c = brickDarkRed
					}
					c = darken(c, 0.65)
					setPixel(img, x, y, c)
				}
			}
		}
	}

	// Base contact shadow
	for x := 0; x < 32; x++ {
		botY := 47 + x/2
		if botY < h {
			setPixel(img, x, botY, color.RGBA{20, 15, 15, 255})
		}
	}
	for x := 32; x < 64; x++ {
		botY := 63 - (x-32)/2
		if botY < h {
			setPixel(img, x, botY, color.RGBA{15, 10, 10, 255})
		}
	}

	saveImg(name, img)
}

func generateIsoTree(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	barkDark := color.RGBA{58, 36, 20, 255}
	barkMid := color.RGBA{88, 56, 32, 255}
	barkLight := color.RGBA{118, 78, 46, 255}

	leafDeepShadow := color.RGBA{16, 42, 20, 255}
	leafMid := color.RGBA{38, 98, 44, 255}
	leafLight := color.RGBA{62, 142, 68, 255}

	// 1. Trunk and Root Flares
	for y := 42; y <= 60; y++ {
		trunkW := 3
		if y >= 54 {
			trunkW = 3 + (y - 54)
		}
		for x := 32 - trunkW; x <= 32+trunkW; x++ {
			if x >= 0 && x < w {
				dx := float64(x - 32)
				var c color.RGBA
				if dx < -1 {
					c = barkLight
				} else if dx > 1 {
					c = barkDark
				} else {
					c = barkMid
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 2. 3-Tier Layered Foliage Domes - Overlapping circles
	type Tier struct {
		cY     int
		radius float64
	}
	tiers := []Tier{
		{cY: 42, radius: 20.0},
		{cY: 26, radius: 15.0},
		{cY: 12, radius: 10.0},
	}

	for _, tier := range tiers {
		for y := tier.cY - int(tier.radius); y <= tier.cY+int(tier.radius); y++ {
			for x := int(32.0 - tier.radius); x <= int(32.0+tier.radius); x++ {
				if x < 0 || x >= w || y < 0 || y >= h {
					continue
				}
				dx := float64(x - 32)
				dy := float64(y - tier.cY)
				if dx*dx+dy*dy <= tier.radius*tier.radius {
					// Flat Shading
					var c color.RGBA
					if dx < -tier.radius*0.2 && dy < -tier.radius*0.2 {
						c = leafLight
					} else if dx > tier.radius*0.3 || dy > tier.radius*0.4 {
						c = leafDeepShadow
					} else {
						c = leafMid
					}

					setPixel(img, x, y, c)
				}
			}
		}
	}

	saveImg(name, img)
}

func generateIsoFence(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	woodLight := color.RGBA{168, 148, 125, 255}
	woodMid := color.RGBA{135, 115, 95, 255}
	woodDark := color.RGBA{85, 70, 55, 255}
	nailColor := color.RGBA{35, 30, 25, 255}

	// 1. Horizontal Rails
	for x := 2; x <= 32; x++ {
		tRailY := int(math.Round(28.0 + float64(x)*0.5))
		bRailY := int(math.Round(40.0 + float64(x)*0.5))
		for dy := 0; dy < 2; dy++ {
			setPixel(img, x, tRailY+dy, woodMid)
			setPixel(img, x, bRailY+dy, woodDark)
		}
	}

	// 2. Vertical Pickets
	picketPositions := []int{5, 9, 13, 17, 21, 25, 29}
	for _, px := range picketPositions {
		baseY := int(math.Round(46.0 + float64(px)*0.5))
		topY := baseY - 24

		setPixel(img, px+1, topY-2, woodLight)
		setPixel(img, px, topY-1, woodLight)
		setPixel(img, px+1, topY-1, woodMid)
		setPixel(img, px+2, topY-1, woodDark)

		for y := topY; y <= baseY; y++ {
			setPixel(img, px, y, woodLight)
			setPixel(img, px+1, y, woodMid)
			setPixel(img, px+2, y, woodDark)
		}

		tRailY := int(math.Round(28.0 + float64(px+1)*0.5))
		bRailY := int(math.Round(40.0 + float64(px+1)*0.5))
		setPixel(img, px+1, tRailY, nailColor)
		setPixel(img, px+1, bRailY, nailColor)
	}

	// 3. Main Corner Post
	for y := 28; y <= 60; y++ {
		setPixel(img, 30, y, woodLight)
		setPixel(img, 31, y, woodLight)
		setPixel(img, 32, y, woodMid)
		setPixel(img, 33, y, woodDark)
		setPixel(img, 34, y, woodDark)
	}
	setPixel(img, 32, 25, woodLight)
	setPixel(img, 31, 26, woodLight)
	setPixel(img, 32, 26, woodMid)
	setPixel(img, 33, 26, woodDark)
	setPixel(img, 30, 27, woodLight)
	setPixel(img, 31, 27, woodLight)
	setPixel(img, 32, 27, woodMid)
	setPixel(img, 33, 27, woodDark)
	setPixel(img, 34, 27, woodDark)

	// Left post
	for y := 16; y <= 48; y++ {
		setPixel(img, 2, y, woodLight)
		setPixel(img, 3, y, woodMid)
		setPixel(img, 4, y, woodDark)
	}
	setPixel(img, 3, 14, woodLight)
	setPixel(img, 2, 15, woodLight)
	setPixel(img, 3, 15, woodMid)
	setPixel(img, 4, 15, woodDark)

	saveImg(name, img)
}

func generateIsoDebris(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	woodTop := color.RGBA{185, 135, 85, 255}
	woodLeft := color.RGBA{150, 105, 65, 255}
	woodRight := color.RGBA{95, 65, 38, 255}
	frameWood := color.RGBA{120, 80, 48, 255}
	metalBracket := color.RGBA{75, 80, 85, 255}

	concreteLight := color.RGBA{160, 160, 155, 255}
	concreteMid := color.RGBA{125, 125, 120, 255}
	concreteDark := color.RGBA{70, 70, 68, 255}
	brickChunk := color.RGBA{140, 55, 40, 255}
	shadowColor := color.RGBA{20, 20, 20, 150}

	// 1. Drop shadow
	for y := 46; y <= 60; y++ {
		for x := 12; x <= 52; x++ {
			dx := float64(x) - 32.0
			dy := float64(y) - 53.0
			if (dx*dx)/(20.0*20.0)+(dy*dy)/(7.0*7.0) <= 1.0 {
				setPixel(img, x, y, shadowColor)
			}
		}
	}

	// 2. 3D Wooden Crate Top Face
	for y := 19; y <= 33; y++ {
		for x := 18; x <= 46; x++ {
			dx := float64(x) - 32.0
			dy := float64(y) - 26.0
			if math.Abs(dx)/14.0+math.Abs(dy)/7.0 <= 1.0 {
				c := woodTop
				if math.Abs(dx)/14.0+math.Abs(dy)/7.0 > 0.80 {
					c = frameWood
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// Left Face
	for x := 18; x <= 32; x++ {
		topY := 26 + (x-18)/2
		botY := 42 + (x-18)/2
		for y := topY; y <= botY; y++ {
			c := woodLeft
			relX := x - 18
			relY := y - topY
			if relX <= 1 || relX >= 13 || relY <= 1 || relY >= 15 {
				c = frameWood
			} else if relX == relY || relX == (16-relY) {
				c = frameWood
			}
			setPixel(img, x, y, c)
		}
	}

	// Right Face
	for x := 32; x <= 46; x++ {
		topY := 33 - (x-32)/2
		botY := 49 - (x-32)/2
		for y := topY; y <= botY; y++ {
			c := woodRight
			relX := x - 32
			relY := y - topY
			if relX <= 1 || relX >= 13 || relY <= 1 || relY >= 15 {
				c = darken(frameWood, 0.75)
			} else if relX == relY || relX == (16-relY) {
				c = darken(frameWood, 0.75)
			}
			setPixel(img, x, y, c)
		}
	}

	corners := [][2]int{
		{18, 26}, {32, 33}, {46, 26}, {32, 19},
		{18, 42}, {32, 49}, {46, 42},
	}
	for _, pt := range corners {
		setPixel(img, pt[0], pt[1], metalBracket)
		setPixel(img, pt[0]+1, pt[1], metalBracket)
		setPixel(img, pt[0], pt[1]+1, metalBracket)
	}

	// 3. Rubble Chunks
	drawRock := func(rx, ry, rw, rh int, baseC, highC, darkC color.RGBA) {
		for y := ry; y < ry+rh; y++ {
			for x := rx; x < rx+rw; x++ {
				dx := float64(x - rx - rw/2)
				dy := float64(y - ry - rh/2)
				if (dx*dx)/float64(rw*rw/4)+(dy*dy)/float64(rh*rh/4) <= 1.0 {
					var c color.RGBA
					if dy < -0.5 {
						c = highC
					} else if dy > 0.5 {
						c = darkC
					} else {
						c = baseC
					}
					setPixel(img, x, y, c)
				}
			}
		}
	}

	drawRock(8, 48, 8, 6, concreteMid, concreteLight, concreteDark)
	drawRock(46, 51, 9, 7, concreteMid, concreteLight, concreteDark)
	drawRock(14, 55, 6, 4, brickChunk, lighten(brickChunk, 1.2), darken(brickChunk, 0.7))
	drawRock(38, 54, 5, 4, brickChunk, lighten(brickChunk, 1.2), darken(brickChunk, 0.7))
	setPixel(img, 24, 56, concreteLight)
	setPixel(img, 25, 56, concreteDark)
	setPixel(img, 44, 46, concreteLight)
	setPixel(img, 45, 46, concreteDark)

	saveImg(name, img)
}

// -------------------------------------------------------------
// 4. ITEMS & EQUIPMENT (16x16)
// -------------------------------------------------------------

func generateFood(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	darkBorder := color.RGBA{45, 48, 52, 255}
	tinLid := color.RGBA{185, 190, 198, 255}
	tinLidHi := color.RGBA{235, 240, 248, 255}
	tinLidSh := color.RGBA{140, 145, 152, 255}
	tinRim := color.RGBA{210, 215, 222, 255}

	labelRed := color.RGBA{200, 38, 30, 255}
	labelRedHi := color.RGBA{240, 80, 72, 255}
	labelRedSh := color.RGBA{135, 20, 16, 255}

	labelGold := color.RGBA{242, 190, 42, 255}
	labelGoldHi := color.RGBA{255, 228, 120, 255}
	labelGoldSh := color.RGBA{175, 128, 20, 255}

	// Can Outline
	drawHLine(img, 5, 10, 2, darkBorder)
	setPixel(img, 4, 3, darkBorder)
	setPixel(img, 11, 3, darkBorder)
	setPixel(img, 3, 4, darkBorder)
	setPixel(img, 12, 4, darkBorder)
	drawVLine(img, 3, 5, 12, darkBorder)
	drawVLine(img, 12, 5, 12, darkBorder)
	drawHLine(img, 4, 11, 14, darkBorder)
	setPixel(img, 3, 13, darkBorder)
	setPixel(img, 12, 13, darkBorder)

	// Lid surface (y=3)
	drawHLine(img, 5, 10, 3, tinLid)
	setPixel(img, 5, 3, tinLidHi)
	setPixel(img, 6, 3, tinLidHi)
	setPixel(img, 10, 3, tinLidSh)

	// Pull tab (y=2..3, x=7..8)
	setPixel(img, 7, 2, color.RGBA{245, 248, 252, 255})
	setPixel(img, 8, 2, color.RGBA{190, 195, 202, 255})
	setPixel(img, 7, 3, color.RGBA{55, 60, 68, 255})
	setPixel(img, 8, 3, color.RGBA{220, 225, 232, 255})

	// Top Rim edge (y=4)
	drawHLine(img, 4, 11, 4, tinRim)
	setPixel(img, 5, 4, color.RGBA{255, 255, 255, 255})
	setPixel(img, 6, 4, color.RGBA{255, 255, 255, 255})
	setPixel(img, 10, 4, tinLidSh)
	setPixel(img, 11, 4, tinLidSh)

	// Exposed top metal strip (y=5)
	drawHLine(img, 4, 11, 5, color.RGBA{170, 176, 184, 255})
	setPixel(img, 5, 5, tinLidHi)

	// Label body (y=6..11)
	for y := 6; y <= 11; y++ {
		drawHLine(img, 4, 11, y, labelRed)
		setPixel(img, 5, y, labelRedHi)
		setPixel(img, 11, y, labelRedSh)
	}

	// Label gold band (y=8..9)
	for y := 8; y <= 9; y++ {
		drawHLine(img, 4, 11, y, labelGold)
		setPixel(img, 5, y, labelGoldHi)
		setPixel(img, 11, y, labelGoldSh)
	}
	// Bean/soup emblem in center
	setPixel(img, 7, 8, color.RGBA{45, 140, 40, 255})
	setPixel(img, 8, 9, color.RGBA{35, 115, 30, 255})

	// Exposed bottom metal strip (y=12)
	drawHLine(img, 4, 11, 12, color.RGBA{160, 166, 174, 255})
	setPixel(img, 5, 12, tinLidHi)
	setPixel(img, 11, 12, tinLidSh)

	// Bottom rim (y=13)
	drawHLine(img, 4, 11, 13, tinRim)
	setPixel(img, 5, 13, color.RGBA{245, 248, 252, 255})
	setPixel(img, 10, 13, tinLidSh)
	setPixel(img, 11, 13, tinLidSh)

	saveImg(name, img)
}

func generateWater(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	capWhite := color.RGBA{248, 250, 255, 255}
	capBlue := color.RGBA{180, 205, 235, 255}
	capSh := color.RGBA{135, 160, 195, 255}
	capBorder := color.RGBA{70, 95, 130, 255}

	glassEdge := color.RGBA{100, 150, 210, 255}
	glassEdgeDark := color.RGBA{35, 75, 140, 255}
	highlight := color.RGBA{240, 250, 255, 255}

	waterBase := color.RGBA{35, 120, 230, 255}
	waterDeep := color.RGBA{20, 70, 180, 255}
	waterLight := color.RGBA{140, 200, 255, 255}

	// White Cap (y=1..3, x=6..9)
	drawHLine(img, 6, 9, 1, capWhite)
	drawHLine(img, 6, 9, 2, capBlue)
	setPixel(img, 6, 2, capWhite)
	setPixel(img, 9, 2, capSh)
	drawHLine(img, 6, 9, 3, capBorder)

	// Neck (y=4, x=6..9)
	setPixel(img, 6, 4, glassEdge)
	setPixel(img, 7, 4, color.RGBA{210, 230, 255, 255})
	setPixel(img, 8, 4, color.RGBA{170, 205, 245, 255})
	setPixel(img, 9, 4, glassEdgeDark)

	// Shoulder (y=5, x=5..10)
	setPixel(img, 5, 5, glassEdge)
	drawHLine(img, 6, 9, 5, color.RGBA{190, 225, 255, 255})
	setPixel(img, 6, 5, highlight)
	setPixel(img, 10, 5, glassEdgeDark)

	// Meniscus / Fill line (y=6, x=4..11)
	setPixel(img, 4, 6, glassEdge)
	drawHLine(img, 5, 10, 6, waterLight)
	setPixel(img, 5, 6, highlight)
	setPixel(img, 6, 6, highlight)
	setPixel(img, 11, 6, glassEdgeDark)

	// Upper Body (y=7, x=4..11)
	setPixel(img, 4, 7, glassEdge)
	drawHLine(img, 5, 10, 7, waterBase)
	setPixel(img, 5, 7, highlight)
	setPixel(img, 10, 7, waterDeep)
	setPixel(img, 11, 7, glassEdgeDark)

	// Contoured Ergonomic Waist (y=8..9, x=5..10)
	for y := 8; y <= 9; y++ {
		setPixel(img, 5, y, glassEdge)
		drawHLine(img, 6, 9, y, waterBase)
		setPixel(img, 6, y, highlight)
		setPixel(img, 9, y, waterDeep)
		setPixel(img, 10, y, glassEdgeDark)
	}

	// Lower Body (y=10..12, x=4..11)
	for y := 10; y <= 12; y++ {
		setPixel(img, 4, y, glassEdge)
		drawHLine(img, 5, 10, y, waterBase)
		setPixel(img, 5, y, highlight)
		setPixel(img, 9, y, waterDeep)
		setPixel(img, 10, y, waterDeep)
		setPixel(img, 11, y, glassEdgeDark)
	}
	// Small air bubble
	setPixel(img, 8, 11, color.RGBA{180, 225, 255, 255})

	// Bottom Base (y=13, x=4..11)
	setPixel(img, 4, 13, glassEdgeDark)
	drawHLine(img, 5, 10, 13, waterDeep)
	setPixel(img, 5, 13, highlight)
	setPixel(img, 11, 13, glassEdgeDark)

	// Base feet / ribs (y=14, x=5..10)
	drawHLine(img, 5, 10, 14, color.RGBA{35, 75, 130, 255})
	setPixel(img, 6, 14, color.RGBA{80, 145, 215, 255})
	setPixel(img, 8, 14, color.RGBA{80, 145, 215, 255})

	saveImg(name, img)
}

func generateWeapon(name string) {
	// Spiked wooden baseball bat with grip wrap
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	darkEdge := color.RGBA{50, 30, 12, 255}
	woodSh := color.RGBA{125, 75, 30, 255}
	woodMid := color.RGBA{185, 128, 65, 255}
	woodHi := color.RGBA{228, 178, 112, 255}

	tapeBase := color.RGBA{225, 222, 210, 255}
	tapeSh := color.RGBA{140, 132, 118, 255}

	steelSpike := color.RGBA{235, 242, 250, 255}
	steelBase := color.RGBA{95, 105, 118, 255}
	blood := color.RGBA{165, 22, 22, 255}

	// Knob (x=2..3, y=13..14)
	setPixel(img, 2, 14, darkEdge)
	setPixel(img, 2, 13, woodSh)
	setPixel(img, 3, 14, woodSh)
	setPixel(img, 3, 13, woodHi)

	// Taped Grip (x=3..6, y=10..13)
	setPixel(img, 3, 12, tapeBase)
	setPixel(img, 4, 13, tapeSh)
	setPixel(img, 4, 12, tapeBase)
	setPixel(img, 4, 11, tapeSh)
	setPixel(img, 5, 12, tapeSh)
	setPixel(img, 5, 11, tapeBase)
	setPixel(img, 5, 10, tapeSh)
	setPixel(img, 6, 11, tapeSh)
	setPixel(img, 6, 10, tapeBase)

	// Throat (y=9, x=7..8)
	setPixel(img, 7, 9, woodMid)
	setPixel(img, 8, 9, woodSh)
	setPixel(img, 6, 9, darkEdge)
	setPixel(img, 7, 10, darkEdge)

	// Mid Barrel (y=6..8, x=8..11)
	setPixel(img, 7, 8, darkEdge)
	setPixel(img, 8, 8, woodMid)
	setPixel(img, 9, 8, woodSh)
	setPixel(img, 8, 7, woodHi)
	setPixel(img, 9, 7, woodMid)
	setPixel(img, 10, 7, woodSh)
	setPixel(img, 9, 6, woodHi)
	setPixel(img, 10, 6, woodMid)
	setPixel(img, 11, 6, woodSh)

	// Upper Barrel & Head (y=2..5, x=10..14)
	setPixel(img, 10, 5, darkEdge)
	setPixel(img, 11, 5, woodHi)
	setPixel(img, 12, 5, woodMid)
	setPixel(img, 13, 5, woodSh)

	setPixel(img, 11, 4, woodHi)
	setPixel(img, 12, 4, woodHi)
	setPixel(img, 13, 4, woodMid)
	setPixel(img, 14, 4, woodSh)

	setPixel(img, 12, 3, woodHi)
	setPixel(img, 13, 3, woodHi)
	setPixel(img, 14, 3, woodMid)

	// Bat Cap (y=2, x=13..14)
	setPixel(img, 13, 2, darkEdge)
	setPixel(img, 14, 2, woodMid)
	setPixel(img, 15, 2, darkEdge)

	// Protruding Spikes
	setPixel(img, 14, 1, steelSpike)
	setPixel(img, 14, 2, blood)
	setPixel(img, 15, 3, blood)

	setPixel(img, 10, 3, steelSpike)
	setPixel(img, 11, 3, steelBase)

	setPixel(img, 14, 5, steelBase)
	setPixel(img, 15, 5, steelSpike)

	setPixel(img, 8, 5, steelSpike)
	setPixel(img, 9, 5, steelBase)

	setPixel(img, 11, 8, steelBase)
	setPixel(img, 12, 8, steelSpike)

	saveImg(name, img)
}

func generateAxe(name string) {
	// Fire axe with curved handle and double-beveled red/steel head
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	darkBorder := color.RGBA{38, 35, 35, 255}
	gripRubber := color.RGBA{42, 44, 48, 255}
	gripHi := color.RGBA{85, 90, 98, 255}

	woodMid := color.RGBA{215, 150, 65, 255}
	woodHi := color.RGBA{245, 190, 115, 255}
	woodSh := color.RGBA{135, 80, 25, 255}

	axeRed := color.RGBA{218, 32, 32, 255}
	axeRedHi := color.RGBA{255, 85, 85, 255}
	axeRedSh := color.RGBA{138, 16, 16, 255}

	steelEye := color.RGBA{85, 92, 102, 255}
	steelBevel := color.RGBA{180, 192, 205, 255}
	steelEdge := color.RGBA{248, 252, 255, 255}

	// Curved Handle (y=5..14, x=2..8)
	setPixel(img, 2, 14, darkBorder)
	setPixel(img, 2, 13, gripRubber)
	setPixel(img, 3, 14, gripRubber)
	setPixel(img, 3, 13, gripHi)
	setPixel(img, 4, 14, darkBorder)

	setPixel(img, 3, 12, darkBorder)
	setPixel(img, 4, 12, woodMid)
	setPixel(img, 4, 11, woodHi)
	setPixel(img, 5, 11, woodSh)

	setPixel(img, 4, 10, woodHi)
	setPixel(img, 5, 10, woodMid)
	setPixel(img, 6, 10, darkBorder)

	setPixel(img, 5, 9, woodHi)
	setPixel(img, 6, 9, woodSh)

	setPixel(img, 5, 8, woodHi)
	setPixel(img, 6, 8, woodMid)

	setPixel(img, 6, 7, woodHi)
	setPixel(img, 7, 7, woodSh)

	setPixel(img, 6, 6, woodHi)
	setPixel(img, 7, 6, woodMid)

	setPixel(img, 7, 5, woodHi)
	setPixel(img, 8, 5, woodSh)

	// Axe Head Socket / Eye (x=7..9, y=3..5)
	fillRect(img, 7, 3, 3, 3, steelEye)
	setPixel(img, 7, 3, color.RGBA{115, 125, 138, 255})
	setPixel(img, 9, 5, color.RGBA{55, 60, 68, 255})

	// Rear Breaching Pick / Poll (x=5..6, y=3..4)
	setPixel(img, 5, 3, steelEdge)
	setPixel(img, 6, 3, steelBevel)
	setPixel(img, 5, 4, darkBorder)
	setPixel(img, 6, 4, steelEye)

	// Main Blade Body (Red Painted, x=9..13, y=1..6)
	setPixel(img, 10, 2, axeRedHi)
	setPixel(img, 11, 2, axeRedHi)
	setPixel(img, 12, 1, axeRedHi)

	for y := 3; y <= 4; y++ {
		setPixel(img, 10, y, axeRed)
		setPixel(img, 11, y, axeRed)
		setPixel(img, 12, y, axeRed)
		setPixel(img, 13, y, axeRed)
	}

	setPixel(img, 10, 5, axeRedSh)
	setPixel(img, 11, 5, axeRedSh)
	setPixel(img, 12, 6, axeRedSh)

	// Beveled Cutting Edge (x=13..15, y=1..6)
	setPixel(img, 13, 1, steelBevel)
	setPixel(img, 14, 1, steelEdge)

	setPixel(img, 13, 2, steelBevel)
	setPixel(img, 14, 2, steelEdge)

	setPixel(img, 14, 3, steelBevel)
	setPixel(img, 15, 3, steelEdge)

	setPixel(img, 14, 4, steelBevel)
	setPixel(img, 15, 4, steelEdge)

	setPixel(img, 13, 5, steelBevel)
	setPixel(img, 14, 5, steelEdge)

	setPixel(img, 13, 6, steelBevel)
	setPixel(img, 14, 6, steelEdge)

	// Outlines
	setPixel(img, 11, 1, darkBorder)
	setPixel(img, 12, 0, darkBorder)
	setPixel(img, 15, 2, darkBorder)
	setPixel(img, 15, 5, darkBorder)
	setPixel(img, 13, 7, darkBorder)
	setPixel(img, 10, 6, darkBorder)

	saveImg(name, img)
}

func generateShotgun(name string) {
	// Pump-action shotgun with wood stock, dark steel receiver, and barrel
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	darkBorder := color.RGBA{28, 30, 35, 255}
	buttRubber := color.RGBA{38, 38, 42, 255}

	woodStock := color.RGBA{145, 88, 42, 255}
	woodHi := color.RGBA{195, 128, 72, 255}
	woodSh := color.RGBA{95, 52, 22, 255}

	steelRec := color.RGBA{68, 74, 85, 255}
	steelRecHi := color.RGBA{115, 125, 140, 255}
	ejectPort := color.RGBA{25, 28, 34, 255}

	pumpWood := color.RGBA{160, 102, 50, 255}
	pumpRib := color.RGBA{78, 45, 20, 255}

	barrelHi := color.RGBA{185, 195, 210, 255}
	barrelMid := color.RGBA{85, 92, 105, 255}
	barrelSh := color.RGBA{45, 50, 60, 255}
	beadSight := color.RGBA{245, 210, 75, 255}

	// Buttpad (x=1..2, y=13..14)
	setPixel(img, 1, 14, darkBorder)
	setPixel(img, 1, 13, buttRubber)
	setPixel(img, 2, 14, buttRubber)

	// Wood Stock (x=2..5, y=10..13)
	setPixel(img, 2, 13, woodStock)
	setPixel(img, 2, 12, woodHi)
	setPixel(img, 3, 13, woodSh)
	setPixel(img, 3, 12, woodStock)
	setPixel(img, 3, 11, woodHi)
	setPixel(img, 4, 12, woodSh)
	setPixel(img, 4, 11, woodStock)
	setPixel(img, 4, 10, woodHi)
	setPixel(img, 5, 11, woodSh)
	setPixel(img, 5, 10, woodStock)

	// Trigger Guard (x=5..6, y=10..11)
	setPixel(img, 5, 11, darkBorder)
	setPixel(img, 6, 10, darkBorder)
	setPixel(img, 6, 9, color.RGBA{150, 155, 165, 255})

	// Steel Receiver (x=5..8, y=7..9)
	setPixel(img, 5, 9, steelRec)
	setPixel(img, 6, 8, steelRecHi)
	setPixel(img, 6, 9, steelRec)
	setPixel(img, 7, 7, steelRecHi)
	setPixel(img, 7, 8, ejectPort)
	setPixel(img, 7, 9, steelRec)
	setPixel(img, 8, 7, steelRecHi)
	setPixel(img, 8, 8, steelRec)

	// Pump Forend (x=8..11, y=6..7)
	setPixel(img, 8, 8, pumpWood)
	setPixel(img, 9, 7, pumpWood)
	setPixel(img, 9, 8, pumpRib)
	setPixel(img, 10, 6, pumpWood)
	setPixel(img, 10, 7, pumpRib)
	setPixel(img, 11, 6, pumpWood)

	// Dual Barrel & Mag Tube (x=9..15, y=2..6)
	for x := 9; x <= 15; x++ {
		setPixel(img, x, 3-(x-9)/3, barrelHi)
		setPixel(img, x, 4-(x-9)/3, barrelMid)
	}
	for x := 9; x <= 13; x++ {
		setPixel(img, x, 5-(x-9)/3, barrelSh)
	}

	// Muzzle Crown & Front Brass Bead Sight (x=14..15, y=1..3)
	setPixel(img, 14, 1, beadSight)
	setPixel(img, 15, 2, barrelMid)
	setPixel(img, 15, 1, darkBorder)
	setPixel(img, 15, 3, darkBorder)

	saveImg(name, img)
}

func generateAmmo(name string) {
	// Green/yellow ammunition box with brass bullet tips
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	darkBorder := color.RGBA{22, 30, 16, 255}
	boxGreen := color.RGBA{65, 88, 42, 255}
	boxGreenHi := color.RGBA{95, 126, 62, 255}
	boxGreenSh := color.RGBA{42, 58, 26, 255}

	stencilYellow := color.RGBA{238, 188, 38, 255}
	stencilYellowHi := color.RGBA{255, 218, 95, 255}
	stencilText := color.RGBA{26, 32, 20, 255}

	brassHi := color.RGBA{255, 235, 130, 255}
	brassMid := color.RGBA{235, 188, 52, 255}
	brassSh := color.RGBA{160, 120, 25, 255}
	copperTip := color.RGBA{215, 105, 45, 255}
	copperTipHi := color.RGBA{255, 160, 95, 255}

	// 4 Cartridges protruding from open top tray (x=3..12, y=2..4)
	bulletsX := []int{3, 6, 9, 12}
	for _, bx := range bulletsX {
		setPixel(img, bx, 2, copperTipHi)
		if bx+1 <= 13 {
			setPixel(img, bx+1, 2, copperTip)
		}
		setPixel(img, bx, 3, brassHi)
		setPixel(img, bx, 4, brassMid)
		if bx+1 <= 13 {
			setPixel(img, bx+1, 3, brassSh)
			setPixel(img, bx+1, 4, brassSh)
		}
	}

	// Box Outline & Top Bevel (y=5)
	drawHLine(img, 2, 13, 5, boxGreenHi)
	setPixel(img, 2, 5, darkBorder)
	setPixel(img, 13, 5, darkBorder)

	// Box Sides and Body (y=6..13)
	for y := 6; y <= 13; y++ {
		drawHLine(img, 3, 12, y, boxGreen)
		setPixel(img, 2, y, darkBorder)
		setPixel(img, 3, y, boxGreenHi)
		setPixel(img, 12, y, boxGreenSh)
		setPixel(img, 13, y, darkBorder)
	}

	// Yellow Stencil Band (y=8..9, x=4..11)
	drawHLine(img, 4, 11, 8, stencilYellowHi)
	drawHLine(img, 4, 11, 9, stencilYellow)
	setPixel(img, 5, 8, stencilText)
	setPixel(img, 7, 8, stencilText)
	setPixel(img, 8, 9, stencilText)
	setPixel(img, 10, 8, stencilText)

	// Corner Rivets
	setPixel(img, 3, 6, color.RGBA{170, 180, 175, 255})
	setPixel(img, 12, 6, color.RGBA{140, 150, 145, 255})
	setPixel(img, 3, 13, color.RGBA{170, 180, 175, 255})
	setPixel(img, 12, 13, color.RGBA{140, 150, 145, 255})

	// Box Bottom Rim (y=14)
	drawHLine(img, 2, 13, 14, darkBorder)

	saveImg(name, img)
}

func generateArmor(name string) {
	// Tactical ballistic Kevlar armor vest with neck opening, shoulder straps, plate carrier pouches
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	darkBorder := color.RGBA{18, 20, 25, 255}
	vestKevlar := color.RGBA{48, 54, 65, 255}
	vestHi := color.RGBA{78, 86, 102, 255}
	vestSh := color.RGBA{30, 34, 42, 255}

	strapHi := color.RGBA{88, 96, 112, 255}
	buckleSteel := color.RGBA{140, 148, 162, 255}

	idPatch := color.RGBA{92, 88, 74, 255}
	idPatchHi := color.RGBA{120, 115, 98, 255}

	molleWeb := color.RGBA{28, 32, 38, 255}
	pouchFlap := color.RGBA{64, 72, 86, 255}
	pouchBody := color.RGBA{38, 44, 54, 255}
	pullTab := color.RGBA{95, 105, 122, 255}

	// Shoulder Straps (y=2..4)
	drawHLine(img, 3, 5, 2, strapHi)
	drawHLine(img, 3, 5, 3, vestKevlar)
	setPixel(img, 3, 3, buckleSteel)
	drawHLine(img, 3, 5, 4, vestKevlar)
	setPixel(img, 2, 2, darkBorder)
	setPixel(img, 2, 3, darkBorder)
	setPixel(img, 2, 4, darkBorder)

	drawHLine(img, 10, 12, 2, strapHi)
	drawHLine(img, 10, 12, 3, vestKevlar)
	setPixel(img, 12, 3, buckleSteel)
	drawHLine(img, 10, 12, 4, vestKevlar)
	setPixel(img, 13, 2, darkBorder)
	setPixel(img, 13, 3, darkBorder)
	setPixel(img, 13, 4, darkBorder)

	// Scooped Neck Opening (x=6..9, y=2..4)
	drawHLine(img, 6, 9, 2, darkBorder)
	drawHLine(img, 6, 9, 3, color.RGBA{22, 25, 30, 255})
	drawHLine(img, 6, 9, 4, molleWeb)

	// Chest Plate Body (y=5..8, x=2..13)
	drawHLine(img, 3, 12, 5, vestKevlar)
	setPixel(img, 2, 5, darkBorder)
	setPixel(img, 13, 5, darkBorder)

	for y := 6; y <= 8; y++ {
		drawHLine(img, 3, 12, y, vestKevlar)
		setPixel(img, 2, y, darkBorder)
		setPixel(img, 3, y, vestHi)
		setPixel(img, 12, y, vestSh)
		setPixel(img, 13, y, darkBorder)
	}

	// Velcro ID Patch on Upper Chest (x=6..9, y=5..6)
	drawHLine(img, 6, 9, 5, idPatchHi)
	drawHLine(img, 6, 9, 6, idPatch)

	// Molle Webbing Straps (y=8, y=10)
	drawHLine(img, 3, 12, 8, molleWeb)
	drawHLine(img, 3, 12, 10, molleWeb)

	// 3 Mag Pouches (y=9..12)
	drawHLine(img, 3, 5, 9, pouchFlap)
	setPixel(img, 4, 9, pullTab)
	for y := 10; y <= 12; y++ {
		drawHLine(img, 3, 5, y, pouchBody)
		setPixel(img, 3, y, vestHi)
	}

	drawHLine(img, 6, 9, 9, pouchFlap)
	setPixel(img, 7, 9, pullTab)
	setPixel(img, 8, 9, pullTab)
	for y := 10; y <= 12; y++ {
		drawHLine(img, 6, 9, y, pouchBody)
	}

	drawHLine(img, 10, 12, 9, pouchFlap)
	setPixel(img, 11, 9, pullTab)
	for y := 10; y <= 12; y++ {
		drawHLine(img, 10, 12, y, pouchBody)
		setPixel(img, 12, y, vestSh)
	}

	// Bottom Waistband Hem (y=13)
	drawHLine(img, 3, 12, 13, vestSh)
	setPixel(img, 2, 13, darkBorder)
	setPixel(img, 13, 13, darkBorder)

	// Bottom Outline (y=14)
	drawHLine(img, 3, 12, 14, darkBorder)

	saveImg(name, img)
}

func generateIsoTent(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	tentGreenLight := color.RGBA{80, 120, 60, 255}
	tentGreenDark := color.RGBA{40, 70, 30, 255}
	tentPole := color.RGBA{180, 180, 180, 255}
	tentDark := color.RGBA{20, 35, 15, 255}

	// Tent Left Face (Triangle)
	for y := 16; y < 48; y++ {
		for x := 16; x < 32; x++ {
			if y > 16+(32-x)*2 && y < 48-(32-x) {
				setPixel(img, x, y, tentGreenLight)
			}
		}
	}
	// Tent Right Face (Triangle)
	for y := 16; y < 48; y++ {
		for x := 32; x < 48; x++ {
			if y > 16+(x-32)*2 && y < 48-(x-32)/2 {
				setPixel(img, x, y, tentGreenDark)
			}
		}
	}
	// Add some details
	setPixel(img, 32, 16, tentPole) // top pole

	// Opening
	for y := 32; y < 48; y++ {
		for x := 28; x < 36; x++ {
			if y > 32+(x-28) && y > 32+(36-x) {
				setPixel(img, x, y, tentDark)
			}
		}
	}

	saveImg(name, img)
}

func generateIsoStump(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bark := color.RGBA{88, 56, 32, 255}
	woodTop := color.RGBA{185, 135, 85, 255}
	barkDark := color.RGBA{58, 36, 20, 255}

	// Base
	for y := 36; y < 44; y++ {
		for x := 24; x < 40; x++ {
			if (x-32)*(x-32)+(y-40)*(y-40)*4 < 64 {
				setPixel(img, x, y, barkDark)
			}
		}
	}

	// Top cut
	for y := 30; y < 38; y++ {
		for x := 26; x < 38; x++ {
			if (x-32)*(x-32)+(y-34)*(y-34)*4 < 36 {
				setPixel(img, x, y, woodTop)
			} else if (x-32)*(x-32)+(y-34)*(y-34)*4 < 49 {
				setPixel(img, x, y, bark)
			}
		}
	}

	saveImg(name, img)
}

func generateIsoMushroom(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	stem := color.RGBA{220, 210, 190, 255}
	capCol := color.RGBA{180, 40, 40, 255}
	spot := color.RGBA{240, 240, 240, 255}

	// Stem
	for y := 36; y < 44; y++ {
		for x := 30; x < 34; x++ {
			setPixel(img, x, y, stem)
		}
	}

	// Cap
	for y := 30; y < 38; y++ {
		for x := 24; x < 40; x++ {
			if (x-32)*(x-32)/2+(y-34)*(y-34)*2 < 16 {
				setPixel(img, x, y, capCol)
				if (x+y)%4 == 0 {
					setPixel(img, x, y, spot)
				}
			}
		}
	}

	saveImg(name, img)
}

func generateIsoSign(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	post := color.RGBA{100, 70, 40, 255}
	board := color.RGBA{150, 110, 70, 255}
	text := color.RGBA{40, 30, 20, 255}

	// Post
	for y := 24; y < 48; y++ {
		for x := 30; x < 34; x++ {
			setPixel(img, x, y, post)
		}
	}

	// Board
	for y := 16; y < 28; y++ {
		for x := 16; x < 48; x++ {
			setPixel(img, x, y, board)
			// Mock text
			if y > 18 && y < 26 && x > 20 && x < 44 && (x+y)%3 != 0 {
				setPixel(img, x, y, text)
			}
		}
	}

	saveImg(name, img)
}

func generateElevationBlock(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	top := color.RGBA{120, 160, 80, 255} // Grass top
	left := color.RGBA{100, 80, 60, 255} // Dirt left
	right := color.RGBA{70, 50, 40, 255} // Dirt right

	// A standard 64x32 iso tile raised by 32px
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			// Top face (y=0..32, center 32,16)
			if y >= 0 && y <= 32 {
				dy := y - 16
				dx := x - 32
				if dx < 0 {
					dx = -dx
				}
				if dy < 0 {
					dy = -dy
				}
				if dx+dy*2 <= 32 {
					setPixel(img, x, y, top)
				}
			}
			// Left face (x=0..32, y=16..48)
			if x >= 0 && x <= 32 && y > 16+x/2 && y <= 48+x/2 {
				setPixel(img, x, y, left)
			}
			// Right face (x=32..64, y=32..64)
			if x >= 32 && x <= 64 && y > 32-(x-32)/2 && y <= 64-(x-32)/2 {
				setPixel(img, x, y, right)
			}
		}
	}

	saveImg(name, img)
}

func generateElevationRamp(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rampSurf := color.RGBA{140, 180, 90, 255} // Grass ramp
	side := color.RGBA{70, 50, 40, 255}       // Dirt side

	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			// Slope rising from bottom-left to top-right
			// We can fake an isometric ramp
			if x >= 0 && x <= 64 && y >= x/2 && y <= 32+x/2 {
				setPixel(img, x, y, rampSurf)
			}
			if x >= 32 && x <= 64 && y > 32+(x-32)/2 && y <= 64 {
				setPixel(img, x, y, side) // fake side
			}
		}
	}

	saveImg(name, img)
}
