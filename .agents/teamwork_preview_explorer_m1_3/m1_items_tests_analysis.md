# Milestone 1 Exploration Report: Items / Weapons / Equipment (64x64) & Asset Test Suite

**Author**: `m1_explorer_3`
**Date**: 2026-08-28
**Scope**: 
1. Procedural Generation Specification for 8 Items/Weapons/Equipment in `cmd/tools/genassets/main.go` at 64x64 resolution.
2. Complete Asset Test Suite Specification across `internal/assets/assets_test.go`, `cmd/tools/genassets/genassets_test.go`, and `internal/assets/assets_stress_test.go` for all 27 assets (256x128 floors, 256x256 obstacles, 64x128 entities, 64x64 items).

---

## 1. Overview of 4x Scaling for Items & Test Suite

Milestone 1 expands all sprite canvases by 4x to match vector-art fidelity:
- **Floor Tiles**: 64x32 → 256x128 (6 textures)
- **Obstacles / Props**: 64x64 → 256x256 (10 textures)
- **Character Entities**: 16x32 → 64x128 (3 textures)
- **Items & Equipment**: 16x16 → 64x64 (8 textures)

Total asset count across all categories is **27 distinct PNG textures**, all generated procedurally in pure Go without external downloads, embedded via `embed.FS` in `internal/assets`, and decoded into `*ebiten.Image` handles.

---

## 2. Items, Weapons & Equipment (64x64) Procedural Art Specifications

In the 2.5D isometric view, items are rendered on the ground with center anchor offset `(-32, -32)`. At 64x64 resolution, each item requires rich internal geometry, directional specular highlights, multi-shade color gradients, structural depth, and a clean 1px dark perimeter contour (`RGBA{20..40, 20..40, 20..40, 255}`) for high contrast on varied floor tiles (grass, dirt, asphalt, wood, concrete, tile).

---

### Item 1: Food (`food.png` - 64x64)
**Subject**: Vintage canned tomato/bean soup with metallic rims, beveled lid, pull-tab ring, vibrant red label with gold foil bands, and center vegetable emblem.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Cylinder Bounds**: Centered horizontally at $X \in [14..49]$ (width = 36px), vertical span $Y \in [8..56]$ (height = 49px).
- **Top Lid Ellipse**: Center at $(31.5, 13.5)$, radii $R_x = 17.5, R_y = 5.5$.
  - Lid surface: $Y \in [8..14]$.
  - Outer chime rim: $Y = 14$, thickness = 2px.
  - Pull-tab ring: $X \in [28..35], Y \in [9..13]$ with silver rivet at $(31.5, 12)$ and shadow cutout.
- **Upper Exposed Metal Strip**: $Y \in [15..18], X \in [14..49]$.
- **Central Label Body**: $Y \in [19..48], X \in [14..49]$ (height = 30px).
  - Main Red Label: $Y \in [19..48]$. Horizontal cylindrical shading: brightest at $X \in [22..28]$, midtone across center, deep shadow at $X \in [44..49]$.
  - Gold Foil Bands: Upper stripe at $Y \in [26..28]$, lower stripe at $Y \in [38..40]$.
  - Central Soup Bowl / Tomato Emblem: Centered at $(31.5, 33.5)$, radius 5px.
- **Lower Exposed Metal Strip**: $Y \in [49..52], X \in [14..49]$.
- **Bottom Rim / Base Ellipse**: Center at $(31.5, 53.5)$, radii $R_x = 17.5, R_y = 5.5$, span $Y \in [51..57]$.

#### Color Palette
```go
darkBorder   := color.RGBA{35, 38, 42, 255}
tinLid       := color.RGBA{185, 190, 198, 255}
tinLidHi     := color.RGBA{240, 245, 252, 255}
tinLidSh     := color.RGBA{130, 135, 142, 255}
tinRim       := color.RGBA{215, 220, 228, 255}
metalBodyHi  := color.RGBA{220, 225, 235, 255}
metalBodyMid := color.RGBA{165, 172, 180, 255}
metalBodySh  := color.RGBA{110, 115, 125, 255}
labelRed     := color.RGBA{205, 35, 28, 255}
labelRedHi   := color.RGBA{245, 75, 65, 255}
labelRedSh   := color.RGBA{125, 18, 14, 255}
labelGold    := color.RGBA{245, 195, 45, 255}
labelGoldHi  := color.RGBA{255, 230, 120, 255}
labelGoldSh  := color.RGBA{170, 125, 18, 255}
emblemGreen  := color.RGBA{45, 145, 40, 255}
emblemYellow := color.RGBA{255, 215, 50, 255}
```

#### Drawing Algorithm
1. Clear 64x64 canvas.
2. Render bottom base ellipse with `metalBodySh` and `tinRim` bevel.
3. For $y \in [15..52]$:
   - For $x \in [14..49]$:
     - Cylindrical intensity $t = \cos\left(\frac{x - 22}{35} \cdot \frac{\pi}{2}\right)$.
     - If $y \in [15..18]$ or $y \in [49..52]$: interpolate `metalBodySh` $\to$ `metalBodyHi` by $t$.
     - If $y \in [19..48]$:
       - Default color: red label (`labelRed`, highlighted at $X \in [20..26]$ with `labelRedHi`, shadowed at $X > 42$ with `labelRedSh`).
       - If $y \in [26..28]$ or $y \in [38..40]$: blend with `labelGold` and `labelGoldHi`.
       - If $(x-31.5)^2 + (y-33.5)^2 \le 5.0^2$: draw emblem (golden border with green tomato/leaf center).
