🕹️ Games in the Collection
1. Snake Game
Description: A modern take on the classic Snake game. Control the snake to eat food, grow longer, and avoid collisions with yourself or the walls.
Features:
Smooth grid-based movement.
Dynamic food spawning and score tracking.
Game-over and restart functionality.
Adjustable speed for gameplay difficulty.

2. Maze Game
Description: Navigate a procedurally generated maze as a player character animated with a GIF. Reach the exit to generate a new maze and start again.

Features:
Procedural maze generation using Depth-First Search (DFS).
Smooth player movement with directional animation.
Animated player character using a sprite sheet (GIF).
Maze regeneration upon reaching the exit.

💡 What I Learned
Programming Concepts
Game Loops:

Managed game states (Update, Draw, and Layout methods in Ebiten).
Controlled animations and movements using frame counters.

Data Structures:
Position Structs: Represented grid-based entities (e.g., snake segments, maze cells).
2D Arrays: Used for procedural maze generation and collision detection.

Randomization:
Implemented random food spawning and maze generation using custom random number generators seeded with the current time.

Collision Detection:
Developed collision checks for boundaries, self-intersections (Snake), and maze walls (Maze Game).

Procedural Generation:
Created dynamic mazes using algorithms like Depth-First Search (DFS) with backtracking.

Animation Handling:
Loaded and cycled through GIF frames for a smooth player character animation.
Controlled animation frame rates independently of game loop timing.
Programming Techniques

Event Handling:
Captured keyboard inputs for game controls.
Implemented directional movement logic and constraints.

Code Modularity:
Separated game logic into reusable structs and functions.
Created distinct functions for maze generation, collision checks, and drawing.

Error Handling:
Used proper error checks for file I/O when loading assets (e.g., GIF files).

Scaling and Transformation:
Applied Ebiten's GeoM transformations for sprite scaling and flipping (e.g., horizontal flipping for player direction).

🛠️ Technologies Used
Ebiten: A simple and efficient library for creating 2D games in Go.
GIF Handling: Leveraged image/gif to load and manipulate animated sprites.
