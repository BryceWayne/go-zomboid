# Orchestrator Execution Plan: R1, R2, R3, R4 Enhancements

## Phase 0: Survey & Scope Mapping (COMPLETED)
- 3 Explorers analyzed the codebase for R1 (Autotiling), R2 & R3 (Equip slot & Chest interaction), R4 (Destruction & Drops) and Testing requirements.
- Findings incorporated into `PROJECT.md`.

## Phase 1: Milestone 1 — Tile Rendering Upgrade & Autotiling (R1)
- Scope: 2D orthogonal autotiling, terrain blending between grass, dirt, concrete, asphalt, floors, and connected walls/fences.
- Loop: Worker -> 2 Reviewers -> 2 Challengers -> Forensic Auditor -> Gate.

## Phase 2: Milestone 2 & Milestone 3 — Equip/Unequip (R2) & Storage Chests (R3)
- Scope: Dedicated 'Equipped' UI slot, item transfer/unequip mechanics, chest inventory persistence, proximity detection, 'E' atomic inventory swap.
- Loop: Worker -> 2 Reviewers -> 2 Challengers -> Forensic Auditor -> Gate.

## Phase 3: Milestone 4 — Environmental Destruction & Wood Drops (R4)
- Scope: Tile durability model, weapon/axe attack chopping collision, solidity/vision clearing, wood item drop spawning & pickup.
- Loop: Worker -> 2 Reviewers -> 2 Challengers -> Forensic Auditor -> Gate.

## Phase 4: Milestone 5 — Comprehensive E2E Verification & Adversarial Hardening
- Scope: Full test suite verification, build check, adversarial stress tests, and forensic integrity audit.
- Final handoff and reporting.