4. Render top lid ellipse at $(31.5, 13.5)$:
   - Lid fill with radial gradient and concentric rim rings at $r=16, 14, 11$.
   - Pull tab at $X \in [28..35], Y \in [9..13]$ with silver loop and drop shadow.
5. Apply `addSelectiveOutline(img, darkBorder)`.

---

### Item 2: Water (`water.png` - 64x64)
**Subject**: Clear ergonomic sports water bottle with ribbed white cap, translucent cyan/blue water volume, curved meniscus fill line, ergonomic contoured waist, internal air bubbles, and petalloid base.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Vertical Alignment**: Centered along $X = 31.5$, span $Y \in [4..59]$ (height = 56px).
- **Sports Cap & Threading Ring**:
  - Main Cap: $X \in [26..37], Y \in [4..11]$. Vertical grip ridges with highlight on left ($X=28$) and shadow on right ($X=36$).
  - Cap Base / Safety Ring: $X \in [25..38], Y \in [12..14]$.
- **Bottle Neck & Taper**:
  - Neck: $X \in [27..36], Y \in [15..19]$.
  - Flared Shoulder: Expanding from width 10px ($X \in [27..36]$ at $Y=19$) to width 32px ($X \in [16..47]$ at $Y=27$).
- **Fluid Meniscus & Water Fill Line**:
  - Curved elliptical meniscus at $Y \in [24..27]$, $X \in [18..45]$.
  - Meniscus refraction highlight line: $Y = 25$, bright white/cyan `RGBA{240, 252, 255, 255}`.
- **Bottle Body with Contoured Waist**:
  - Upper Body: $Y \in [27..35], X \in [16..47]$ (width = 32px).
  - Ergonomic Waist Indentation: $Y \in [35..45]$, narrowing to $X \in [19..44]$ (width = 26px) with 3 horizontal grip ridges.
  - Lower Body: $Y \in [45..54]$, expanding back to $X \in [16..47]$ (width = 32px).
- **Water Volume & Specular Highlights**:
  - Fluid color: Gradient from luminous cyan `RGBA{40, 160, 245, 255}` on left to deep cobalt `RGBA{18, 65, 175, 255}` on right.
  - Primary vertical reflection streak: $X \in [20..22], Y \in [26..53]$.
  - Micro-bubbles: Circles at $(26, 33)$ ($r=1.5$), $(38, 42)$ ($r=2.0$), $(29, 48)$ ($r=1.2$).
- **Petalloid Base & Bottom Ribs**:
  - Base Feet: $Y \in [55..59], X \in [17..46]$ with 5-point structural mold ridges.

#### Color Palette
```go
darkBorder    := color.RGBA{25, 45, 75, 255}
capWhite      := color.RGBA{245, 248, 255, 255}
capBlue       := color.RGBA{170, 200, 235, 255}
capSh         := color.RGBA{115, 145, 185, 255}
glassEdge     := color.RGBA{95, 155, 220, 255}
glassEdgeDark := color.RGBA{30, 68, 135, 255}
highlight     := color.RGBA{245, 252, 255, 255}
waterBase     := color.RGBA{38, 130, 235, 255}
waterDeep     := color.RGBA{18, 65, 175, 255}
waterLight    := color.RGBA{145, 210, 255, 255}
waterGlow     := color.RGBA{85, 185, 255, 255}
bubbleColor   := color.RGBA{210, 240, 255, 255}
```

#### Drawing Algorithm
1. Compute profile half-width $W(y)$ for each $y \in [4..59]$.
2. Fill bottle volume for $x \in [31.5 - W(y), 31.5 + W(y)]$:
   - If $y < 14$: draw cap with vertical ribbed shading.
   - If $y \ge 14$ and $y < 25$: draw empty translucent bottle neck and shoulder.
   - If $y \ge 25$: draw water volume with horizontal cylindrical gradient (`waterLight` $\to$ `waterBase` $\to$ `waterDeep`).
3. Draw primary vertical specular reflection streak along left curvature ($x \approx 31.5 - W(y) + 3$).
4. Draw 3 horizontal gripping ribs across waist ($y = 37, 40, 43$) with highlight on upper edge and shadow on lower edge.
5. Render air bubbles as small circles with bright top-left specular dot.
6. Render base mold ribs at $y \in [55..59]$.
7. Apply `addSelectiveOutline(img, darkBorder)`.

---

### Item 3: Weapon / Spiked Bat (`weapon.png` - 64x64)
**Subject**: Weathered wooden baseball bat oriented diagonally, featuring layered white athletic tape grip, turned wooden knob, smooth tapered ash barrel with woodgrain streaks, hardened steel spikes/nails driven through the head, and crimson blood drops.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Diagonal Orientation**: Vector from knob at $(8, 56)$ to top cap at $(55, 8)$ (angle $\theta \approx -45^\circ$, length $\approx 67\text{px}$).
- **Knob / Pommel**: Center $(8, 56)$, radius 4.5px. Flared wooden lip with dark end cap.
- **Taped Grip**: Along centerline from $(10, 54)$ to $(23, 41)$ (length 18px, thickness 6px).
  - Overlapping tape wraps: Spiraling angled seams every 3px with tape base `RGBA{230, 226, 215, 255}` and shadow folds `RGBA{145, 138, 122, 255}`.
