## 2026-08-29T15:12:30Z
You are teamwork_preview_explorer_survey_1.
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1.
Please read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md.

Task:
Conduct an in-depth technical survey of the asset pipeline:
1. Explore the `/home/bryce/code/go-zomboid/context/` directory. Enumerate every single file, its filename, image type, dimensions (if inspectable), and category/theme.
2. Explore `/home/bryce/code/go-zomboid/cmd/tools/genassets` and search the entire codebase for references to `genassets` (e.g. in `go:generate`, `Makefile`, `README`, scripts, etc.).
3. Explore `/home/bryce/code/go-zomboid/internal/assets/` (especially `assets.go` and `images/`). Understand how images are embedded/loaded (e.g. `//go:embed`, `ebiten.NewImageFromImage`, decode PNG), which `ebiten.Image` variables exist currently, and what variables/functions will be needed for the new PNG assets in `context/`.
4. Document the exact file copying and code refactoring plan needed for R1 and R2.

Write your comprehensive survey report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1/survey.md`.
Also write a structured `handoff.md` and update `progress.md` with your status. Send a message when done.
