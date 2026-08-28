# Milestone 1 - Item, Weapon, Armor Sprites and `internal/assets` Integration Handoff Report

## 1. Observation

### Codebase and Architecture Inspection
1. **Existing Asset Generation (`cmd/tools/genassets/main.go`)**:
   - Lines 22–24:
     ```go
     generateWeapon("weapon.png")
     generateItem("food.png", color.RGBA{255, 165, 0, 255}) // Orange food
     generateItem("water.png", color.RGBA{0, 191, 255, 255}) // Blue water
     ```
   - Current item generator `generateItem` (lines 33–41) merely draws a flat 8x8 colored square inside a 16x16 canvas:
     ```go
     func generateItem(name string, c color.RGBA) {
         img := image.NewRGBA(image.Rect(0, 0, 16, 16))
         for x := 4; x < 12; x++ {
             for y := 4; y < 12; y++ {
                 img.Set(x, y, c)
             }
         }
         saveImg(name, img)
     }
     ```
   - Current weapon generator `generateWeapon` (lines 208–217) draws a single diagonal brown line without shading, spikes, grip wrap, or details.
   - Missing item generators in `cmd/tools/genassets`: `axe.png`, `shotgun.png`, `ammo.png`, `armor.png`.
   - Missing tile generators in `cmd/tools/genassets`: `asphalt.png`, `concrete.png`, `tile_floor.png`, `fence.png`, `debris.png`.

2. **Existing Asset Exporters (`internal/assets/assets.go`)**:
   - Lines 16–28:
     ```go
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
     ```
   - Lines 30–42: `Load()` only loads the 11 legacy images.
   - Currently missing 9 exported variables specified in `PROJECT.md` §Interface Contracts:
     - Items & Weapons: `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`
     - Environment Tiles: `AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `FenceImage`, `DebrisImage`

3. **Requirement Reference**:
   - `ORIGINAL_REQUEST.md` §R1 & §R2: Procedural image generation in `cmd/tools/genassets` without external dependencies. Expand weapons (axe, shotgun) and armor system.
   - `PROJECT.md` Milestone 1: Rich pixel-art generation for 16x16 items (`food.png`, `water.png`, `weapon.png`, `axe.png`, `shotgun.png`, `ammo.png`, `armor.png`) and registration in `internal/assets/assets.go`.

---

## 2. Logic Chain

1. **Pixel Art Readability & Ground Contrast**:
   - In 2.5D isometric view, 16x16 items are dropped on varied floor tiles (vibrant green grass, brown dirt, gray asphalt/concrete, interior wood/tile floors).
   - Flat squares or unbordered sprites blend into the terrain and are hard to distinguish.
   - Therefore, all 16x16 items require:
     - A dark outer perimeter outline (1px dark border) for maximum contrast against any terrain or UI inventory background.
     - Directional lighting highlights (top/left bright edges, bottom/right shadow tones) to establish 3D volumetric depth.
     - Distinct thematic silhouettes (e.g. cylindrical can with pull tab, contoured hourglass bottle with meniscus, diagonal spiked club with medical grip tape wrap, red double-beveled fire axe with rear pick, blued-steel shotgun with walnut stock and brass front bead, olive ammo box with brass bullet tips, dark ballistic Kevlar plate carrier vest with shoulder straps and pouches).

2. **Item-by-Item Procedural Algorithms**:
   - **`food.png`**:
     - Cylindrical tin can (x: 3..12, y: 2..14).
     - Silver metallic top rim & lid (y=2..4) with a metallic pull tab ring at (7..8, 2..3).
     - Exposed tin bands at top and bottom (y=5 and y=12).
     - Classic red soup/canned label (y=6..11) with a horizontal gold band (y=8..9) and centered emblem.
     - Left cylindrical specular highlight strip (x=5) and right shade (x=11).
   - **`water.png`**:
     - Plastic water bottle (x: 4..11, y: 1..14).
     - White molded screw cap (y=1..3, x=6..9) with neck seal ring.
     - Transparent plastic neck & tapered shoulders (y=4..5).
     - Water meniscus reflection line (y=6).
     - Ergonomic contoured waist (y=8..9 narrower by 1px on left/right).
     - Vivid mineral blue water fill with vertical specular reflection on the left (x=5) and air bubble at (8, 11).
     - Base ribs (y=14).
   - **`weapon.png` (Spiked Baseball Bat)**:
     - Diagonally angled from handle knob (2, 14) to barrel tip (14, 2).
     - Textured off-white athletic tape grip wrap (x=3..6, y=10..13) with diagonal groove lines.
     - Tapered hickory/ash wood body with 3-tone shading (highlight, core grain, shadow edge).
     - Protruding forged steel nails/spikes poking through the barrel at multiple angles with sharp metallic highlights.
     - Zombie blood spatter accent on the top spike and cap.
   - **`axe.png` (Fire Axe)**:
     - Ergonomic curved handle from rubberized base grip (2..3, 13..15) up to collar (8, 5).
     - Forged steel collar/eye (7..9, 3..5).
     - Pointed steel breaching pick / poll (5..6, 3..4) for breaching doors.
     - Fire-engine high-visibility red blade flare (x=10..13, y=1..6).
     - Double-beveled razor-sharp steel cutting bit (x=13..15, y=1..6) with mirror highlight.
   - **`shotgun.png` (Pump-Action Shotgun)**:
     - Full-profile diagonal layout from rubber buttpad (1, 14) to muzzle (15, 2).
     - Contoured walnut wood buttstock and wrist grip (x=2..5, y=10..13).
     - Dark matte blued-steel receiver (x=5..8, y=7..9) with ejection port recess and trigger guard.
     - Ribbed wood/polymer pump forend slide (x=8..11, y=6..7).
     - Dual-tube assembly: top smoothbore 12GA barrel (x=9..15) with specular line and lower tubular magazine (x=9..13).
     - Front brass bead sight at (14, 1).
   - **`ammo.png` (Ammunition Box)**:
     - Heavy-duty military olive-drab box casing (x: 2..13, y: 5..14).
     - Open top tray revealing 4 protruding cartridges (x=3, 6, 9, 12, y=2..4) with copper FMJ bullet tips and bright brass cases.
     - Box body with top bevel, reinforced corner rivets, and bright yellow/gold commercial stencil band (y=8..9) with caliber text stippling.
   - **`armor.png` (Tactical Ballistic Kevlar Vest)**:
     - Tactical plate carrier vest (x: 2..13, y: 2..14).
     - Scooped neck cavity cutout (x=6..9, y=2..4) with dark inner shadow and reinforced collar piping.
     - Padded shoulder straps (x=3..5 and x=10..12, y=2..4) with steel quick-release buckles.
     - Angular ballistic nylon Kevlar strike-face (x=2..13, y=5..8) with chest velcro ID patch panel (x=6..9, y=5..6).
     - Horizontal Molle webbing straps (y=8, y=10) and 3 tactical magazine/utility pouches (y=9..12) with retention flaps and pull tabs.
     - Reinforced waistband hem and dark perimeter outline.

3. **`internal/assets/assets.go` Export Integration**:
   - Must export 20 global `*ebiten.Image` handles:
     - Existing: `PlayerImage`, `ZombieImage`, `RunnerImage`, `GrassImage`, `DirtImage`, `WoodImage`, `WallImage`, `TreeImage`, `WeaponImage`, `FoodImage`, `WaterImage`
     - New Items: `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`
     - New Tiles/Props: `AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `FenceImage`, `DebrisImage`
   - `Load()` must load each corresponding file from `images/<name>.png`.