- **Tapered Throat**: Centerline from $(23, 41)$ to $(34, 30)$ (expanding thickness from 6px to 9px).
- **Ash/Maple Barrel & Sweet Spot**: Centerline from $(34, 30)$ to $(53, 11)$ (thickness expanding up to 13px).
  - Cylindrical lighting perpendicular to $-45^\circ$ diagonal: Top-left flank highlighted with golden ash `RGBA{240, 192, 125}`, center honey `RGBA{195, 135, 70}`, bottom-right flank deep walnut `RGBA{120, 68, 25}`.
  - Longitudinal woodgrain streaks drawn along barrel axis.
- **Bat End Cap**: Rounded dome at $(54, 9)$ (radius 5.5px).
- **Protruding Steel Spikes**: 6 metallic spikes driven through barrel:
  - Spike 1: At $(40, 22)$, pointing top-left (tip at $(34, 16)$).
  - Spike 2: At $(44, 18)$, pointing top-right (tip at $(48, 12)$).
  - Spike 3: At $(47, 15)$, pointing bottom-left (tip at $(43, 21)$).
  - Spike 4: At $(51, 11)$, pointing top-left (tip at $(46, 5)$).
  - Spike 5: At $(37, 25)$, pointing bottom-right (tip at $(43, 29)$).
  - Spike 6: At $(49, 13)$, pointing bottom-right (tip at $(56, 17)$).
  - Each spike has dark steel base collar `RGBA{90, 100, 115}`, bright needle point `RGBA{245, 250, 255}`, and specular line.
- **Blood Stains**: Crimson splatter drops `RGBA{168, 22, 22, 255}` dripping from head spikes at $(46, 5), (55, 18), (50, 8)$.

#### Color Palette
```go
darkBorder := color.RGBA{45, 28, 15, 255}
woodSh     := color.RGBA{115, 65, 25, 255}
woodMid    := color.RGBA{185, 128, 65, 255}
woodHi     := color.RGBA{238, 188, 122, 255}
woodCore   := color.RGBA{210, 155, 90, 255}
tapeBase   := color.RGBA{230, 226, 215, 255}
tapeSh     := color.RGBA{145, 138, 122, 255}
tapeHi     := color.RGBA{250, 248, 242, 255}
steelSpike := color.RGBA{245, 250, 255, 255}
steelMid   := color.RGBA{170, 182, 196, 255}
steelBase  := color.RGBA{90, 100, 115, 255}
blood      := color.RGBA{168, 22, 22, 255}
bloodDark  := color.RGBA{110, 14, 14, 255}
```

#### Drawing Algorithm
1. Parametrize bat centerline $C(s) = (1-s) \cdot (8, 56) + s \cdot (54, 9)$ for $s \in [0.0, 1.0]$.
2. Compute normal vector $\vec{n} = (1/\sqrt{2}, 1/\sqrt{2})$ and bat radius $R(s)$:
   - For $s \in [0.0, 0.05]$: Knob ($R = 4.5$).
   - For $s \in [0.05, 0.30]$: Grip ($R = 3.2$).
   - For $s \in [0.30, 0.50]$: Throat ($R = 3.2 \to 4.8$).
   - For $s \in [0.50, 0.95]$: Barrel ($R = 4.8 \to 6.5$).
   - For $s \in [0.95, 1.0]$: Cap ($R = 6.5 \to 0$).
3. For each pixel $(x, y)$, compute distance to centerline and offset $d_{\perp}$ along $\vec{n}$:
   - If $|d_{\perp}| \le R(s)$:
     - In grip zone: draw angled tape wraps with `tapeBase` and spiral shade lines.
     - In wood zone: compute lighting $u = (d_{\perp} + R) / (2R)$ to blend `woodHi` $\to$ `woodMid` $\to$ `woodSh`, and add subtle grain noise.
4. Draw 6 spikes by rendering angled line segments from barrel center outward to tips with needle highlights.
5. Add blood drops on spike tips and impact barrel zone.
6. Apply `addSelectiveOutline(img, darkBorder)`.

---

### Item 4: Fire Axe (`axe.png` - 64x64)
**Subject**: Heavy municipal fire axe with curved hickory handle, rubberized fawn-foot grip, forged steel eye collar with wedge pin, rear breaching pick/poll, high-visibility red enamel blade body, and mirror-polished curved cutting edge.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Hickory Handle**: Curves from fawn-foot knob at $(10, 58)$ through $(19, 44) \to (26, 32) \to (34, 18)$ (length $\approx 54\text{px}$, width $\approx 6\text{px}$).
  - Rubber base grip: $X \in [8..16], Y \in [48..58]$ with cross-hatch texture.
  - Wood shaft: Smooth golden hickory grain with upper highlight and underside shadow.
- **Axe Head Eye / Socket**: Heavy steel collar centered at $(33, 18)$, width 10px, height 12px ($X \in [28..38], Y \in [12..24]$). Steel top wedge pin visible at $(33, 12)$.
- **Rear Breaching Pick / Spike**: Extends backward/left from eye:
  - Base at $(28, 16)$, tapering to sharp chisel point at $(15, 16)$.
  - Thickness tapers from 6px at eye to 1px at tip, beveled top highlight.
