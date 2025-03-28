package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SpriteBrowser handles browsing and selecting sprites
type SpriteBrowser struct {
	editor      *Editor
	window      fyne.Window
	currentPath string
}

// NewSpriteBrowser creates a new sprite browser
func NewSpriteBrowser(editor *Editor) *SpriteBrowser {
	return &SpriteBrowser{
		editor: editor,
		window: fyne.CurrentApp().NewWindow("Sprite Browser"),
	}
}

// BrowseCategory opens a window to browse sprites in a category
func (sb *SpriteBrowser) BrowseCategory(category SpriteCategory) {
	// Get the absolute path to the assets directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("ERROR: Failed to get working directory: %v\n", err)
		dialog.ShowError(fmt.Errorf("Failed to get working directory: %v", err), sb.window)
		return
	}

	// Use the working directory to build the absolute path
	path := filepath.Join(workDir, PixelSpacesPath, string(category))
	sb.currentPath = path

	fmt.Printf("DEBUG: Browsing category %s at path: %s\n", category, path)

	// Make sure the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("ERROR: Directory not found: %s (error: %v)\n", path, err)
		dialog.ShowInformation("Directory Not Found",
			fmt.Sprintf("The directory %s was not found. Check asset paths.", path),
			sb.editor.window)
		sb.editor.statusLabel.SetText("Error: Directory not found: " + path)
		return
	}

	fmt.Printf("DEBUG: Directory exists: %s\n", path)

	// Set window title
	sb.window.SetTitle(fmt.Sprintf("Sprite Browser - %s", category))

	// Create content
	content := sb.createBrowserContent(path)

	// Set window content
	sb.window.SetContent(content)
	sb.window.Resize(fyne.NewSize(800, 600))
	sb.window.Show()
}

// createBrowserContent creates content for sprite browser
func (sb *SpriteBrowser) createBrowserContent(path string) fyne.CanvasObject {
	// Create a label showing the current path
	pathLabel := widget.NewLabel("Path: " + path)
	pathLabel.Alignment = fyne.TextAlignCenter

	// Create a scroll container for the sprites
	scroll := container.NewScroll(sb.createSpriteGrid(path))

	// Create navigation buttons
	backButton := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		sb.window.Close()
	})

	// Create reload button
	reloadButton := widget.NewButtonWithIcon("Reload", theme.ViewRefreshIcon(), func() {
		sb.window.SetContent(sb.createBrowserContent(path))
	})

	// Create description label
	descLabel := widget.NewLabel("Select a sprite to use in your level")

	// Create the main container
	return container.NewBorder(
		container.NewVBox(pathLabel, container.NewHBox(backButton, reloadButton), descLabel),
		nil, nil, nil,
		scroll,
	)
}

// createSpriteGrid creates a grid of sprite buttons
func (sb *SpriteBrowser) createSpriteGrid(path string) fyne.CanvasObject {
	fmt.Printf("DEBUG: Creating sprite grid for path: %s\n", path)

	// Get list of files in directory
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("ERROR: Failed to read directory: %s (error: %v)\n", path, err)
		return widget.NewLabel("Error: " + err.Error())
	}

	fmt.Printf("DEBUG: Found %d entries in directory\n", len(entries))

	// Create grid for sprites
	grid := container.NewAdaptiveGrid(4)

	// Add sprites to grid
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Skip non-PNG files
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
			continue
		}

		// Create full path to sprite
		spritePath := filepath.Join(path, entry.Name())

		// Clean up name for display
		displayName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		displayName = strings.ReplaceAll(displayName, "_", " ")

		fmt.Printf("DEBUG: Adding sprite: %s (path: %s)\n", displayName, spritePath)

		// Add sprite button to grid
		grid.Add(sb.createSpriteButton(spritePath, displayName))
	}

	// If no sprites found, show a message
	if len(grid.Objects) == 0 {
		fmt.Printf("WARNING: No PNG sprites found in directory: %s\n", path)
		return widget.NewLabel("No PNG sprites found in this directory")
	}

	return grid
}

