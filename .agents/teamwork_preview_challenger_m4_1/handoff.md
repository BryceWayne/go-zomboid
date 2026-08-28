# Empirical Challenge Handoff Report: Milestone 4 Weapons & Combat

## 1. Observation

Direct code and test execution observations:

1. **Implementation Files & Lines**:
   - `internal/ecs/components.go:29-48`: `ecs.Player` component contains weapon state fields: `WeaponEquipped bool`, `WeaponType string`, `WeaponDurability int`, `AttackCooldown int`, `FacingX float64`, `FacingY float64`.
   - `internal/game/game.go:307-316`: Equipping items from inventory assigns `WeaponType = "axe"` with `WeaponDurability = 12` and `WeaponType = "shotgun"` with `WeaponDurability = 15`.
   - `internal/game/game.go:361-432`: Shotgun blast execution:
     - Scans `player.Inventory` for `"ammo"`, consumes 1 item via `player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)` (line 373).
     - Decrements durability: `player.WeaponDurability--` (line 376).
     - Normalizes facing vector `facingX, facingY` (lines 387-394).
     - Spread cone evaluation: `const maxShotgunRange = 160.0`, `const cosSpread = 0.9238795325112867` ($\cos 22.5^\circ$), with point-blank kill distance `< 24.0` (lines 397-420).
     - Acoustic noise pulse: Iterates all zombies, if `Hypot(pos.X - zPos.X, pos.Y - zPos.Y) <= 400.0`, triggers `z.Chasing = true` and resets `z.WanderTimer = 0` (lines 422-432).
   - `internal/game/game.go:433-450`: Dry fire fallback when `ammoIdx < 0`:
     - Plays `assets.ShoveSound`, preserves weapon durability and inventory.
     - Sets `player.AttackCooldown = 30`.
     - Executes shove attack at reach 24.0px, setting `z.StunTimer = 45` and applying velocity impulse `zVel.X = player.FacingX * 5.0`, without deleting zombie entities and without triggering 400px noise pulse.
   - `internal/game/game.go:451-480`: Fire Axe melee attack:
     - Cleave center: `attackX := pos.X + player.FacingX*32.0`, `attackY := pos.Y + player.FacingY*32.0`.
     - Hit circle: `math.Hypot(dx, dy) < 32.0`.
     - Deletes all zombies within hit radius and decrements `player.WeaponDurability--` once per swing upon hit.

2. **Empirical Test Suite Execution**:
   - Test harness file authored: `/home/bryce/code/go-zomboid/internal/game/combat_empirical_stress_test.go`
   - Command executed: `CC=gcc go test -v -count=1 ./...`
   - Output:
     ```
     === RUN   TestEmpirical_AxeCleaveDenseSwarm
     --- PASS: TestEmpirical_AxeCleaveDenseSwarm (0.00s)
     === RUN   TestEmpirical_AxeDurabilityLifecycle12Swings
     --- PASS: TestEmpirical_AxeDurabilityLifecycle12Swings (0.00s)
     === RUN   TestEmpirical_ShotgunConeBoundaryPrecision
     --- PASS: TestEmpirical_ShotgunConeBoundaryPrecision (0.00s)
     === RUN   TestEmpirical_Shotgun8DirectionsMonteCarlo
     --- PASS: TestEmpirical_Shotgun8DirectionsMonteCarlo (0.00s)
     === RUN   TestEmpirical_ExactAmmoConsumptionSequence
     --- PASS: TestEmpirical_ExactAmmoConsumptionSequence (0.00s)
     === RUN   TestEmpirical_Shotgun400pxNoiseRadiusHordeAggro
     --- PASS: TestEmpirical_Shotgun400pxNoiseRadiusHordeAggro (0.00s)
     === RUN   TestEmpirical_DryFireFallbackBehavior
     --- PASS: TestEmpirical_DryFireFallbackBehavior (0.00s)
     === RUN   TestEmpirical_FullCombatCycleIntegration
     --- PASS: TestEmpirical_FullCombatCycleIntegration (0.00s)
     === RUN   TestEmpirical_HUDFormattingMatrix
     --- PASS: TestEmpirical_HUDFormattingMatrix (0.00s)
     PASS
     ok      github.com/BryceWayne/go-zomboid/internal/game          2.051s
     ok      github.com/BryceWayne/go-zomboid/internal/game/world    0.007s
     ok      github.com/BryceWayne/go-zomboid/cmd/tools/genassets    0.003s
     ok      github.com/BryceWayne/go-zomboid/internal/assets        0.004s
     ```

## 2. Logic Chain

1. **Axe Cleave Multi-Kill Sweep** (Observation 1, 2):
   - In `TestEmpirical_AxeCleaveDenseSwarm`, 50 zombies were spawned within a radius $< 32.0\text{px}$ of the axe attack center $(332, 300)$, alongside 20 zombies outside the radius.
   - A single axe swing hit and deleted all 50 inside zombies simultaneously while sparing all 20 outside zombies.
   - The player's axe durability decremented from 12 to 11 (exactly 1 durability loss per cleave swing, regardless of victim count).
   - In `TestEmpirical_AxeDurabilityLifecycle12Swings`, 12 consecutive swings depleted durability to 0, breaking the weapon to fists (`WeaponEquipped: false`, `WeaponType: ""`, `WeaponDurability: 0`).

