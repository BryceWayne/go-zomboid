# Original User Request

## Initial Request — 2026-08-28T12:12:28-05:00

You are the Project Orchestrator (teamwork_preview_orchestrator_1).

Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_1
Project root: /home/bryce/code/go-zomboid
Original Request File: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md

User Goal & Requirements:
Enhance the gameplay of the `go-zomboid` project by upgrading sprites, improving the environment generation, implementing armor mechanics, and expanding weapons.
Integrity mode: demo

Requirements:
- R1. Procedural Sprite Enhancements: Enhance the existing sprites using the procedural image generation in `cmd/tools/genassets`. Do not download external assets.
- R2. Environment and Items Update: Expand the procedural town generation and implement at least one new weapon type and an armor system that mitigates damage.

Acceptance Criteria:
- Running `go run ./cmd/tools/genassets` executes without errors and generates new/updated sprite files in `internal/assets/images`.
- Running `CC=gcc go test ./...` passes all tests.
- Running `CC=gcc go run ./cmd/game` successfully launches the game, and the new mechanics (armor, new weapons) do not crash the Ebitengine loop.

Please orchestrate this project: create your plan, dispatch specialist subagents, monitor progress, write your progress.md / BRIEFING.md, ensure all tests and acceptance criteria pass, and report back via send_message when complete.

## Follow-up — 2026-08-28T18:47:14Z

# Teamwork Project Prompt — Draft

> Status: Launched
> Goal: Craft prompt → get user approval → delegate to teamwork_preview
> Requested team: full team

Significantly improve the game's resolution to achieve extremely smooth, high-fidelity tiles that exactly match the Dribbble vector art. Overhaul the combat visualization by implementing bezier curves for attack dynamics.

Working directory: `/home/bryce/code/go-zomboid`
Integrity mode: benchmark

## Requirements

### R1. High-Fidelity Tile Resolution
Quadruple the base tile pixel size in `cmd/tools/genassets` (from 64x32 to 256x128 for floors, and proportionally scale entities/walls/props). Ensure the geometric overlays (chevrons, pebbles, overlapping circles) scale up perfectly to maintain the smooth vector style. 

### R2. Engine Isometric Math Upgrade
Update the engine math in `internal/game/world` and `internal/game/game.go` (`TileSize`, `WorldToIso`, `IsoToWorld`, speed coefficients) to seamlessly support the new 4x texture resolution without breaking map generation or movement speed.

### R3. Bezier Curve Combat Dynamics
Implement dynamic weapon swing trails/arcs (like a swoosh when swinging the axe) using Bezier Curves in the `DrawSystem`. The curve should dynamically trace the attack arc based on the player's facing direction and mouse click.

## Acceptance Criteria

### Verification
- [ ] Running `go run ./cmd/tools/genassets` executes without errors and generates 256x128 high-fidelity sprites in `internal/assets/images`.
- [ ] The `TileSize` constant and coordinate math are correctly updated, and the map generates properly.
- [ ] Running `CC=gcc go test ./...` passes all tests.
- [ ] Running `CC=gcc go run ./cmd/game` successfully launches the game. Attacking with the axe produces a visible bezier curve trail on the screen.