// createSpriteButton creates a button with a sprite preview
func (sb *SpriteBrowser) createSpriteButton(path, name string) fyne.CanvasObject {
	// Check if the file exists and print debug info
	_, err := os.Stat(path)
	if err != nil {
		fmt.Printf("ERROR: File not found: %s (error: %v)\n", path, err)
		errLabel := widget.NewLabel("File not found:\n" + path)
		errLabel.Wrapping = fyne.TextWrapWord

		return container.NewVBox(
			errLabel,
			widget.NewButton("Try load anyway", func() {
				sb.selectSprite(path)
			}),
		)
	}

	fmt.Printf("INFO: Loading sprite: %s\n", path)

	// First create a container with a label
	vbox := container.NewVBox(
		widget.NewLabel(name),
	)

	// Create a placeholder for the image
	placeholder := canvas.NewRectangle(color.RGBA{150, 150, 150, 255})
	placeholder.SetMinSize(fyne.NewSize(100, 100))
	vbox.Add(placeholder)

	// Create a URI from the path and try to load image
	uri := storage.NewFileURI(path)
	fmt.Printf("INFO: URI: %s\n", uri.String())

	// Try to load the image
	img := canvas.NewImageFromURI(uri)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(100, 100))

	// This is a hack to try to make the image load properly
	go func() {
		// Wait a bit for the image to load
		time.Sleep(200 * time.Millisecond)

		// Replace the placeholder with the image
		if len(vbox.Objects) > 1 && vbox.Objects[1] == placeholder {
			vbox.Objects[1] = img
			vbox.Refresh()
			fmt.Printf("INFO: Replaced placeholder with image for %s\n", path)
		}
	}()

	// Create a button to select this sprite
	button := widget.NewButton("Select", func() {
		sb.selectSprite(path)
	})
	vbox.Add(button)

	// Add a label with the file path for debugging
	pathLabel := widget.NewLabel(path)
	pathLabel.TextStyle = fyne.TextStyle{Italic: true, Monospace: true}
	pathLabel.Wrapping = fyne.TextWrapWord
	vbox.Add(pathLabel)

	return vbox
}

// createSpritePreview creates a preview for a single sprite
func (sb *SpriteBrowser) createSpritePreview(path string) fyne.CanvasObject {
	// Create a URI from the path
	uri := storage.NewFileURI(path)

	// Read the image
	img := canvas.NewImageFromURI(uri)
	img.FillMode = canvas.ImageFillContain

	// Create a label for the name
	label := widget.NewLabel(filepath.Base(path))

	// Create a button to select this sprite
	selectButton := widget.NewButton("Select This Sprite", func() {
		sb.selectSprite(path)
	})

	// Back button
	backButton := widget.NewButton("Back to Browser", func() {
		dirPath := filepath.Dir(path)
		content := sb.createBrowserContent(dirPath)
		sb.currentPath = dirPath
		sb.window.SetContent(content)
	})

	// Create a container for the sprite preview
	return container.NewBorder(
		container.NewVBox(
			label,
			container.NewHBox(selectButton, backButton),
		),
		nil, nil, nil,
		img,
	)
}

// selectSprite selects a sprite and closes the browser
func (sb *SpriteBrowser) selectSprite(path string) {
	fmt.Printf("DEBUG: Selecting sprite: %s\n", path)

	// Check if the file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("ERROR: File not found: %s (error: %v)\n", path, err)
		dialog.ShowError(fmt.Errorf("File not found: %s\nError: %v", path, err), sb.window)
		return
	}

	// Log file info
	fmt.Printf("DEBUG: File exists: %s (size: %d bytes)\n", path, fileInfo.Size())

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("ERROR: Failed to get absolute path for %s: %v\n", path, err)
	} else {
		fmt.Printf("DEBUG: Absolute path: %s\n", absPath)
		// Use absolute path if available
		path = absPath
	}

	// Set the selected sprite in the editor
	sb.editor.selectedSprite = path

	// Update status
	sb.editor.statusLabel.SetText("Selected sprite: " + filepath.Base(path))

	// Close the browser window
	sb.window.Close()

	// Refresh editor properties panel to show the selected sprite
	sb.editor.updatePropertiesPanel()

	fmt.Printf("DEBUG: Selected sprite and updated properties panel\n")
}

// isImageFile checks if a file is an image based on extension
func isImageFile(filename string) bool {
	extension := strings.ToLower(filepath.Ext(filename))
	return extension == ".png" || extension == ".jpg" || extension == ".jpeg" || extension == ".gif"
}
