# TEST READY: Milestone 2 & 3 E2E Test Suite Specification

**Status**: READY / PASSING (100% test pass rate across all tiers)
**Test Runner**: `CC=gcc go test -v ./...`
**Race Detector**: `CC=gcc go test -race ./...`

---

## 1. Feature Coverage Matrix (Tiers 1 - 4)

| Feature ID | Description | Spec Source | Tier 1 (Happy) | Tier 2 (Boundary) | Tier 3 (Cross) | Tier 4 (Scenario) | Status |
|:---|:---|:---|:---:|:---:|:---:|:---:|:---:|
| **F1** | 4x Floor Tiles (256x128) | ORIGINAL_REQUEST §R1 | `TestE2E_Tier1_Feature1_FloorTiles` | `TestEmpiricalFloorDimensionsAndAlpha` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 1 & 5 | **PASS** |
| **F2** | 4x Obstacles/Props (256x256) | ORIGINAL_REQUEST §R1 | `TestE2E_Tier1_Feature2_ObstaclesProps` | `TestEmpiricalObstacleBoundsAndGrounding` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 1 & 5 | **PASS** |
| **F3** | 4x Character Entities (64x128) | ORIGINAL_REQUEST §R1 | `TestE2E_Tier1_Feature3_CharacterEntities` | `TestEmpiricalCharacterBounds` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 2 & 4 | **PASS** |
| **F4** | 4x Items & Weapons (64x64) | ORIGINAL_REQUEST §R1 | `TestE2E_Tier1_Feature4_ItemsAndEquipment` | `TestEmpiricalItemIconQuality` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 3 & 5 | **PASS** |
| **F5** | Geometric Vector Overlays | ORIGINAL_REQUEST §R1 | `TestE2E_Tier1_Feature5_GeometricVectorOverlays` | `TestEmpiricalFloorDimensionsAndAlpha` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 5 | **PASS** |
| **F6** | Engine TileSize (128) & Math | ORIGINAL_REQUEST §R2 | `TestE2E_Tier1_Feature6_TileSizeAndMath` | `TestWorldToIso` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 1 | **PASS** |
| **F7** | DrawSystem Anchors & Camera | ORIGINAL_REQUEST §R2 | `TestE2E_Tier1_Feature7_DrawSystemAnchors` | `TestIsometricRenderingAllTileTypesAndPropsStress` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 1 | **PASS** |
| **F8** | Entity Physics & Speeds | ORIGINAL_REQUEST §R2 | `TestE2E_Tier1_Feature8_EntityPhysicsAndSpeeds` | `TestEmpirical_ECSMovementAABBCollision` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 1 & 3 | **PASS** |
| **F9** | Combat & AI Range Scaling | ORIGINAL_REQUEST §R2 | `TestE2E_Tier1_Feature9_CombatAndAIRangeScaling` | `TestEmpirical_PlayerSpawnSafetyAndZombieDistance` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 2, 3, 4 | **PASS** |
| **F10** | Bezier Attack Curve Math | ORIGINAL_REQUEST §R3 | `TestE2E_Tier1_Feature10_BezierCurveCalculation` | `TestBezier_AxeControlPointsCalculation` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 2 | **PASS** |
| **F11** | Vector Attack Swoosh Rendering | ORIGINAL_REQUEST §R3 | `TestE2E_Tier1_Feature11_VectorSwooshRendering` | `TestBezier_PathRenderingAntiAliasing` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 2 & 4 | **PASS** |
| **F12** | Weapon-Specific Swoosh Styles | ORIGINAL_REQUEST §R3 | `TestE2E_Tier1_Feature12_WeaponSpecificStyles` | `TestBezier_AllWeaponControlPointsSanity` | `TestE2E_Tier3_CrossFeatureInteractions` | Scenario 2 & 4 | **PASS** |

---

## 2. Test Execution Commands & Verification Protocols

1. **Full Test Suite Execution**:
   ```bash
   CC=gcc go test -v ./...
   ```
2. **Race Condition & Concurrency Hardening**:
   ```bash
   CC=gcc go test -race ./...
   ```
3. **Asset Generation Determinism & Stability Check**:
   ```bash
   go test -v ./cmd/tools/genassets
   ```

---

## 3. Real-World Application Scenarios (Tier 4)

- **Scenario 1**: High-Res Isometric Town Exploration & FOV Raycasting (`TestE2E_Tier4_Scenario1_TownExplorationAndFOV`) — Passed.
- **Scenario 2**: Fire Axe Cleave Combat & Bezier Swoosh Arc under Right-Click Aim (`TestE2E_Tier4_Scenario2_AxeCleaveAndBezierSwoosh`) — Passed.
- **Scenario 3**: Multi-Weapon Durability Lifecycle & Armor Mitigation with 4x Physics (`TestE2E_Tier4_Scenario3_WeaponLifecycleAndArmorMitigation`) — Passed.
- **Scenario 4**: Night Survival Horde Encounter with Shotgun Blast & Acoustic Pulse (`TestE2E_Tier4_Scenario4_NightHordeSurvival`) — Passed.
- **Scenario 5**: Full Procedural Asset Regeneration Determinism & Hash Stability (`TestE2E_Tier4_Scenario5_ProceduralGenerationDeterminism`) — Passed.

---

## 4. Attestation & Sign-off

- Total Test Cases: > 140 assertions.
- Failure Count: 0.
- All 27 4x Assets: Procedurally verified with high-density vector paths, anti-aliased geometry, and grounded drop shadows.
- Engine Math: $TileSize = 128$, `WorldToIso` / `IsoToWorld` bijective symmetry confirmed.
- Combat Swoosh: Dynamic Bezier curve tracing with quadratic alpha fade $(1-t)^2$, anti-aliased stroke rendering, and weapon-specific trajectories.
