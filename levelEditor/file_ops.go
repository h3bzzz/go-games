package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// FileOperations handles file-related operations for the level editor
type FileOperations struct {
	editor *Editor
}

// NewFileOperations creates a new file operations handler
func NewFileOperations(editor *Editor) *FileOperations {
	return &FileOperations{
		editor: editor,
	}
}

// NewLevel creates a new level and prompts for dimensions
func (f *FileOperations) NewLevel() {
	// Create dialog for new level dimensions
	nameEntry := widget.NewEntry()
	nameEntry.Text = "New Level"

	widthEntry := widget.NewEntry()
	widthEntry.Text = "60"

	heightEntry := widget.NewEntry()
	heightEntry.Text = "40"

	items := []*widget.FormItem{
		widget.NewFormItem("Level Name", nameEntry),
		widget.NewFormItem("Width", widthEntry),
		widget.NewFormItem("Height", heightEntry),
	}

	dialog.ShowForm("New Level", "Create", "Cancel", items, func(confirm bool) {
		if confirm {
			// Parse dimensions
			var width, height int
			fmt.Sscanf(widthEntry.Text, "%d", &width)
			fmt.Sscanf(heightEntry.Text, "%d", &height)

			// Create new level
			f.editor.currentLevel = NewLevel(nameEntry.Text, width, height)
			f.editor.statusLabel.SetText("Created new level: " + nameEntry.Text)
		}
	}, f.editor.window)
}

// OpenLevel opens a level file
func (f *FileOperations) OpenLevel() {
	// Create file open dialog
	dlg := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, f.editor.window)
			return
		}
		if read == nil {
			return // User cancelled
		}

		// Read file content
		defer read.Close()
		data, err := io.ReadAll(read)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to read file: %v", err), f.editor.window)
			return
		}

		// Parse JSON
		var level Level
		err = json.Unmarshal(data, &level)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to parse level data: %v", err), f.editor.window)
			return
		}

		// Set the level
		f.editor.currentLevel = &level
		f.editor.statusLabel.SetText("Opened level: " + level.Name)
	}, f.editor.window)

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	dlg.Show()
}

// SaveLevel saves the current level
func (f *FileOperations) SaveLevel() {
	// Create file save dialog
	dlg := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, f.editor.window)
			return
		}
		if write == nil {
			return // User cancelled
		}

		// Serialize level to JSON
		data, err := json.MarshalIndent(f.editor.currentLevel, "", "  ")
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to serialize level: %v", err), f.editor.window)
			return
		}

		// Write to file
		_, err = write.Write(data)
		write.Close()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save file: %v", err), f.editor.window)
			return
		}

		f.editor.statusLabel.SetText("Level saved successfully")
	}, f.editor.window)

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	dlg.SetFileName(f.editor.currentLevel.Name + ".json")
	dlg.Show()
}

// ExportLevel exports the level to the game format
func (f *FileOperations) ExportLevel() {
	// Create file save dialog
	dlg := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, f.editor.window)
			return
		}
		if write == nil {
			return // User cancelled
		}

		// Serialize level to JSON (game format)
		data, err := json.MarshalIndent(f.editor.currentLevel, "", "  ")
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to serialize level: %v", err), f.editor.window)
			return
		}

		// Write to file
		_, err = write.Write(data)
		write.Close()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to export file: %v", err), f.editor.window)
			return
		}

		f.editor.statusLabel.SetText("Level exported successfully")
	}, f.editor.window)

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	dlg.SetFileName(f.editor.currentLevel.Name + "_export.json")
	dlg.Show()
}

// SelectBackground lets the user select a background image
func (f *FileOperations) SelectBackground() {
	// Create file open dialog
	dlg := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, f.editor.window)
			return
		}
		if read == nil {
			return // User cancelled
		}

		// Set the background path
		f.editor.currentLevel.Background = read.URI().Path()
		f.editor.statusLabel.SetText("Background selected: " + filepath.Base(read.URI().Path()))

		// Update the canvas to show the new background
		f.editor.updateEditorCanvas()
	}, f.editor.window)

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png"}))
	dlg.Show()
}
