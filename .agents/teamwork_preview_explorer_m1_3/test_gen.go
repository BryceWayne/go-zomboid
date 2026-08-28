package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	if x >= 0 && x < 16 && y >= 0 && y < 16 {
		img.Set(x, y, c)
	}
}

func drawHLine(img *image.RGBA, x0, x1, y int, c color.RGBA) {
	for x := x0; x <= x1; x++ {
		setPixel(img, x, y, c)
	}
}

func drawVLine(img *image.RGBA, x, y0, y1 int, c color.RGBA) {
	for y := y0; y <= y1; y++ {
		setPixel(img, x, y, c)
	}
}

func fillRect(img *image.RGBA, x0, y0, w, h int, c color.RGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			setPixel(img, x, y, c)
		}
	}
}

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

	// Bat Shaft & Barrel (x=6..14, y=2..9)
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

	// Spikes protruding
	// Top spike & blood
	setPixel(img, 14, 1, steelSpike)
	setPixel(img, 14, 2, blood)
	setPixel(img, 15, 3, blood)

	// Left spike
	setPixel(img, 10, 3, steelSpike)
	setPixel(img, 11, 3, steelBase)

	// Right spike
	setPixel(img, 14, 5, steelBase)
	setPixel(img, 15, 5, steelSpike)

	// Mid spike
	setPixel(img, 8, 5, steelSpike)
	setPixel(img, 9, 5, steelBase)

	// Lower spike
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
	// Base rubber grip (y=13..15, x=2..4)
	setPixel(img, 2, 14, darkBorder)
	setPixel(img, 2, 13, gripRubber)
	setPixel(img, 3, 14, gripRubber)
	setPixel(img, 3, 13, gripHi)
	setPixel(img, 4, 14, darkBorder)

	// Wood shaft curve
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

	// Rear Breaching Spike / Poll (x=5..6, y=3..4)
	setPixel(img, 5, 3, steelEdge)
	setPixel(img, 6, 3, steelBevel)
	setPixel(img, 5, 4, darkBorder)
	setPixel(img, 6, 4, steelEye)

	// Main Blade Body (Red Painted, x=9..13, y=1..6)
	// Top flare (y=1..2)
	setPixel(img, 10, 2, axeRedHi)
	setPixel(img, 11, 2, axeRedHi)
	setPixel(img, 12, 1, axeRedHi)

	// Mid blade (y=3..4)
	for y := 3; y <= 4; y++ {
		setPixel(img, 10, y, axeRed)
		setPixel(img, 11, y, axeRed)
		setPixel(img, 12, y, axeRed)
		setPixel(img, 13, y, axeRed)
	}

	// Bottom flare (y=5..6)
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
	setPixel(img, 6, 9, color.RGBA{150, 155, 165, 255}) // trigger

	// Steel Receiver (x=5..8, y=7..9)
	setPixel(img, 5, 9, steelRec)
	setPixel(img, 6, 8, steelRecHi)
	setPixel(img, 6, 9, steelRec)
	setPixel(img, 7, 7, steelRecHi)
	setPixel(img, 7, 8, ejectPort) // Ejection port!
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
	// Top Barrel (y=3..4)
	for x := 9; x <= 15; x++ {
		setPixel(img, x, 3-(x-9)/3, barrelHi)   // Top highlight line
		setPixel(img, x, 4-(x-9)/3, barrelMid)  // Core barrel
	}
	// Under-barrel Mag Tube (y=4..5)
	for x := 9; x <= 13; x++ {
		setPixel(img, x, 5-(x-9)/3, barrelSh)
	}

	// Muzzle Crown & Front Brass Bead Sight (x=14..15, y=1..3)
	setPixel(img, 14, 1, beadSight) // Brass bead
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

	// 4 Bullet Tips protruding from open top tray (x=3..12, y=2..4)
	bulletsX := []int{3, 6, 9, 12}
	for _, bx := range bulletsX {
		// Copper FMJ Tip (y=2)
		setPixel(img, bx, 2, copperTipHi)
		if bx+1 <= 13 {
			setPixel(img, bx+1, 2, copperTip)
		}
		// Brass Case (y=3..4)
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
	// Stencil caliber text dots
	setPixel(img, 5, 8, stencilText)
	setPixel(img, 7, 8, stencilText)
	setPixel(img, 8, 9, stencilText)
	setPixel(img, 10, 8, stencilText)

	// Corner Rivets (metal reinforcements)
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
	// Left strap (x=3..5)
	drawHLine(img, 3, 5, 2, strapHi)
	drawHLine(img, 3, 5, 3, vestKevlar)
	setPixel(img, 3, 3, buckleSteel) // Left buckle
	drawHLine(img, 3, 5, 4, vestKevlar)
	setPixel(img, 2, 2, darkBorder)
	setPixel(img, 2, 3, darkBorder)
	setPixel(img, 2, 4, darkBorder)

	// Right strap (x=10..12)
	drawHLine(img, 10, 12, 2, strapHi)
	drawHLine(img, 10, 12, 3, vestKevlar)
	setPixel(img, 12, 3, buckleSteel) // Right buckle
	drawHLine(img, 10, 12, 4, vestKevlar)
	setPixel(img, 13, 2, darkBorder)
	setPixel(img, 13, 3, darkBorder)
	setPixel(img, 13, 4, darkBorder)

	// Scooped Neck Opening (x=6..9, y=2..4)
	drawHLine(img, 6, 9, 2, darkBorder)
	drawHLine(img, 6, 9, 3, color.RGBA{22, 25, 30, 255})
	drawHLine(img, 6, 9, 4, molleWeb) // Collar reinforcement

	// Chest Plate Body (y=5..8, x=2..13)
	// y=5 (x=3..12)
	drawHLine(img, 3, 12, 5, vestKevlar)
	setPixel(img, 2, 5, darkBorder)
	setPixel(img, 13, 5, darkBorder)
	// y=6..8 (x=2..13)
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
	// Pouch 1 (Left, x=3..5)
	drawHLine(img, 3, 5, 9, pouchFlap)
	setPixel(img, 4, 9, pullTab)
	for y := 10; y <= 12; y++ {
		drawHLine(img, 3, 5, y, pouchBody)
		setPixel(img, 3, y, vestHi)
	}

	// Pouch 2 (Center, x=6..9)
	drawHLine(img, 6, 9, 9, pouchFlap)
	setPixel(img, 7, 9, pullTab)
	setPixel(img, 8, 9, pullTab)
	for y := 10; y <= 12; y++ {
		drawHLine(img, 6, 9, y, pouchBody)
	}

	// Pouch 3 (Right, x=10..12)
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

func saveImg(name string, img *image.RGBA) {
	path := "/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_3/" + name
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

func main() {
	generateFood("food.png")
	generateWater("water.png")
	generateWeapon("weapon.png")
	generateAxe("axe.png")
	generateShotgun("shotgun.png")
	generateAmmo("ammo.png")
	generateArmor("armor.png")
}
