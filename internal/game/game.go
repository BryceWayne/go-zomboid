package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strings"

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
	camera    *Camera
	dm        *DungeonMaster
	updateSys *UpdateSystem
	drawSys   *DrawSystem
	timeOfDay float64
}

func NewGame() *Game {
	assets.InitAudio()
	g := &Game{} 
	g.Reset()
	return g
}

func (g *Game) Reset() {
	g.timeOfDay = 8.0 // Morning!

	w := arkecs.NewWorld()
	gameMap := world.NewMap(100, 100)

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)

	// Create Player at safe designated residential spawn
	playerStartX := gameMap.PlayerSpawn.X
	playerStartY := gameMap.PlayerSpawn.Y

	// Initialize and snap camera immediately to player spawn
	g.camera = NewCamera()
	g.camera.Snap(playerStartX, playerStartY)

	playerMap.NewEntity(
		&ecs.Player{
			Health:    100.0,
			Hunger:    100.0,
			Thirst:    100.0,
			Inventory: []string{},
			FacingX:   1, 
			FacingY:   0,
		},
		&ecs.Position{X: playerStartX, Y: playerStartY},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{
			Color: color.RGBA{R: 0, G: 255, B: 0, A: 255},
			W:     64,
			H:     128,
		},
		&ecs.Collider{Width: 64, Height: 64},
	)

	// Spawn contextual loot items from map
	for _, loot := range gameMap.LootSpawns {
		itemMap.NewEntity(
			&ecs.Item{Type: loot.Type},
			&ecs.Position{X: loot.X, Y: loot.Y},
		)
	}

	// Give the player an antidote near their feet at spawn
	itemMap.NewEntity(
		&ecs.Item{Type: "antidote"},
		&ecs.Position{X: playerStartX + 16, Y: playerStartY + 16},
	)


	// Spawn zombies from pre-validated non-colliding coordinates
	for _, zs := range gameMap.ZombieSpawns {
		isRunner := rand.Float64() < 0.2 // 20% chance to be a runner
		speed := 4.0 + rand.Float64()*2.0
		if isRunner {
			speed = 8.8 + rand.Float64()*1.6
		}

		zombieMap.NewEntity(
			&ecs.Zombie{
				Speed:       speed,
				IsRunner:    isRunner,
				WanderTimer: rand.Intn(120),
			},
			&ecs.Position{X: zs.X, Y: zs.Y},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{
				Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
				W:     64,
				H:     128,
			},
			&ecs.Collider{Width: 64, Height: 64},
		)
	}

	g.world = w
	g.gameMap = gameMap
	g.dm = NewDungeonMaster(w, gameMap)
	g.updateSys = NewUpdateSystem(w, gameMap)
	g.updateSys.dm = g.dm
	g.updateSys.camera = g.camera
	g.updateSys.timeOfDay = g.timeOfDay
	g.drawSys = NewDrawSystem(w, gameMap)
	g.drawSys.dm = g.dm
	g.drawSys.camera = g.camera
}

func (g *Game) Update() error {
	// Advance time: Extended day cycle! 5 real minutes for 24 hours. (24 / (60 * 5 * 60 frames))
	g.timeOfDay += 24.0 / (60.0 * 5.0 * 60.0) 
	if g.timeOfDay >= 24.0 {
		g.timeOfDay -= 24.0
	}

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

	g.updateSys.timeOfDay = g.timeOfDay
	g.updateSys.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 15, G: 15, B: 15, A: 255})
	g.drawSys.Draw(screen, g.timeOfDay)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1280, 720
}

// Camera represents an orthogonal 2D top-down camera with smooth exponential lerping towards a target Cartesian position.
type Camera struct {
	X, Y             float64
	TargetX, TargetY float64
	LerpFactor       float64
	Initialized      bool
}

func NewCamera() *Camera {
	return &Camera{
		LerpFactor: 0.10,
	}
}

func (c *Camera) Snap(targetX, targetY float64) {
	c.X = targetX
	c.Y = targetY
	c.TargetX = targetX
	c.TargetY = targetY
	c.Initialized = true
}

func (c *Camera) Update(targetX, targetY float64) {
	c.TargetX = targetX
	c.TargetY = targetY
	if !c.Initialized {
		c.Snap(targetX, targetY)
		return
	}
	dx := c.TargetX - c.X
	dy := c.TargetY - c.Y
	if math.Hypot(dx, dy) < 0.01 {
		c.X = c.TargetX
		c.Y = c.TargetY
		return
	}
	c.X += dx * c.LerpFactor
	c.Y += dy * c.LerpFactor
}

const DefaultZoom = 0.5

func ScreenToIso(screenX, screenY, camX, camY float64) (isoX, isoY float64) {
	isoX = camX + (screenX-640.0)/DefaultZoom
	isoY = camY + (screenY-360.0)/DefaultZoom
	return
}

func ScreenToWorld(screenX, screenY, camX, camY float64) (wx, wy float64) {
	wx = camX + (screenX-640.0)/DefaultZoom
	wy = camY + (screenY-360.0)/DefaultZoom
	return
}

func WorldToScreen(wx, wy, camX, camY float64) (screenX, screenY float64) {
	screenX = (wx-camX)*DefaultZoom + 640.0
	screenY = (wy-camY)*DefaultZoom + 360.0
	return
}

// -- Systems --

type UpdateSystem struct {
	world     *arkecs.World
	gameMap   *world.Map
	camera    *Camera
	dm        *DungeonMaster
	timeOfDay float64

	playerFilter *arkecs.Filter3[ecs.Player, ecs.Position, ecs.Velocity]
	zombieFilter *arkecs.Filter3[ecs.Zombie, ecs.Position, ecs.Velocity]
	moveFilter   *arkecs.Filter3[ecs.Position, ecs.Velocity, ecs.Collider]
	itemFilter   *arkecs.Filter2[ecs.Item, ecs.Position]
}