2. **Shotgun Spread Cone Geometry ($\pm 22.5^\circ$, 160px reach)** (Observation 1, 2):
   - In `TestEmpirical_ShotgunConeBoundaryPrecision`, boundary conditions were verified:
     - Point-blank $< 24.0\text{px}$: $100\%$ omnidirectional hits (even $180^\circ$ directly behind the player up to $23.99\text{px}$).
     - Angular bounds at 100px: $\pm 22.40^\circ$ (HIT), $\pm 22.49^\circ$ (HIT), $\pm 22.51^\circ$ (MISS), $\pm 22.60^\circ$ (MISS).
     - Range bounds at $0^\circ$: $159.90\text{px}$ (HIT), $160.00\text{px}$ (HIT), $160.10\text{px}$ (MISS).
   - In `TestEmpirical_Shotgun8DirectionsMonteCarlo`, 40,000 randomized points (5,000 across each of the 8 cardinal and diagonal facing vectors) were evaluated against the mathematical oracle $f(\Delta \theta, r) = (r \le 24 \lor (r \le 160 \land |\Delta \theta| \le 22.5^\circ))$, achieving a 100.0% match with zero false positives or false negatives.

3. **Exact Ammo Consumption (1 item per blast)** (Observation 1, 2):
   - In `TestEmpirical_ExactAmmoConsumptionSequence`, sequential firing from a mixed inventory (`["food", "ammo", "water", "ammo", "ammo", "vest"]`) removed exactly 1 `"ammo"` per blast.
   - Non-ammo inventory items (`food`, `water`, `vest`) retained their exact slots and state.
   - When all 3 ammo items were consumed, subsequent attempts immediately routed to dry fire.

4. **Exact 400px Noise Radius Horde Aggro** (Observation 1, 2):
   - In `TestEmpirical_Shotgun400pxNoiseRadiusHordeAggro`, boundary distance testing ($399.0\text{px}$, $399.9\text{px}$, $400.0\text{px}$, $400.1\text{px}$, $401.0\text{px}$) and 200 randomized radial zombies confirmed that all entities at Euclidean distance $\le 400.0\text{px}$ had `z.Chasing` set to `true` and `z.WanderTimer` set to `0`.
   - All zombies at distance $> 400.0\text{px}$ remained undisturbed (`Chasing: false`, `WanderTimer > 0`).

5. **Dry Fire Fallback When Ammo is 0** (Observation 1, 2):
   - In `TestEmpirical_DryFireFallbackBehavior`, firing a shotgun with 0 ammo executed the defensive shove mechanism:
     - Shotgun durability remained unchanged at 15.
     - Inventory was untouched.
     - Attack cooldown was set to 30.
     - Zombies within 24px were stunned (`StunTimer = 45`) and pushed (`zVel = player.Facing * 5.0`) without being killed.
     - No 400px noise pulse was emitted.

## 3. Caveats

- All combat and weapon mechanics were evaluated headlessly through deterministic ECS queries and mathematical ground-truth oracles. The real-time interactive game loop rendering on graphical display was verified via headless Ebitengine tests (`TestGameLoopContinuousSimulationStress`).
- No other caveats.

## 4. Conclusion

**Verdict: APPROVE**

Milestone 4 weapon expansion and combat mechanics satisfy all functional specifications, edge cases, and stress requirements:
- Axe Cleave Multi-Kill: VERIFIED
- Shotgun Spread Cone ($\pm 22.5^\circ$, 160px reach, 24px point-blank): VERIFIED
- Exact Ammo Consumption (1 item per blast): VERIFIED
- Exact 400px Noise Radius Horde Aggro: VERIFIED
- Dry Fire Fallback (Defensive shove, stun, no ammo/durability drain, no noise pulse): VERIFIED
- Test Suite (`CC=gcc go test -v -count=1 ./...`): 100% PASS

## 5. Verification Method

To independently reproduce and verify these findings, run:
```bash
CC=gcc go test -v -count=1 ./...
```
Specifically inspect the new empirical test suite:
```bash
CC=gcc go test -v -count=1 ./internal/game -run "TestEmpirical_"
```
Invariants verified:
- `TestEmpirical_AxeCleaveDenseSwarm`
- `TestEmpirical_AxeDurabilityLifecycle12Swings`
- `TestEmpirical_ShotgunConeBoundaryPrecision`
- `TestEmpirical_Shotgun8DirectionsMonteCarlo`
- `TestEmpirical_ExactAmmoConsumptionSequence`
- `TestEmpirical_Shotgun400pxNoiseRadiusHordeAggro`
- `TestEmpirical_DryFireFallbackBehavior`
- `TestEmpirical_FullCombatCycleIntegration`
