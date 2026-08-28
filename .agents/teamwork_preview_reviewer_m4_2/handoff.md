# Milestone 4 Review Report: Weapon Expansion & Combat Mechanics

## 1. Observation

1. **ECS Data Model**:
   - `internal/ecs/components.go:28-48`: `ecs.Player` contains `WeaponEquipped bool`, `WeaponType string`, and `WeaponDurability int`.
2. **Combat & Inventory Execution**:
   - `internal/game/game.go:278-331`: Hotbar inventory equipping supports `"weapon"` (bat, 5 durability), `"axe"` (fire axe, 12 durability), `"shotgun"` (shotgun, 15 durability), and preserves `"ammo"` without direct hotbar consumption.
   - `internal/game/game.go:361-450`: Shotgun checks for `"ammo"` in `player.Inventory`, consumes 1 item per shot, applies 160px spread cone raycasting ($\pm 22.5^\circ$, $\cos \theta \ge 0.9238795325112867$) with point-blank kill threshold ($< 24\text{px}$), decrements 1 durability, and broadcasts a 400.0px acoustic noise pulse that alerts wandering zombies (`z.Chasing = true`, `z.WanderTimer = 0`). Dry fire gracefully falls back to mechanical click (`assets.ShoveSound`) and 24px butt shove without durability loss.
   - `internal/game/game.go:451-480`: Fire Axe implements 32.0px reach and 32.0px radius cleave sweep, eliminating all zombies within the arc while consuming 1 durability per swing on hit (0 on miss).
   - `internal/game/game.go:481-529`: Spiked Bat implements 24.0px reach and 24.0px radius single/multi-hit. Unarmed shove applies 24.0px reach knockback and stun (`StunTimer = 45`) without destroying entities.
   - `internal/game/game.go:1085-1103`: HUD accurately renders weapon type, durability count, and remaining shotgun ammo count at $Y = 95$ without overlapping other status bars.
3. **Automated Verification**:
   - `CC=gcc go test -v -count=1 ./...` executed with exit code 0 (all test suites passing).
   - `CC=gcc go vet ./...` executed with exit code 0 (0 warnings or errors).
   - `CC=gcc go build -o bin/game ./cmd/game` produced valid binary `bin/game`.
   - `go run ./cmd/tools/genassets` generated all 20 required PNG assets cleanly.

---

## 2. Logic Chain

1. **Vector Normalization & Cone Trigonometry**:
   - Facing vector `(FacingX, FacingY)` is normalized by `Hypot(FacingX, FacingY)`.
   - The dot product `(fx*dx + fy*dy) / dist` accurately yields $\cos \theta$. Testing against $\cos(22.5^\circ) \approx 0.9238795$ creates a precise $45^\circ$ total cone of fire.
   - Normalization fallback handles zero-length facing vectors safely by defaulting to $(1.0, 0.0)$, preventing NaN division.
2. **Point-Blank Protection**:
   - In tight combat situations ($dist < 24\text{px}$), checking $dist < 24.0$ guarantees that zombies overlapping the player model are eliminated by the blast regardless of angle, while zombies outside 24px and off-angle are spared.
3. **Inventory & Ammo Integrity**:
   - Shotgun ammo consumption directly removes the first `"ammo"` slice element in `player.Inventory`.
   - When 0 ammo is present, dry fire executes a defensive butt shove with 30-tick cooldown and 0 durability penalty.
   - Hotbar keys 1-9 do not consume or equip ammo directly, preventing accidental loss.
4. **Cleave Durability Balance**:
   - Cleave attacks decrement durability once per successful swing (`if hitZombies { player.WeaponDurability-- }`), rather than per zombie killed. This allows the Fire Axe (12 durability) to clear hordes efficiently.
5. **HUD Formatting**:
   - Shotgun dynamically computes `ammoCount` from inventory items, rendering `Weapon: SHOTGUN (%d hits | Ammo: %d)`.
   - Non-ranged weapons render `Weapon: %s (%d hits)`, and unarmed state falls back to `Weapon: NONE (Fists)`.

---

## 3. Caveats

No caveats. All edge cases, mathematical constraints, audio triggers, and component lifecycles have been tested and verified across headless unit tests and empirical stress benchmarks.

---

## 4. Conclusion

**Verdict: APPROVE**

Milestone 4 (Weapon Expansion & Combat Mechanics Implementation) is fully verified, robust, and free of defects or regressions. All 6 review dimensions and edge case scenarios pass cleanly.

