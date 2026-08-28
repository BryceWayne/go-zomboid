package game

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	arkecs "github.com/mlange-42/ark/ecs"
)

type Game struct {
	world     *arkecs.World
	gameMap   *world.Map
	updateSys *UpdateSystem
	drawSys   *DrawSystem
}

func NewGame() *Game {
	g := &Game{}
	g.Reset()
	return g
}

func (g *Game) Reset() {
	w := arkecs.NewWorld()
	gameMap := world.NewMap(50, 50)

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)

	// Create Player
	playerMap.NewEntity(
		&ecs.Player{Health: 100.0, FacingX: 1, FacingY: 0},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{
			Color: color.RGBA{R: 0, G: 255, B: 0, A: 255},
			W:     16,
			H:     16,
		},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Create Weapons on map
	for i := 0; i < 5; i++ {
		itemMap.NewEntity(
			&ecs.Item{Type: "weapon"},
			&ecs.Position{X: float64(100 + rand.Intn(400)), Y: float64(100 + rand.Intn(400))},
		)
	}

	// Create Zombies
	for i := 0; i < 25; i++ {
		isRunner := rand.Float64() < 0.2 // 20% chance to be a runner
		speed := 1.0 + rand.Float64()*0.5
		if isRunner {
			speed = 2.8 + rand.Float64()*0.5
		}

		zombieMap.NewEntity(
			&ecs.Zombie{
				Speed:       speed,
				IsRunner:    isRunner,
				WanderTimer: rand.Intn(120),
			},
			&ecs.Position{X: float64(200 + rand.Intn(800)), Y: float64(200 + rand.Intn(800))},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{
				Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
				W:     16,
				H:     16,
			},
			&ecs.Collider{Width: 16, Height: 16},
		)
	}

	g.world = w
	g.gameMap = gameMap
	g.updateSys = NewUpdateSystem(w, gameMap)
	g.drawSys = NewDrawSystem(w, gameMap)
}

func (g *Game) Update() error {
	// Check for restart if dead
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		var isDead bool
		pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
		for pq.Next() {
			p := pq.Get()
			if p.Dead {
				isDead = true
			}
		}
		if isDead {
			g.Reset()
			return nil
		}
	}

	g.updateSys.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 15, G: 15, B: 15, A: 255})
	g.drawSys.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}

// -- Systems --

type UpdateSystem struct {
	world   *arkecs.World
	gameMap *world.Map
	
	playerFilter *arkecs.Filter3[ecs.Player, ecs.Position, ecs.Velocity]
	zombieFilter *arkecs.Filter3[ecs.Zombie, ecs.Position, ecs.Velocity]
	moveFilter   *arkecs.Filter3[ecs.Position, ecs.Velocity, ecs.Collider]
	itemFilter   *arkecs.Filter2[ecs.Item, ecs.Position]
}

func NewUpdateSystem(w *arkecs.World, m *world.Map) *UpdateSystem {
	return &UpdateSystem{
		world:        w,
		gameMap:      m,
		playerFilter: arkecs.NewFilter3[ecs.Player, ecs.Position, ecs.Velocity](w),
		zombieFilter: arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w),
		moveFilter:   arkecs.NewFilter3[ecs.Position, ecs.Velocity, ecs.Collider](w),
		itemFilter:   arkecs.NewFilter2[ecs.Item, ecs.Position](w),
	}
}

func (s *UpdateSystem) Update() {
	s.processItems()
	s.processInputAndCombat()
	s.processZombies()
	s.processMovement()
}

func (s *UpdateSystem) processItems() {
	var pX, pY float64
	var pEnt arkecs.Entity
	var found bool
	
	pq := s.playerFilter.Query()
	for pq.Next() {
		_, pPos, _ := pq.Get()
		pX, pY = pPos.X, pPos.Y
		pEnt = pq.Entity()
		found = true
	}
	if !found {
		return
	}

	var toRemove []arkecs.Entity
	qItem := s.itemFilter.Query()
	for qItem.Next() {
		item, iPos := qItem.Get()
		ent := qItem.Entity()
		
		dx := pX - iPos.X
		dy := pY - iPos.Y
		if math.Sqrt(dx*dx + dy*dy) < 16.0 {
			if item.Type == "weapon" {
				pMap := arkecs.NewMap1[ecs.Player](s.world)
				if player := pMap.Get(pEnt); player != nil {
					player.WeaponEquipped = true
				}
				toRemove = append(toRemove, ent)
			}
		}
	}

	for _, ent := range toRemove {
		s.world.RemoveEntity(ent)
	}
}

