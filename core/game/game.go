package game

import "time"

func (g *GameState) MakeMove(from, to Position) MoveResult {
	if !IsValidBoardPosition(from) || !IsValidBoardPosition(to) {
		return InvalidMove
	}

	piece := g.Board[from.Y][from.X]

	if piece == Empty {
		return InvalidMove
	}

	if (g.CurrentTurn == WhitePlayer && !IsPieceWhite(piece)) ||
		(g.CurrentTurn == BlackPlayer && !IsPieceBlack(piece)) {
		return InvalidMove
	}

	var validMoves []Position

	if piece == WhiteKing || piece == BlackKing {
		validMoves = GetPieceMovesWithGameState(g.Board, from, g)
	} else {
		validMoves = GetPieceMoves(g.Board, from)
	}

	moveValid := false
	for _, move := range validMoves {
		if move.X == to.X && move.Y == to.Y {
			moveValid = true
			break
		}
	}

	if !moveValid {
		return InvalidMove
	}

	capturedPiece := g.Board[to.Y][to.X]

	isCastling := false
	if piece == WhiteKing || piece == BlackKing {
		if to.X-from.X == 2 {
			isCastling = true
			if piece == WhiteKing {
				g.Board[0][5] = WhiteRook
				g.Board[0][7] = Empty
			} else {
				g.Board[7][5] = BlackRook
				g.Board[7][7] = Empty
			}
		} else if from.X-to.X == 2 {
			isCastling = true
			if piece == WhiteKing {
				g.Board[0][3] = WhiteRook
				g.Board[0][0] = Empty
			} else {
				g.Board[7][3] = BlackRook
				g.Board[7][0] = Empty
			}
		}
	}

	g.Board[to.Y][to.X] = piece
	g.Board[from.Y][from.X] = Empty

	if piece == WhiteKing {
		g.WhiteKingMoved = true
	} else if piece == BlackKing {
		g.BlackKingMoved = true
	} else if piece == WhiteRook {
		if from.X == 0 && from.Y == 0 {
			g.WhiteRookAMoved = true
		} else if from.X == 7 && from.Y == 0 {
			g.WhiteRookHMoved = true
		}
	} else if piece == BlackRook {
		if from.X == 0 && from.Y == 7 {
			g.BlackRookAMoved = true
		} else if from.X == 7 && from.Y == 7 {
			g.BlackRookHMoved = true
		}
	}

	g.CurrentTurn = 1 - g.CurrentTurn

	isCheck := IsInCheck(g.Board, g.CurrentTurn)
	isCheckmate := false

	if isCheck {
		isCheckmate = IsCheckmate(g.Board, g.CurrentTurn)
	}

	move := Move{
		From:      from,
		To:        to,
		Piece:     piece,
		Captured:  capturedPiece,
		Check:     isCheck,
		Checkmate: isCheckmate,
		Castling:  isCastling,
	}

	g.MoveHistory = append(g.MoveHistory, move)

	if isCheckmate {
		if g.CurrentTurn == WhitePlayer {
			g.GameStatus = BlackWon
		} else {
			g.GameStatus = WhiteWon
		}
		return Checkmate
	}

	isStalemate := IsStalemate(g.Board, g.CurrentTurn)

	if isStalemate {
		g.GameStatus = GameDraw
		return Stalemate
	}

	now := time.Now()
	if g.TimerActive {
		elapsed := now.Sub(g.LastMoveTime)
		if g.CurrentTurn == WhitePlayer {
			g.BlackPlayerTime -= elapsed
		} else {
			g.WhitePlayerTime -= elapsed
		}
	}
	g.LastMoveTime = now

	if isCheck {
		return Check
	}

	if isCastling {
		return Castling
	}

	return ValidMove
}

