package gui

import (
	"fmt"
	"image/color"
	"os"

	"github.com/h3bzzz/go-chess/core/game"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type ChessBoard struct {
	game         *game.GameState
	squares      [8][8]*ChessSquare
	container    *fyne.Container
	pieceManager *PieceManager
	theme        string

	draggedPiece      int
	dragStartPosition game.Position
	highlightedMoves  []game.Position
	isDragging        bool
	dragObj           *DraggablePiece
	overlay           *fyne.Container
}

type ChessSquare struct {
	widget.Button
	board    *ChessBoard
	position game.Position
}

func NewChessSquare(board *ChessBoard, pos game.Position) *ChessSquare {
	square := &ChessSquare{
		board:    board,
		position: pos,
	}
	square.ExtendBaseWidget(square)

	square.OnTapped = func() {
		if board.isDragging {
			board.handleDrop(pos)
		} else {
			board.handleSquareClick(pos)
		}
	}

	return square
}

func (s *ChessSquare) MouseDown(me *desktop.MouseEvent) {
	s.board.handleMouseDown(s.position)
}

type DraggablePiece struct {
	widget.BaseWidget
	resource fyne.Resource
	board    *ChessBoard
}

func NewDraggablePiece(res fyne.Resource, board *ChessBoard) *DraggablePiece {
	piece := &DraggablePiece{
		resource: res,
		board:    board,
	}
	piece.ExtendBaseWidget(piece)
	return piece
}

func (d *DraggablePiece) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromResource(d.resource)
	img.FillMode = canvas.ImageFillContain

	return widget.NewSimpleRenderer(img)
}

func (d *DraggablePiece) Dragged(e *fyne.DragEvent) {
	d.Move(fyne.NewPos(e.Position.X-30, e.Position.Y-30))
}

func (d *DraggablePiece) DragEnd() {
	if d.board != nil {
		d.board.handleDragEnd()
	}
}

func (d *DraggablePiece) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

var (
	lightSquareColor = color.RGBA{240, 217, 181, 255} // Light brown
	darkSquareColor  = color.RGBA{181, 136, 99, 255}  // Dark brown

	selectedSquareColor    = color.RGBA{186, 202, 68, 255}  // Highlighted green
	possibleMoveColor      = color.RGBA{106, 135, 77, 255}  // Darker green
	lastMoveHighlightColor = color.RGBA{206, 210, 107, 255} // Light yellow
)

// Create a custom style method to apply square colors
func (b *ChessBoard) applySquareStyle(square *ChessSquare, isSelected bool, isPossibleMove bool) {
	isLightSquare := (square.position.X+square.position.Y)%2 == 0

	if isSelected {
		square.Importance = widget.HighImportance
	} else if isPossibleMove {
		square.Importance = widget.WarningImportance
	} else if isLightSquare {
		square.Importance = widget.LowImportance
	} else {
		square.Importance = widget.MediumImportance
	}
}

func NewChessBoard(chessGame *game.GameState, theme string) *ChessBoard {
	board := &ChessBoard{
		game:  chessGame,
		theme: theme,
	}

	board.container = container.NewWithoutLayout()
	mainBoard := container.NewGridWithColumns(8)
	board.overlay = container.NewWithoutLayout()

	board.container.Add(mainBoard)
	board.container.Add(board.overlay)

	board.pieceManager = NewPieceManager(theme)
	if err := board.pieceManager.LoadPieces(); err != nil {
		fmt.Printf("Error loading pieces: %v\n", err)
	}

	folder := "chess"
	if theme == "green" {
		folder = "chess_green"
	} else if theme == "pink" {
		folder = "chess_pink"
	}

	assetDir := "assets/" + folder
	fmt.Printf("Checking asset directory: %s\n", assetDir)
	files, err := os.ReadDir(assetDir)
	if err != nil {
		fmt.Printf("Error reading asset directory: %v\n", err)
	} else {
		fmt.Println("Available asset files:")
		for _, file := range files {
			fmt.Printf("  - %s\n", file.Name())
		}
	}

	for y := 7; y >= 0; y-- {
		for x := 0; x < 8; x++ {
			square := NewChessSquare(board, game.Position{X: x, Y: y})

			isLightSquare := (x+y)%2 == 0
			if isLightSquare {
				square.Importance = widget.LowImportance
			} else {
				square.Importance = widget.MediumImportance
			}

			square.Resize(fyne.NewSize(60, 60))

			board.squares[y][x] = square
			mainBoard.Add(square)
		}
	}

	boardSize := fyne.NewSize(480, 480) // 8 squares * 60px
	mainBoard.Resize(boardSize)
	board.container.Resize(boardSize)

	board.UpdateDisplay()

	return board
}