- **Main Blade Body (Red Painted Enamel)**:
  - Extends forward/right from eye: $X \in [36..52], Y \in [8..28]$.
  - Flared wedge profile expanding vertically from 10px at eye to 22px near edge.
  - Glossy red finish: specular reflection stripe at $Y \in [10..14]$, deep red core, dark red underside shadow.
- **Beveled Cutting Edge / Bit**:
  - Curved crescent cutting edge from $(50, 6)$ through $(57, 17)$ to $(50, 29)$ (span 24px).
  - Forged transition bevel ($X \in [48..53]$) with satin steel sheen `RGBA{180, 195, 210, 255}`.
  - Mirror-polished razor edge apex ($X \in [54..58]$) with brilliant white highlight `RGBA{255, 255, 255, 255}`.

#### Color Palette
```go
darkBorder := color.RGBA{32, 30, 32, 255}
gripRubber := color.RGBA{38, 40, 45, 255}
gripHi     := color.RGBA{75, 80, 90, 255}
woodMid    := color.RGBA{220, 155, 70, 255}
woodHi     := color.RGBA{250, 195, 125, 255}
woodSh     := color.RGBA{135, 80, 25, 255}
axeRed     := color.RGBA{220, 32, 32, 255}
axeRedHi   := color.RGBA{255, 88, 88, 255}
axeRedSh   := color.RGBA{135, 15, 15, 255}
steelEye   := color.RGBA{85, 92, 102, 255}
steelBevel := color.RGBA{180, 195, 210, 255}
steelEdge  := color.RGBA{250, 253, 255, 255}
```

#### Drawing Algorithm
1. Draw curved handle centerline using quadratic bezier from $(10, 58)$ via $(20, 42)$ to $(34, 18)$.
2. Draw rubber grip at base $(Y \in [48..58])$ with ribbed cross-hatch texture.
3. Draw hickory wood along remaining handle with directional light shading.
4. Draw forged steel eye collar at $X \in [28..38], Y \in [12..24]$.
5. Draw tapered rear breaching pick from $(28, 16)$ to $(15, 16)$ with beveled steel shading.
6. Draw flared axe blade body with red enamel gradient from $X=36$ to $X=52$.
7. Draw crescent cutting bit from $Y=6$ to $Y=29$ with transition bevel and brilliant white edge apex.
8. Apply `addSelectiveOutline(img, darkBorder)`.

---

### Item 5: Shotgun (`shotgun.png` - 64x64)
**Subject**: Tactical pump-action 12-gauge shotgun oriented diagonally, featuring walnut buttstock with rubber recoil pad, blued steel receiver with ejection port, ribbed pump forend slide, parallel barrel & magazine tube, and front brass bead sight.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Diagonal Orientation**: From buttpad at $(6, 52)$ to muzzle crown at $(58, 14)$ (length $\approx 65\text{px}$).
- **Rubber Recoil Buttpad**: Plate at $X \in [6..9], Y \in [49..55]$ (thickness 3px, height 7px) with mounting screws.
- **Walnut Buttstock & Pistol Grip**:
  - Stock body: $X \in [9..22], Y \in [38..52]$. Ergonomic comb curve on top, pistol grip sweep on bottom with checkered traction stippling.
- **Trigger Guard & Steel Trigger**:
  - Guard loop at $X \in [22..26], Y \in [36..41]$, silver curved trigger at $(24, 38)$.
- **Blued Steel Receiver**:
  - Milled box at $X \in [22..34], Y \in [26..37]$ (width 12px, height 11px).
  - Darkened ejection port cutout at $X \in [26..31], Y \in [28..31]$ with visible brass shell rim inside.
  - Receiver pin rivets at $(24, 33)$ and $(31, 33)$.
- **Corrugated Wooden/Polymer Pump Forend**:
  - Cylindrical slide around mag tube at $X \in [34..44], Y \in [20..29]$ (length 11px, thickness 7px).
  - 6 vertical gripping ribs with highlight on left ridge and deep shadow in groove.
- **Over/Under Barrel & Tubular Magazine**:
  - Top 12-Gauge Barrel: $X \in [34..58], Y \in [13..25]$ (thickness 4px) with ventilated top rib.
  - Bottom Magazine Tube: $X \in [34..52], Y \in [17..28]$ (thickness 3.5px) with knurled end cap at $(52, 20)$.
  - Gunmetal cylindrical shading with bright top reflection line.
- **Muzzle Crown & Brass Bead Sight**:
  - Muzzle opening at $(58, 14)$ with dark bore hole.
  - Front brass bead sight at $(56, 12)$ (`RGBA{250, 215, 60, 255}`).

#### Color Palette
```go
darkBorder := color.RGBA{24, 26, 30, 255}
buttRubber := color.RGBA{38, 38, 42, 255}
buttScrew  := color.RGBA{140, 145, 155, 255}
woodStock  := color.RGBA{148, 90, 44, 255}
woodHi     := color.RGBA{200, 132, 75, 255}
woodSh     := color.RGBA{95, 52, 22, 255}
steelRec   := color.RGBA{68, 74, 85, 255}
steelRecHi := color.RGBA{118, 128, 145, 255}
ejectPort  := color.RGBA{22, 24, 28, 255}
pumpWood   := color.RGBA{160, 102, 50, 255}
pumpRib    := color.RGBA{75, 42, 18, 255}
barrelHi   := color.RGBA{190, 200, 215, 255}
barrelMid  := color.RGBA{85, 92, 105, 255}
barrelSh   := color.RGBA{42, 48, 56, 255}
beadSight  := color.RGBA{250, 215, 60, 255}
```

