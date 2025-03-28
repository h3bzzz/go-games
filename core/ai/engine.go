package ai

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/h3bzzz/go-chess/core/game"
)

var PieceValues = map[int]int{
	game.WhitePawn:   100,
	game.BlackPawn:   100,
	game.WhiteKnight: 320,
	game.BlackKnight: 320,
	game.WhiteBishop: 330,
	game.BlackBishop: 330,
	game.WhiteRook:   500,
	game.BlackRook:   500,
	game.WhiteQueen:  900,
	game.BlackQueen:  900,
	game.WhiteKing:   20000,
	game.BlackKing:   20000,
	game.Empty:       0,
}

// Add special protection values for key pieces
var PieceProtectionValue = map[int]int{
	game.WhitePawn:   50,
	game.BlackPawn:   50,
	game.WhiteKnight: 200,
	game.BlackKnight: 200,
	game.WhiteBishop: 200,
	game.BlackBishop: 200,
	game.WhiteRook:   350,
	game.BlackRook:   350,
	game.WhiteQueen:  800, // Queen gets extra protection value
	game.BlackQueen:  800, // Queen gets extra protection value
	game.WhiteKing:   5000,
	game.BlackKing:   5000,
}

type ChessAI struct {
	gameState   *game.GameState
	playerColor int
	difficulty  int // 1-3: easy, medium, hard
	random      *rand.Rand
	aiColor     int
}

func NewChessAI(gameState *game.GameState, playerColor int, difficulty int) *ChessAI {
	return &ChessAI{
		gameState:   gameState,
		playerColor: playerColor,
		difficulty:  difficulty,
		random:      rand.New(rand.NewSource(time.Now().UnixNano())),
		aiColor:     playerColor,
	}
}

func (ai *ChessAI) MakeMove() bool {
	currentTurn := ai.gameState.CurrentTurn
	aiColor := game.WhitePlayer
	if ai.playerColor == game.WhitePlayer {
		aiColor = game.BlackPlayer
	}

	if currentTurn != aiColor {
		return false
	}

	fmt.Printf("AI (%s) is thinking...\n", getColorName(aiColor))

	allMoves := ai.getAllPossibleMoves(aiColor)
	if len(allMoves) == 0 {
		fmt.Println("AI has no legal moves!")
		return false
	}

	var selectedMove Move
	switch ai.difficulty {
	case 1:
		selectedMove = ai.getRandomMove(allMoves)
	case 2:
		selectedMove = ai.getMediumMove(allMoves)
	case 3:
		selectedMove = ai.getHardMove(allMoves)
	default:
		selectedMove = ai.getRandomMove(allMoves)
	}

	fmt.Printf("AI moves %d,%d -> %d,%d\n", selectedMove.From.X, selectedMove.From.Y, selectedMove.To.X, selectedMove.To.Y)
	result := ai.gameState.MakeMove(selectedMove.From, selectedMove.To)
	return result != game.InvalidMove
}

type Move struct {
	From         game.Position
	To           game.Position
	Piece        int
	CapturePiece int
	Score        int
}

func (ai *ChessAI) getAllPossibleMoves(color int) []Move {
	var allMoves []Move

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			pos := game.Position{X: x, Y: y}
			piece := ai.gameState.GetPieceAtPosition(pos)

			if piece == game.Empty {
				continue
			}

			isPieceWhite := game.IsPieceWhite(piece)
			if (color == game.WhitePlayer && !isPieceWhite) ||
				(color == game.BlackPlayer && isPieceWhite) {
				continue
			}

			moves := ai.gameState.GetPossibleMoves(pos)
			for _, movePos := range moves {
				capturePiece := ai.gameState.GetPieceAtPosition(movePos)
				move := Move{
					From:         pos,
					To:           movePos,
					Piece:        piece,
					CapturePiece: capturePiece,
				}
				allMoves = append(allMoves, move)
			}
		}
	}

	return allMoves
}

