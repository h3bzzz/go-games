
# Go Chess

A full-featured chess game built with Go and the Fyne UI toolkit. This project implements chess rules, AI opponents with multiple difficulty levels, and a clean modern interface.

![Chess Game Screenshot](https://github.com/h3bzzz/go-chess/raw/main/assets/screenshots/game.png)

## Features

- Complete chess rule implementation including castling, en passant, promotion
- AI opponent with three difficulty levels:
  - Easy: Makes reasonable moves but misses opportunities
  - Medium: Captures pieces and avoids obvious traps
  - Hard: Uses positional strategy and looks for tactical opportunities
- Multiple board themes (Classic, Green, Pink)
- Game timer with clock for timed games
- Move history tracking
- Drag and drop piece movement

## Screenshots

### Classic Theme
![Classic Theme](https://github.com/h3bzzz/go-chess/raw/main/assets/screenshots/classic.png)

### Green Theme
![Green Theme](https://github.com/h3bzzz/go-chess/raw/main/assets/screenshots/green.png)

### Pink Theme
![Pink Theme](https://github.com/h3bzzz/go-chess/raw/main/assets/screenshots/pink.png)

## How It Works

### Game Logic

The chess implementation uses a standard 8x8 board represented as a 2D array, with integer constants for pieces. The game rules logic includes:

- Position validation to ensure moves are within board boundaries
- Piece movement patterns implemented for all chess pieces
- Special rules like castling and check detection
- Game state tracking for win/loss/draw conditions

### AI Implementation

The "AI" uses a score-based evaluation approach:

- Piece values: Traditional chess piece values (pawns: 100, knights/bishops: ~330, rooks: 500, queen: 900)
- Position evaluation: Bonuses for controlling the center and developing pieces early
- Tactical awareness: The AI recognizes checks, captures, and avoids losing pieces
- Protection values: Extra safeguards to prevent sacrificing high-value pieces like the queen

The difficulty levels modify how deeply the AI evaluates positions and how much randomness is introduced to its choices:

- Easy: Makes safe moves but has a large random element
- Medium: Captures pieces when possible and protects its pieces
- Hard: Evaluates positions more thoroughly and uses proper opening principles

### UI Implementation

The UI is built with the Fyne toolkit, a cross-platform GUI library for Go:

- Responsive grid-based board layout
- Drag and drop functionality for piece movement
- Theme switching capabilities
- Game state display showing current player, check status, etc.
- Clock display for timed games

## Technology Stack

- Go (Golang) - Core game logic and application
- Fyne - UI toolkit for cross-platform GUI
- Custom asset management for multiple themes
- Go concurrency for AI calculations

## Building and Running

### Prerequisites

- Go 1.16 or later
- Fyne dependencies (see [Fyne getting started](https://developer.fyne.io/started/))

### Installation

```bash
# Clone the repository
git clone https://github.com/h3bzzz/go-chess.git

# Navigate to the project directory
cd go-chess

# Build the project
go build ./cmd/main.go

# Run the game
./main
```

## Future Improvements

- Network play functionality
- PGN import/export for game notation
- Opening book for the AI
- Deeper AI evaluation with actual minimax or alpha-beta pruning
- Customizable time controls

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

Made with ♟️ by h3bzzz 
=======
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

# Gopher Platform Runner

A platform-based runner game featuring the Go gopher mascot traversing randomly generated levels.

## Features

- 10 progressively difficult levels with randomized platform generation
- Platform physics with jumping, double-jumping, wall-jumping, and dashing abilities
- Different platform types (normal, moving, breaking, bouncy)
- Collectibles to gather for points
- Obstacles to avoid
- Lives system and score tracking

## Controls

- **Left/Right Arrow Keys or A/D**: Move left/right
- **Space, W, or Up Arrow**: Jump (press again in mid-air for double jump)
- **Shift or E**: Dash
- **F1**: Toggle debug mode
- **Esc**: Pause game

## How to Run

Navigate to the mazeGame directory and run:

```bash
go run *.go
```

Or build an executable:

```bash
cd mazeGame
go build
./mazeGame
```

## Progression

- **Level 1-2**: Basic platforming
- **Level 3**: Unlocks double jump
- **Level 4-5**: Introduces moving and breaking platforms
- **Level 5**: Unlocks wall jump
- **Level 6-7**: Increases obstacles and platform difficulty
- **Level 7**: Unlocks dash ability
- **Level 8-10**: Maximum challenge with all mechanics

## Development

This game was created as an overhaul of a simple maze game, transformed into a full-featured platformer with physics, level progression, and advanced game mechanics.

Future improvements may include:
- Additional animations
- Sound effects and music
- More varied obstacle types
- Boss encounters
- Level editor

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
>>>>>>> 70257313caf3ae719091bbfc2a01d0b8bfd2b87d