func (s *UpdateSystem) processInputAndCombat() {
	var toRemoveZombies []arkecs.Entity

	query := s.playerFilter.Query()
	for query.Next() {
		player, pos, vel := query.Get()
		
		if player.Dead {
			vel.X, vel.Y = 0, 0
			continue
		}

		if player.Infected {
			player.Health -= 0.5
			if player.Health <= 0 {
				player.Dead = true
			}
		}

		speed := 3.0
		vel.X, vel.Y = 0, 0

		if !player.Dead {
			if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
				vel.Y -= speed
			}
			if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
				vel.Y += speed
			}
			if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
				vel.X -= speed
			}
			if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
				vel.X += speed
			}

			// Update facing
			if vel.X != 0 || vel.Y != 0 {
				player.FacingX = vel.X / speed
				player.FacingY = vel.Y / speed
			}

			// Combat
			if player.AttackCooldown > 0 {
				player.AttackCooldown--
			}
			if ebiten.IsKeyPressed(ebiten.KeySpace) && player.WeaponEquipped && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30 // Half second cooldown

				attackX := pos.X + player.FacingX*24
				attackY := pos.Y + player.FacingY*24
				
				zQuery := s.zombieFilter.Query()
				for zQuery.Next() {
					_, zPos, _ := zQuery.Get()
					ent := zQuery.Entity()
					
					dx := attackX - zPos.X
					dy := attackY - zPos.Y
					if math.Sqrt(dx*dx + dy*dy) < 24.0 { // Hit radius
						toRemoveZombies = append(toRemoveZombies, ent)
					}
				}
			}
		}
	}

	for _, ent := range toRemoveZombies {
		s.world.RemoveEntity(ent)
	}
}

func (s *UpdateSystem) processZombies() {
	var playerX, playerY float64
	var playerMoving, playerDead bool
	var playerEnt arkecs.Entity
	var foundPlayer bool
	
	pq := s.playerFilter.Query()
	for pq.Next() {
		p, pPos, pVel := pq.Get()
		playerX, playerY = pPos.X, pPos.Y
		playerMoving = pVel.X != 0 || pVel.Y != 0
		playerDead = p.Dead
		playerEnt = pq.Entity()
		foundPlayer = true
	}
	
	if !foundPlayer {
		return
	}

	noiseRadius := 50.0
	if playerMoving {
		noiseRadius = 200.0
	}
	visionRadius := 150.0
	if playerDead {
		noiseRadius, visionRadius = 0, 0 
	}

	type zombieData struct {
		id  arkecs.Entity
		pos *ecs.Position
	}
	var others []zombieData
	
	qGather := s.zombieFilter.Query()
	for qGather.Next() {
		_, pos, _ := qGather.Get()
		others = append(others, zombieData{id: qGather.Entity(), pos: pos})
	}

	separationRadius := 20.0
	separationForce := 2.0

	query := s.zombieFilter.Query()
	for query.Next() {
		zombie, pos, vel := query.Get()
		entityID := query.Entity()
		
		dx := playerX - pos.X
		dy := playerY - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		
		var moveX, moveY float64
		
		// Infection Check
		if dist < 14.0 && !playerDead {
			pMap := arkecs.NewMap1[ecs.Player](s.world)
			if playerComp := pMap.Get(playerEnt); playerComp != nil {
				playerComp.Infected = true
			}
		}

		if dist < noiseRadius || dist < visionRadius {
			zombie.Chasing = true
		} else if dist > 400.0 || playerDead { 
			zombie.Chasing = false
		}

		if zombie.Chasing && dist > 0 {
			moveX = (dx / dist) * zombie.Speed
			moveY = (dy / dist) * zombie.Speed
		} else {
			zombie.WanderTimer--
			if zombie.WanderTimer <= 0 {
				zombie.WanderTimer = 60 + rand.Intn(120)
				angle := rand.Float64() * 2 * math.Pi
				zombie.WanderDirX = math.Cos(angle)
				zombie.WanderDirY = math.Sin(angle)
			}
			wanderSpeed := zombie.Speed * 0.4
			moveX = zombie.WanderDirX * wanderSpeed
			moveY = zombie.WanderDirY * wanderSpeed
		}
		
		var sepX, sepY float64
		for _, other := range others {
			if other.id == entityID {
				continue
			}
			
			odx := pos.X - other.pos.X
			ody := pos.Y - other.pos.Y
			odist := math.Sqrt(odx*odx + ody*ody)
			
			if odist > 0 && odist < separationRadius {
				pushStrength := (separationRadius - odist) / separationRadius
				sepX += (odx / odist) * pushStrength * separationForce
				sepY += (ody / odist) * pushStrength * separationForce
			}
		}
		
		vel.X = moveX + sepX
		vel.Y = moveY + sepY
	}
}

