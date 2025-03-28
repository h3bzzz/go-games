package game

import "time"

// Piece constants
const (
	Empty = iota
	WhitePawn
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
)

// Player constants
const (
	WhitePlayer = iota
	BlackPlayer
)

// Position represents a square on the chess board
type Position struct {
	X, Y int
}

// MoveResult represents the result of a move attempt
type MoveResult int

const (
	InvalidMove MoveResult = iota
	ValidMove
	Check
	Checkmate
	Stalemate
	Draw
	Castling
)

// GameStatus represents the status of the game
type GameStatus int

const (
	InProgress GameStatus = iota
	WhiteWon
	BlackWon
	GameDraw
)

// Move represents a chess move
type Move struct {
	From      Position
	To        Position
	Piece     int
	Captured  int
	Promotion int
	Check     bool
	Checkmate bool
	Castling  bool
}

// GameState represents the current state of a chess game
type GameState struct {
	Board            [8][8]int
	CurrentTurn      int
	MoveHistory      []Move
	WhiteKingMoved   bool
	BlackKingMoved   bool
	WhiteRookAMoved  bool
	WhiteRookHMoved  bool
	BlackRookAMoved  bool
	BlackRookHMoved  bool
	GameStatus       GameStatus
	WhitePlayerTime  time.Duration
	BlackPlayerTime  time.Duration
	LastMoveTime     time.Time
	TimerActive      bool
	SelectedPosition *Position
}
