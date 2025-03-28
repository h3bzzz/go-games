# Go Chess

A full-featured chess game built with Go and the Fyne UI toolkit. This project implements chess rules, AI opponents with multiple difficulty levels, and a clean modern interface.

## Features

- Complete chess rule implementation including castling, en passant, promotion
- AI opponent with three difficulty levels:
  - Easy: Makes reasonable moves but misses opportunities
  - Medium: Captures pieces and avoids obvious traps
  - Hard: Uses positional strategy and looks for tactical opportunities
- Multiple board themes
- Game timer with clock for timed games
- Move history tracking
- Drag and drop piece movement

## How It Works

### Game Logic

The chess implementation uses a standard 8x8 board represented as a 2D array, with integer constants for pieces. The game rules logic includes:

- Position validation to ensure moves are within board boundaries
- Piece movement patterns implemented for all chess pieces
- Special rules like castling and check detection
- Game state tracking for win/loss/draw conditions

### AI Implementation

The AI uses a score-based evaluation approach:

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