---

## 5. Verification Method

To independently reproduce and verify all results:

```bash
# 1. Run full test suite with clean cache
CC=gcc go test -v -count=1 ./...

# 2. Run static analysis
CC=gcc go vet ./...

# 3. Run combat empirical stress tests
CC=gcc go test -v -run "TestCombat.*" ./internal/game

# 4. Compile game binary
CC=gcc go build -o bin/game ./cmd/game

# 5. Run asset generation tool
go run ./cmd/tools/genassets
```

---

## Quality Review

### Verdict: APPROVE

### Findings: None (0 Critical, 0 Major, 0 Minor)

### Verified Claims
- ECS `Player` component weapon archetype fields $\rightarrow$ verified via `TestCombat_ECSComponentWeaponFields` $\rightarrow$ PASS
- Inventory hotbar weapon equipping and stat overrides $\rightarrow$ verified via `TestCombat_EquipWeaponsFromInventory` & `TestCombatStress_RapidWeaponSwitchingHotbar` $\rightarrow$ PASS
- Fire Axe multi-target cleave and durability loss $\rightarrow$ verified via `TestCombat_AxeCleaveMultiTargetKill` & `TestCombatStress_DenseHordeDurabilityDepletion` $\rightarrow$ PASS
- Shotgun ammo requirement and spread cone geometry $\rightarrow$ verified via `TestCombat_ShotgunConeReachHit` & `TestCombatStress_PointBlankVsConeAngleHits` $\rightarrow$ PASS
- Shotgun dry fire mechanical click and defensive shove $\rightarrow$ verified via `TestCombat_ShotgunOutOfAmmoDryFire` & `TestCombatStress_EmptyVsFullInventoryAmmoConsumption` $\rightarrow$ PASS
- Acoustic noise pulse (400px) swarm alert $\rightarrow$ verified via `TestCombat_ShotgunNoisePulseAlertsSwarm` $\rightarrow$ PASS
- Zero durability weapon breakdown to fists $\rightarrow$ verified via `TestCombat_WeaponDurabilityBreakdownOnZeroHits` $\rightarrow$ PASS
- HUD string formatting and ammo counting $\rightarrow$ verified via `TestCombatStress_HUDFormattingMatrix` $\rightarrow$ PASS

### Coverage Gaps: None
### Unverified Items: None

---

## Adversarial Review

### Challenge Summary
**Overall Risk Assessment**: LOW

### Challenges & Stress Test Results

1. **Diagonal Facing Vector Normalization**:
   - *Challenge*: Moving diagonally produces `(FacingX, FacingY) = (1, 1)` with length $\sqrt{2}$. If unnormalized, dot product calculations would distort the spread cone.
   - *Result*: Normalization via `math.Hypot` ensures exact $\pm 22.5^\circ$ cone across all diagonal and cardinal angles. $\rightarrow$ PASS
2. **Point-Blank (<24px) vs Spread Cone Angle**:
   - *Challenge*: Zombies flanking or behind the player at distance $< 24\text{px}$ might fail the angle check and bite the player during a shotgun blast.
   - *Result*: Explicit $dist < 24.0$ branch destroys all point-blank threats in $360^\circ$ around player. Targets beyond 24px require strict cone alignment. $\rightarrow$ PASS
3. **Empty vs Full Inventory Ammo Consumption**:
   - *Challenge*: Out-of-ammo firing or rapid firing with 9 ammo boxes might cause slice index out of bounds or negative durability.
   - *Result*: 9 consecutive shots consume ammo elements sequentially down to 0, whereupon dry fire cleanly activates. $\rightarrow$ PASS
4. **Durability Depletion in Dense 50-Zombie Horde**:
   - *Challenge*: Cleaving into a massive horde of 50 zombies could deplete 50 durability on a single swing and instantly break the weapon.
   - *Result*: Per-swing durability decrement decrements durability by exactly 1 per swing regardless of target count. $\rightarrow$ PASS
5. **Hotbar Rapid Weapon Switching**:
   - *Challenge*: Rapidly equipping different weapons could corrupt durability or equip ammo boxes.
   - *Result*: Re-equipping immediately overwrites `WeaponType` and sets new `WeaponDurability`. Hotbar ignores `"ammo"` items. $\rightarrow$ PASS

### Unchallenged Areas: None