func (ai *ChessAI) getRandomMove(moves []Move) Move {
	// Score moves - even in easy mode we need to avoid obvious queen sacrifices
	for i := range moves {
		moves[i].Score = 0

		tempBoard := copyBoard(ai.gameState.Board)
		from := moves[i].From
		to := moves[i].To
		piece := tempBoard[from.Y][from.X]
		capturedPiece := tempBoard[to.Y][to.X]

		// Make temporary move
		tempBoard[to.Y][to.X] = piece
		tempBoard[from.Y][from.X] = game.Empty

		// Don't move into check
		if game.IsInCheck(tempBoard, ai.aiColor) {
			moves[i].Score -= 500
		}

		// Add special protection for the queen
		if piece == game.WhiteQueen || piece == game.BlackQueen {
			// Check if the queen would be under attack
			if isPositionUnderAttack(tempBoard, to, ai.playerColor) {
				// If we're not capturing a piece of similar value, heavily penalize queen moves to attacked squares
				if PieceValues[capturedPiece] < 600 {
					moves[i].Score -= 800
				}
			}
		}

		// Restore the board
		tempBoard[from.Y][from.X] = piece
		tempBoard[to.Y][to.X] = capturedPiece
	}

	// Separate good moves from potentially bad ones
	var goodMoves []Move
	var okayMoves []Move

	for _, move := range moves {
		if move.Score >= 0 {
			goodMoves = append(goodMoves, move)
		} else if move.Score > -400 {
			okayMoves = append(okayMoves, move)
		}
	}

	// Choose from good moves if possible
	if len(goodMoves) > 0 {
		return goodMoves[ai.random.Intn(len(goodMoves))]
	}

	// If no good moves, choose from okay moves
	if len(okayMoves) > 0 {
		return okayMoves[ai.random.Intn(len(okayMoves))]
	}

	// If no okay moves, pick any move that doesn't totally sacrifice the queen
	var safeMoves []Move
	for _, move := range moves {
		if move.Score > -700 {
			safeMoves = append(safeMoves, move)
		}
	}

	if len(safeMoves) > 0 {
		return safeMoves[ai.random.Intn(len(safeMoves))]
	}

	// If we have no choice, pick any move
	return moves[ai.random.Intn(len(moves))]
}

func (ai *ChessAI) getMediumMove(moves []Move) Move {
	for i := range moves {
		moves[i].Score = 0

		if moves[i].CapturePiece != game.Empty {
			moves[i].Score += PieceValues[moves[i].CapturePiece]
		}

		tempBoard := copyBoard(ai.gameState.Board)
		from := moves[i].From
		to := moves[i].To
		piece := tempBoard[from.Y][from.X]
		capturedPiece := tempBoard[to.Y][to.X]

		tempBoard[to.Y][to.X] = piece
		tempBoard[from.Y][from.X] = game.Empty

		if game.IsInCheck(tempBoard, ai.aiColor) {
			moves[i].Score -= 1000
		}

		if (to.X >= 2 && to.X <= 5) && (to.Y >= 2 && to.Y <= 5) {
			moves[i].Score += 10
		}

		if isPositionUnderAttack(tempBoard, to, ai.playerColor) {
			if piece == game.WhiteQueen || piece == game.BlackQueen {
				moves[i].Score -= 900 + PieceProtectionValue[piece]
			} else if PieceValues[piece] > PieceValues[capturedPiece] {
				moves[i].Score -= PieceProtectionValue[piece]
			}
		}

		tempBoard[from.Y][from.X] = piece
		tempBoard[to.Y][to.X] = capturedPiece
	}

	sortMovesByScore(moves)

	numToConsider := 3
	if len(moves) < numToConsider {
		numToConsider = len(moves)
	}

	return moves[ai.random.Intn(numToConsider)]
}

func sortMovesByScore(moves []Move) {
	for i := 0; i < len(moves)-1; i++ {
		for j := i + 1; j < len(moves); j++ {
			if moves[j].Score > moves[i].Score {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}
}

func copyBoard(board [8][8]int) [8][8]int {
	var newBoard [8][8]int
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			newBoard[i][j] = board[i][j]
		}
	}
	return newBoard
}

