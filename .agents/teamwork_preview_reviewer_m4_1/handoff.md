# Milestone 4 Review & Adversarial Stress-Test Report

## 1. Observation

### 1.1 Direct Source Code Observations
1. **ECS Data Model**:
   - `internal/ecs/components.go:28-48`: `ecs.Player` contains `WeaponEquipped bool`, `WeaponType string`, and `WeaponDurability int` conforming to `PROJECT.md:63-84`.
2. **Inventory Usage & Equipping**:
   - `internal/game/game.go:277-331`:
     - Keys 1-9 check `player.Inventory[useItemIdx]`.
     - `"weapon"` sets `WeaponEquipped = true`, `WeaponType = "weapon"`, `WeaponDurability = 5`.
     - `"axe"` sets `WeaponEquipped = true`, `WeaponType = "axe"`, `WeaponDurability = 12`.
     - `"shotgun"` sets `WeaponEquipped = true`, `WeaponType = "shotgun"`, `WeaponDurability = 15`.
     - `"ammo"` items do not trigger `used = true`, remaining safely in inventory for ranged combat.
3. **Weapon Combat Mechanics**:
   - `internal/game/game.go:357-536`:
     - **Shotgun**:
       - Searches inventory for `"ammo"`. If present, consumes 1 ammo item and decrements durability.
       - Raycasts a spread cone with range 160.0px, cone half-angle $22.5^\circ$ ($\cos \theta \ge 0.9238795325112867$), and point-blank override $< 24.0\text{px}$.
       - Emits a 400.0px acoustic noise pulse alerting all zombies within range (`z.Chasing = true`, `z.WanderTimer = 0`).
       - If out of ammo, triggers a dry-fire butt shove with 24.0px reach, 45-frame stun, and knockback velocity without losing durability or alerting the swarm.
     - **Fire Axe**:
       - Extended 32.0px reach and 32.0px radius cleave sweep, killing all zombies caught in the arc. Decrements 1 durability on hit and breaks to unarmed fists on reaching 0.
     - **Spiked Bat / Club**:
       - 24.0px reach and 24.0px radius melee attack, decrements 1 durability on hit.
     - **Unarmed (Fists)**:
       - 24.0px reach, applies 45-frame stun and $5.0\times$ knockback velocity along player facing vector.
4. **Visual Reticle & HUD**:
   - `internal/game/game.go:1005-1027`:
     - Semi-transparent reticle tints: Orange `(1.0, 0.6, 0.2, 0.8)` for shotgun, Red-Orange `(1.0, 0.2, 0.2, 0.8)` for axe, Red `(1.0, 0.0, 0.0, 0.7)` for club, Yellow `(1.0, 1.0, 0.0, 0.7)` for fists.
   - `internal/game/game.go:1084-1104`:
     - Formats HUD text at Y=95:
       - Shotgun: `Weapon: SHOTGUN (%d hits | Ammo: %d)` (dynamically scans inventory for `"ammo"` count).
       - Axe: `Weapon: AXE (%d hits)`
       - Club: `Weapon: WEAPON (%d hits)`
       - Unarmed: `Weapon: NONE (Fists)`

### 1.2 Automated Test & Build Execution
- Running `CC=gcc go test -v -run "TestCombat_.*" ./internal/game`:
  - 16/16 combat unit tests passed in 0.058s.
- Running `CC=gcc go test -count=1 -v ./...`:
  - All test packages (`cmd/tools/genassets`, `internal/assets`, `internal/game`, `internal/game/world`) passed with exit code 0.
- Running `CC=gcc go build -o bin/game ./cmd/game`:
  - Compiles cleanly into `bin/game` with exit code 0.
- Running `go run ./cmd/tools/genassets`:
  - Executes successfully and generates all 20 required pixel-art texture assets.

---

## 2. Logic Chain

1. **Integrity & Authenticity**:
   - Verified that the source code does not contain hardcoded test fixtures, mocked bypasses, dummy stubs, or simulated outputs.
   - All mechanics execute real ECS component operations, geometric vector math, and state transformations.
2. **Interface Conformance**:
   - Verified that `ecs.Player` matches the schema prescribed in `PROJECT.md`.
   - Verified hotbar input integration, inventory slot removals, and weapon durability cascades.
3. **Adversarial Robustness & Edge Cases**:
   - Zero facing vector division safeguard: `facingLen < 0.001` fallback protects against NaN / division-by-zero during shotgun firing.
   - Point-blank shotgun coverage: Targets closer than 24.0px are hit regardless of angular deviation.
   - Out-of-ammo dry-fire stability: Does not leak durability or trigger zombie horde noise.
   - Lateral cleave coverage: Fire axe hits multiple targets spread across a wide lateral arc in a single swing.
   - Ammo safety: Hotbar activation ignores ammo items, preventing accidental consumption or improper weapon slot states.

---

## 3. Caveats

No caveats. All requirements in `PROJECT.md` and `ORIGINAL_REQUEST.md` for Milestone 4 are fulfilled, fully tested, and stable.

---

## 4. Conclusion

**Verdict: APPROVE**

Milestone 4 (Weapon Expansion & Combat Mechanics Implementation) is fully verified, robust, and ready for integration into Milestone 5.

---

## 5. Verification Method

To reproduce and independently verify the results:

1. **Run full unit test suite**:
   ```bash
   CC=gcc go test -count=1 -v ./...
   ```
   *Expected Output*: Exit code 0, all tests pass.

2. **Run combat unit tests specifically**:
   ```bash
   CC=gcc go test -v -run "TestCombat_.*" ./internal/game
   ```
   *Expected Output*: 16/16 tests pass.

3. **Verify game compilation**:
   ```bash
   CC=gcc go build -o bin/game ./cmd/game
   ```
   *Expected Output*: Exit code 0, binary created at `bin/game`.

4. **Verify procedural asset pipeline**:
   ```bash
   go run ./cmd/tools/genassets
   ```
   *Expected Output*: 20 assets generated in `internal/assets/images`.
