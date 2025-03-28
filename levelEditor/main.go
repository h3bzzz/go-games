package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Create the application with ID
	a := app.NewWithID("io.fyne.gopherleveleditor")
	window := a.NewWindow("Gopher Level Editor")
	window.Resize(fyne.NewSize(1024, 768))

	// Create editor
	editor := NewEditor(window)

	// Setup the editor UI
	editor.BuildUI()

	// Show the window and run the app
	window.ShowAndRun()
}