func NewUpdateSystem(w *arkecs.World, m *world.Map) *UpdateSystem {
	return &UpdateSystem{
		world:        w,
		gameMap:      m,
		dm:           NewDungeonMaster(w, m),
		timeOfDay:    8.0,
		playerFilter: arkecs.NewFilter3[ecs.Player, ecs.Position, ecs.Velocity](w),
		zombieFilter: arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w),
		moveFilter:   arkecs.NewFilter3[ecs.Position, ecs.Velocity, ecs.Collider](w),
		itemFilter:   arkecs.NewFilter2[ecs.Item, ecs.Position](w),
	}
}

func (s *UpdateSystem) Update() {
	var playerPos world.FloatPoint
	var hasPlayer bool

	// Update FOV & Camera
	pq := s.playerFilter.Query()
	for pq.Next() {
		_, pPos, _ := pq.Get()
		playerPos = world.FloatPoint{X: pPos.X, Y: pPos.Y}
		hasPlayer = true
		if s.camera != nil {
			s.camera.Update(pPos.X, pPos.Y)
		}
		s.gameMap.CalculateFOV(pPos.X, pPos.Y, 22) // 22 tile vision radius
	}

	if s.dm != nil && hasPlayer {
		s.dm.Update(s.timeOfDay, playerPos)
	}

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

	pMap := arkecs.NewMap1[ecs.Player](s.world)
	player := pMap.Get(pEnt)
	if player == nil || player.Dead {
		return
	}

	var toRemove []arkecs.Entity
	qItem := s.itemFilter.Query()
	for qItem.Next() {
		item, iPos := qItem.Get()
		ent := qItem.Entity()
		
		dx := pX - iPos.X
		dy := pY - iPos.Y
		if math.Sqrt(dx*dx + dy*dy) < 64.0 {
			if len(player.Inventory) < 9 {
				player.Inventory = append(player.Inventory, item.Type)
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
			drain := 0.05 // Lose 3 health per second (takes ~33 seconds to die)
			if player.ArmorEquipped && player.ArmorDefense > 0 {
				drain *= (1.0 - player.ArmorDefense)
			}
			player.Health -= drain
			if player.Health <= 0 {
				player.Dead = true
			}
		}

		// Drain Hunger and Thirst
		if !player.Dead {
			player.Hunger -= 0.003
			player.Thirst -= 0.005
			
			if player.Hunger < 0 { player.Hunger = 0 }
			if player.Thirst < 0 { player.Thirst = 0 }

			if player.Hunger == 0 || player.Thirst == 0 {
				player.Health -= 0.05
				if player.Health <= 0 {
					player.Dead = true
				}
			}
		}

		speed := 12.0
		vel.X, vel.Y = 0, 0

		if !player.Dead {
			// Inventory Usage (Keys 1-9)
			for i := ebiten.Key1; i <= ebiten.Key9; i++ {
				if ebiten.IsKeyPressed(i) { // Or maybe just detect a tap, but let's use IsKeyPressed and just delete the item immediately. Wait, this will trigger multiple times if held.
				}
			}
			
			// A better way to handle inventory: Inpututil is standard for just pressed.
			// However, since we don't have inpututil imported, we can do a simple loop and remove the item.
			// Actually, inpututil is part of ebiten. Let's import it if we can, or just use a small cooldown or check.
			// Let's just assume one tap removes it.
			
			// Let's implement inventory use
			useItemIdx := -1
			if ebiten.IsKeyPressed(ebiten.Key1) { useItemIdx = 0 }
			if ebiten.IsKeyPressed(ebiten.Key2) { useItemIdx = 1 }
			if ebiten.IsKeyPressed(ebiten.Key3) { useItemIdx = 2 }
			if ebiten.IsKeyPressed(ebiten.Key4) { useItemIdx = 3 }
			if ebiten.IsKeyPressed(ebiten.Key5) { useItemIdx = 4 }
			if ebiten.IsKeyPressed(ebiten.Key6) { useItemIdx = 5 }
			if ebiten.IsKeyPressed(ebiten.Key7) { useItemIdx = 6 }
			if ebiten.IsKeyPressed(ebiten.Key8) { useItemIdx = 7 }
			if ebiten.IsKeyPressed(ebiten.Key9) { useItemIdx = 8 }
			
			if useItemIdx >= 0 && useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30 // Small cooldown so it doesn't instantly consume everything if held
				t := player.Inventory[useItemIdx]
				
				used := false
				if t == "food" && player.Hunger < 100 {
					player.Hunger += 50
					if player.Hunger > 100 { player.Hunger = 100 }
					used = true
				} else if t == "antidote" && player.Infected {
					player.Infected = false
					used = true
				} else if t == "water" && player.Thirst < 100 {
					player.Thirst += 50
					if player.Thirst > 100 { player.Thirst = 100 }
					used = true
				} else if t == "weapon" {
					player.WeaponEquipped = true
					player.WeaponType = "weapon"
					player.WeaponDurability = 5
					used = true
				} else if t == "axe" {
					player.WeaponEquipped = true
					player.WeaponType = "axe"
					player.WeaponDurability = 12
					used = true
				} else if t == "shotgun" {
					player.WeaponEquipped = true
					player.WeaponType = "shotgun"
					player.WeaponDurability = 15
					used = true
				} else if t == "armor" || t == "vest" {
					player.ArmorEquipped = true
					player.ArmorType = "vest"
					player.ArmorDefense = 0.50
					player.ArmorDurability = 10
					player.ArmorMaxDurability = 10
					player.InfectionResist = 0.70
					used = true
				}
				
				if used {
					// Remove item from inventory
					player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
				}
			}

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

			// Mouse Input
			camX := 0.0
			camY := 0.0
			if s.camera != nil {
				camX = s.camera.X
				camY = s.camera.Y
			} else {
				camX = pos.X
				camY = pos.Y
			}
			mx, my := ebiten.CursorPosition()
			mouseWorldX, mouseWorldY := ScreenToWorld(float64(mx), float64(my), camX, camY)

			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				dx := mouseWorldX - pos.X
				dy := mouseWorldY - pos.Y
				dist := math.Hypot(dx, dy)
				if dist > speed {
					vel.X = (dx / dist) * speed
					vel.Y = (dy / dist) * speed
				} else if dist > 0 {
					vel.X = dx
					vel.Y = dy
				}
			}

			// Update facing
			if vel.X != 0 || vel.Y != 0 {
				player.FacingX = vel.X / speed
				player.FacingY = vel.Y / speed
			}

			isAttacking := ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyX) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
				dx := mouseWorldX - pos.X
				dy := mouseWorldY - pos.Y
				dist := math.Hypot(dx, dy)
				if dist > 0.001 {
					player.FacingX = dx / dist
					player.FacingY = dy / dist
				}
			}

			// Combat
			if player.AttackCooldown > 0 {
				player.AttackCooldown--
			}
			
			if isAttacking && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30 // Half second cooldown

				if player.WeaponEquipped && player.WeaponType == "shotgun" {
					// Check for ammo in inventory
					ammoIdx := -1
					for idx, itm := range player.Inventory {
						if itm == "ammo" {
							ammoIdx = idx
							break
						}
					}

					if ammoIdx >= 0 {
						// Consume 1 ammo item
						player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)

						// Deduct shotgun durability
						player.WeaponDurability--
						if player.WeaponDurability <= 0 {
							player.WeaponEquipped = false
							player.WeaponType = ""
							player.WeaponDurability = 0
						}

						// Play gunshot blast sound
						assets.PlaySound(assets.HitSound)

						// Normalize facing vector
						facingLen := math.Hypot(player.FacingX, player.FacingY)
						facingX, facingY := player.FacingX, player.FacingY
						if facingLen < 0.001 {
							facingX, facingY = 1.0, 0.0
						} else {
							facingX /= facingLen
							facingY /= facingLen
						}

						// Shotgun Spread Cone (Range: 640px, Angle: +-22.5 degrees, Point-blank < 96px)
						const maxShotgunRange = 640.0
						const cosSpread = 0.9238795325112867

						zQuery := s.zombieFilter.Query()
						for zQuery.Next() {
							_, zPos, _ := zQuery.Get()
							ent := zQuery.Entity()

							dx := zPos.X - pos.X
							dy := zPos.Y - pos.Y
							dist := math.Hypot(dx, dy)

							if dist <= maxShotgunRange {
								if dist < 96.0 {
									// Point-blank kill
									toRemoveZombies = append(toRemoveZombies, ent)
								} else {
									cosAngle := (facingX*dx + facingY*dy) / dist
									if cosAngle >= cosSpread {
										toRemoveZombies = append(toRemoveZombies, ent)
									}
								}
							}
						}

						// Acoustic Noise Pulse: Alerts all wandering zombies within 1600.0px
						noiseQuery := s.zombieFilter.Query()
						for noiseQuery.Next() {
							z, zPos, _ := noiseQuery.Get()
							zdx := pos.X - zPos.X
							zdy := pos.Y - zPos.Y
							if math.Hypot(zdx, zdy) <= 1600.0 {
								z.Chasing = true
								z.WanderTimer = 0
							}
						}
					} else {
						// Dry Fire / Out of Ammo: Mechanical click & defensive butt shove
						assets.PlaySound(assets.ShoveSound)

						attackX := pos.X + player.FacingX*96.0
						attackY := pos.Y + player.FacingY*96.0
						zQuery := s.zombieFilter.Query()
						for zQuery.Next() {
							z, zPos, zVel := zQuery.Get()
							dx := attackX - zPos.X
							dy := attackY - zPos.Y
							if math.Hypot(dx, dy) < 96.0 {
								z.StunTimer = 45
								zVel.X = player.FacingX * 20.0
								zVel.Y = player.FacingY * 20.0
							}
						}
					}
				} else if player.WeaponEquipped && player.WeaponType == "axe" {
					// Fire Axe Melee Attack: Cleave reach 128.0px, radius 128.0px
					attackX := pos.X + player.FacingX*128.0
					attackY := pos.Y + player.FacingY*128.0
					hitZombies := false

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						_, zPos, _ := zQuery.Get()
						ent := zQuery.Entity()

						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 128.0 {
							hitZombies = true
							toRemoveZombies = append(toRemoveZombies, ent)
						}
					}

					if hitZombies {
						assets.PlaySound(assets.HitSound)
						player.WeaponDurability--
						if player.WeaponDurability <= 0 {
							player.WeaponEquipped = false
							player.WeaponType = ""
							player.WeaponDurability = 0
						}
					} else {
						assets.PlaySound(assets.ShoveSound)
					}
				} else if player.WeaponEquipped {
					// Standard Melee Attack (Bat/Club): Reach 96.0px, radius 96.0px
					attackX := pos.X + player.FacingX*96.0
					attackY := pos.Y + player.FacingY*96.0
					hitZombies := false

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						_, zPos, _ := zQuery.Get()
						ent := zQuery.Entity()

						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 96.0 {
							hitZombies = true
							toRemoveZombies = append(toRemoveZombies, ent)
						}
					}

					if hitZombies {
						assets.PlaySound(assets.HitSound)
						player.WeaponDurability--
						if player.WeaponDurability <= 0 {
							player.WeaponEquipped = false
							player.WeaponType = ""
							player.WeaponDurability = 0
						}
					} else {
						assets.PlaySound(assets.ShoveSound)
					}
				} else {
					// Unarmed Shove: Reach 96.0px, radius 96.0px
					attackX := pos.X + player.FacingX*96.0
					attackY := pos.Y + player.FacingY*96.0

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						z, zPos, zVel := zQuery.Get()
						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 96.0 {
							z.StunTimer = 45
							zVel.X = player.FacingX * 20.0
							zVel.Y = player.FacingY * 20.0
						}
					}
					assets.PlaySound(assets.ShoveSound)
				}
			}
		}
	}

	for _, ent := range toRemoveZombies {
		if s.dm != nil {
			posMap := arkecs.NewMap1[ecs.Position](s.world)
			if zPos := posMap.Get(ent); zPos != nil {
				s.dm.HandleZombieDeath(zPos.X, zPos.Y)
			}
		}
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

	speedMult, noiseMult, visionMult := 1.0, 1.0, 1.0
	if s.dm != nil {
		speedMult, noiseMult, visionMult = s.dm.GetAggressionModifiers(s.timeOfDay)
	}

	noiseRadius := 200.0 * noiseMult
	if playerMoving {
		noiseRadius = 800.0 * noiseMult
	}
	visionRadius := 600.0 * visionMult
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

	separationRadius := 80.0
	separationForce := 8.0

	query := s.zombieFilter.Query()
	for query.Next() {
		zombie, pos, vel := query.Get()
		entityID := query.Entity()
		
		dx := playerX - pos.X
		dy := playerY - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		
		var moveX, moveY float64
		
		if zombie.StunTimer > 0 {
			zombie.StunTimer--
			vel.X *= 0.85 // Apply friction to the shove pushback
			vel.Y *= 0.85
			continue // Skip standard AI logic while stunned
		}

		// Infection Check & Armor Deflection
		if dist < 56.0 && !playerDead {
			pMap := arkecs.NewMap1[ecs.Player](s.world)
			if playerComp := pMap.Get(playerEnt); playerComp != nil {
				if playerComp.ArmorEquipped {
					// Deflection roll against player's InfectionResist (e.g. 0.70 = 70% chance to block infection)
					if !playerComp.Infected {
						if rand.Float64() < playerComp.InfectionResist {
							// Deflected! Armor blocked the zombie bite/scratch.
						} else {
							// Deflection failed! Zombie bite penetrated the armor.
							playerComp.Infected = true
						}
					}
					// Deduct armor durability on contact hit
					playerComp.ArmorDurability--
					if playerComp.ArmorDurability <= 0 {
						// Armor broke under the zombie attack!
						playerComp.ArmorEquipped = false
						playerComp.ArmorType = ""
						playerComp.ArmorDefense = 0.0
						playerComp.ArmorDurability = 0
						playerComp.ArmorMaxDurability = 0
						playerComp.InfectionResist = 0.0
					}
				} else {
					// Unarmored player takes immediate infection on contact
					playerComp.Infected = true
				}
			}
		}

		if dist < noiseRadius || dist < visionRadius {
			zombie.Chasing = true
		} else if dist > 1600.0 || playerDead { 
			zombie.Chasing = false
		}

		effectiveSpeed := zombie.Speed * speedMult
		if zombie.Chasing && dist > 0 {
			moveX = (dx / dist) * effectiveSpeed
			moveY = (dy / dist) * effectiveSpeed
		} else {
			zombie.WanderTimer--
			if zombie.WanderTimer <= 0 {
				zombie.WanderTimer = 60 + rand.Intn(120)
				angle := rand.Float64() * 2 * math.Pi
				zombie.WanderDirX = math.Cos(angle)
				zombie.WanderDirY = math.Sin(angle)
			}
			wanderSpeed := effectiveSpeed * 0.4
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
	camera     *Camera
	dm         *DungeonMaster
	itemFilter *arkecs.Filter2[ecs.Item, ecs.Position]
}

func NewDrawSystem(w *arkecs.World, m *world.Map) *DrawSystem {
	return &DrawSystem{
		world:      w,
		gameMap:    m,
		dm:         NewDungeonMaster(w, m),
		itemFilter: arkecs.NewFilter2[ecs.Item, ecs.Position](w),
	}
}

func WorldToIso(wx, wy float64) (isoX, isoY float64) {
	return wx, wy
}

func IsoToWorld(isoX, isoY float64) (wx, wy float64) {
	return isoX, isoY
}

func (s *DrawSystem) Draw(screen *ebiten.Image, timeOfDay float64) {
	var camX, camY float64
	var playerX, playerY float64
	var playerDead, playerInfected bool
	var playerHealth float64
	var playerHunger float64
	var playerThirst float64
	var playerInventory []string
	var hasWeapon bool
	var playerWeaponType string
	var attackCooldown int
	var playerDurability int
	var playerFacingX, playerFacingY float64
	var hasArmor bool
	var armorDurability int
	var armorMaxDurability int
	var armorDefense float64

	pq := arkecs.NewFilter2[ecs.Player, ecs.Position](s.world).Query()
	for pq.Next() {
		p, pPos := pq.Get()
		playerX, playerY = pPos.X, pPos.Y
		playerDead = p.Dead
		playerInfected = p.Infected
		playerHealth = p.Health
		playerHunger = p.Hunger
		playerThirst = p.Thirst
		playerInventory = p.Inventory
		hasWeapon = p.WeaponEquipped
		playerWeaponType = p.WeaponType
		attackCooldown = p.AttackCooldown
		playerDurability = p.WeaponDurability
		playerFacingX = p.FacingX
		playerFacingY = p.FacingY
		hasArmor = p.ArmorEquipped
		armorDurability = p.ArmorDurability
		armorMaxDurability = p.ArmorMaxDurability
		armorDefense = p.ArmorDefense
	}

	if s.camera != nil {
		camX = s.camera.X
		camY = s.camera.Y
	} else {
		camX, camY = playerX, playerY
	}

	visionRadius := 2200.0

	// 2. Draw Ground Tiles (Rectangular 2D orthogonal, seamless, zero gaps)
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
			if dx*dx+dy*dy > visionRadius*visionRadius {
				continue
			}

			idx := y*s.gameMap.Width + x
			if !s.gameMap.Visible[idx] && !s.gameMap.Explored[idx] {
				continue
			}

			var img *ebiten.Image
			switch t {
			case world.TileGrass, world.TileTree, world.TileFence, world.TileTent, world.TileElevationBlock, world.TileRamp, world.TileStump, world.TileMushroom, world.TileSign,
				world.TileBench, world.TileChest, world.TileSculpture, world.TileBush, world.TileFlower, world.TileStone:
				img = assets.GrassImage
			case world.TileDirt:
				img = assets.DirtImage
			case world.TileWoodFloor:
				img = assets.WoodImage
			case world.TileAsphalt:
				img = assets.AsphaltImage
			case world.TileConcrete, world.TileDebris:
				img = assets.ConcreteImage
			case world.TileTileFloor:
				img = assets.TileFloorImage
			}

			if img == nil {
				continue
			}

			bounds := img.Bounds()
			imgW := float64(bounds.Dx())
			imgH := float64(bounds.Dy())
			if imgW <= 0 || imgH <= 0 {
				continue
			}

			scaleX := (float64(world.TileSize) / imgW) * DefaultZoom
			scaleY := (float64(world.TileSize) / imgH) * DefaultZoom
			screenX, screenY := WorldToScreen(worldX, worldY, camX, camY)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(screenX, screenY)

			if !s.gameMap.Visible[idx] && s.gameMap.Explored[idx] {
				op.ColorScale.Scale(0.2, 0.2, 0.3, 1) // Memory tint
			}

			screen.DrawImage(img, op)
		}
	}

	type Renderable struct {
		Image *ebiten.Image
		Depth float64
		Op    *ebiten.DrawImageOptions
	}
	var sprites []Renderable

	// Add walls, trees, fences & debris (Top-down anchors and Depth = worldY + TileSize)
	for y := 0; y < s.gameMap.Height; y++ {
		for x := 0; x < s.gameMap.Width; x++ {
			t := s.gameMap.GetTile(x, y)
			if t == world.TileWall || t == world.TileTree || t == world.TileFence || t == world.TileDebris ||
				t == world.TileTent || t == world.TileElevationBlock || t == world.TileRamp || t == world.TileStump || t == world.TileMushroom || t == world.TileSign ||
				t == world.TileBench || t == world.TileChest || t == world.TileSculpture || t == world.TileBush || t == world.TileFlower || t == world.TileStone {
				worldX := float64(x * world.TileSize)
				worldY := float64(y * world.TileSize)

				dx := worldX - playerX
				dy := worldY - playerY
				if dx*dx+dy*dy > visionRadius*visionRadius {
					continue
				}

				idx := y*s.gameMap.Width + x
				if !s.gameMap.Visible[idx] && !s.gameMap.Explored[idx] {
					continue
				}

				var img *ebiten.Image
				switch t {
				case world.TileWall:
					img = assets.WallImage
				case world.TileTree:
					img = assets.TreeImage
				case world.TileFence:
					img = assets.FenceImage
				case world.TileDebris:
					img = assets.DebrisImage
				case world.TileTent:
					img = assets.TentImage
				case world.TileElevationBlock:
					img = assets.ElevationBlockImage
				case world.TileRamp:
					img = assets.ElevationRampImage
				case world.TileStump:
					img = assets.StumpImage
				case world.TileMushroom:
					img = assets.MushroomImage
				case world.TileSign:
					img = assets.SignImage
				case world.TileBench:
					img = assets.BenchImage
				case world.TileChest:
					img = assets.ChestImage
				case world.TileSculpture:
					img = assets.SculptureImage
				case world.TileBush:
					img = assets.BushImage
				case world.TileFlower:
					img = assets.FlowerImage
				case world.TileStone:
					img = assets.StoneImage
				}

				if img == nil {
					continue
				}

				bounds := img.Bounds()
				imgW := float64(bounds.Dx())
				imgH := float64(bounds.Dy())
				if imgW <= 0 || imgH <= 0 {
					continue
				}

				scaleX := (float64(world.TileSize) / imgW) * DefaultZoom
				scaleY := (float64(world.TileSize) / imgH) * DefaultZoom
				screenX, screenY := WorldToScreen(worldX, worldY, camX, camY)

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(scaleX, scaleY)
				op.GeoM.Translate(screenX, screenY)

				if !s.gameMap.Visible[idx] && s.gameMap.Explored[idx] {
					op.ColorScale.Scale(0.2, 0.2, 0.3, 1) // Memory tint
				}

				sprites = append(sprites, Renderable{
					Image: img,
					Depth: worldY + float64(world.TileSize),
					Op:    op,
				})
			}
		}
	}

	// Add items (Centered at item position, Depth = iPos.Y)
	qItem := s.itemFilter.Query()
	for qItem.Next() {
		item, iPos := qItem.Get()

		dx := iPos.X - playerX
		dy := iPos.Y - playerY
		if dx*dx+dy*dy > visionRadius*visionRadius {
			continue
		}

		tx := int(iPos.X) / world.TileSize
		ty := int(iPos.Y) / world.TileSize
		if tx >= 0 && tx < s.gameMap.Width && ty >= 0 && ty < s.gameMap.Height {
			if !s.gameMap.Visible[ty*s.gameMap.Width+tx] {
				continue // Can't see item in darkness
			}
		}

		img := assets.WeaponImage
		switch item.Type {
		case "food":
			img = assets.FoodImage
		case "water":
			img = assets.WaterImage
		case "weapon":
			img = assets.WeaponImage
		case "axe":
			img = assets.AxeImage
		case "shotgun":
			img = assets.ShotgunImage
		case "ammo":
			img = assets.AmmoImage
		case "armor":
			img = assets.ArmorImage
		case "antidote":
			img = assets.AntidoteImage
		}

		if img == nil {
			continue
		}

		bounds := img.Bounds()
		imgW := float64(bounds.Dx())
		imgH := float64(bounds.Dy())

		screenX, screenY := WorldToScreen(iPos.X, iPos.Y, camX, camY)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-imgW/2.0, -imgH/2.0)
		op.GeoM.Scale(DefaultZoom, DefaultZoom)
		op.GeoM.Translate(screenX, screenY)

		sprites = append(sprites, Renderable{
			Image: img,
			Depth: iPos.Y,
			Op:    op,
		})
	}

	// Add entities (Centered at entity position, Depth = pos.Y)
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
			if dx*dx+dy*dy > visionRadius*visionRadius {
				continue
			}

			tx := int(pos.X) / world.TileSize
			ty := int(pos.Y) / world.TileSize
			if tx >= 0 && tx < s.gameMap.Width && ty >= 0 && ty < s.gameMap.Height {
				if !s.gameMap.Visible[ty*s.gameMap.Width+tx] {
					continue // Zombies in darkness are invisible!
				}
			}
		}

		img := assets.PlayerImage
		op := &ebiten.DrawImageOptions{}

		if isPlayer {
			if playerDead {
				op.ColorScale.Scale(0.3, 0.3, 0.3, 1)
			} else if playerInfected {
				pulse := 0.5 + 0.5*math.Sin(float64(playerHealth))
				op.ColorScale.Scale(float32(pulse), 1, float32(pulse), 1)
			} else if hasArmor {
				// Tactical Armor Visual Tint: Steel-Blue metallic highlight
				op.ColorScale.Scale(0.75, 0.85, 1.25, 1.0)
			}
			if attackCooldown > 20 {
				op.ColorScale.Scale(2, 2, 2, 1)
			}
		} else if z := zMap.Get(ent); z != nil {
			if z.IsRunner {
				img = assets.RunnerImage
			} else {
				img = assets.ZombieImage
			}
			// Stun tint
			if z.StunTimer > 0 {
				op.ColorScale.Scale(1.5, 1.5, 2.5, 1) // Bluish flash when stunned
			}
		}

		if img == nil {
			continue
		}

		bounds := img.Bounds()
		imgW := float64(bounds.Dx())
		imgH := float64(bounds.Dy())

		screenX, screenY := WorldToScreen(pos.X, pos.Y, camX, camY)

		op.GeoM.Translate(-imgW/2.0, -imgH/2.0)
		op.GeoM.Scale(DefaultZoom, DefaultZoom)
		op.GeoM.Translate(screenX, screenY)

		sprites = append(sprites, Renderable{
			Image: img,
			Depth: pos.Y,
			Op:    op,
		})
	}

	if !playerDead {
		// Draw facing indicator (Depth = targetY)
		targetX := playerX + playerFacingX*80.0
		targetY := playerY + playerFacingY*80.0

		screenX, screenY := WorldToScreen(targetX, targetY, camX, camY)

		op := &ebiten.DrawImageOptions{}

		// Semi-transparent indicator
		if hasWeapon {
			if playerWeaponType == "shotgun" {
				op.ColorScale.Scale(1.0, 0.6, 0.2, 0.8) // Orange for shotgun
			} else if playerWeaponType == "axe" {
				op.ColorScale.Scale(1.0, 0.2, 0.2, 0.8) // Red-orange for axe
			} else {
				op.ColorScale.Scale(1.0, 0.0, 0.0, 0.7) // Red for club/weapon
			}
		} else {
			op.ColorScale.Scale(1.0, 1.0, 0.0, 0.7) // Yellow if shove
		}

		bounds := assets.PlayerImage.Bounds()
		imgW := float64(bounds.Dx())
		imgH := float64(bounds.Dy())

		op.GeoM.Translate(-imgW/2.0, -imgH/2.0)
		op.GeoM.Scale(DefaultZoom*0.5, DefaultZoom*0.5)
		op.GeoM.Translate(screenX, screenY)

		sprites = append(sprites, Renderable{
			Image: assets.PlayerImage,
			Depth: targetY,
			Op:    op,
		})
	}

	// Depth Sorting: Natural top-down vertical occlusion (lower Y drawn in front of higher Y)
	sort.SliceStable(sprites, func(i, j int) bool {
		return sprites[i].Depth < sprites[j].Depth
	})

	for _, s := range sprites {
		screen.DrawImage(s.Image, s.Op)
	}

	// 4. Bezier Curve Combat Swoosh Trails
	if !playerDead && attackCooldown > 16 {
		s.DrawAttackSwingArc(screen, playerX, playerY, playerFacingX, playerFacingY, playerWeaponType, attackCooldown, camX, camY)
	}

	// 5. Lighting / Day-Night Cycle
	var tint color.RGBA
	var alpha float64
	if s.dm != nil {
		tint, alpha = s.dm.GetAmbientLighting(timeOfDay)
	} else {
		alpha = 0.45 + 0.45*math.Cos((timeOfDay/24.0)*math.Pi*2)
		tint = color.RGBA{R: 0, G: 0, B: 15, A: 255}
	}
	if alpha > 0.01 {
		overlayColor := color.RGBA{
			R: tint.R,
			G: tint.G,
			B: tint.B,
			A: uint8(alpha * 255),
		}
		vector.DrawFilledRect(screen, 0, 0, 1280, 720, overlayColor, false)
	}

	// 6. UI Rendering
	vector.DrawFilledRect(screen, 10, 10, 200, 20, color.RGBA{100, 0, 0, 255}, false)
	hpWidth := float32(playerHealth / 100.0 * 200.0)
	if hpWidth < 0 {
		hpWidth = 0
	}
	vector.DrawFilledRect(screen, 10, 10, hpWidth, 20, color.RGBA{0, 255, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Health", 15, 12)

	// Hunger Bar
	vector.DrawFilledRect(screen, 10, 35, 200, 15, color.RGBA{100, 50, 0, 255}, false)
	hungerW := float32(playerHunger / 100.0 * 200.0)
	if hungerW < 0 {
		hungerW = 0
	}
	vector.DrawFilledRect(screen, 10, 35, hungerW, 15, color.RGBA{255, 140, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Hunger", 15, 35)

	// Thirst Bar
	vector.DrawFilledRect(screen, 10, 55, 200, 15, color.RGBA{0, 0, 100, 255}, false)
	thirstW := float32(playerThirst / 100.0 * 200.0)
	if thirstW < 0 {
		thirstW = 0
	}
	vector.DrawFilledRect(screen, 10, 55, thirstW, 15, color.RGBA{0, 191, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Thirst", 15, 55)

	// Armor Bar (Y: 75, H: 15)
	vector.DrawFilledRect(screen, 10, 75, 200, 15, color.RGBA{30, 45, 60, 255}, false)
	armorW := float32(0)
	if armorMaxDurability > 0 && armorDurability > 0 {
		armorW = float32(float64(armorDurability) / float64(armorMaxDurability) * 200.0)
		if armorW > 200 {
			armorW = 200
		}
	}
	if armorW > 0 {
		vector.DrawFilledRect(screen, 10, 75, armorW, 15, color.RGBA{70, 130, 180, 255}, false) // Steel Blue
	}
	if hasArmor && armorDurability > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Armor: %d/%d (Def: %d%%)", armorDurability, armorMaxDurability, int(armorDefense*100)), 15, 75)
	} else {
		ebitenutil.DebugPrintAt(screen, "Armor: NONE", 15, 75)
	}

	// Weapon Text (Repositioned to Y: 95)
	if hasWeapon && playerDurability > 0 {
		wType := strings.ToUpper(playerWeaponType)
		if wType == "" {
			wType = "WEAPON"
		}
		if playerWeaponType == "shotgun" {
			ammoCount := 0
			for _, item := range playerInventory {
				if item == "ammo" {
					ammoCount++
				}
			}
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Weapon: %s (%d hits | Ammo: %d)", wType, playerDurability, ammoCount), 10, 95)
		} else {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Weapon: %s (%d hits)", wType, playerDurability), 10, 95)
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "Weapon: NONE (Fists)", 10, 95)
	}

	// Inventory UI
	ebitenutil.DebugPrintAt(screen, "Inventory (Press 1-9 to use):", 550, 10)
	for i := 0; i < 9; i++ {
		// Draw slot background
		y := 30 + (i * 25)
		vector.DrawFilledRect(screen, 550, float32(y), 200, 20, color.RGBA{50, 50, 50, 200}, false)

		text := fmt.Sprintf("%d: [Empty]", i+1)
		if i < len(playerInventory) {
			text = fmt.Sprintf("%d: %s", i+1, playerInventory[i])
		}
		ebitenutil.DebugPrintAt(screen, text, 555, y+2)
	}

	// Infected Status (Repositioned to Y: 115)
	if playerInfected && !playerDead {
		ebitenutil.DebugPrintAt(screen, "INFECTED!", 10, 115)
	}
	if playerDead {
		ebitenutil.DebugPrintAt(screen, "YOU DIED\n(Press 'R' to restart)", 350, 280)
	}
}

// DrawAttackSwingArc renders dynamic Bezier curves for melee weapons, shove shockwaves, and shotgun blasts.
func (s *DrawSystem) DrawAttackSwingArc(screen *ebiten.Image, playerX, playerY float64, facingX, facingY float64, weaponType string, attackCooldown int, camX, camY float64) {
	if attackCooldown <= 16 || attackCooldown > 30 {
		return
	}

	// Normalized swing progress t in [0.0, 1.0] across 14 active swing frames (30 -> 16)
	t := float64(30-attackCooldown) / 14.0
	alpha := float32((1.0 - t) * (1.0 - t))
	if alpha <= 0.01 {
		return
	}

	facingLen := math.Hypot(facingX, facingY)
	if facingLen < 0.001 {
		facingX, facingY = 1.0, 0.0
	} else {
		facingX /= facingLen
		facingY /= facingLen
	}
	baseAngle := math.Atan2(facingY, facingX)

	if weaponType == "shotgun" {
		// Shotgun: Radial blast cone lines + expanding wavefront Bezier arc
		const halfSpread = 0.39269908169872414 // 22.5 deg in radians

		// 1. Expanding wavefront Bezier arc
		waveRadius := float64(80.0 + t*240.0)
		wAngle0 := baseAngle - halfSpread
		wAngle1 := baseAngle + halfSpread

		wp0X := playerX + waveRadius*0.85*math.Cos(wAngle0)
		wp0Y := playerY + waveRadius*0.85*math.Sin(wAngle0)
		wp1X := playerX + waveRadius*1.15*math.Cos(baseAngle)
		wp1Y := playerY + waveRadius*1.15*math.Sin(baseAngle)
		wp2X := playerX + waveRadius*0.85*math.Cos(wAngle1)
		wp2Y := playerY + waveRadius*0.85*math.Sin(wAngle1)

		sx0, sy0 := WorldToScreen(wp0X, wp0Y, camX, camY)
		sx1, sy1 := WorldToScreen(wp1X, wp1Y, camX, camY)
		sx2, sy2 := WorldToScreen(wp2X, wp2Y, camX, camY)

		screen0X := float32(sx0)
		screen0Y := float32(sy0)
		screen1X := float32(sx1)
		screen1Y := float32(sy1)
		screen2X := float32(sx2)
		screen2Y := float32(sy2)

		var arcPath vector.Path
		arcPath.MoveTo(screen0X, screen0Y)
		arcPath.QuadTo(screen1X, screen1Y, screen2X, screen2Y)

		// Outer glow
		outerColor := color.RGBA{R: 255, G: 140, B: 0, A: uint8(255 * alpha * 0.7)}
		outerDrawOpts := &vector.DrawPathOptions{AntiAlias: true}
		outerDrawOpts.ColorScale.ScaleWithColor(outerColor)
		vector.StrokePath(screen, &arcPath, &vector.StrokeOptions{
			Width:    5.0,
			LineCap:  vector.LineCapRound,
			LineJoin: vector.LineJoinRound,
		}, outerDrawOpts)

		// Inner core
		coreColor := color.RGBA{R: 255, G: 255, B: 200, A: uint8(255 * alpha * 0.95)}
		coreDrawOpts := &vector.DrawPathOptions{AntiAlias: true}
		coreDrawOpts.ColorScale.ScaleWithColor(coreColor)
		vector.StrokePath(screen, &arcPath, &vector.StrokeOptions{
			Width:    1.5,
			LineCap:  vector.LineCapRound,
			LineJoin: vector.LineJoinRound,
		}, coreDrawOpts)

		// 2. Muzzle blast radial traces
		psx, psy := WorldToScreen(playerX, playerY, camX, camY)
		pxScreen := float32(psx)
		pyScreen := float32(psy)

		numRays := 7
		for i := 0; i < numRays; i++ {
			rayFraction := float64(i)/float64(numRays-1) - 0.5 // -0.5 to +0.5
			rayAngle := baseAngle + rayFraction*2.0*halfSpread
			rayDist := float64(120.0 + (1.0-math.Abs(rayFraction)*0.4)*200.0)

			rxWorld := playerX + rayDist*math.Cos(rayAngle)
			ryWorld := playerY + rayDist*math.Sin(rayAngle)
			rsx, rsy := WorldToScreen(rxWorld, ryWorld, camX, camY)
			rxScreen := float32(rsx)
			ryScreen := float32(rsy)

			rayColor := color.RGBA{R: 255, G: 200, B: 50, A: uint8(255 * alpha * 0.6)}
			vector.StrokeLine(screen, pxScreen, pyScreen, rxScreen, ryScreen, 1.5, rayColor, true)
		}
		return
	}

	// Melee weapons & shove: Sweeping quadratic Bezier arc
	var deltaTheta float64
	var rIn, rApex, rOut float64
	var outerWidth, coreWidth float32
	var outerColor, coreColor color.Color

	switch weaponType {
	case "axe":
		// Fire Axe: Wide cleave fiery red-orange
		deltaTheta = 2.0 // ~115 degrees
		rIn = 40.0
		rApex = 140.0
		rOut = 120.0
		outerWidth = 14.0
		coreWidth = 4.0
		outerColor = color.RGBA{R: 255, G: 69, B: 0, A: uint8(255 * alpha * 0.65)}
		coreColor = color.RGBA{R: 255, G: 230, B: 120, A: uint8(255 * alpha * 0.95)}
	case "weapon":
		// Spiked Club / Bat: Cool royal blue / cyan motion trail
		deltaTheta = 1.6 // ~90 degrees
		rIn = 30.0
		rApex = 105.0
		rOut = 90.0
		outerWidth = 10.0
		coreWidth = 3.0
		outerColor = color.RGBA{R: 65, G: 105, B: 225, A: uint8(255 * alpha * 0.55)}
		coreColor = color.RGBA{R: 200, G: 240, B: 255, A: uint8(255 * alpha * 0.9)}
	default:
		// Unarmed shove / fists: Amber / bright shockwave
		deltaTheta = 1.2 // ~70 degrees
		rIn = 20.0
		rApex = 100.0
		rOut = 80.0
		outerWidth = 8.0
		coreWidth = 2.5
		outerColor = color.RGBA{R: 255, G: 191, B: 0, A: uint8(255 * alpha * 0.45)}
		coreColor = color.RGBA{R: 255, G: 255, B: 240, A: uint8(255 * alpha * 0.85)}
	}

	// Compute World Control Points P0, P1, P2
	theta0 := baseAngle - deltaTheta/2.0
	theta1 := baseAngle + deltaTheta/2.0

	p0x := playerX + rIn*math.Cos(theta0)
	p0y := playerY + rIn*math.Sin(theta0)

	p1x := playerX + rApex*math.Cos(baseAngle)
	p1y := playerY + rApex*math.Sin(baseAngle)

	p2x := playerX + rOut*math.Cos(theta1)
	p2y := playerY + rOut*math.Sin(theta1)

	// Transform directly to screen space
	sx0, sy0 := WorldToScreen(p0x, p0y, camX, camY)
	sx1, sy1 := WorldToScreen(p1x, p1y, camX, camY)
	sx2, sy2 := WorldToScreen(p2x, p2y, camX, camY)

	screen0X := float32(sx0)
	screen0Y := float32(sy0)
	screen1X := float32(sx1)
	screen1Y := float32(sy1)
	screen2X := float32(sx2)
	screen2Y := float32(sy2)

	var path vector.Path
	path.MoveTo(screen0X, screen0Y)
	path.QuadTo(screen1X, screen1Y, screen2X, screen2Y)

	// Pass 1: Outer Glow
	outerDrawOpts := &vector.DrawPathOptions{AntiAlias: true}
	outerDrawOpts.ColorScale.ScaleWithColor(outerColor)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{
		Width:    outerWidth * 0.5,
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}, outerDrawOpts)

	// Pass 2: Bright Core
	coreDrawOpts := &vector.DrawPathOptions{AntiAlias: true}
	coreDrawOpts.ColorScale.ScaleWithColor(coreColor)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{
		Width:    coreWidth * 0.5,
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}, coreDrawOpts)
}
