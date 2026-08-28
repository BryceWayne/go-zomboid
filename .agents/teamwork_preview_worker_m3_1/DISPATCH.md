## 2026-08-28T17:33:30Z
You are a Worker subagent (teamwork_preview_worker_m3_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m3_1
Project root: /home/bryce/code/go-zomboid
Original Request: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan: /home/bryce/code/go-zomboid/PROJECT.md

Scope: Milestone 3 - Armor System & Damage Mitigation Implementation

Explorer Handoff References:
- Armor ECS & Equipping: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m3_1/handoff.md
- Damage & Infection Mitigation: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m3_2/handoff.md
- HUD & Test Suite: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m3_3/handoff.md

Write Ownership:
You own and may modify:
- `internal/ecs/components.go`
- `internal/game/game.go`
- `internal/game/armor_test.go`

Tasks:
1. Update `internal/ecs/components.go`:
   - Add armor fields to `ecs.Player`: `ArmorEquipped bool`, `ArmorType string`, `ArmorDefense float64`, `ArmorDurability int`, `ArmorMaxDurability int`, `InfectionResist float64`.
2. Update `internal/game/game.go`:
   - In `processInputAndCombat()`:
     - When slot key (1-9) with `"armor"` or `"vest"` is activated, equip armor: `ArmorEquipped = true`, `ArmorType = "vest"`, `ArmorDefense = 0.50`, `ArmorDurability = 10`, `ArmorMaxDurability = 10`, `InfectionResist = 0.70`. Remove `"armor"` from inventory and set `player.AttackCooldown = 30`.
     - In health drain loop, when `player.Infected`, scale drain by `(1.0 - player.ArmorDefense)` if armored.
   - In `processZombies()`:
     - On zombie contact (`dist < 14.0 && !playerDead`):
       - If `playerComp.ArmorEquipped`:
         - If `!playerComp.Infected`: roll `rand.Float64() < playerComp.InfectionResist` to deflect; if roll fails, set `playerComp.Infected = true`.
         - Decrement `playerComp.ArmorDurability--`.
         - If `playerComp.ArmorDurability <= 0`: break armor (`ArmorEquipped = false`, `ArmorType = ""`, `ArmorDefense = 0`, `ArmorDurability = 0`, `ArmorMaxDurability = 0`, `InfectionResist = 0`).
       - If `!playerComp.ArmorEquipped`: set `playerComp.Infected = true`.
   - In `DrawSystem.Draw()`:
     - Add Armor durability bar below Thirst bar at `Y=75, W=200, H=15` with Steel Blue `color.RGBA{70, 130, 180, 255}`.
     - Add Armor text: `fmt.Sprintf("Armor: %d/%d (Def: %d%%)", player.ArmorDurability, player.ArmorMaxDurability, int(player.ArmorDefense*100))`.
     - Shift Weapon text to `Y=95` and Infected status text to `Y=115`.
     - When drawing player entity, apply metallic armor tint `op.ColorScale.Scale(0.75, 0.85, 1.25, 1.0)` if `player.ArmorEquipped`.
3. Create `internal/game/armor_test.go` implementing comprehensive unit tests for all armor mechanics (ECS fields, equip, re-equip refresh, unarmored direct infection, armored deflection, deflection failure, armor breakage at 0, infection drain mitigation).
4. Run `CC=gcc go test -v ./...` and `CC=gcc go build -o bin/game ./cmd/game` to verify all tests and builds pass.
5. Document your implementation in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m3_1/handoff.md`.
