package test

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func RunImageTest() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
	} else {
		fmt.Printf("Current working directory: %s\n", cwd)
	}

	testApp := app.New()
	window := testApp.NewWindow("Chess Image Test")
	window.Resize(fyne.NewSize(600, 600))

	chessFolderPath := filepath.Join("assets", "chess")
	files, err := os.ReadDir(chessFolderPath)
	if err != nil {
		fmt.Printf("Error reading chess folder: %v\n", err)
	} else {
		fmt.Printf("Chess folder contents (%s):\n", chessFolderPath)
		for _, file := range files {
			fmt.Printf("  - %s\n", file.Name())
		}
	}

	content := container.NewVBox(
		widget.NewLabel("Chess Piece Images Test"),
	)

	whitePieces := container.NewHBox()

	pawnPath := filepath.Join(chessFolderPath, "white_pawn.png")
	fmt.Printf("Loading pawn from: %s\n", pawnPath)

	pawnRes, err := fyne.LoadResourceFromPath(pawnPath)
	if err != nil {
		fmt.Printf("Error loading pawn: %v\n", err)
		whitePieces.Add(widget.NewLabel("Pawn Error"))
	} else {
		pawnImg := canvas.NewImageFromResource(pawnRes)
		pawnImg.SetMinSize(fyne.NewSize(50, 50))
		pawnImg.FillMode = canvas.ImageFillContain
		whitePieces.Add(pawnImg)
	}

	pieceTypes := []string{"pawn", "knight", "bishop", "rook", "queen", "king"}
	for _, piece := range pieceTypes {
		piecePath := filepath.Join(chessFolderPath, "white_"+piece+".png")
		pieceRes, err := fyne.LoadResourceFromPath(piecePath)
		if err != nil {
			fmt.Printf("Error loading %s: %v\n", piece, err)
			continue
		}

		pieceImg := canvas.NewImageFromResource(pieceRes)
		pieceImg.SetMinSize(fyne.NewSize(50, 50))
		pieceImg.FillMode = canvas.ImageFillContain
		whitePieces.Add(pieceImg)
	}

	content.Add(whitePieces)

	boardRow := container.NewHBox()
	for i := 0; i < 8; i++ {
		btn := widget.NewButton("", nil)
		if i%2 == 0 {
			btn.Importance = widget.LowImportance
		} else {
			btn.Importance = widget.MediumImportance
		}

		if i%2 == 0 {
			piecePath := filepath.Join(chessFolderPath, "white_pawn.png")
			pieceRes, err := fyne.LoadResourceFromPath(piecePath)
			if err == nil {
				btn.SetIcon(pieceRes)
			}
		}

		boardRow.Add(btn)
	}
	content.Add(boardRow)

	window.SetContent(content)
	window.ShowAndRun()
}
