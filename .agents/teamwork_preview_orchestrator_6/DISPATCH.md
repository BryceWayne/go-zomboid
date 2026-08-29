## 2026-08-29T16:49:17Z

You are the Project Orchestrator for the go-zomboid enhancement project.
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_6.
Please refer to the authoritative user request in /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md (specifically the latest request under ## 2026-08-29T16:48:41Z).
The workspace directory is /home/bryce/code/go-zomboid.

Key Requirements:
- R1: Tile Rendering Upgrade (autotiling for 2D orthogonal grid, terrain blending between grass, dirt, and walls to eliminate harsh borders).
- R2: Equip/Unequip Items (dedicated 'Equipped' UI slot, item transfer between inventory and equipped slot).
- R3: Storage Chest Interaction (hotkey 'E' near chest swaps entire player inventory with chest contents).
- R4: Environmental Destruction (chop wooden barriers with weapons/axes, drop wood/resource items).

Verification Criteria:
- `CC=gcc go test ./...` passes all tests.
- `CC=gcc go build -o bin/game ./cmd/game` compiles without errors.
- Player can equip weapons into the dedicated 'Equipped' UI slot.
- Pressing 'E' near chest swaps inventory without data loss or crashes.
- Attacking wooden barriers destroys them and drops collectible resources.
- Terrain rendering utilizes autotiling logic to blend tile edges seamlessly.