---

## 3. Caveats

1. **Character Sprites and Environment Tiles Responsibility**:
   - Character entities (`player.png`, `zombie.png`, `runner.png`) are specified in Explorer m1_1's scope.
   - Environment tiles (`grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png`, `wall.png`, `tree.png`, `fence.png`, `debris.png`) are specified in Explorer m1_2's scope.
   - When updating `cmd/tools/genassets/main.go`, the implementer should integrate the item generation functions alongside the character and tile functions into a unified generator.
2. **Headless Environment Graphics Initialization**:
   - In automated test pipelines without an active X11/Wayland display server, `ebiten.NewImageFromImage` requires Ebitengine's headless support or decoding standard `image.Image`.

---

## 4. Conclusion & Recommended Code Implementations

### A. Procedural Generation Functions for `cmd/tools/genassets/main.go`

```go
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
```

---

### B. Updated `internal/assets/assets.go` Complete Drop-in Replacement

```go
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
	WallImage   *ebiten.Image
	TreeImage   *ebiten.Image
	FenceImage  *ebiten.Image
	DebrisImage *ebiten.Image

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
```

---

## 5. Verification Method

To verify the implementation independently:

1. **Asset Generation Verification**:
   ```bash
   go run ./cmd/tools/genassets
   ```
   *Expected Result*: Exits with code 0 and logs successful generation for all 20 PNG textures in `internal/assets/images/`, including `food.png`, `water.png`, `weapon.png`, `axe.png`, `shotgun.png`, `ammo.png`, and `armor.png`.

2. **Asset Embedding & Compilation Verification**:
   ```bash
   CC=gcc go test ./...
   ```
   *Expected Result*: All package tests compile and pass with exit code 0.

3. **Runtime Integration Verification**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   *Expected Result*: Launches the game window without missing asset panics, correctly displaying world tiles and items.

4. **Invalidation Conditions**:
   - Any missing image file referenced by `loadEbitenImage` causes a fatal panic on startup.
   - Any transparent or off-by-one pixel indexing outside `[0, 15]` in item generation.
