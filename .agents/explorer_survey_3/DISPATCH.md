## 2026-08-29T15:53:56Z
You are Explorer 3 investigating the go-zomboid testing infrastructure and acceptance criteria.
Working directory: /home/bryce/code/go-zomboid/.agents/explorer_survey_3
Original request path: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md

Mission & Scope:
1. Read `/home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md`.
2. Investigate all existing tests across the codebase (`go test ./...`), map generation tests, logic tests, and mocks.
3. Investigate `cmd/game` and how the application launches, initializes graphics/window (e.g., raylib, ebiten, or custom engine), handles input, and updates ticks.
4. Identify all tests that will be broken by the transition from Isometric to 2D Orthogonal math, and detail exactly how they need to be updated.
5. Design the test strategy and acceptance criteria checklist for:
   - Orthogonal coordinate conversions and map generation tests.
   - Dungeon Master wave spawning, loot distribution, and day/night transitions.
   - Headless test execution and E2E verification.
6. Write your complete analysis and findings to `/home/bryce/code/go-zomboid/.agents/explorer_survey_3/handoff.md` and send a completion message to parent.
