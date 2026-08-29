## 2026-08-29T16:49:35Z
You are Explorer 2 surveying the codebase for Requirements R2 (Equip/Unequip Items) and R3 (Storage Chest Interaction).
Your working directory is /home/bryce/code/go-zomboid/.agents/explorer_survey_r2_r3_1.
Read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md (especially section ## 2026-08-29T16:48:41Z).
Investigate the codebase in /home/bryce/code/go-zomboid:
1. Examine `internal/ecs/components.go`, `internal/game/game.go` (`UpdateSystem`, `DrawSystem`, UI rendering, input handling, inventory management).
2. Analyze how Player inventory is currently represented, how items are selected/used, and how UI displays inventory.
3. Investigate how storage chests (`TileChest` or chest entities) are represented in the world, how proximity is detected, and how interaction (hotkey 'E') can swap the player's entire inventory with the chest's inventory without data loss or crashes.
4. Investigate how a dedicated 'Equipped' UI slot should be added to the UI, how equipping an item transfers it from main inventory to the equipped slot, and how unequipping returns it to inventory.
5. Write your comprehensive survey findings and implementation proposal to `/home/bryce/code/go-zomboid/.agents/explorer_survey_r2_r3_1/handoff.md` and send a message back when done.
