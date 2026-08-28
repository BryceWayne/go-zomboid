## 2026-08-28T17:31:35Z
Scope: Milestone 3 - Armor Data Structures & Equipping Integration
Task:
1. Formulate exact modifications to `internal/ecs/components.go` for `ecs.Player` armor fields (`ArmorEquipped`, `ArmorType`, `ArmorDefense`, `ArmorDurability`, `ArmorMaxDurability`, `InfectionResist`).
2. Formulate the equipping logic in `internal/game/game.go:processInputAndCombat()` when a player selects an `"armor"` inventory slot (key 1-9), setting attributes, consuming the item, and setting cooldown.
3. Formulate pure Go code and document in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m3_1/handoff.md`.
When done, message your parent.
