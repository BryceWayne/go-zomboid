# Procedural Art Style Guide

## Overview
This document outlines the target aesthetic for the next iteration of the project based on the new reference art, and provides actionable steps for updating our procedural generation math in `cmd/tools/genassets/main.go`.

## Visual Analysis
* **Art Style**: The reference art features a highly stylized, minimalist, "vector-like" aesthetic. It emphasizes clean geometric shapes, flat shading, and zero high-frequency texture noise.
* **Color Palette**: The palette is vibrant and pastel-leaning. Greens are soft mint and emerald, browns are warm and smooth. Contrast is achieved through distinct color blocking rather than granular gradients.
* **Isometric Projection Angle**: The scene uses a standard 2:1 pixel-art isometric projection (dimetric projection). The top surface width is twice its height, fitting perfectly into the existing 64x32 tile grid.
* **Level of Detail**: Very low and intentional. Detail is added through distinct, separated shapes (e.g., a few plus-shaped flowers, simple chevron/triangle grass blades, overlapping circles for tree canopies) rather than realistic textures.

## Procedural Math Adjustments (`cmd/tools/genassets/main.go`)
To align the asset generator with this new aesthetic, the following adjustments must be made:

1. **Eliminate Noise and Stippling**: 
   * **Action**: Remove high-frequency noise algorithms (e.g., `math.Sin(u*18.0) * math.Cos(v*18.0)`) and random `rng.Float64()` stippling found in functions like `generateGrass`, `generateDirt`, and `generateAsphalt`.
   * **Replacement**: Use solid, flat base colors for the ground tiles.

2. **Implement Geometric Overlays**:
   * **Action**: Instead of generating detail by randomly altering individual pixel colors to blend them, draw discrete geometric shapes on top of the flat base.
   * **Replacement**: Add helper functions to draw simple shapes: chevrons (v-shapes) for grass blades, 3x3 "plus" shapes or small circles for flowers, and distinct rounded rectangles for pebbles.

3. **Flat Shading and Extrusions**:
   * **Action**: Update the 3D depth/extrusion logic (e.g., the dirt block underneath the grass).
   * **Replacement**: Use solid blocks of shadow and midtones instead of dark noisy gradients. Add simple rectangular "rocks" embedded in the dirt depth faces, just like the reference art.

4. **Refine Shadows and Highlights**:
   * **Action**: Replace smooth gradient blending (e.g., `blend(c1, c2, t)`) with a stepped, toon-shading approach. Shadows and highlights should be discrete, solid color zones.
   * **Replacement**: Shadows (like drop shadows from trees or stumps) should be solid polygons (e.g., flat, semi-transparent black or a darker base color) rather than soft blurred shapes.

5. **Update Color Primitives and Entities**:
   * **Action**: Shift the hardcoded `color.RGBA` definitions to match the softer, more vibrant pastel values. 
   * **Replacement**: Radically simplify the `drawMatrix` arrays for entities (`generatePlayer`, `generateZombie`, etc.) to remove noisy detail (like "striated muscles") in favor of large, flat, recognizable blocks of color.

## Additional Insights (Reference Art 2)

Based on the second reference image, we can further refine our approach to entities, multi-tile structures, and elevation handling:

### 1. Entities & Characters
* **Minimalist Features:** Entities (like slimes, skulls, and floating creatures) are built from simple geometric primitives (circles, capsules) with flat colors. They lack outlines and complex shading, relying on a single, clean shadow tone on one side to convey volume.
* **Floating & Drop Shadows:** Grounding is achieved using simple, detached oval drop shadows underneath entities. This reinforces the clean vector style without cluttering the ground plane.

### 2. Multi-Tile Structures & Environment Objects
* **Pure Geometry:** Props like trees, stumps, and signs are constructed from pure geometric forms (e.g., cones, cylinders, spheres). Trees feature perfectly smooth conical, droplet, or spherical canopies resting on simple cylindrical trunks.
* **Distinct Overlays:** Elements remain distinct rather than blended. Grass tufts, mushrooms, and pebbles are clearly defined, colored shapes placed cleanly on top of the base terrain.

### 3. Elevation & Terrain
* **Solid Elevation Blocks:** Terrain height is represented by distinct solid blocks. Elevation changes (like hills or cliffs) clearly show the top surface layer (grass) and a solid, contrasting under-layer (dirt) in sheer vertical steps.
* **Consistent Flat Shading:** The vertical faces of terrain blocks are flat-shaded based on orientation. The left vertical face is typically a lighter shade, while the right vertical face is noticeably darker, establishing a clear, consistent global light source without gradients.
* **Smooth Ramps:** Sloped terrain is rendered as a continuous, smooth ramp connecting two elevation levels, maintaining the distinct flat shading for the vertical dirt edges underneath the slope.
