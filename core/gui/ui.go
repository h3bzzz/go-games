package gui

import (
	"fmt"
	"time"

	"github.com/h3bzzz/go-chess/core/ai"
	"github.com/h3bzzz/go-chess/core/game"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ChessUI struct {
	game           *game.GameState
	window         fyne.Window
	board          *ChessBoard
	status         *widget.Label
	whiteTime      *widget.Label
	blackTime      *widget.Label
	content        *fyne.Container
	timer          *time.Ticker
	timerChan      chan bool
	aiManager      *ai.AIManager
	aiEnabledCheck *widget.Check
	aiColorSelect  *widget.Select
	aiDiffSelect   *widget.Select
}

func NewChessUI(chessGame *game.GameState, window fyne.Window) *ChessUI {
	ui := &ChessUI{
		game:      chessGame,
		window:    window,
		status:    widget.NewLabel("White to move"),
		whiteTime: widget.NewLabel("30:00"),
		blackTime: widget.NewLabel("30:00"),
		timerChan: make(chan bool),
	}

	ui.board = NewChessBoard(chessGame, "classic")
	ui.aiManager = ai.NewAIManager(chessGame)

	// Register callback for AI moves
	ui.aiManager.SetMoveCallback(func() {
		// Queue an update to run on the main thread
		ui.window.Canvas().Refresh(ui.board.GetContainer())
		ui.board.UpdateDisplay()
		ui.updateStatus()
		fmt.Println("Board display updated after AI move")
	})

	ui.createLayout()
	ui.startTimer()

	return ui
}

func (ui *ChessUI) GetContent() fyne.CanvasObject {
	return ui.content
}

func (ui *ChessUI) createLayout() {
	newGameBtn := widget.NewButton("New Game", func() {
		ui.newGame()
	})

	undoBtn := widget.NewButton("Undo Move", func() {
		ui.undoMove()
	})

	testImageBtn := widget.NewButton("Test Images", func() {
		ui.testImages()
	})

	fmt.Println("Initializing theme selector...")
	themes := []string{"Classic", "Green", "Pink"}
	themeSelector := widget.NewSelect(themes, func(selectedTheme string) {
		fmt.Printf("Theme selected from dropdown: %s\n", selectedTheme)
		ui.changeTheme(selectedTheme)
	})
	themeSelector.SetSelected("Classic")
	fmt.Println("Theme selector initialized with Classic theme")

	// AI Controls
	ui.aiEnabledCheck = widget.NewCheck("Enable AI", func(enabled bool) {
		ui.aiManager.SetEnabled(enabled)
	})

	ui.aiColorSelect = widget.NewSelect([]string{"White", "Black"}, func(color string) {
		if color == "White" {
			ui.aiManager.SetAIColor(game.WhitePlayer)
		} else {
			ui.aiManager.SetAIColor(game.BlackPlayer)
		}
	})
	ui.aiColorSelect.SetSelected("Black")

	ui.aiDiffSelect = widget.NewSelect([]string{"Easy", "Medium", "Hard"}, func(level string) {
		var difficulty int
		switch level {
		case "Easy":
			difficulty = 1
		case "Medium":
			difficulty = 2
		case "Hard":
			difficulty = 3
		default:
			difficulty = 2
		}
		ui.aiManager.SetDifficulty(difficulty)
	})
	ui.aiDiffSelect.SetSelected("Medium")

	aiControls := container.NewVBox(
		widget.NewLabel("AI Settings"),
		ui.aiEnabledCheck,
		container.NewHBox(widget.NewLabel("AI plays:"), ui.aiColorSelect),
		container.NewHBox(widget.NewLabel("Difficulty:"), ui.aiDiffSelect),
	)

	// Layout components
	header := container.NewHBox(
		ui.status,
		layout.NewSpacer(),
		widget.NewLabel("Theme:"),
		themeSelector,
		undoBtn,
		newGameBtn,
		testImageBtn,
	)

	whiteLabel := widget.NewLabel("White:")
	blackLabel := widget.NewLabel("Black:")

	footer := container.NewHBox(
		whiteLabel,
		ui.whiteTime,
		layout.NewSpacer(),
		blackLabel,
		ui.blackTime,
	)

	// Create a right panel for AI controls
	rightPanel := container.NewVBox(
		aiControls,
	)

	// Main layout with board in center and AI controls on right
	mainContainer := container.NewBorder(
		nil, nil, nil, rightPanel, ui.board.GetContainer(),
	)

	// Overall layout with header and footer
	ui.content = container.NewBorder(
		header,        // top
		footer,        // bottom
		nil,           // left
		nil,           // right
		mainContainer, // center
	)
}

func (ui *ChessUI) updateStatus() {
	ui.status.SetText(ui.game.GetGameStatus())

	whiteTimeStr := formatTime(ui.game.GetRemainingTime(game.WhitePlayer))
	blackTimeStr := formatTime(ui.game.GetRemainingTime(game.BlackPlayer))

	ui.whiteTime.SetText(whiteTimeStr)
	ui.blackTime.SetText(blackTimeStr)

	if ui.game.IsGameOver() {
		ui.stopTimer()
	}
}

func formatTime(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (ui *ChessUI) startTimer() {
	ui.game.StartTimer()
	ui.timer = time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ui.timer.C:
				ui.updateStatus()
			case <-ui.timerChan:
				return
			}
		}
	}()
}