func (s *UpdateSystem) processMovement() {
	query := s.moveFilter.Query()
	for query.Next() {
		pos, vel, col := query.Get()
		
		if vel.X != 0 {
			if !s.gameMap.IsColliding(pos.X+vel.X, pos.Y, col.Width, col.Height) {
				pos.X += vel.X
			}
		}
		if vel.Y != 0 {
			if !s.gameMap.IsColliding(pos.X, pos.Y+vel.Y, col.Width, col.Height) {
				pos.Y += vel.Y
			}
		}
	}
}

type DrawSystem struct {
	world      *arkecs.World
	gameMap    *world.Map
	itemFilter *arkecs.Filter2[ecs.Item, ecs.Position]
}

func NewDrawSystem(w *arkecs.World, m *world.Map) *DrawSystem {
	return &DrawSystem{
		world:      w,
		gameMap:    m,
		itemFilter: arkecs.NewFilter2[ecs.Item, ecs.Position](w),
	}
}

func WorldToIso(wx, wy float64) (isoX, isoY float64) {
	isoX = wx - wy
	isoY = (wx + wy) / 2.0
	return
}

func (s *DrawSystem) Draw(screen *ebiten.Image) {
	var camX, camY float64
	var playerX, playerY float64
	var playerDead, playerInfected bool
	var playerHealth float64
	var hasWeapon bool
	var attackCooldown int
	
	pq := arkecs.NewFilter2[ecs.Player, ecs.Position](s.world).Query()
	for pq.Next() {
		p, pPos := pq.Get()
		playerX, playerY = pPos.X, pPos.Y
		playerDead = p.Dead
		playerInfected = p.Infected
		playerHealth = p.Health
		hasWeapon = p.WeaponEquipped
		attackCooldown = p.AttackCooldown
		
		isoX, isoY := WorldToIso(pPos.X, pPos.Y)
		camX = isoX - 400
		camY = isoY - 300
	}

	visionRadius := 250.0

	// 2. Draw Ground Tiles
	for y := 0; y < s.gameMap.Height; y++ {
		for x := 0; x < s.gameMap.Width; x++ {
			t := s.gameMap.GetTile(x, y)
			if t == world.TileWall {
				continue
			}
			
			worldX := float64(x * world.TileSize)
			worldY := float64(y * world.TileSize)

			dx := worldX - playerX
			dy := worldY - playerY
			if dx*dx + dy*dy > visionRadius*visionRadius {
				continue
			}

			isoX, isoY := WorldToIso(worldX, worldY)
			
			drawX := isoX - 32 - camX
			drawY := isoY - 0 - camY

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(drawX, drawY)

			switch t {
			case world.TileGrass, world.TileTree:
				screen.DrawImage(assets.GrassImage, op)
			case world.TileDirt:
				screen.DrawImage(assets.DirtImage, op)
			case world.TileWoodFloor:
				screen.DrawImage(assets.WoodImage, op)
			}
		}
	}

	type Renderable struct {
		Image *ebiten.Image
		Depth float64
		Op    *ebiten.DrawImageOptions
	}
	var sprites []Renderable

	// Add walls
	for y := 0; y < s.gameMap.Height; y++ {
		for x := 0; x < s.gameMap.Width; x++ {
			t := s.gameMap.GetTile(x, y)
			if t == world.TileWall || t == world.TileTree {
				worldX := float64(x * world.TileSize)
				worldY := float64(y * world.TileSize)
				
				dx := worldX - playerX
				dy := worldY - playerY
				if dx*dx + dy*dy > visionRadius*visionRadius {
					continue
				}

				isoX, isoY := WorldToIso(worldX, worldY)
				
				var img *ebiten.Image
				if t == world.TileWall {
					img = assets.WallImage
				} else {
					img = assets.TreeImage
				}
				
				drawX := isoX - 32 - camX
				drawY := isoY - 32 - camY

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(drawX, drawY)

				sprites = append(sprites, Renderable{
					Image: img,
					Depth: worldX + worldY,
					Op:    op,
				})
			}
		}
	}

	// Add items
	qItem := s.itemFilter.Query()
	for qItem.Next() {
		_, iPos := qItem.Get()
		
		dx := iPos.X - playerX
		dy := iPos.Y - playerY
		if dx*dx + dy*dy > visionRadius*visionRadius {
			continue
		}

		isoX, isoY := WorldToIso(iPos.X, iPos.Y)
		drawX := isoX - 8 - camX
		drawY := isoY - 8 - camY

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(drawX, drawY)

		sprites = append(sprites, Renderable{
			Image: assets.WeaponImage,
			Depth: iPos.X + iPos.Y,
			Op:    op,
		})
	}

	// Add entities
	pMap := arkecs.NewMap1[ecs.Player](s.world)
	zMap := arkecs.NewMap1[ecs.Zombie](s.world)
	
	drawQuery := arkecs.NewFilter2[ecs.Position, ecs.Sprite](s.world).Query()
	for drawQuery.Next() {
		pos, _ := drawQuery.Get()
		ent := drawQuery.Entity()
		
		isPlayer := pMap.Get(ent) != nil
		if !isPlayer {
			dx := pos.X - playerX
			dy := pos.Y - playerY
			if dx*dx + dy*dy > visionRadius*visionRadius {
				continue
			}
		}

		isoX, isoY := WorldToIso(pos.X, pos.Y)
		
		drawX := isoX - 8 - camX
		drawY := isoY - 32 - camY

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(drawX, drawY)

		img := assets.PlayerImage
		
		if isPlayer {
			if playerDead {
				op.ColorScale.Scale(0.3, 0.3, 0.3, 1)
			} else if playerInfected {
				pulse := 0.5 + 0.5*math.Sin(float64(playerHealth))
				op.ColorScale.Scale(float32(pulse), 1, float32(pulse), 1)
			}
			
			// Draw swing effect
			if attackCooldown > 20 {
				// Flash white while swinging
				op.ColorScale.Scale(2, 2, 2, 1)
			}
		} else if z := zMap.Get(ent); z != nil {
			if z.IsRunner {
				img = assets.RunnerImage
			} else {
				img = assets.ZombieImage
			}
		}

		sprites = append(sprites, Renderable{
			Image: img,
			Depth: pos.X + pos.Y,
			Op:    op,
		})
	}

	sort.SliceStable(sprites, func(i, j int) bool {
		return sprites[i].Depth < sprites[j].Depth
	})

	for _, s := range sprites {
		screen.DrawImage(s.Image, s.Op)
	}

	// 4. UI Rendering
	
	// Health Bar Background
	vector.DrawFilledRect(screen, 10, 10, 200, 20, color.RGBA{100, 0, 0, 255}, false)
	// Health Bar Fill
	hpWidth := float32(playerHealth / 100.0 * 200.0)
	if hpWidth < 0 {
		hpWidth = 0
	}
	vector.DrawFilledRect(screen, 10, 10, hpWidth, 20, color.RGBA{0, 255, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Health", 15, 12)

	if hasWeapon {
		ebitenutil.DebugPrintAt(screen, "Weapon: EQUIPPED (Press SPACE to attack)", 10, 35)
	} else {
		ebitenutil.DebugPrintAt(screen, "Weapon: NONE (Find a weapon on the map!)", 10, 35)
	}

	if playerInfected && !playerDead {
		ebitenutil.DebugPrintAt(screen, "INFECTED!", 10, 55)
	}
	if playerDead {
		ebitenutil.DebugPrintAt(screen, "YOU DIED\n(Press 'R' to restart)", 350, 280)
	}
}
