package game

import "time"

// IsValidBoardPosition checks if a position is within the boundaries of the chess board
func IsValidBoardPosition(pos Position) bool {
	return pos.X >= 0 && pos.X < 8 && pos.Y >= 0 && pos.Y < 8
}

// IsPieceWhite checks if a piece is white
func IsPieceWhite(piece int) bool {
	return piece >= WhitePawn && piece <= WhiteKing
}

// IsPieceBlack checks if a piece is black
func IsPieceBlack(piece int) bool {
	return piece >= BlackPawn && piece <= BlackKing
}

// InitializeBoard returns a new chess board with pieces in starting positions
func InitializeBoard() [8][8]int {
	var board [8][8]int

	// Place pawns
	for i := 0; i < 8; i++ {
		board[1][i] = WhitePawn
		board[6][i] = BlackPawn
	}

	// Place other pieces
	// White pieces
	board[0][0] = WhiteRook
	board[0][1] = WhiteKnight
	board[0][2] = WhiteBishop
	board[0][3] = WhiteQueen
	board[0][4] = WhiteKing
	board[0][5] = WhiteBishop
	board[0][6] = WhiteKnight
	board[0][7] = WhiteRook

	// Black pieces
	board[7][0] = BlackRook
	board[7][1] = BlackKnight
	board[7][2] = BlackBishop
	board[7][3] = BlackQueen
	board[7][4] = BlackKing
	board[7][5] = BlackBishop
	board[7][6] = BlackKnight
	board[7][7] = BlackRook

	return board
}

// NewGame creates a new chess game with default settings
func NewGame() *GameState {
	return &GameState{
		Board:           InitializeBoard(),
		CurrentTurn:     WhitePlayer,
		MoveHistory:     []Move{},
		WhitePlayerTime: 30 * time.Minute,
		BlackPlayerTime: 30 * time.Minute,
		LastMoveTime:    time.Now(),
		GameStatus:      InProgress,
	}
}

// IsInCheck determines if the specified player is in check
func IsInCheck(board [8][8]int, playerTurn int) bool {
	// Find the king
	var kingPos Position
	kingPiece := WhiteKing
	if playerTurn == BlackPlayer {
		kingPiece = BlackKing
	}

	// Locate the king
	kingFound := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if board[y][x] == kingPiece {
				kingPos = Position{x, y}
				kingFound = true
				break
			}
		}
		if kingFound {
			break
		}
	}

	// Check if any opponent piece can capture the king
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			piece := board[y][x]
			if piece != Empty && ((playerTurn == WhitePlayer && IsPieceBlack(piece)) ||
				(playerTurn == BlackPlayer && IsPieceWhite(piece))) {
				moves := GetPieceMoves(board, Position{x, y})
				for _, move := range moves {
					if move.X == kingPos.X && move.Y == kingPos.Y {
						return true
					}
				}
			}
		}
	}

	return false
}

// IsCheckmate determines if the specified player is in checkmate
func IsCheckmate(board [8][8]int, playerTurn int) bool {
	// If not in check, can't be checkmate
	if !IsInCheck(board, playerTurn) {
		return false
	}

	// Check if any move can get the player out of check
	return !canEscapeCheck(board, playerTurn)
}

// canEscapeCheck determines if the player can make any move to escape check
func canEscapeCheck(board [8][8]int, playerTurn int) bool {
	// Try all possible moves for all pieces of the player
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			piece := board[y][x]
			if piece != Empty && ((playerTurn == WhitePlayer && IsPieceWhite(piece)) ||
				(playerTurn == BlackPlayer && IsPieceBlack(piece))) {
				moves := GetPieceMoves(board, Position{x, y})
				for _, move := range moves {
					// Make a temporary move
					tempBoard := board
					tempPiece := tempBoard[y][x]
					tempBoard[y][x] = Empty
					tempBoard[move.Y][move.X] = tempPiece

					// Check if still in check after the move
					if !IsInCheck(tempBoard, playerTurn) {
						return true
					}
				}
			}
		}
	}

	return false
}

// IsStalemate determines if the current position is a stalemate
func IsStalemate(board [8][8]int, playerTurn int) bool {
	// If in check, it's not stalemate
	if IsInCheck(board, playerTurn) {
		return false
	}

	// Check if player has any legal moves
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			piece := board[y][x]
			if piece != Empty && ((playerTurn == WhitePlayer && IsPieceWhite(piece)) ||
				(playerTurn == BlackPlayer && IsPieceBlack(piece))) {
				moves := GetPieceMoves(board, Position{x, y})
				if len(moves) > 0 {
					return false
				}
			}
		}
	}

	return true
}