#### Drawing Algorithm
1. Render buttpad and screws at $(6..9, 49..55)$.
2. Render walnut stock from $(9, 49)$ to $(22, 38)$ with curved comb and textured grip.
3. Render trigger guard and silver trigger.
4. Render milled steel receiver box with top highlight, bottom shadow, ejection port, and pin rivets.
5. Render pump forend slide with 6 vertical traction grooves.
6. Render parallel top barrel and bottom magazine tube with cylindrical gunmetal gradients.
7. Render muzzle crown, bore opening, and front brass bead sight.
8. Apply `addSelectiveOutline(img, darkBorder)`.

---

### Item 6: Ammo Box (`ammo.png` - 64x64)
**Subject**: Heavy-gauge olive drab military ammunition can with hinged lid open, revealing a top tray of standing brass cartridges with copper jacket tips, yellow stencil military markings, side carry handle latch, and steel corner rivets.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Ammo Can Body**: Centered horizontally, $X \in [10..53]$ (width 44px), $Y \in [20..58]$ (height 39px).
- **Open Top Tray & Protruding Cartridges**:
  - Tray opening: $Y \in [18..22], X \in [12..51]$.
  - 6 Standing Rifle/Pistol Cartridges protruding upward at $X = 15, 21, 27, 33, 39, 45$:
    - Copper jacketed bullet tips: $Y \in [8..13]$ (pointed apex at $Y=8$, width 3px).
    - Polished brass casings: $Y \in [13..21]$ (width 4px) with bright left specular line and extractor rim.
- **Olive Drab Metal Can Exterior**:
  - Top Beveled Rim: $Y \in [21..24], X \in [10..53]$.
  - Main Front Wall: $Y \in [24..54], X \in [10..53]$.
    - Stamped recessed center panel: $X \in [14..49], Y \in [26..52]$ (indented 1px with drop shadow).
  - Yellow Military Stencil Band: $Y \in [32..40], X \in [16..47]$:
    - Background stencil plate / text: "9MM BALL" / "100 RDS" markings (`RGBA{240, 190, 40, 255}` with dark stenciled lettering `RGBA{25, 32, 20, 255}`).
  - Corner Rivets & Hardware: Steel rivets at $(13, 25), (50, 25), (13, 51), (50, 51)$. Side handle hinge bracket at $X \in [8..11], Y \in [33..43]$.
- **Bottom Base Rim**: $Y \in [54..58], X \in [10..53]$ with floor contact shadow.

#### Color Palette
```go
darkBorder    := color.RGBA{22, 30, 16, 255}
boxGreen      := color.RGBA{68, 90, 45, 255}
boxGreenHi    := color.RGBA{100, 132, 65, 255}
boxGreenSh    := color.RGBA{40, 55, 25, 255}
stencilYellow := color.RGBA{240, 190, 40, 255}
stencilText   := color.RGBA{25, 32, 20, 255}
brassHi       := color.RGBA{255, 238, 135, 255}
brassMid      := color.RGBA{235, 190, 55, 255}
brassSh       := color.RGBA{155, 115, 25, 255}
copperTip     := color.RGBA{215, 105, 45, 255}
copperTipHi   := color.RGBA{255, 165, 98, 255}
rivetColor    := color.RGBA{170, 180, 175, 255}
```

#### Drawing Algorithm
1. Render 6 brass cartridges at $X \in \{15, 21, 27, 33, 39, 45\}$ from $Y=8$ to $Y=21$ with copper bullet tips and brass casing highlights.
2. Render top rim of ammo box at $Y \in [21..24]$.
3. Render front wall of ammo can at $Y \in [24..54]$ with olive drab enamel shading.
4. Draw stamped recessed panel with top-left highlight and bottom-right inner shadow.
5. Draw yellow stencil band with military typography stippling across $Y \in [32..40]$.
6. Draw 4 corner steel rivets and side latch bracket.
7. Render bottom base rim at $Y \in [54..58]$.
8. Apply `addSelectiveOutline(img, darkBorder)`.

---

### Item 7: Armor (`armor.png` - 64x64)
**Subject**: Tactical ballistic Kevlar plate carrier vest with padded shoulder straps, quick-release steel buckles, scooped ergonomic neckline, modular chest rig with hook-and-loop Velcro ID patch, laser-cut MOLLE webbing grid, and 3 lower magazine utility pouches with elastic pull-tabs.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Overall Silhouette**: Centered horizontally, $X \in [10..53]$ (width 44px), $Y \in [6..58]$ (height 53px).
- **Padded Shoulder Straps**:
  - Left Strap: $X \in [12..21], Y \in [6..18]$ with steel buckle at $(14, 12)$.
  - Right Strap: $X \in [42..51], Y \in [6..18]$ with steel buckle at $(48, 12)$.
  - Heavy 1000D Cordura texture with top highlight and outer edge shadow.
