# Milestone 4 Handoff: Weapon Expansion & Combat Mechanics Implementation

## 1. Observation

### 1.1 Existing Component Definition
In `internal/ecs/components.go:28-48`:
`ecs.Player` struct originally had `WeaponEquipped bool` and `WeaponDurability int`, but was missing `WeaponType string` to distinguish weapon archetypes (`"axe"`, `"shotgun"`, `"weapon"`).

### 1.2 Existing Input & Combat Processing
In `internal/game/game.go:277-388`:
- Inventory slot equipping (keys 1-9) previously only recognized `"food"`, `"water"`, `"weapon"`, and `"armor"`, without handling `"axe"`, `"shotgun"`, or preserving `"ammo"`.
- Combat previously used fixed 24.0px reach and hit radius without distinction for weapon types or ammo requirements, and did not support shotgun spread cone raycasting or 400px acoustic noise pulse.
- HUD weapon line at Y=95 rendered generic `"Weapon: EQUIPPED (Durability: %d)"` without weapon archetype names or shotgun ammo counters.

### 1.3 Execution and Verification Evidence
Running `CC=gcc go test -v -run "TestCombat_.*" ./internal/game`:
```
=== RUN   TestCombat_ECSComponentWeaponFields
--- PASS: TestCombat_ECSComponentWeaponFields (0.00s)
=== RUN   TestCombat_EquipWeaponsFromInventory
=== RUN   TestCombat_EquipWeaponsFromInventory/weapon
=== RUN   TestCombat_EquipWeaponsFromInventory/axe
=== RUN   TestCombat_EquipWeaponsFromInventory/shotgun
--- PASS: TestCombat_EquipWeaponsFromInventory (0.00s)
=== RUN   TestCombat_AxeCleaveMultiTargetKill
--- PASS: TestCombat_AxeCleaveMultiTargetKill (0.00s)
=== RUN   TestCombat_AxeVsBatReachComparison
--- PASS: TestCombat_AxeVsBatReachComparison (0.00s)
=== RUN   TestCombat_UnarmedFistShove
--- PASS: TestCombat_UnarmedFistShove (0.00s)
=== RUN   TestCombat_ShotgunAmmoRequirementAndConsumption
--- PASS: TestCombat_ShotgunAmmoRequirementAndConsumption (0.00s)
=== RUN   TestCombat_ShotgunConeReachHit
--- PASS: TestCombat_ShotgunConeReachHit (0.00s)
=== RUN   TestCombat_ShotgunOutOfAmmoDryFire
--- PASS: TestCombat_ShotgunOutOfAmmoDryFire (0.00s)
=== RUN   TestCombat_ShotgunNoisePulseAlertsSwarm
--- PASS: TestCombat_ShotgunNoisePulseAlertsSwarm (0.00s)
=== RUN   TestCombat_WeaponDurabilityBreakdownOnZeroHits
=== RUN   TestCombat_WeaponDurabilityBreakdownOnZeroHits/Club
=== RUN   TestCombat_WeaponDurabilityBreakdownOnZeroHits/Axe
=== RUN   TestCombat_WeaponDurabilityBreakdownOnZeroHits/Shotgun
--- PASS: TestCombat_WeaponDurabilityBreakdownOnZeroHits (0.00s)
=== RUN   TestCombat_MultiHitDegradationLoop
=== RUN   TestCombat_MultiHitDegradationLoop/weapon
=== RUN   TestCombat_MultiHitDegradationLoop/axe
=== RUN   TestCombat_MultiHitDegradationLoop/shotgun
--- PASS: TestCombat_MultiHitDegradationLoop (0.00s)
=== RUN   TestCombat_HUDFormattingAndAmmoCount
=== RUN   TestCombat_HUDFormattingAndAmmoCount/Unarmed_Fists
=== RUN   TestCombat_HUDFormattingAndAmmoCount/Spiked_Club_Equipped
=== RUN   TestCombat_HUDFormattingAndAmmoCount/Fire_Axe_Equipped
=== RUN   TestCombat_HUDFormattingAndAmmoCount/Shotgun_with_3_Ammo
=== RUN   TestCombat_HUDFormattingAndAmmoCount/Shotgun_with_0_Ammo
=== RUN   TestCombat_HUDFormattingAndAmmoCount/Broken_Weapon_Fallback
--- PASS: TestCombat_HUDFormattingAndAmmoCount (0.00s)
=== RUN   TestCombat_AmmoNotDirectlyEquippable
--- PASS: TestCombat_AmmoNotDirectlyEquippable (0.00s)
=== RUN   TestCombat_ReEquipOverridesStats
--- PASS: TestCombat_ReEquipOverridesStats (0.00s)
=== RUN   TestCombat_ShotgunDiagonalFacingSpreadCone
--- PASS: TestCombat_ShotgunDiagonalFacingSpreadCone (0.00s)
=== RUN   TestCombat_FireAxeWideAngleLateralCleave
--- PASS: TestCombat_FireAxeWideAngleLateralCleave (0.00s)
PASS
```
Running `CC=gcc go test -v ./...`:
All tests pass across all packages (`cmd/tools/genassets`, `internal/assets`, `internal/game`, `internal/game/world`).

