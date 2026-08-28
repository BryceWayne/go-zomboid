# Handoff Report: Weapon HUD, UI & Combat Test Suite (Milestone 4)

## 1. Observation

### 1.1 Existing HUD & UI Rendering
- **Location**: `internal/game/game.go:572-612` (State extraction) and `internal/game/game.go:887-962` (UI rendering).
- **Current Weapon HUD Implementation** (`internal/game/game.go:934-940`):
  ```go
  // Weapon Text (Repositioned to Y: 95)
  if hasWeapon {
      ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Weapon: EQUIPPED (Durability: %d) (Press SPACE/X to attack)", playerDurability), 10, 95)
  } else {
      ebitenutil.DebugPrintAt(screen, "Weapon: NONE (Press SPACE/X to shove zombies back)", 10, 95)
  }
  ```
- **Observations on Weapon State in HUD**:
  1. `DrawSystem.Draw()` queries `ecs.Player` for `hasWeapon = p.WeaponEquipped` and `playerDurability = p.WeaponDurability`.
  2. It does not extract or display the specific `p.WeaponType` (`"axe"`, `"shotgun"`, `"weapon"`).
  3. When a shotgun is equipped, it does not scan `player.Inventory` to show remaining `"ammo"` count.
  4. The facing indicator at lines 860-866 only differentiates armed vs unarmed (`op.ColorScale.Scale(1, 0, 0, 0.7)` vs `(1, 1, 0, 0.7)`), without weapon-specific color coding (e.g. orange for shotgun, fire-red for axe).

### 1.2 Existing ECS Data Structures & Inventory
- **Location**: `internal/ecs/components.go:28-47`
- `ecs.Player`: Contains `Health`, `Hunger`, `Thirst`, `Inventory` (`[]string`), `WeaponEquipped` (`bool`), `WeaponDurability` (`int`), `ArmorEquipped` (`bool`), `ArmorType` (`string`), `ArmorDefense` (`float64`), `ArmorDurability` (`int`), `ArmorMaxDurability` (`int`), `InfectionResist` (`float64`), `AttackCooldown` (`int`), `Dead` (`bool`), `Infected` (`bool`), `FacingX`, `FacingY`.
- `PROJECT.md:71` specifies adding `WeaponType string` to `ecs.Player` to identify the active weapon archetype (`"axe"`, `"shotgun"`, `"weapon"`).

### 1.3 Combat Mechanics Baseline & Test Coverage
- **Location**: `internal/game/game.go:277-320` (Inventory use) and `345-388` (Combat processing).
- **Current Test Files**:
  - `internal/game/game_test.go`: Tests coordinate projection (`TestWorldToIso`) and game initialization / contextual spawns.
  - `internal/game/armor_test.go` and `internal/game/armor_empirical_challenge_test.go`: 11+ unit and empirical tests verifying armor equipping, deflection probability, durability degradation, and infection drain mitigation.
- **Combat Test Gap**: No dedicated `internal/game/combat_test.go` exists yet to rigorously validate:
  1. Fire Axe multi-target cleave simultaneous kills and single durability cost.
  2. Fire Axe 32px extended reach vs Spiked Bat 24px reach.
  3. Shotgun ammo requirement and single-ammo inventory consumption per shot.
  4. Shotgun $\pm 22.5^\circ$ spread cone geometry up to 160px reach.
  5. Shotgun out-of-ammo dry-fire behavior (no durability loss, no zombie kills, no noise alert).
  6. Shotgun 400px acoustic noise pulse alerting wandering horde zombies into active chase.
  7. Weapon durability degradation to 0 and seamless fallback to unarmed fist shove/stun.

---

## 2. Logic Chain

### 2.1 Weapon HUD Status Updates in `DrawSystem.Draw()`
1. **State Extraction**:
   In `DrawSystem.Draw()` (`internal/game/game.go:580-608`):
   Extract `playerWeaponType = p.WeaponType` alongside `hasWeapon = p.WeaponEquipped`, `playerDurability = p.WeaponDurability`, and `playerInventory = p.Inventory`.

2. **Shotgun Ammo Counting**:
   When `playerWeaponType == "shotgun"`, iterate through `playerInventory` to count occurrences of `"ammo"`:
   ```go
   ammoCount := 0
   for _, item := range playerInventory {
       if item == "ammo" {
           ammoCount++
       }
   }
   ```

3. **Formatted Weapon Text Rules**:
   - **Equipped Weapon (`hasWeapon && playerDurability > 0`)**:
     - Extract normalized uppercase name: `wType := strings.ToUpper(playerWeaponType)`. If empty, default to `"WEAPON"`.
     - **If Shotgun**:
       `fmt.Sprintf("Weapon: %s (%d hits | Ammo: %d)", wType, playerDurability, ammoCount)`
     - **If Axe or Club**:
       `fmt.Sprintf("Weapon: %s (%d hits)", wType, playerDurability)`
   - **Unarmed / Fists (`!hasWeapon || playerDurability <= 0`)**:
     `"Weapon: NONE (Fists)"`

