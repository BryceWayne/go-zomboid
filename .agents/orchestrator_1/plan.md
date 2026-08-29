# Orchestration Plan: Orthogonal Engine & Dungeon Master System

## Objective
1. Rewrite coordinate math and `DrawSystem` from Isometric to 2D Orthogonal (top-down) grid.
2. Tile RPG Maker / 2D assets seamlessly without black spaces/gaps.
3. Implement Dungeon Master simulation system (wave spawning, dynamic difficulty, randomized loot drops, day/night cycle with ambient lighting and enemy aggression scaling).
4. Update and pass all tests (`CC=gcc go test ./...`) and ensure seamless execution of `CC=gcc go run ./cmd/game`.

## Phases & Steps
- **Phase 0: Codebase & Specification Survey**
  - Explorer 1: Engine architecture, coordinate math (IsoToWorld, WorldToIso), DrawSystem, camera, rendering pipeline, asset tiling.
  - Explorer 2: Game loop, world generation, entities/ECS, enemy AI, inventory/loot system, lighting/ambient shaders.
  - Explorer 3: Test suite review (`go test ./...`), build scripts, test infrastructure, edge cases, acceptance verification.
- **Phase 1: Project Plan & Test Infrastructure Specification**
  - Synthesize survey findings into `PROJECT.md` and `TEST_INFRA.md`.
  - Feature Inventory & Interface Contracts definition.
- **Phase 2: Milestone Decomposition & Execution**
  - **Milestone 1: 2D Orthogonal Coordinate Math & Map Data Layer**
  - **Milestone 2: 2D Orthogonal DrawSystem & Asset Tiling Layer**
  - **Milestone 3: Dungeon Master System (Day/Night cycle & Ambient Lighting)**
  - **Milestone 4: Dungeon Master System (Dynamic Zombie Spawning, Aggression & Loot Drops)**
  - **Milestone 5: E2E Test Suite Creation & Verification Track**
- **Phase 3: Integration & Final E2E Verification**
  - Pass 100% of unit, integration, and E2E test suites with `CC=gcc go test ./...`.
  - Verify game startup and runtime behavior (`CC=gcc go run ./cmd/game`).
  - Adversarial hardening & Forensic Audit verification.
- **Phase 4: Final Synthesis & Human Reporting**