- **Scooped Neck Opening**: Smooth curved neckline centered at $X=31.5$, $Y \in [6..18]$, inner collar shadow `RGBA{20, 22, 28, 255}`.
- **Chest Plate Carrier Body**:
  - Front Plate Pocket: $X \in [14..49], Y \in [18..54]$ (trapezoidal plate silhouette).
  - Upper Velcro ID / Morale Patch: $X \in [22..41], Y \in [19..25]$ (tan/olive hook fabric with embroidered cross border).
  - Laser-Cut MOLLE Webbing Rows: 3 horizontal webbing bands across $X \in [15..48]$ at $Y = 27..28, 33..34, 39..40$ with vertical bar-tack stitches every 6px.
- **3 Front Magazine / Utility Pouches**:
  - Left Pouch: $X \in [14..24], Y \in [41..53]$.
  - Center Pouch: $X \in [26..37], Y \in [41..53]$.
  - Right Pouch: $X \in [39..49], Y \in [41..53]$.
  - Each pouch has top flap at $Y \in [41..43]$, central pull-tab `RGBA{100, 110, 128, 255}`, and side accordion depth.
- **Side Elastic Cummerbund & Waistband Hem**:
  - Ribbed cummerbund panels at sides ($X \in [10..14]$ and $X \in [49..53]$).
  - Bottom edge binding hem at $Y \in [54..58]$.

#### Color Palette
```go
darkBorder  := color.RGBA{18, 20, 25, 255}
vestKevlar  := color.RGBA{48, 54, 65, 255}
vestHi      := color.RGBA{80, 88, 105, 255}
vestSh      := color.RGBA{28, 32, 40, 255}
strapHi     := color.RGBA{92, 100, 118, 255}
buckleSteel := color.RGBA{145, 155, 170, 255}
idPatch     := color.RGBA{95, 90, 75, 255}
idPatchHi   := color.RGBA{125, 120, 102, 255}
molleWeb    := color.RGBA{26, 30, 36, 255}
pouchFlap   := color.RGBA{66, 75, 90, 255}
pouchBody   := color.RGBA{38, 44, 54, 255}
pullTab     := color.RGBA{100, 110, 128, 255}
```

#### Drawing Algorithm
1. Render shoulder straps at $X \in [12..21]$ and $X \in [42..51]$ for $Y \in [6..18]$ with steel buckles.
2. Render scooped neckline cutout with dark inner depth.
3. Render main chest plate carrier trapezoid from $Y=18$ to $Y=54$.
4. Render upper chest Velcro ID patch with highlight border.
5. Render 3 horizontal MOLLE webbing rows with bar-tack stitching columns.
6. Render 3 distinct magazine pouches across lower plate carrier with flaps and pull tabs.
7. Render side cummerbund adjusters and bottom hem binding.
8. Apply `addSelectiveOutline(img, darkBorder)`.

---

### Item 8: Antidote (`antidote.png` - 64x64)
**Subject**: High-tech bioluminescent antidote ampoule / glass vial with natural cork stopper, flanged glass collar lip, narrow neck, rounded cylindrical laboratory flask body, glowing emerald/neon-green antidote fluid with curved meniscus, rising micro-bubbles, etched volume graduation marks, and glass caustic specular highlights.

#### Coordinate Geometry & Proportions (64x64 Canvas)
- **Overall Alignment**: Centered horizontally at $X = 31.5$, span $Y \in [6..58]$ (height 53px).
- **Cork Stopper**: $X \in [26..37], Y \in [6..15]$ (width 12px, height 10px). Textured cork grain with beveled top corners, cork pores, and shadow under lip.
- **Glass Lip & Narrow Neck**:
  - Heavy flanged glass collar lip: $X \in [24..39], Y \in [15..19]$.
  - Neck: $X \in [26..37], Y \in [19..25]$.
- **Laboratory Ampoule Body**:
  - Expanding rounded shoulder: $Y \in [25..31]$, widening to $X \in [14..49]$ (width 36px).
  - Cylindrical body: $X \in [14..49], Y \in [31..52]$.
  - Convex rounded bottom glass base: $Y \in [52..58]$.
- **Bioluminescent Antidote Liquid**:
  - Fluid volume occupies $X \in [17..46], Y \in [29..55]$.
  - Curved fluid meniscus at $Y \in [29..32]$ with intense lime-green surface glow `RGBA{130, 255, 150, 255}`.
  - Multi-layer luminous fluid gradient:
    - Intense electric neon core: $X \in [26..37], Y \in [33..51]$ (`RGBA{65, 250, 85, 255}`).
    - Emerald depth periphery: $X \in [17..25]$ and $X \in [38..46]$ (`RGBA{22, 165, 42, 255}`).
    - Settled bottom sediment: $Y \in [52..55]$ (`RGBA{12, 105, 28, 255}`).
  - 4 Rising micro-bubbles: At $(24, 46)$ ($r=1.5$), $(37, 39)$ ($r=2.0$), $(30, 43)$ ($r=1.2$), $(38, 49)$ ($r=1.0$) with translucent halos.
- **Glass Refraction & Etched Graduations**:
  - Vertical primary specular reflection: $X \in [17..20], Y \in [25..54]$ (brilliant white/cyan streak).
  - Etched volumetric measurement graduation tick lines on right wall at $Y = 35, 41, 47$ ($X \in [43..46]$).
  - Secondary rim backlight along right glass perimeter ($X \in [46..48]$).