func (g *GameState) GetPossibleMoves(pos Position) []Position {
	if !IsValidBoardPosition(pos) {
		return []Position{}
	}

	piece := g.Board[pos.Y][pos.X]

	if piece == Empty {
		return []Position{}
	}

	if (g.CurrentTurn == WhitePlayer && !IsPieceWhite(piece)) ||
		(g.CurrentTurn == BlackPlayer && !IsPieceBlack(piece)) {
		return []Position{}
	}

	var allMoves []Position
	if piece == WhiteKing || piece == BlackKing {
		allMoves = GetPieceMovesWithGameState(g.Board, pos, g)
	} else {
		allMoves = GetPieceMoves(g.Board, pos)
	}

	// Filter out moves that leave the king in check
	var legalMoves []Position
	for _, move := range allMoves {
		// Create a proper copy of the board
		tempBoard := copyBoard(g.Board)
		tempPiece := tempBoard[pos.Y][pos.X]
		tempBoard[pos.Y][pos.X] = Empty
		tempBoard[move.Y][move.X] = tempPiece

		// Check if king is still in check after the move
		if !IsInCheck(tempBoard, g.CurrentTurn) {
			legalMoves = append(legalMoves, move)
		}
	}

	return legalMoves
}

// Helper function to copy a board
func copyBoard(board [8][8]int) [8][8]int {
	var newBoard [8][8]int
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			newBoard[i][j] = board[i][j]
		}
	}
	return newBoard
}

func (g *GameState) SelectPosition(pos Position) bool {
	if !IsValidBoardPosition(pos) {
		g.SelectedPosition = nil
		return false
	}

	piece := g.Board[pos.Y][pos.X]

	if piece != Empty && ((g.CurrentTurn == WhitePlayer && IsPieceWhite(piece)) ||
		(g.CurrentTurn == BlackPlayer && IsPieceBlack(piece))) {
		g.SelectedPosition = &Position{X: pos.X, Y: pos.Y}
		return true
	}

	if g.SelectedPosition != nil {
		result := g.MakeMove(*g.SelectedPosition, pos)
		g.SelectedPosition = nil
		return result != InvalidMove
	}

	g.SelectedPosition = nil
	return false
}

func (g *GameState) UndoLastMove() bool {
	if len(g.MoveHistory) == 0 {
		return false
	}

	lastMove := g.MoveHistory[len(g.MoveHistory)-1]

	g.Board[lastMove.From.Y][lastMove.From.X] = lastMove.Piece
	g.Board[lastMove.To.Y][lastMove.To.X] = lastMove.Captured

	g.GameStatus = InProgress

	g.CurrentTurn = 1 - g.CurrentTurn

	g.MoveHistory = g.MoveHistory[:len(g.MoveHistory)-1]

	return true
}

func (g *GameState) GetGameStatus() string {
	switch g.GameStatus {
	case WhiteWon:
		return "White won by checkmate"
	case BlackWon:
		return "Black won by checkmate"
	case GameDraw:
		return "Game ended in a draw"
	default:
		if IsInCheck(g.Board, g.CurrentTurn) {
			if g.CurrentTurn == WhitePlayer {
				return "White is in check"
			} else {
				return "Black is in check"
			}
		}
		if g.CurrentTurn == WhitePlayer {
			return "White to move"
		} else {
			return "Black to move"
		}
	}
}

func (g *GameState) GetPieceAtPosition(pos Position) int {
	if !IsValidBoardPosition(pos) {
		return Empty
	}
	return g.Board[pos.Y][pos.X]
}

func (g *GameState) StartTimer() {
	if !g.TimerActive {
		g.LastMoveTime = time.Now()
		g.TimerActive = true
	}
}

func (g *GameState) StopTimer() {
	if g.TimerActive {
		now := time.Now()
		elapsed := now.Sub(g.LastMoveTime)
		if g.CurrentTurn == WhitePlayer {
			g.WhitePlayerTime -= elapsed
		} else {
			g.BlackPlayerTime -= elapsed
		}
		g.TimerActive = false
	}
}

func (g *GameState) GetRemainingTime(player int) time.Duration {
	if player == WhitePlayer {
		return g.WhitePlayerTime
	}
	return g.BlackPlayerTime
}

func (g *GameState) IsGameOver() bool {
	return g.GameStatus != InProgress
}