4. **HUD Position Layout**:
   - Health Bar: $Y = 10$
   - Hunger Bar: $Y = 35$
   - Thirst Bar: $Y = 55$
   - Armor Bar: $Y = 75$
   - Weapon Text: $Y = 95$
   - Infected Status: $Y = 115$

5. **Visual Weapon Reticle Color Scaling** (`DrawSystem.Draw:860-866`):
   - Shotgun: Orange highlight `op.ColorScale.Scale(1.0, 0.6, 0.2, 0.8)`
   - Axe: Fire-Red highlight `op.ColorScale.Scale(1.0, 0.2, 0.2, 0.8)`
   - Club: Standard Red highlight `op.ColorScale.Scale(1.0, 0.0, 0.0, 0.7)`
   - Unarmed: Yellow shove indicator `op.ColorScale.Scale(1.0, 1.0, 0.0, 0.7)`

### 2.2 Comprehensive Combat Test Suite Design (`internal/game/combat_test.go`)
The test suite is structured into 12 deterministic test functions:

1. `TestCombat_ECSComponentWeaponFields`:
   - Validates that `ecs.Player` correctly stores and initializes `WeaponEquipped`, `WeaponType`, and `WeaponDurability`.

2. `TestCombat_EquipWeaponsFromInventory`:
   - Validates slot key usage (1-9) for `"weapon"` (durability 5), `"axe"` (durability 12), and `"shotgun"` (durability 15), verifying inventory item removal and attack cooldown application.

3. `TestCombat_AxeCleaveMultiTargetKill`:
   - Spawns 3 zombies in front of player within the 32px reach arc (e.g. angle $-15^\circ, 0^\circ, +15^\circ$).
   - Executes axe attack.
   - Asserts: All 3 zombies are killed/removed in a single swing, durability decrements from 12 to 11 (1 swing = 1 durability, regardless of targets killed), and `WeaponEquipped` remains true.

4. `TestCombat_AxeVsBatReachComparison`:
   - Places a zombie at distance $X = 152\text{px}$ from player at $(100, 100)$ (distance to bat attack center is $28\text{px} > 24\text{px}$, distance to axe attack center is $24\text{px} < 32\text{px}$).
   - Asserts bat misses while axe hits and kills.

5. `TestCombat_UnarmedFistShove`:
   - Player with `WeaponEquipped = false` attacks adjacent zombie ($< 24\text{px}$).
   - Asserts: Zombie entity is NOT deleted, `StunTimer` is set to 45, and pushback velocity is applied: $\vec{V} = (FacingX \times 5.0, FacingY \times 5.0)$.

6. `TestCombat_ShotgunAmmoRequirementAndConsumption`:
   - Player with shotgun and inventory `["ammo", "ammo", "food"]` fires at zombie.
   - Asserts: Exactly 1 `"ammo"` item is consumed (inventory becomes `["ammo", "food"]`), durability decreases by 1 ($15 \to 14$), and zombie in cone is killed.

7. `TestCombat_ShotgunConeReachHit`:
   - Tests geometric spread cone ($\pm 22.5^\circ$, reach $160\text{px}$, threshold $\cos(22.5^\circ) \approx 0.92388$).
   - Spawns 5 zombies:
     - Z1: $(200, 100)$ (direct center $0^\circ$, $100\text{px}$) $\to$ Killed.
     - Z2: $(220, 120)$ (flank $9.5^\circ$, $121\text{px}$) $\to$ Killed.
     - Z3: $(200, 180)$ (flank $38.7^\circ$, $128\text{px}$) $\to$ Spared (outside cone).
     - Z4: $(300, 100)$ (direct center $0^\circ$, $200\text{px}$) $\to$ Spared (out of range).
     - Z5: $(50, 100)$ (behind player) $\to$ Spared.
   - Asserts precise angular and radial filtering.

8. `TestCombat_ShotgunOutOfAmmoDryFire`:
   - Player with shotgun and NO ammo in inventory `["food", "water"]` attempts to fire.
   - Asserts: No zombies are killed, durability is NOT decremented (remains 15), inventory is untouched, attack cooldown is applied (30 frames), and no gunshot noise pulse is emitted.

9. `TestCombat_ShotgunNoisePulseAlertsSwarm`:
   - Player at $(200, 200)$ fires shotgun with ammo.
   - Evaluates 3 wandering zombies (`Chasing: false`, `WanderTimer > 0`):
     - Z1 at $(350, 200)$ (distance $150\text{px} \le 400\text{px}$) $\to$ Alerted (`Chasing: true`, `WanderTimer: 0`).
     - Z2 at $(550, 200)$ (distance $350\text{px} \le 400\text{px}$) $\to$ Alerted (`Chasing: true`, `WanderTimer: 0`).
     - Z3 at $(750, 200)$ (distance $550\text{px} > 400\text{px}$) $\to$ Unaffected (`Chasing: false`, `WanderTimer: 100`).
   - Asserts acoustic radius boundary $400.0\text{px}$ works reliably.

10. `TestCombat_WeaponDurabilityBreakdownOnZeroHits`:
    - Tests Club, Axe, and Shotgun starting at durability 1.
    - Final hit decrements durability to 0.
    - Asserts: `WeaponEquipped` transitions to `false`, `WeaponType` transitions to `""`, and durability resets to 0.

