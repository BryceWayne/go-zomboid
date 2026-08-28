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
			W:     16,
			H:     16,
		},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Spawn contextual loot items from map
	for _, loot := range gameMap.LootSpawns {
		itemMap.NewEntity(
			&ecs.Item{Type: loot.Type},
			&ecs.Position{X: loot.X, Y: loot.Y},
		)
	}


	// Spawn zombies from pre-validated non-colliding coordinates
	for _, zs := range gameMap.ZombieSpawns {
		isRunner := rand.Float64() < 0.2 // 20% chance to be a runner
		speed := 1.0 + rand.Float64()*0.5
		if isRunner {
			speed = 2.2 + rand.Float64()*0.4
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

	g.updateSys.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 15, G: 15, B: 15, A: 255})
	g.drawSys.Draw(screen, g.timeOfDay)
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
	// Update FOV
	pq := s.playerFilter.Query()
	for pq.Next() {
		_, pPos, _ := pq.Get()
		s.gameMap.CalculateFOV(pPos.X, pPos.Y, 15) // 15 tile vision radius
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
		if math.Sqrt(dx*dx + dy*dy) < 16.0 {
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

		speed := 3.0
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
			isoX, isoY := WorldToIso(pos.X, pos.Y)
			camX := isoX - 400
			camY := isoY - 300
			mx, my := ebiten.CursorPosition()
			mouseIsoX := float64(mx) + camX
			mouseIsoY := float64(my) + camY
			mouseWorldX, mouseWorldY := IsoToWorld(mouseIsoX, mouseIsoY)

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

						// Shotgun Spread Cone (Range: 160px, Angle: +-22.5 degrees, Point-blank < 24px)
						const maxShotgunRange = 160.0
						const cosSpread = 0.9238795325112867

						zQuery := s.zombieFilter.Query()
						for zQuery.Next() {
							_, zPos, _ := zQuery.Get()
							ent := zQuery.Entity()

							dx := zPos.X - pos.X
							dy := zPos.Y - pos.Y
							dist := math.Hypot(dx, dy)

							if dist <= maxShotgunRange {
								if dist < 24.0 {
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

						// Acoustic Noise Pulse: Alerts all wandering zombies within 400.0px
						noiseQuery := s.zombieFilter.Query()
						for noiseQuery.Next() {
							z, zPos, _ := noiseQuery.Get()
							zdx := pos.X - zPos.X
							zdy := pos.Y - zPos.Y
							if math.Hypot(zdx, zdy) <= 400.0 {
								z.Chasing = true
								z.WanderTimer = 0
							}
						}
					} else {
						// Dry Fire / Out of Ammo: Mechanical click & defensive butt shove
						assets.PlaySound(assets.ShoveSound)

						attackX := pos.X + player.FacingX*24.0
						attackY := pos.Y + player.FacingY*24.0
						zQuery := s.zombieFilter.Query()
						for zQuery.Next() {
							z, zPos, zVel := zQuery.Get()
							dx := attackX - zPos.X
							dy := attackY - zPos.Y
							if math.Hypot(dx, dy) < 24.0 {
								z.StunTimer = 45
								zVel.X = player.FacingX * 5.0
								zVel.Y = player.FacingY * 5.0
							}
						}
					}
				} else if player.WeaponEquipped && player.WeaponType == "axe" {
					// Fire Axe Melee Attack: Cleave reach 32.0px, radius 32.0px
					attackX := pos.X + player.FacingX*32.0
					attackY := pos.Y + player.FacingY*32.0
					hitZombies := false

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						_, zPos, _ := zQuery.Get()
						ent := zQuery.Entity()

						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 32.0 {
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
					// Standard Melee Attack (Bat/Club): Reach 24.0px, radius 24.0px
					attackX := pos.X + player.FacingX*24.0
					attackY := pos.Y + player.FacingY*24.0
					hitZombies := false

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						_, zPos, _ := zQuery.Get()
						ent := zQuery.Entity()

						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 24.0 {
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
					// Unarmed Shove
					attackX := pos.X + player.FacingX*24.0
					attackY := pos.Y + player.FacingY*24.0

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						z, zPos, zVel := zQuery.Get()
						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 24.0 {
							z.StunTimer = 45
							zVel.X = player.FacingX * 5.0
							zVel.Y = player.FacingY * 5.0
						}
					}
					assets.PlaySound(assets.ShoveSound)
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
		
		if zombie.StunTimer > 0 {
			zombie.StunTimer--
			vel.X *= 0.85 // Apply friction to the shove pushback
			vel.Y *= 0.85
			continue // Skip standard AI logic while stunned
		}

		// Infection Check & Armor Deflection
		if dist < 14.0 && !playerDead {
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

func IsoToWorld(isoX, isoY float64) (wx, wy float64) {
	wx = isoY + isoX/2.0
	wy = isoY - isoX/2.0
	return
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
			
			idx := y*s.gameMap.Width + x
			if !s.gameMap.Visible[idx] && !s.gameMap.Explored[idx] {
				continue
			}

			isoX, isoY := WorldToIso(worldX, worldY)
			drawX := isoX - 32 - camX
			drawY := isoY - 0 - camY

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(drawX, drawY)
			
			if !s.gameMap.Visible[idx] && s.gameMap.Explored[idx] {
				op.ColorScale.Scale(0.2, 0.2, 0.3, 1) // Memory tint
			}

			switch t {
			case world.TileGrass, world.TileTree, world.TileFence, world.TileTent, world.TileElevationBlock, world.TileRamp, world.TileStump, world.TileMushroom, world.TileSign:
				screen.DrawImage(assets.GrassImage, op)
			case world.TileDirt:
				screen.DrawImage(assets.DirtImage, op)
			case world.TileWoodFloor:
				screen.DrawImage(assets.WoodImage, op)
			case world.TileAsphalt:
				screen.DrawImage(assets.AsphaltImage, op)
			case world.TileConcrete, world.TileDebris:
				screen.DrawImage(assets.ConcreteImage, op)
			case world.TileTileFloor:
				screen.DrawImage(assets.TileFloorImage, op)
			}
		}
	}

	type Renderable struct {
		Image *ebiten.Image
		Depth float64
		Op    *ebiten.DrawImageOptions
	}
	var sprites []Renderable

	// Add walls, trees, fences & debris
	for y := 0; y < s.gameMap.Height; y++ {
		for x := 0; x < s.gameMap.Width; x++ {
			t := s.gameMap.GetTile(x, y)
			if t == world.TileWall || t == world.TileTree || t == world.TileFence || t == world.TileDebris ||
				t == world.TileTent || t == world.TileElevationBlock || t == world.TileRamp || t == world.TileStump || t == world.TileMushroom || t == world.TileSign {
				worldX := float64(x * world.TileSize)
				worldY := float64(y * world.TileSize)
				
				dx := worldX - playerX
				dy := worldY - playerY
				if dx*dx + dy*dy > visionRadius*visionRadius {
					continue
				}

				idx := y*s.gameMap.Width + x
				if !s.gameMap.Visible[idx] && !s.gameMap.Explored[idx] {
					continue
				}

				isoX, isoY := WorldToIso(worldX, worldY)
				
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
				}

				if img == nil {
					continue
				}
				
				drawX := isoX - 32 - camX
				drawY := isoY - 32 - camY

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(drawX, drawY)

				if !s.gameMap.Visible[idx] && s.gameMap.Explored[idx] {
					op.ColorScale.Scale(0.2, 0.2, 0.3, 1) // Memory tint
				}

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
		item, iPos := qItem.Get()
		
		dx := iPos.X - playerX
		dy := iPos.Y - playerY
		if dx*dx + dy*dy > visionRadius*visionRadius {
			continue
		}

		tx := int(iPos.X) / world.TileSize
		ty := int(iPos.Y) / world.TileSize
		if tx >= 0 && tx < s.gameMap.Width && ty >= 0 && ty < s.gameMap.Height {
			if !s.gameMap.Visible[ty*s.gameMap.Width+tx] {
				continue // Can't see item in darkness
			}
		}

		isoX, isoY := WorldToIso(iPos.X, iPos.Y)
		drawX := isoX - 8 - camX
		drawY := isoY - 8 - camY

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(drawX, drawY)

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

		sprites = append(sprites, Renderable{
			Image: img,
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

			tx := int(pos.X) / world.TileSize
			ty := int(pos.Y) / world.TileSize
			if tx >= 0 && tx < s.gameMap.Width && ty >= 0 && ty < s.gameMap.Height {
				if !s.gameMap.Visible[ty*s.gameMap.Width+tx] {
					continue // Zombies in darkness are invisible!
				}
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

		sprites = append(sprites, Renderable{
			Image: img,
			Depth: pos.X + pos.Y,
			Op:    op,
		})
	}

	if !playerDead {
		// Draw facing indicator
		targetX := playerX + playerFacingX*20.0
		targetY := playerY + playerFacingY*20.0
		
		isoX, isoY := WorldToIso(targetX, targetY)
		drawX := isoX - 4 - camX
		drawY := isoY - 4 - camY

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(drawX, drawY)
		
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
			op.ColorScale.Scale(1.0, 1.0, 0.0, 0.7)     // Yellow if shove
		}

		sprites = append(sprites, Renderable{
			Image: assets.PlayerImage, // Use player image block for now, scaled down? Wait, we can scale it.
			Depth: targetX + targetY,
			Op:    op,
		})
		
		// Let's actually scale the indicator to be small (4x4)
		op.GeoM.Reset()
		op.GeoM.Scale(0.25, 0.25)
		op.GeoM.Translate(drawX, drawY)
	}

	sort.SliceStable(sprites, func(i, j int) bool {
		return sprites[i].Depth < sprites[j].Depth
	})

	for _, s := range sprites {
		screen.DrawImage(s.Image, s.Op)
	}

	// 5. Lighting / Day-Night Cycle
	// timeOfDay goes 0.0 to 24.0. Noon is 12, Midnight is 0 or 24.
	// alpha goes from 0.0 (noon) to 0.90 (midnight)
	alpha := 0.45 + 0.45*math.Cos((timeOfDay/24.0)*math.Pi*2)
	if alpha > 0.05 {
		vector.DrawFilledRect(screen, 0, 0, 800, 600, color.RGBA{0, 0, 15, uint8(alpha * 255)}, false)
	}

	// 6. UI Rendering
	vector.DrawFilledRect(screen, 10, 10, 200, 20, color.RGBA{100, 0, 0, 255}, false)
	hpWidth := float32(playerHealth / 100.0 * 200.0)
	if hpWidth < 0 { hpWidth = 0 }
	vector.DrawFilledRect(screen, 10, 10, hpWidth, 20, color.RGBA{0, 255, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Health", 15, 12)

	// Hunger Bar
	vector.DrawFilledRect(screen, 10, 35, 200, 15, color.RGBA{100, 50, 0, 255}, false)
	hungerW := float32(playerHunger / 100.0 * 200.0)
	if hungerW < 0 { hungerW = 0 }
	vector.DrawFilledRect(screen, 10, 35, hungerW, 15, color.RGBA{255, 140, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Hunger", 15, 35)

	// Thirst Bar
	vector.DrawFilledRect(screen, 10, 55, 200, 15, color.RGBA{0, 0, 100, 255}, false)
	thirstW := float32(playerThirst / 100.0 * 200.0)
	if thirstW < 0 { thirstW = 0 }
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
