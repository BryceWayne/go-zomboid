# Handoff Report: Character Procedural Sprites (Milestone 1)

**Agent**: `teamwork_preview_explorer_m1_1`  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_1`  
**Target Milestone**: Milestone 1 - Procedural Sprite Enhancements (Character Entities)  

---

## 1. Observation

1. **Current Entity Generation in `cmd/tools/genassets/main.go:43-57`**:
   ```go
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
   ```
   Currently, `player.png` (solid bright green), `zombie.png` (solid dark green), and `runner.png` (solid red) are flat 16x32 rectangular blocks with four arbitrary black dots.

2. **Asset Embedding & Loading in `internal/assets/assets.go:17-42`**:
   - `var PlayerImage, ZombieImage, RunnerImage *ebiten.Image`
   - `Load()` decodes embedded PNG files from `internal/assets/images/*.png`.

3. **Rendering & Projection in `internal/game/game.go:754-788`**:
   - Character rendering translates sprites at:
     ```go
     drawX := isoX - 8 - camX
     drawY := isoY - 32 - camY
     ```
   - This anchors the 16x32 entity sprite with its bottom-center at `(isoX, isoY)` corresponding to the entity's foot placement on the 64x32 isometric ground tile.

4. **Project Requirements**:
   - `PROJECT.md:13,32,52`: Milestone 1 requires anatomical pixel-art character sprites for `player.png`, `zombie.png`, and `runner.png` with clothes, hair, wounds, muscle tone, and directional shading.
   - `ORIGINAL_REQUEST.md:16,20-22`: Asset generation must be implemented in pure Go within `cmd/tools/genassets` without external dependencies.

---

## 2. Logic Chain

1. **Spatial Constraint Analysis (16x32 Resolution)**:
   - Width = 16 pixels ($x \in [0, 15]$)
   - Height = 32 pixels ($y \in [0, 31]$)
   - Anchor point is $(8, 31)$ (bottom center).
   - Anatomical proportions in 16x32 pixel art:
     - **Head & Hair**: $y = 1 \dots 7$ (7 pixels high)
     - **Neck & Collar**: $y = 7 \dots 8$ (1-2 pixels)
     - **Torso & Arms**: $y = 8 \dots 16$ (9 pixels high, shoulders span $x = 3 \dots 12$)
     - **Belt & Waistline**: $y = 16 \dots 17$ (1-2 pixels)
     - **Legs & Pants**: $y = 17 \dots 26$ (10 pixels high, left leg $x = 4 \dots 7$, right leg $x = 8 \dots 11$)
     - **Boots / Feet**: $y = 27 \dots 30$ (4 pixels high, tread $x = 3 \dots 12$)
     - **Ground Contact Shadow**: $y = 31$ (subtle 8-pixel wide semi-transparent shadow)

2. **Character Aesthetic & Anatomical Design**:

   ### A. `player.png` — Human Survivor
   - **Head/Face**: Natural peachy skin tone (`#E0A97A`), dark brown parted hair with crown highlight (`#6E472A`), distinct eyes with white sclera (`#F0F0F0`) and navy pupils (`#1E3F66`), nose bridge, and jawline.
   - **Torso/Shirt**: Flannel/utility shirt in teal/slate blue (`#2A6F8F`) with top-left highlight (`#4292B8`), collar lapel, central button seam, and rolled-up sleeves showing skin forearms and hands at sides ($x = 2 \dots 3$ and $x = 12 \dots 13$).
   - **Waist**: Dark leather belt (`#3D2314`) with a polished brass/gold buckle (`#E5B834`) at $x = 7 \dots 8$.
   - **Legs/Pants**: Sturdy denim jeans (`#34495E`) with knee highlights (`#4B6584`), inner seam shadows, and cuff folds.
   - **Boots**: Tough leather combat boots (`#3B2818`) with reinforced soles (`#151515`).
   - **Lighting**: Consistent top-left directional lighting (left side + top edges highlighted, right side + bottom shaded).

   ### B. `zombie.png` — Decayed Shambler
   - **Posture**: Asymmetrical, limping shambler stance. Head tilted slightly to the right ($x+1$).
   - **Flesh & Head**: Sickly rotting green/grey skin (`#5E7A58`), dark sunken eye sockets with milky/yellow dead eyes (`#E2E874`), exposed decaying jaw with yellowed teeth (`#DCD4B8`), matted dark hair patches (`#252525`).
   - **Torso & Wounds**: Tattered, blood-splattered faded shirt (`#7D7565`). Open chest wound at $x = 7 \dots 9, y = 10 \dots 12$ exposing 3 ivory rib bones (`#DDD8C4`) surrounded by coagulated crimson blood (`#781414`).
   - **Arms**: Asymmetric poses — left arm dangles limply at side ($x = 1 \dots 3, y = 9 \dots 15$), right arm raised forward in reaching/clawing attack pose ($x = 12 \dots 15, y = 8 \dots 12$).
   - **Legs & Feet**: Left leg wearing frayed trousers down to a worn boot; right pant leg torn off at knee ($y = 21$), exposing bloody wounded kneecap (`#8A2020`) and bare rotting green leg down to a clawed decaying foot.

   ### C. `runner.png` — Feral Crimson Mutated Zombie
   - **Posture**: Low-slung, hunched aggressive predator stance.
   - **Flesh & Anatomy**: Peeled scalp with exposed cranium bone (`#B8AFA0`), raw sinewy crimson muscle fibers (`#961C1C` base, `#C82828` arterial highlight, `#520A0A` deep crevice shadow).
   - **Face**: Piercing glowing red/orange eyes (`#FF1515` eye, `#FF9E1B` core) that flare in the dark, gaping predator maw with sharp fangs.
   - **Spine & Back**: Exposed ivory vertebrae ridge (`#D8D0C0`) protruding along the hunched spine ($x = 6 \dots 7, y = 7 \dots 12$).
   - **Limbs**: Elongated muscular arms lunging forward ending in razor-sharp obsidian talons (`#1A0505`). Heavy coiled sprinter thighs and tensed calves on digitigrade clawed feet.

3. **Helper Drawing Primitives Architecture**:
   To make sprite generation clean, robust, and maintainable, `cmd/tools/genassets/main.go` should incorporate a dedicated suite of pixel-art drawing primitives:
   - `setPixel(img *image.RGBA, x, y int, c color.RGBA)`: Bounds-checked pixel setter.
   - `fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA)`: Fills a rectangular block.
   - `drawHLine(img *image.RGBA, x1, x2, y int, c color.RGBA)` & `drawVLine(img *image.RGBA, x, y1, y2 int, c color.RGBA)`.
   - `drawShadedRect(img *image.RGBA, x, y, w, h int, base, highlight, shadow color.RGBA)`: Fills rect with top-left highlight and bottom-right shadow.
   - `darken(c color.RGBA, factor float64) color.RGBA` & `lighten(c color.RGBA, factor float64) color.RGBA`.
   - `blend(c1, c2 color.RGBA, t float64) color.RGBA`.
   - `drawMatrix(img *image.RGBA, startX, startY int, rows []string, palette map[rune]color.RGBA)`: Renders matrix templates.
   - `addSelectiveOutline(img *image.RGBA, outlineColor color.RGBA)`: Adds 1px darkened boundary contour around non-transparent pixels so sprites read clearly against all tile backgrounds.

---

## 3. Recommended Code Specification for Implementation

### A. Drawing Primitives & Helper Code

```go
package main

import (
	"image"
	"image/color"
	"math"
)

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

	// Copy existing pixels
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			temp.SetRGBA(x, y, img.RGBAAt(x, y))
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if temp.RGBAAt(x, y).A == 0 {
				// Check 4-way neighbors
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
```

---

### B. `generatePlayer("player.png")` Specification

```go
func generatePlayer(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 32))

	// Palettes
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
		"................", // 0
		".....hhhhh......", // 1 Hair crown
		"....hHHHHHd.....", // 2 Hair top
		"....HHHHHHHd....", // 3 Forehead
		"....HkEeKkEed...", // 4 Eyes
		"....sKKKKKKs....", // 5 Cheeks & nose
		".....sKKKKs.....", // 6 Mouth & chin
		"......sKKs......", // 7 Neck
		"...tTTTTTTTT....", // 8 Shoulders
		"..ttTTTTTTTTr...", // 9 Upper chest
		"..ttTTcTTTTTr...", // 10 Plaid seam & pocket
		"..ttTTcTTTTTr...", // 11 Mid torso
		"..KKTTcTTTTTr...", // 12 Forearm & shirt
		"..KKTTcTTTTKK...", // 13 Hands & shirt hem
		"..ssTTTTTTTss...", // 14 Hands lower
		"....BBBBGBBB....", // 15 Belt with buckle
		"....pPPPPPPq....", // 16 Hips
		"....pPPPPPPq....", // 17 Upper crotch
		"....pPPq.pPPq...", // 18 Thigh split
		"....pPPq.pPPq...", // 19 Thighs
		"....pPPq.pPPq...", // 20 Above knees
		"....pPPq.pPPq...", // 21 Knees
		"....pPPq.pPPq...", // 22 Below knees
		"....pPPq.pPPq...", // 23 Shins
		"....pPPq.pPPq...", // 24 Calves
		"....pPPq.pPPq...", // 25 Ankles
		"....fFFq.fFFq...", // 26 Boot tops
		"....fFFq.fFFq...", // 27 Boot ankles
		"...ffFFq.ffFFq..", // 28 Boot feet
		"...OOOOq.OOOOq..", // 29 Soles
		"...OOOOq.OOOOq..", // 30 Soles bottom
		"....SSSSSSSS....", // 31 Ground shadow
	}

	drawMatrix(img, 0, 0, rows, palette)
	addSelectiveOutline(img, color.RGBA{26, 24, 28, 255})
	saveImg(name, img)
}
```

---

### C. `generateZombie("zombie.png")` Specification

```go
func generateZombie(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 32))

	// Palettes
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
		"................", // 0
		"......h.h.h.....", // 1 Ragged hair tufts
		".....hhhhhh.....", // 2 Scalp
		".....zZZZZZx....", // 3 Rotting forehead
		".....ZYZdZYx....", // 4 Feral yellow eyes
		".....ZZZZZZx....", // 5 Rotting face
		"......xMWdZx....", // 6 Open jaw with exposed tooth
		"......xZZZx.....", // 7 Neck
		"....tTTTTTTTr...", // 8 Shoulders
		"..zZtTTTTTTTrZ..", // 9 Left limp arm / Right reach arm
		"..zZtTTBWWdTrZZ.", // 10 Exposed rib bone & blood
		"..xZtTTBWWdTr.ZZ", // 11 Ribcage gore & reaching hand
		"..xZtTTBddTTr...", // 12 Torn bloody shirt hem
		"..xxTT.TTrrr....", // 13 Ragged hem fringe
		"..BB...TTr......", // 14 Dangling bloody claw
		"....PPPPPPqq....", // 15 Torn pants waist
		"....PPPPPPqq....", // 16 Hips
		"....PPPP.Pqq....", // 17 Leg split
		"....PPPP.Pqq....", // 18 Upper thighs
		"....PPPP.Pqq....", // 19 Mid thighs
		"....PPPP.Bqq....", // 20 Right pants torn off here
		"....PPPP.zBz....", // 21 Right knee: exposed bloody gore
		"....PPPP.ZZx....", // 22 Left pants continue; Right bare rotting leg
		"....PPPP.ZZx....", // 23 Left pants / Right green shin
		"....PPPP.ZZx....", // 24 Left pants / Right green shin
		"....PPPP.ZZx....", // 25 Left pants cuff
		"....FFFF.ZZx....", // 26 Left boot / Right rotting ankle
		"....FFFF.ZZx....", // 27 Left boot
		"...FFFFF.ZZZx...", // 28 Left boot foot / Right bare decayed foot
		"...FFFFF.BdBB...", // 29 Left boot sole / Right bloody clawed toes
		"...FFFFF.dddd...", // 30 Soles & gore
		"....SSSSSSSS....", // 31 Ground shadow
	}

	drawMatrix(img, 0, 0, rows, palette)
	addSelectiveOutline(img, color.RGBA{20, 24, 18, 255})
	saveImg(name, img)
}
```

---

### D. `generateRunner("runner.png")` Specification

```go
func generateRunner(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 32))

	// Palettes
	boneWhite := color.RGBA{216, 208, 192, 255}
	boneShadow := color.RGBA{138, 130, 115, 255}

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
		'w': boneShadow,
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
		"................", // 0
		"................", // 1
		"......WWWW......", // 2 Exposed cranium
		".....WWWWWW.....", // 3 Skull plate
		".....ammMmMd....", // 4 Striated facial muscle
		".....mEYaEYm....", // 5 Glowing burning red/orange eyes
		".....aMDDDMa....", // 6 Snarling predatory maw
		"....amMWWMmad...", // 7 Spine ridge begins on hunched neck
		"...ammMWWMmmad..", // 8 Hunched upper back & shoulders
		"..aMmmMWWMmmmMD.", // 9 Muscular spine furrow
		"..MmmamWWmammmD.", // 10 Striated back & lunging shoulders
		"cCMmmamWWmammmMC", // 11 Extended razor claw talons
		"CCMmmamDDmammmCC", // 12 Claws & tensed flanks
		".CDmmammmmammmD.", // 13 Forearms
		"..DmmammmmammmD.", // 14 Striated waist
		"..DDmma..ammDD..", // 15 Coiled hip split
		"..Dmmaa..aammD..", // 16 Sprinter quadriceps
		".Dmmmaa..aammmD.", // 17 Sprinter quadriceps
		".Dmmmaa..aammmD.", // 18 Muscular thigh bulk
		".Dmmmaa..aammmD.", // 19 Coiled flexed knees
		"..Dmma....ammD..", // 20 Knee joints
		"..Dmma....ammD..", // 21 Upper tensed calves
		"..Dmma....ammD..", // 22 Calves
		"..Dmma....ammD..", // 23 Lower calves / Achilles tendons
		"..Dmma....ammD..", // 24 Digitigrade ankle transition
		"..Dmma....ammD..", // 25 Ankle transition
		"..Dmma....ammD..", // 26 Foot arch
		"..cCCd....cCCd..", // 27 Clawed feet
		"..CCCC....CCCC..", // 28 Talons gripping ground
		"..CCCC....CCCC..", // 29 Talons
		"..cCCc....cCCc..", // 30 Claw tips
		"...SSSSSSSSSS...", // 31 Ground shadow
	}

	drawMatrix(img, 0, 0, rows, palette)
	addSelectiveOutline(img, color.RGBA{30, 4, 4, 255})
	saveImg(name, img)
}
```

---

## 4. Caveats

1. **Static Sprites**: The current engine uses single 16x32 frame sprites without multi-frame walk-cycle animation. The proposed layout uses a balanced, dynamic 3/4 isometric stance that renders naturally while moving and standing still.
2. **Armor Visual Layering in M3**: When Armor items are implemented in Milestone 3, the torso coordinates identified here ($x \in [3, 12], y \in [8, 16]$) provide the exact target area for dynamic armor overlays or chest tinting.
3. **No External Libraries**: All code uses standard Go `image`, `image/color`, `image/png`, conforming strictly to project constraints.

---

## 5. Conclusion

- The character entity sprite architecture has been fully formulated with exact pixel matrices, rich retro color palettes, anatomical zones, lighting models, and reusable pure Go drawing primitives.
- The developer assigned to Milestone 1 can directly integrate these algorithms into `cmd/tools/genassets/main.go`.

---

## 6. Verification Method

1. **Asset Generation Verification**:
   ```bash
   go run ./cmd/tools/genassets
   ```
   Inspect generated PNGs in `internal/assets/images/` (`player.png`, `zombie.png`, `runner.png`) to ensure they are 16x32 RGBA PNGs.
2. **Unit Test Suite**:
   ```bash
   CC=gcc go test ./...
   ```
   Ensure all tests compile and pass.
3. **Interactive & Game Loop Launch**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   Verify that entities display correctly in the isometric world view with proper depth sorting and no graphical artifacts.