#### Color Palette
```go
darkBorder  := color.RGBA{12, 18, 15, 255}
corkMid     := color.RGBA{150, 110, 65, 255}
corkHi      := color.RGBA{185, 145, 95, 255}
corkSh      := color.RGBA{105, 72, 38, 255}
glassHi     := color.RGBA{235, 245, 255, 255}
glassMid    := color.RGBA{160, 205, 230, 255}
glassSh     := color.RGBA{75, 110, 135, 255}
liquidCore  := color.RGBA{65, 250, 85, 255}
liquidGlow  := color.RGBA{130, 255, 150, 255}
liquidMid   := color.RGBA{32, 195, 55, 255}
liquidDark  := color.RGBA{18, 135, 35, 255}
liquidDeep  := color.RGBA{10, 85, 22, 255}
bubbleGlow  := color.RGBA{180, 255, 195, 255}
gradLine    := color.RGBA{220, 245, 230, 255}
```

#### Drawing Algorithm
1. Render textured cork stopper at $X \in [26..37], Y \in [6..15]$ with cork grain speckles and beveled cap.
2. Render glass lip and neck at $X \in [24..39], Y \in [15..25]$.
3. Fill vial body from $Y=25$ to $Y=58$ with rounded base curvature.
4. Render glowing antidote fluid inside vial from $Y=29$ to $Y=55$ with radial luminous glow around core.
5. Render curved meniscus with intense lime glow line at $Y \in [29..32]$.
6. Render 4 glowing micro-bubbles in fluid volume.
7. Draw white measurement graduation marks on right glass wall at $Y \in \{35, 41, 47\}$.
8. Draw primary left specular reflection streak down $X \in [17..20]$ and right rim backlight.
9. Apply `addSelectiveOutline(img, darkBorder)`.

---

## 3. Comprehensive Asset Test Suite Analysis

The project includes three asset test suites across `internal/assets` and `cmd/tools/genassets`. Every test must be updated to validate all **27 assets** under the new 4x dimensions:
- **Floor Tiles (6)**: 256x128
- **Vertical Obstacles & Props (10)**: 256x256
- **Character Entities (3)**: 64x128
- **Items & Equipment (8)**: 64x64

---

### Test Suite 1: `internal/assets/assets_test.go`

#### A. `TestEmbeddedAssetDimensionsAndValidity(t *testing.T)`
1. **Asset Count Assertion**:
   - Current: `if len(expectedAssets) != 20`
   - **Target**: `if len(expectedAssets) != 27`
2. **Table Test Entries (27 assets)**:
   ```go
   expectedAssets := []struct {
       path   string
       width  int
       height int
   }{
       // Character Entities (64x128)
       {"images/player.png", 64, 128},
       {"images/zombie.png", 64, 128},
       {"images/runner.png", 64, 128},

       // Floor Tiles (256x128)
       {"images/grass.png", 256, 128},
       {"images/dirt.png", 256, 128},
       {"images/wood.png", 256, 128},
       {"images/asphalt.png", 256, 128},
       {"images/concrete.png", 256, 128},
       {"images/tile_floor.png", 256, 128},

       // Vertical Obstacles & Props (256x256)
       {"images/wall.png", 256, 256},
       {"images/tree.png", 256, 256},
       {"images/fence.png", 256, 256},
       {"images/debris.png", 256, 256},
       {"images/tent.png", 256, 256},
       {"images/stump.png", 256, 256},
       {"images/mushroom.png", 256, 256},
       {"images/sign.png", 256, 256},
       {"images/elevation_block.png", 256, 256},
       {"images/elevation_ramp.png", 256, 256},

       // Items, Weapons & Equipment (64x64)
       {"images/food.png", 64, 64},
       {"images/water.png", 64, 64},
       {"images/weapon.png", 64, 64},
       {"images/axe.png", 64, 64},
       {"images/shotgun.png", 64, 64},
       {"images/ammo.png", 64, 64},
       {"images/armor.png", 64, 64},
       {"images/antidote.png", 64, 64},
   }
   ```
3. **Assertions**: Verifies format is PNG, bounds match `tc.width` and `tc.height`, and non-transparent pixel count > 0.

#### B. `TestAssetsLoadAllPointersNonNil(t *testing.T)`
1. **Handle Count Assertion**:
   - Current: `if len(handles) != 20`
   - **Target**: `if len(handles) != 27`
