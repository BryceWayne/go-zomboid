# Go Zomboid

A 2.5D isometric zombie survival game built in Go, using [Ebitengine](https://ebitengine.org/) and [Ark ECS](https://github.com/mlange-42/ark).

## Features

* **2.5D Isometric Engine**: Mathematical projection of grid coordinates to an isometric perspective, complete with Y-depth sorting so entities render properly behind or in front of walls and trees.
* **Entity Component System**: Powered by Ark ECS for performant querying and system processing.
* **Zombie Swarm AI**: Zombies feature wandering logic, hearing/vision aggro, and boids-style separation flocking to prevent stacking.
* **Procedural Assets**: The game assets are generated programmatically via the included `genassets` tool, creating isometric tiles, trees, and entities with randomized noise textures.
* **Combat System**: Find and equip weapons, project hitboxes, and fight off the horde.

## Getting Started

### Prerequisites

You need Go 1.21+ installed.

If you are on Linux, Ebitengine requires C compiler and ALSA/X11 development libraries:

```sh
sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config
```

### Running the Game

To play the game, run:

```sh
go run ./cmd/game
```

*Note: If you are using a Conda environment that intercepts the C compiler, you may need to explicitly specify `CC=gcc` before the command.*

### Generating Assets

If you modify the asset generation script, you can regenerate all PNG images by running:

```sh
go run ./cmd/tools/genassets
```

## Controls

* **W/A/S/D** or **Arrow Keys**: Move
* **Spacebar**: Attack (requires equipped weapon)
* **R**: Restart game (upon death)