func (ai *ChessAI) getHardMove(moves []Move) Move {
	for i := range moves {
		moves[i].Score = PieceValues[moves[i].CapturePiece]

		tempBoard := copyBoard(ai.gameState.Board)
		from := moves[i].From
		to := moves[i].To
		piece := tempBoard[from.Y][from.X]
		capturedPiece := tempBoard[to.Y][to.X]

		if piece == game.WhiteQueen || piece == game.BlackQueen {
			if capturedPiece != game.WhiteQueen && capturedPiece != game.BlackQueen {
				moves[i].Score -= 900 + (PieceValues[piece]-PieceValues[capturedPiece])*2
			}
		} else if PieceValues[piece] > PieceValues[capturedPiece]+20 {
			moves[i].Score -= (PieceValues[piece] - PieceValues[capturedPiece])
		}

		tempBoard[to.Y][to.X] = piece
		tempBoard[from.Y][from.X] = game.Empty

		if isPositionUnderAttack(tempBoard, to, ai.playerColor) {
			if piece == game.WhiteQueen || piece == game.BlackQueen {
				moves[i].Score -= 900 + PieceProtectionValue[piece]
			} else {
				moves[i].Score -= PieceProtectionValue[piece]
			}
		}

		opponentColor := game.WhitePlayer
		if ai.playerColor == game.WhitePlayer {
			opponentColor = game.BlackPlayer
		}

		if game.IsInCheck(tempBoard, opponentColor) {
			moves[i].Score += 60

			if game.IsCheckmate(tempBoard, opponentColor) {
				// Massive bonus for checkmate
				moves[i].Score += 20000
			}
		}

		// Avoid moving into check at all costs
		if game.IsInCheck(tempBoard, ai.aiColor) {
			// Heavy penalty for moving into check
			moves[i].Score -= 1500
		}

		// Center control bonus
		centerBonus := 0
		if (to.X >= 2 && to.X <= 5) && (to.Y >= 2 && to.Y <= 5) {
			centerBonus = 15
			// Extra bonus for the 4 center squares
			if (to.X >= 3 && to.X <= 4) && (to.Y >= 3 && to.Y <= 4) {
				centerBonus = 25
			}
			moves[i].Score += centerBonus
		}

		// Development bonus in the opening
		if len(ai.gameState.MoveHistory) < 10 {
			// Encourage developing minor pieces early
			if piece == game.WhiteKnight || piece == game.BlackKnight ||
				piece == game.WhiteBishop || piece == game.BlackBishop {
				// Check if the piece is moving from its starting position
				isStartingPos := false
				if piece == game.WhiteKnight && (from.Y == 0 && (from.X == 1 || from.X == 6)) {
					isStartingPos = true
				} else if piece == game.BlackKnight && (from.Y == 7 && (from.X == 1 || from.X == 6)) {
					isStartingPos = true
				} else if piece == game.WhiteBishop && (from.Y == 0 && (from.X == 2 || from.X == 5)) {
					isStartingPos = true
				} else if piece == game.BlackBishop && (from.Y == 7 && (from.X == 2 || from.X == 5)) {
					isStartingPos = true
				}

				if isStartingPos {
					moves[i].Score += 40 // Bonus for developing pieces
				}
			}
		}

		// Restore the board
		tempBoard[from.Y][from.X] = piece
		tempBoard[to.Y][to.X] = capturedPiece
	}

	// Find the highest scoring move
	bestMove := moves[0]
	bestScore := moves[0].Score

	for _, move := range moves[1:] {
		if move.Score > bestScore {
			bestMove = move
			bestScore = move.Score
		}
	}

	return bestMove
}

func isPositionUnderAttack(board [8][8]int, pos game.Position, playerColor int) bool {
	opponentColor := game.BlackPlayer
	if playerColor == game.BlackPlayer {
		opponentColor = game.WhitePlayer
	}

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			piece := board[y][x]
			if piece == game.Empty {
				continue
			}

			isPieceWhite := game.IsPieceWhite(piece)
			if (opponentColor == game.WhitePlayer && !isPieceWhite) ||
				(opponentColor == game.BlackPlayer && isPieceWhite) {
				continue
			}

			position := game.Position{X: x, Y: y}
			moves := game.GetPieceMoves(board, position)
			for _, move := range moves {
				if move.X == pos.X && move.Y == pos.Y {
					return true
				}
			}
		}
	}

	return false
}

func getColorName(color int) string {
	if color == game.WhitePlayer {
		return "White"
	}
	return "Black"
}