2. **Handles Table (27 pointers)**:
   - Entities (3): `PlayerImage`, `ZombieImage`, `RunnerImage` (wantW: 64, wantH: 128)
   - Floors (6): `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage` (wantW: 256, wantH: 128)
   - Obstacles & Props (10): `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`, `TentImage`, `StumpImage`, `MushroomImage`, `SignImage`, `ElevationBlockImage`, `ElevationRampImage` (wantW: 256, wantH: 256)
   - Items (8): `FoodImage`, `WaterImage`, `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `AntidoteImage` (wantW: 64, wantH: 64)
3. **Assertions**: Verifies pointer is non-nil after `Load()` and dimensions match `wantW` and `wantH`.

---

### Test Suite 2: `internal/assets/assets_stress_test.go`

#### A. `TestFloorTileIsometricBounds(t *testing.T)`
- **Current Parameters (64x32)**:
  - Bounds check: `bounds.Dx() != 64 || bounds.Dy() != 32`
  - Center: `centerX = 31.5, centerY = 15.5`
  - Radii: `radiusX = 32.5, radiusY = 16.5`
  - Loops: `for y := 0; y < 32`, `for x := 0; x < 64`
- **Updated Parameters (256x128)**:
  - Bounds check: `bounds.Dx() != 256 || bounds.Dy() != 128`
  - Center: `centerX = 127.5, centerY = 63.5`
  - Radii: `radiusX = 128.5, radiusY = 64.5`
  - Loops: `for y := 0; y < 128; y++`, `for x := 0; x < 256; x++`
  - Normalized distance calculation:
    $$\text{dist} = \frac{|x - 127.5|}{128.5} + \frac{|y - 63.5|}{64.5}$$
  - Anti-aliasing / pixel boundary tolerance remains $\le 1.15$.

#### B. `TestCharacterGroundAnchor(t *testing.T)`
- **Current Parameters (16x32)**:
  - Height = 32, rows checked `y \in [28..31]` (`for y := 28; y < 32`, `for x := 0; x < 16`).
- **Updated Parameters (64x128)**:
  - Height = 128, width = 64.
  - Ground anchor rows: Check bottom 16 rows (`y \in [112..127]`, i.e. `for y := 112; y < 128; y++`, `for x := 0; x < 64; x++`).
  - Ensures feet / shadow pixels exist at the base to prevent floating entities.

#### C. `TestItemOutlineContrast(t *testing.T)`
- **Current Parameters (16x16, 7 items)**:
  - Missing `images/antidote.png`.
  - Loops: `for y := 0; y < 16`, `for x := 0; x < 16`.
  - Solid pixel threshold: `solidCount < 20` (out of 256).
- **Updated Parameters (64x64, 8 items)**:
  - Include all 8 items: `images/food.png`, `images/water.png`, `images/weapon.png`, `images/axe.png`, `images/shotgun.png`, `images/ammo.png`, `images/armor.png`, `images/antidote.png`.
  - Loops: `for y := 0; y < 64; y++`, `for x := 0; x < 64; x++`.
  - Solid pixel threshold: $64 \times 64 = 4096\text{px}$ (16x area). Threshold scaled from 20 to $\ge 320\text{px}$ (`if solidCount < 320`).
  - Dark contour threshold: Luminance $\text{lum} = 0.299R + 0.587G + 0.114B < 80$, verify `darkContourCount > 0`.

#### D. `TestAssetsLoadIdempotency(t *testing.T)`
- Check all 27 pointers for non-nil across multiple consecutive `Load()` calls (include `TentImage`, `StumpImage`, `MushroomImage`, `SignImage`, `ElevationBlockImage`, `ElevationRampImage`, `AntidoteImage`).

---

### Test Suite 3: `cmd/tools/genassets/genassets_test.go`

#### A. `expectedAssetFiles` Registration
- Update table to contain all **27 assets**:
  - 3 Characters @ 64x128
  - 6 Floor Tiles @ 256x128 (`isIso: true`)
  - 10 Obstacles & Props @ 256x256
  - 8 Items & Equipment @ 64x64
- Update count assertion: `if len(expectedAssetFiles) != 27`.

#### B. `TestAssetRegenerationDeterminism(t *testing.T)`
- Runs `go run ./cmd/tools/genassets` for 3 iterations from repo root.
- Verifies SHA-256 hash stability for all 27 asset files.

#### C. `TestAssetDimensionsAndIntegrity(t *testing.T)`
- Verifies format is `png`.
- Verifies `Dx() == tc.width` and `Dy() == tc.height`.
- Computes fill ratio $\text{nonTransparent} / (W \times H) \ge 0.05$ (at least 5% non-transparent pixel fill).

---

## 4. Implementation Plan & File Modifications

| File | Target Changes |
|---|---|
| `cmd/tools/genassets/main.go` | Replace item generator functions (`generateFood`, `generateWater`, `generateWeapon`, `generateAxe`, `generateShotgun`, `generateAmmo`, `generateArmor`, `generateAntidote`) with 64x64 canvas implementations adhering to the mathematical geometries, multi-tone color palettes, and `addSelectiveOutline`. |
| `cmd/tools/genassets/genassets_test.go` | Update `expectedAssetFiles` table to 27 items with 256x128, 256x256, 64x128, 64x64 dimensions. Update count check to 27. |
| `internal/assets/assets_test.go` | Update `TestEmbeddedAssetDimensionsAndValidity` table to 27 items with new dimensions; update `TestAssetsLoadAllPointersNonNil` handles table to 27 pointers with new dimensions. |
| `internal/assets/assets_stress_test.go` | Update `TestFloorTileIsometricBounds` diamond center/radii for 256x128; update `TestCharacterGroundAnchor` rows for 64x128; update `TestItemOutlineContrast` loop, threshold, and item list for 64x64 (8 items); update `TestAssetsLoadIdempotency` to check all 27 pointers. |

---

## 5. Verification Commands

To verify the asset scaling and tests once implemented:
1. `go run ./cmd/tools/genassets` (must succeed and generate all 27 high-fidelity PNGs in `internal/assets/images`)
2. `CC=gcc go test -v ./cmd/tools/genassets/...` (determinism and dimension checks pass)
3. `CC=gcc go test -v ./internal/assets/...` (validity, bounds, grounding, and contrast stress tests pass)
4. `CC=gcc go test ./...` (full repository test suite passes)
