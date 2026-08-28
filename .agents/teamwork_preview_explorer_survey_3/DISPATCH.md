## 2026-08-28T17:12:49Z
Investigate the codebase with a focus on Items, Weapons, Armor, Combat, and Damage mechanics, plus existing test suites.
Analyze:
1. Current items, inventory, weapon types, combat logic, player stats, zombie attacks, and damage calculation.
2. What weapon types exist today, and how a new weapon type (e.g. ranged weapons like crossbow/shotgun/pistol, or new melee types like spear/axe/chainsaw) can be implemented.
3. How an armor system (damage mitigation, armor slots/equipment, durability/defense calculation, UI indicator) should be architected and integrated.
4. Current test suites (`go test ./...`), build commands, Ebitengine game loop startup (`cmd/game`), and testing harness.