11. `TestCombat_MultiHitDegradationLoop`:
    - Full degradation loop:
      - Club: 5 hits until destruction.
      - Axe: 12 hits until destruction.
      - Shotgun: 15 hits until destruction.
    - Asserts monotonic $-1$ decrements and breakage on final hit.

12. `TestCombat_HUDFormattingAndAmmoCount`:
    - Table-driven test verifying HUD string formatting across all archetypes:
      - Unarmed: `"Weapon: NONE (Fists)"`
      - Club: `"Weapon: WEAPON (5 hits)"`
      - Axe: `"Weapon: AXE (12 hits)"`
      - Shotgun with 3 Ammo: `"Weapon: SHOTGUN (15 hits | Ammo: 3)"`
      - Shotgun with 0 Ammo: `"Weapon: SHOTGUN (8 hits | Ammo: 0)"`
      - Broken weapon: `"Weapon: NONE (Fists)"`

---

## 3. Caveats

1. **Ebitengine Headless Testing**: Ebitengine graphical and audio rendering routines operate in non-blocking / mockable headless mode during `go test ./...`. All unit tests execute purely in-memory against the Ark ECS world without opening desktop windows.
2. **Inventory Ammo Stacking**: In the current design, ammunition is stored as individual discrete `"ammo"` item entries in `player.Inventory` (each box providing 1 shot or 1 ammo token). Future expansions could support an internal integer `WeaponAmmo` clip pool, but counting discrete `"ammo"` items maintains consistency with the 9-slot inventory contract.
3. **Sound Playback During Tests**: `assets.PlaySound` handles nil/headless audio contexts safely without throwing errors.

---

## 4. Conclusion & Concrete Implementation Deliverables

### 4.1 Deliverable 1: `internal/ecs/components.go`
Add `WeaponType string` to `ecs.Player`:
```go
type Player struct {
	Health             float64
	Hunger             float64
	Thirst             float64
	Inventory          []string
	WeaponEquipped     bool
	WeaponType         string // "weapon", "axe", "shotgun"
	WeaponDurability   int
	ArmorEquipped      bool
	ArmorType          string
	ArmorDefense       float64
	ArmorDurability    int
	ArmorMaxDurability int
	InfectionResist    float64
	AttackCooldown     int
	Dead               bool
	Infected           bool
	FacingX            float64
	FacingY            float64
}
```

### 4.2 Deliverable 2: `internal/game/game.go` Weapon HUD Updates
In `DrawSystem.Draw()`:
```go
// 1. Variable extraction in player query:
var playerWeaponType string
pq := arkecs.NewFilter2[ecs.Player, ecs.Position](s.world).Query()
for pq.Next() {
    p, pPos := pq.Get()
    // ...
    hasWeapon = p.WeaponEquipped
    playerWeaponType = p.WeaponType
    playerDurability = p.WeaponDurability
    playerInventory = p.Inventory
    // ...
}

// 2. Facing Reticle Color Scaling (lines 860-866):
if hasWeapon {
    if playerWeaponType == "shotgun" {
        op.ColorScale.Scale(1, 0.6, 0.2, 0.8) // Orange for shotgun
    } else if playerWeaponType == "axe" {
        op.ColorScale.Scale(1, 0.2, 0.2, 0.8) // Red-orange for axe
    } else {
        op.ColorScale.Scale(1, 0, 0, 0.7)     // Red for club
    }
} else {
    op.ColorScale.Scale(1, 1, 0, 0.7)         // Yellow for fists
}

// 3. Weapon Status Text (lines 934-940):
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
```

### 4.3 Deliverable 3: Complete Proposed Artifacts
The complete, ready-to-run files have been authored in this subagent folder:
- **Patch file**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/proposed_m4_hud.patch`
- **Unit test suite**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/proposed_combat_test.go`

---

## 5. Verification Method

To independently verify the Weapon HUD calculations and the complete combat test suite:

1. **Run Unit Tests across all packages**:
   ```bash
   CC=gcc go test -v ./...
   ```
   *Expected Outcome*: All test packages pass with exit code 0.

2. **Verify Combat Test Suite Directly**:
   ```bash
   CC=gcc go test -v -run "TestCombat_.*" ./internal/game
   ```
   *Expected Outcome*: All 12 combat tests pass, verifying axe cleave, shotgun ammo consumption, spread cone reach, out-of-ammo dry fire, 400px noise pulse swarm alert, and durability breakdown.

3. **Verify Interactive HUD Rendering in Game Loop**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   *Expected Outcome*:
   - Unarmed state displays `"Weapon: NONE (Fists)"` at $Y = 95$.
   - Equipping Spiked Bat (slot key) displays `"Weapon: WEAPON (5 hits)"`.
   - Equipping Fire Axe (slot key) displays `"Weapon: AXE (12 hits)"`.
   - Equipping Shotgun (slot key) displays `"Weapon: SHOTGUN (15 hits | Ammo: N)"` where $N$ reflects total `"ammo"` items in inventory.
   - Firing shotgun decrements ammo count on HUD dynamically.