Running `CC=gcc go build -o bin/game ./cmd/game`:
Compiles cleanly with exit code 0.

---

## 2. Logic Chain

1. **ECS Data Model**:
   - Added `WeaponType string` to `ecs.Player` in `internal/ecs/components.go`.
   - Allows differentiation of weapon behaviors: `"weapon"` (bat), `"axe"` (fire axe), `"shotgun"` (shotgun), and `""` (unarmed).

2. **Inventory Hotbar Usage**:
   - In `internal/game/game.go:processInputAndCombat()`:
     - `"weapon"`: equips bat (`WeaponEquipped = true`, `WeaponType = "weapon"`, `WeaponDurability = 5`).
     - `"axe"`: equips axe (`WeaponEquipped = true`, `WeaponType = "axe"`, `WeaponDurability = 12`).
     - `"shotgun"`: equips shotgun (`WeaponEquipped = true`, `WeaponType = "shotgun"`, `WeaponDurability = 15`).
     - `"ammo"`: ignored on hotbar press (`used = false`), preserving ammo in inventory for shotgun firing.

3. **Melee & Ranged Combat Processing**:
   - **Shotgun**:
     - Scans `player.Inventory` for `"ammo"`.
     - When ammo is found: consumes 1 `"ammo"` item, decrements durability (breaks at 0), plays `HitSound`, evaluates spread cone ($\le 160\text{px}$, $\cos \theta \ge 0.9238795325112867$ corresponding to $\pm 22.5^\circ$, point-blank $< 24\text{px}$) removing all hit zombies, and emits a 400.0px acoustic noise pulse setting `z.Chasing = true` and `z.WanderTimer = 0` for all horde zombies within 400px.
     - When out of ammo: plays `ShoveSound` (mechanical click), performs close-quarters 24px butt shove (`StunTimer = 45`, knockback velocity), without durability loss or noise pulse.
   - **Fire Axe**:
     - Extended 32.0px reach and 32.0px radius cleave sweep, kills all zombies in sweep, decrements 1 durability on hit, plays `HitSound`.
   - **Spiked Bat / Club**:
     - 24.0px reach and 24.0px radius, kills zombies in reach, decrements 1 durability on hit, plays `HitSound`.
   - **Unarmed**:
     - 24.0px reach, applies stun (`StunTimer = 45`) and pushback velocity (`player.Facing * 5.0`), plays `ShoveSound`.
   - **Weapon Breakage**:
     - When durability reaches 0, transitions to unarmed: `WeaponEquipped = false`, `WeaponType = ""`, `WeaponDurability = 0`.

4. **HUD & Reticle UI Updates**:
   - Reticle color scale: orange `(1.0, 0.6, 0.2, 0.8)` for shotgun, red-orange `(1.0, 0.2, 0.2, 0.8)` for axe, red `(1.0, 0.0, 0.0, 0.7)` for bat, yellow `(1.0, 1.0, 0.0, 0.7)` for fists.
   - Weapon HUD text at $Y = 95$:
     - Shotgun: `Weapon: SHOTGUN (%d hits | Ammo: %d)`
     - Axe: `Weapon: AXE (%d hits)`
     - Bat/Club: `Weapon: WEAPON (%d hits)`
     - Unarmed: `Weapon: NONE (Fists)`

5. **Deterministic Unit Test Suite**:
   - `internal/game/combat_test.go` implements 16 comprehensive unit tests validating component fields, equipping, axe cleave, reach comparison, unarmed shove, shotgun ammo consumption, spread cone math, dry fire shove, 400px noise alert, zero durability breakdown, multi-hit loops, HUD formatting, and edge cases.

---

## 3. Caveats

No caveats. All mechanics conform exactly to `PROJECT.md` interface contracts and design specifications. All headless unit tests and full graphical builds pass without errors.

---

## 4. Conclusion

Milestone 4 (Weapon Expansion & Combat Mechanics Implementation) is fully completed and verified.
Modified files:
- `internal/ecs/components.go`
- `internal/game/game.go`
- `internal/game/combat_test.go`

---

## 5. Verification Method

1. Run unit test suite:
   ```bash
   CC=gcc go test -v ./...
   ```
   *Expected Output*: Exit code 0, all tests pass.

2. Run combat unit tests specifically:
   ```bash
   CC=gcc go test -v -run "TestCombat_.*" ./internal/game
   ```
   *Expected Output*: 16/16 tests pass.

3. Compile game binary:
   ```bash
   CC=gcc go build -o bin/game ./cmd/game
   ```
   *Expected Output*: Exit code 0, binary created at `bin/game`.
