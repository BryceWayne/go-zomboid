package world

import "math"



type TileType int

const (
	TileGrass TileType = iota
	TileWall
	TileDirt
	TileWoodFloor
	TileTree
)

const TileSize = 32

type Map struct {
	Width, Height int
	Tiles         []TileType
	Visible       []bool
	Explored      []bool
}

func NewMap(width, height int) *Map {
	m := &Map{
		Width:    width,
		Height:   height,
		Tiles:    make([]TileType, width*height),
		Visible:  make([]bool, width*height),
		Explored: make([]bool, width*height),
	}
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileGrass)
			}
		}
	}
	
	// Create a dirt road
	for y := 5; y < height-5; y++ {
		m.SetTile(20, y, TileDirt)
		m.SetTile(21, y, TileDirt)
	}

	// Create a wooden house
	for y := 10; y < 20; y++ {
		for x := 10; x < 18; x++ {
			if x == 10 || x == 17 || y == 10 || y == 19 {
				if !(x == 17 && y == 15) { // Doorway
					m.SetTile(x, y, TileWall)
				} else {
					m.SetTile(x, y, TileWoodFloor)
				}
			} else {
				m.SetTile(x, y, TileWoodFloor)
			}
		}
	}
	
	// Randomly spawn some trees
	importRand := true
	if importRand {
		for i := 0; i < 40; i++ {
			tx := 5 + (i * 7) % 35
			ty := 5 + (i * 13) % 40
			if m.GetTile(tx, ty) == TileGrass {
				m.SetTile(tx, ty, TileTree)
			}
		}
	}

	return m
}

func (m *Map) CalculateFOV(playerX, playerY float64, radiusTiles int) {
	// Reset visibility
	for i := range m.Visible {
		m.Visible[i] = false
	}

	px := int(playerX) / TileSize
	py := int(playerY) / TileSize

	if px < 0 || px >= m.Width || py < 0 || py >= m.Height {
		return
	}

	m.Visible[py*m.Width+px] = true
	m.Explored[py*m.Width+px] = true

	// Cast rays in 360 degrees
	rays := radiusTiles * 8 // Number of rays scales with radius
	for i := 0; i < rays; i++ {
		angle := (float64(i) / float64(rays)) * 2 * 3.1415926535
		dirX := math.Cos(angle)
		dirY := math.Sin(angle)

		cx, cy := float64(px)+0.5, float64(py)+0.5
		for step := 0; step < radiusTiles; step++ {
			cx += dirX
			cy += dirY

			tx, ty := int(cx), int(cy)
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
				break
			}

			m.Visible[ty*m.Width+tx] = true
			m.Explored[ty*m.Width+tx] = true

			// Stop ray if we hit a wall
			if m.GetTile(tx, ty) == TileWall {
				break
			}
		}
	}
}

func (m *Map) GetTile(x, y int) TileType {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return TileWall
	}
	return m.Tiles[y*m.Width+x]
}

func (m *Map) SetTile(x, y int, t TileType) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}
	m.Tiles[y*m.Width+x] = t
}

// Map drawing has been moved to game.go for Y-sorting with entities

func (m *Map) IsColliding(rectX, rectY, rectW, rectH float64) bool {
	minTileX := int(rectX) / TileSize
	minTileY := int(rectY) / TileSize
	maxTileX := int(rectX+rectW) / TileSize
	maxTileY := int(rectY+rectH) / TileSize

	for y := minTileY; y <= maxTileY; y++ {
		for x := minTileX; x <= maxTileX; x++ {
			t := m.GetTile(x, y)
			if t == TileWall || t == TileTree {
				return true
			}
		}
	}
	return false
}