func (ui *ChessUI) stopTimer() {
	ui.game.StopTimer()
	if ui.timer != nil {
		ui.timer.Stop()
		ui.timerChan <- true
	}
}

func (ui *ChessUI) newGame() {
	ui.stopTimer()

	ui.game = game.NewGame()

	ui.board.game = ui.game
	ui.board.UpdateDisplay()

	ui.aiManager.Stop()
	ui.aiManager = ai.NewAIManager(ui.game)
	if ui.aiEnabledCheck.Checked {
		ui.aiManager.SetEnabled(true)
	}

	ui.updateStatus()

	ui.startTimer()
}

func (ui *ChessUI) undoMove() {
	if ui.game.UndoLastMove() {
		ui.board.UpdateDisplay()
		ui.updateStatus()
	}
}

func (ui *ChessUI) changeTheme(theme string) {
	if ui == nil || ui.board == nil {
		fmt.Println("Cannot change theme: UI or board is nil")
		return
	}

	fmt.Printf("Theme selected: %s\n", theme)

	themeMap := map[string]string{
		"Classic": "chess",
		"Green":   "chess_green",
		"Pink":    "chess_pink",
	}

	folderTheme := themeMap[theme]
	fmt.Printf("Changing to theme folder: %s\n", folderTheme)

	err := ui.board.ChangeTheme(folderTheme)
	if err != nil {
		fmt.Printf("Error changing theme: %v\n", err)
		return
	}

	if ui.content != nil {
		ui.content.Refresh()
		fmt.Println("UI refreshed with new theme")
	}
}

func (ui *ChessUI) testImages() {
	testWindow := fyne.CurrentApp().NewWindow("Image Test")
	testWindow.Resize(fyne.NewSize(400, 400))

	var imgContainer *fyne.Container

	directImg := canvas.NewImageFromFile("assets/chess/white_king.png")
	directImg.SetMinSize(fyne.NewSize(100, 100))
	directImg.FillMode = canvas.ImageFillContain

	res, err := fyne.LoadResourceFromPath("assets/chess/black_king.png")
	var resourceImg *canvas.Image
	if err != nil {
		fmt.Printf("Error loading test resource: %v\n", err)
		errLabel := widget.NewLabel("Failed to load: " + err.Error())
		imgContainer = container.NewVBox(
			widget.NewLabel("Error loading images:"),
			errLabel,
		)
	} else {
		resourceImg = canvas.NewImageFromResource(res)
		resourceImg.SetMinSize(fyne.NewSize(100, 100))
		resourceImg.FillMode = canvas.ImageFillContain

		imgContainer = container.NewHBox(
			container.NewVBox(
				widget.NewLabel("Direct loading:"),
				directImg,
			),
			container.NewVBox(
				widget.NewLabel("Resource loading:"),
				resourceImg,
			),
		)
	}

	testWindow.SetContent(imgContainer)
	testWindow.Show()
}