func (b *ChessBoard) handleMouseDown(pos game.Position) {
	piece := b.game.GetPieceAtPosition(pos)
	if piece == game.Empty {
		return
	}

	isWhitePiece := game.IsPieceWhite(piece)
	if (isWhitePiece && b.game.CurrentTurn != game.WhitePlayer) ||
		(!isWhitePiece && b.game.CurrentTurn != game.BlackPlayer) {
		return
	}

	b.draggedPiece = piece
	b.dragStartPosition = pos
	b.isDragging = true

	b.highlightedMoves = b.game.GetPossibleMoves(pos)
	b.highlightPossibleMoves()

	resource := b.pieceManager.GetResource(piece)
	b.dragObj = NewDraggablePiece(resource, b)
	b.dragObj.Resize(fyne.NewSize(60, 60))

	btnPos := b.squares[pos.Y][pos.X].Position()
	b.dragObj.Move(btnPos)

	b.overlay.Add(b.dragObj)
	b.overlay.Refresh()
}

func (b *ChessBoard) handleDragEnd() {
	if !b.isDragging {
		return
	}

	b.isDragging = false
	b.overlay.Remove(b.dragObj)
	b.overlay.Refresh()
	b.dragObj = nil

	b.clearHighlightedMoves()

	b.UpdateDisplay()
}

func (b *ChessBoard) handleDrop(targetPos game.Position) {
	if !b.isDragging {
		return
	}

	isValidMove := false
	for _, move := range b.highlightedMoves {
		if move.X == targetPos.X && move.Y == targetPos.Y {
			isValidMove = true
			break
		}
	}

	if isValidMove {
		b.game.MakeMove(b.dragStartPosition, targetPos)
	}

	b.handleDragEnd()
}

func (b *ChessBoard) highlightPossibleMoves() {
	for _, move := range b.highlightedMoves {
		b.applySquareStyle(b.squares[move.Y][move.X], false, true)
	}
}

func (b *ChessBoard) clearHighlightedMoves() {
	for _, move := range b.highlightedMoves {
		b.applySquareStyle(b.squares[move.Y][move.X], false, false)
	}
	b.highlightedMoves = nil
}

func (b *ChessBoard) GetContainer() fyne.CanvasObject {
	return b.container
}

func (b *ChessBoard) UpdateDisplay() {
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			pos := game.Position{X: x, Y: y}
			piece := b.game.GetPieceAtPosition(pos)

			resource := b.pieceManager.GetResource(piece)
			if resource != nil {
				b.squares[y][x].SetIcon(resource)
			} else {
				b.squares[y][x].SetIcon(nil)
			}

			isSelected := b.game.SelectedPosition != nil &&
				b.game.SelectedPosition.X == x &&
				b.game.SelectedPosition.Y == y

			isPossibleMove := false
			if b.game.SelectedPosition != nil {
				moves := b.game.GetPossibleMoves(*b.game.SelectedPosition)
				for _, move := range moves {
					if move.X == x && move.Y == y {
						isPossibleMove = true
						break
					}
				}
			}

			b.applySquareStyle(b.squares[y][x], isSelected, isPossibleMove)
		}
	}
}

func (b *ChessBoard) handleSquareClick(pos game.Position) {
	b.game.SelectPosition(pos)

	b.UpdateDisplay()
}

func (b *ChessBoard) ChangeTheme(theme string) error {
	if b == nil || b.pieceManager == nil {
		return fmt.Errorf("cannot change theme: board or pieceManager is nil")
	}

	fmt.Printf("ChessBoard: Changing theme from '%s' to '%s'\n", b.theme, theme)

	b.theme = theme

	err := b.pieceManager.ChangeTheme(theme)
	if err != nil {
		return fmt.Errorf("failed to change theme: %v", err)
	}

	b.UpdateDisplay()
	fmt.Printf("ChessBoard: Board display updated with theme '%s'\n", theme)

	return nil
}
