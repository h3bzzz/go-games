package main

import (
	"fmt"
	"os"

	"github.com/h3bzzz/go-chess/core/game"
	"github.com/h3bzzz/go-chess/core/gui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	chessGame := game.NewGame()

	chessApp := app.New()

	window := chessApp.NewWindow("Chess Game")
	window.Resize(fyne.NewSize(800, 600))

	_, err := os.Stat("assets/chess")
	if err != nil {
		fmt.Printf("Error accessing assets directory: %v\n", err)
		fmt.Println("Make sure you run the app from the project root directory!")
		fmt.Println("Current working directory:")
		pwd, _ := os.Getwd()
		fmt.Println(pwd)
		window.SetContent(container.NewWithoutLayout())
		window.Show()
		return
	}

	chessUI := gui.NewChessUI(chessGame, window)

	if chessUI != nil {
		window.SetContent(chessUI.GetContent())
	} else {
		fmt.Println("Error: Failed to create chess UI")
		window.SetContent(container.NewWithoutLayout())
	}

	window.ShowAndRun()
}
