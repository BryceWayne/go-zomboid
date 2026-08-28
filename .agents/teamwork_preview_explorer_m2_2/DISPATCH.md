## 2026-08-28T17:24:25Z
Scope: Milestone 2 - Multi-Room Building Archetypes & Interior Floorplans
Task:
1. Read the original request, project plan, and survey handoff.
2. Design the procedural building generator for `internal/game/world/map.go`:
   - Archetypes: Suburban Residential House (living room, bedroom, kitchen, bathroom), Grocery/Convenience Store (sales floor, storage room), Police Station (office, armory/holding cell), Pharmacy/Clinic (consultation room, medical storage), Warehouse (open storage, crates).
   - Interior floorplans: TileWoodFloor for houses, TileTileFloor for commercial/police/pharmacy, TileConcrete for warehouse.
   - Wall perimeters (TileWall), interior partition walls, doors, windows/openings.
3. Formulate pure Go data structures and building synthesis functions.
4. Document findings and proposed code in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_2/handoff.md`.
When done, message your parent.
