package game

func GetPieceMoves(board [8][8]int, pos Position) []Position {
	piece := board[pos.Y][pos.X]
	var moves []Position

	switch piece {
	case WhitePawn, BlackPawn:
		moves = getPawnMoves(board, pos)
	case WhiteKnight, BlackKnight:
		moves = getKnightMoves(board, pos)
	case WhiteBishop, BlackBishop:
		moves = getBishopMoves(board, pos)
	case WhiteRook, BlackRook:
		moves = getRookMoves(board, pos)
	case WhiteQueen, BlackQueen:
		moves = getQueenMoves(board, pos)
	case WhiteKing, BlackKing:
		moves = getKingMoves(board, pos)
	}

	return moves
}

func GetPieceMovesWithGameState(board [8][8]int, pos Position, gameState *GameState) []Position {
	piece := board[pos.Y][pos.X]

	if piece == WhiteKing || piece == BlackKing {
		return GetKingMovesWithCastling(board, pos, gameState)
	}

	return GetPieceMoves(board, pos)
}

func getPawnMoves(board [8][8]int, pos Position) []Position {
	var moves []Position
	piece := board[pos.Y][pos.X]
	isWhite := IsPieceWhite(piece)

	dir := 1
	startRow := 1
	if !isWhite {
		dir = -1
		startRow = 6
	}

	newPos := Position{pos.X, pos.Y + dir}
	if IsValidBoardPosition(newPos) && board[newPos.Y][newPos.X] == Empty {
		moves = append(moves, newPos)

		if pos.Y == startRow {
			doublePos := Position{pos.X, pos.Y + 2*dir}
			if IsValidBoardPosition(doublePos) && board[doublePos.Y][doublePos.X] == Empty {
				moves = append(moves, doublePos)
			}
		}
	}

	for _, dx := range []int{-1, 1} {
		capturePos := Position{pos.X + dx, pos.Y + dir}
		if IsValidBoardPosition(capturePos) {
			targetPiece := board[capturePos.Y][capturePos.X]
			if targetPiece != Empty && (isWhite != IsPieceWhite(targetPiece)) {
				moves = append(moves, capturePos)
			}
		}
	}

	return moves
}

func getKnightMoves(board [8][8]int, pos Position) []Position {
	var moves []Position
	piece := board[pos.Y][pos.X]
	isWhite := IsPieceWhite(piece)

	knightMoves := []Position{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}

	for _, offset := range knightMoves {
		newPos := Position{pos.X + offset.X, pos.Y + offset.Y}
		if IsValidBoardPosition(newPos) {
			targetPiece := board[newPos.Y][newPos.X]
			if targetPiece == Empty || (isWhite != IsPieceWhite(targetPiece)) {
				moves = append(moves, newPos)
			}
		}
	}

	return moves
}

func getBishopMoves(board [8][8]int, pos Position) []Position {
	var moves []Position
	piece := board[pos.Y][pos.X]
	isWhite := IsPieceWhite(piece)

	directions := []Position{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newPos := Position{pos.X + i*dir.X, pos.Y + i*dir.Y}
			if !IsValidBoardPosition(newPos) {
				break
			}

			targetPiece := board[newPos.Y][newPos.X]
			if targetPiece == Empty {
				moves = append(moves, newPos)
			} else {
				if isWhite != IsPieceWhite(targetPiece) {
					moves = append(moves, newPos)
				}
				break
			}
		}
	}

	return moves
}

func getRookMoves(board [8][8]int, pos Position) []Position {
	var moves []Position
	piece := board[pos.Y][pos.X]
	isWhite := IsPieceWhite(piece)

	directions := []Position{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newPos := Position{pos.X + i*dir.X, pos.Y + i*dir.Y}
			if !IsValidBoardPosition(newPos) {
				break
			}

			targetPiece := board[newPos.Y][newPos.X]
			if targetPiece == Empty {
				moves = append(moves, newPos)
			} else {
				if isWhite != IsPieceWhite(targetPiece) {
					moves = append(moves, newPos)
				}
				break
			}
		}
	}

	return moves
}

func getQueenMoves(board [8][8]int, pos Position) []Position {
	bishopMoves := getBishopMoves(board, pos)
	rookMoves := getRookMoves(board, pos)

	return append(bishopMoves, rookMoves...)
}

func getKingMoves(board [8][8]int, pos Position) []Position {
	var moves []Position
	piece := board[pos.Y][pos.X]
	isWhite := IsPieceWhite(piece)

	directions := []Position{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}

	for _, dir := range directions {
		newPos := Position{pos.X + dir.X, pos.Y + dir.Y}
		if IsValidBoardPosition(newPos) {
			targetPiece := board[newPos.Y][newPos.X]
			if targetPiece == Empty || (isWhite != IsPieceWhite(targetPiece)) {
				moves = append(moves, newPos)
			}
		}
	}

	return moves
}

func GetKingMovesWithCastling(board [8][8]int, pos Position, gameState *GameState) []Position {
	moves := getKingMoves(board, pos)
	piece := board[pos.Y][pos.X]
	isWhite := IsPieceWhite(piece)

	if isWhite {
		if !gameState.WhiteKingMoved && !gameState.WhiteRookAMoved {
			if board[0][1] == Empty && board[0][2] == Empty && board[0][3] == Empty {
				moves = append(moves, Position{2, 0})
			}
		}
		if !gameState.WhiteKingMoved && !gameState.WhiteRookHMoved {
			if board[0][5] == Empty && board[0][6] == Empty {
				moves = append(moves, Position{6, 0})
			}
		}
	} else {
		if !gameState.BlackKingMoved && !gameState.BlackRookAMoved {
			if board[7][1] == Empty && board[7][2] == Empty && board[7][3] == Empty {
				moves = append(moves, Position{2, 7})
			}
		}
		if !gameState.BlackKingMoved && !gameState.BlackRookHMoved {
			if board[7][5] == Empty && board[7][6] == Empty {
				moves = append(moves, Position{6, 7})
			}
		}
	}

	return moves
}
