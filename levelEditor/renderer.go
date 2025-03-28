package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// GameObjectRenderer handles rendering game objects on the canvas
type GameObjectRenderer struct {
	editor *Editor
}

// NewGameObjectRenderer creates a new renderer
func NewGameObjectRenderer(editor *Editor) *GameObjectRenderer {
	return &GameObjectRenderer{
		editor: editor,
	}
}

// RenderLevel renders all objects in the level on the canvas
func (r *GameObjectRenderer) RenderLevel() fyne.CanvasObject {
	// Create a container for all game objects
	objectsContainer := container.NewWithoutLayout()

	// Base is either the background image or a colored rectangle
	var base fyne.CanvasObject
	if r.editor.currentLevel.Background != "" && fileExists(r.editor.currentLevel.Background) {
		// Use background image
		bgURI := storage.NewFileURI(r.editor.currentLevel.Background)
		bgImage := canvas.NewImageFromURI(bgURI)
		bgImage.FillMode = canvas.ImageFillStretch
		bgImage.Resize(fyne.NewSize(
			float32(r.editor.currentLevel.Width*r.editor.gridSize),
			float32(r.editor.currentLevel.Height*r.editor.gridSize),
		))
		base = bgImage
	} else {
		// Use colored rectangle
		rect := canvas.NewRectangle(color.RGBA{40, 40, 60, 255})
		rect.Resize(fyne.NewSize(
			float32(r.editor.currentLevel.Width*r.editor.gridSize),
			float32(r.editor.currentLevel.Height*r.editor.gridSize),
		))
		base = rect
	}

	// Add the base to the container
	objectsContainer.Add(base)
	base.Move(fyne.NewPos(0, 0))

	// Draw grid lines
	r.drawGrid(objectsContainer)

	// Draw platforms
	for _, platform := range r.editor.currentLevel.Platforms {
		r.renderPlatform(objectsContainer, platform)
	}

	// Draw enemies
	for _, enemy := range r.editor.currentLevel.Enemies {
		r.renderEnemy(objectsContainer, enemy)
	}

	// Draw collectibles
	for _, item := range r.editor.currentLevel.Collectibles {
		r.renderItem(objectsContainer, item)
	}

	// Draw player start position
	r.renderStartPosition(objectsContainer)

	// Draw level exit
	r.renderExitPosition(objectsContainer)

	return objectsContainer
}

// drawGrid draws grid lines on the canvas
func (r *GameObjectRenderer) drawGrid(container *fyne.Container) {
	gridColor := color.RGBA{100, 100, 100, 100} // Transparent gray

	// Draw vertical lines
	for x := 0; x <= r.editor.currentLevel.Width; x++ {
		line := canvas.NewLine(gridColor)
		line.StrokeWidth = 1
		xPos := float32(x * r.editor.gridSize)
		line.Position1 = fyne.NewPos(xPos, 0)
		line.Position2 = fyne.NewPos(xPos, float32(r.editor.currentLevel.Height*r.editor.gridSize))
		container.Add(line)
	}

	// Draw horizontal lines
	for y := 0; y <= r.editor.currentLevel.Height; y++ {
		line := canvas.NewLine(gridColor)
		line.StrokeWidth = 1
		yPos := float32(y * r.editor.gridSize)
		line.Position1 = fyne.NewPos(0, yPos)
		line.Position2 = fyne.NewPos(float32(r.editor.currentLevel.Width*r.editor.gridSize), yPos)
		container.Add(line)
	}
}

// renderPlatform renders a platform on the canvas
func (r *GameObjectRenderer) renderPlatform(container *fyne.Container, platform Platform) {
	// Calculate position and size
	x := float32(platform.X * r.editor.gridSize)
	y := float32(platform.Y * r.editor.gridSize)
	width := float32(platform.Width * r.editor.gridSize)
	height := float32(platform.Height * r.editor.gridSize)

	var obj fyne.CanvasObject

	// Use sprite if available, otherwise use colored rectangle
	if platform.SpritePath != "" {
		// Check if file exists and log details
		_, err := os.Stat(platform.SpritePath)
		if err != nil {
			// Log error but continue
			fmt.Printf("ERROR: Failed to load platform sprite: %s (error: %v)\n", platform.SpritePath, err)
		} else {
			fmt.Printf("DEBUG: Loading platform sprite: %s\n", platform.SpritePath)

			// Try to get absolute path
			absPath, err := filepath.Abs(platform.SpritePath)
			if err == nil {
				fmt.Printf("DEBUG: Absolute path: %s\n", absPath)
				platform.SpritePath = absPath
			}

			uri := storage.NewFileURI(platform.SpritePath)
			fmt.Printf("DEBUG: Platform sprite URI: %s\n", uri.String())

			img := canvas.NewImageFromURI(uri)
			img.FillMode = canvas.ImageFillStretch
			img.Resize(fyne.NewSize(width, height))
			obj = img

			// Add rectangle outline to show bounds
			rect := canvas.NewRectangle(color.RGBA{0, 0, 255, 100})
			rect.Resize(fyne.NewSize(width, height))
			container.Add(rect)
			rect.Move(fyne.NewPos(x, y))
		}
	}

	// If we didn't create an image object, create a fallback
	if obj == nil {
		// Color based on platform type
		var platformColor color.RGBA

		switch platform.Type {
		case "moving":
			platformColor = color.RGBA{150, 150, 255, 255} // Blue for moving
		case "breaking":
			platformColor = color.RGBA{255, 150, 150, 255} // Red for breaking
		case "bouncy":
			platformColor = color.RGBA{150, 255, 150, 255} // Green for bouncy
		default:
			platformColor = color.RGBA{200, 200, 200, 255} // Gray for normal
		}

		rect := canvas.NewRectangle(platformColor)
		rect.Resize(fyne.NewSize(width, height))
		obj = rect

		fmt.Printf("DEBUG: Using fallback rectangle for platform at %d,%d\n", platform.X, platform.Y)
	}

	container.Add(obj)
	obj.Move(fyne.NewPos(x, y))

	// Add a label for platform type and size
	details := fmt.Sprintf("%s %dx%d\n", platform.Type, platform.Width, platform.Height)
	if platform.SpritePath != "" {
		details += filepath.Base(platform.SpritePath)
	}

	label := widget.NewLabel(details)
	label.TextStyle.Bold = true
	container.Add(label)
	label.Move(fyne.NewPos(x+5, y+5))
}

// renderEnemy renders an enemy on the canvas
func (r *GameObjectRenderer) renderEnemy(container *fyne.Container, enemy Enemy) {
	// Calculate position and size
	x := float32(enemy.X * r.editor.gridSize)
	y := float32(enemy.Y * r.editor.gridSize)
	size := float32(r.editor.gridSize)

	var obj fyne.CanvasObject

	// Use sprite if available, otherwise use colored circle
	if enemy.SpritePath != "" {
		// Check if file exists and log details
		_, err := os.Stat(enemy.SpritePath)
		if err != nil {
			// Log error but continue
			fmt.Printf("ERROR: Failed to load enemy sprite: %s (error: %v)\n", enemy.SpritePath, err)
		} else {
			fmt.Printf("DEBUG: Loading enemy sprite: %s\n", enemy.SpritePath)

			// Try to get absolute path
			absPath, err := filepath.Abs(enemy.SpritePath)
			if err == nil {
				fmt.Printf("DEBUG: Absolute path: %s\n", absPath)
				enemy.SpritePath = absPath
			}

			uri := storage.NewFileURI(enemy.SpritePath)
			fmt.Printf("DEBUG: Enemy sprite URI: %s\n", uri.String())

			img := canvas.NewImageFromURI(uri)
			img.FillMode = canvas.ImageFillContain
			img.Resize(fyne.NewSize(size, size))
			obj = img

			// Add rectangle outline to show bounds
			rect := canvas.NewRectangle(color.RGBA{255, 0, 0, 100})
			rect.Resize(fyne.NewSize(size, size))
			container.Add(rect)
			rect.Move(fyne.NewPos(x, y))
		}
	}

	// If we didn't create an image object, create a fallback
	if obj == nil {
		// Color based on enemy type
		enemyColor := color.RGBA{255, 0, 0, 255} // Red for default

		circle := canvas.NewCircle(enemyColor)
		circle.Resize(fyne.NewSize(size, size))
		obj = circle

		fmt.Printf("DEBUG: Using fallback circle for enemy at %d,%d\n", enemy.X, enemy.Y)
	}

	container.Add(obj)
	obj.Move(fyne.NewPos(x, y))

	// Add a label for enemy type and sprite path
	details := fmt.Sprintf("%s\n", enemy.Type)
	if enemy.SpritePath != "" {
		details += filepath.Base(enemy.SpritePath)
	}

	label := widget.NewLabel(details)
	label.TextStyle.Bold = true
	container.Add(label)
	label.Move(fyne.NewPos(x, y+size))
}

// renderItem renders a collectible item on the canvas
func (r *GameObjectRenderer) renderItem(container *fyne.Container, item Item) {
	// Calculate position and size
	x := float32(item.X * r.editor.gridSize)
	y := float32(item.Y * r.editor.gridSize)
	size := float32(r.editor.gridSize * 3 / 4) // Slightly smaller than grid

	var obj fyne.CanvasObject

	// Use sprite if available, otherwise use colored shape
	if item.SpritePath != "" {
		// Check if file exists and log details
		_, err := os.Stat(item.SpritePath)
		if err != nil {
			// Log error but continue
			fmt.Printf("ERROR: Failed to load item sprite: %s (error: %v)\n", item.SpritePath, err)
		} else {
			fmt.Printf("DEBUG: Loading item sprite: %s\n", item.SpritePath)

			// Try to get absolute path
			absPath, err := filepath.Abs(item.SpritePath)
			if err == nil {
				fmt.Printf("DEBUG: Absolute path: %s\n", absPath)
				item.SpritePath = absPath
			}

			uri := storage.NewFileURI(item.SpritePath)
			fmt.Printf("DEBUG: Item sprite URI: %s\n", uri.String())

			img := canvas.NewImageFromURI(uri)
			img.FillMode = canvas.ImageFillContain
			img.Resize(fyne.NewSize(size, size))
			obj = img

			// Add rectangle outline to show bounds
			rect := canvas.NewRectangle(color.RGBA{0, 255, 0, 100})
			rect.Resize(fyne.NewSize(size, size))
			container.Add(rect)
			rect.Move(fyne.NewPos(x+(float32(r.editor.gridSize)-size)/2, y+(float32(r.editor.gridSize)-size)/2))
		}
	}

	// If we didn't create an image object, create a fallback
	if obj == nil {
		// Color based on item type
		var itemColor color.RGBA

		switch item.Type {
		case "gem":
			itemColor = color.RGBA{0, 0, 255, 255} // Blue for gem
		case "key":
			itemColor = color.RGBA{255, 255, 0, 255} // Yellow for key
		default:
			itemColor = color.RGBA{255, 215, 0, 255} // Gold for coin/default
		}

		// Create a star-like shape using a rectangle
		rect := canvas.NewRectangle(itemColor)
		rect.Resize(fyne.NewSize(size, size))
		obj = rect

		fmt.Printf("DEBUG: Using fallback shape for item at %d,%d\n", item.X, item.Y)
	}

	container.Add(obj)
	// Center in grid
	obj.Move(fyne.NewPos(
		x+(float32(r.editor.gridSize)-size)/2,
		y+(float32(r.editor.gridSize)-size)/2,
	))

	// Add a label for item type and value
	details := fmt.Sprintf("%s (%d)\n", item.Type, item.Value)
	if item.SpritePath != "" {
		details += filepath.Base(item.SpritePath)
	}

	label := widget.NewLabel(details)
	label.TextStyle.Bold = true
	container.Add(label)
	label.Move(fyne.NewPos(x, y+float32(r.editor.gridSize)))
}

// renderStartPosition renders the player start position
func (r *GameObjectRenderer) renderStartPosition(container *fyne.Container) {
	// Calculate position
	x := float32(r.editor.currentLevel.StartX * r.editor.gridSize)
	y := float32(r.editor.currentLevel.StartY * r.editor.gridSize)
	size := float32(r.editor.gridSize)

	// Draw a green circle for start position
	circle := canvas.NewCircle(color.RGBA{0, 255, 0, 255})
	circle.Resize(fyne.NewSize(size, size))
	container.Add(circle)
	circle.Move(fyne.NewPos(x, y))

	// Add a label
	label := widget.NewLabel("START")
	label.TextStyle.Bold = true
	container.Add(label)
	label.Move(fyne.NewPos(x, y+size))
}

// renderExitPosition renders the level exit
func (r *GameObjectRenderer) renderExitPosition(container *fyne.Container) {
	// Calculate position
	x := float32(r.editor.currentLevel.ExitX * r.editor.gridSize)
	y := float32(r.editor.currentLevel.ExitY * r.editor.gridSize)
	size := float32(r.editor.gridSize)

	// Draw a blue circle for exit position
	circle := canvas.NewCircle(color.RGBA{0, 0, 255, 255})
	circle.Resize(fyne.NewSize(size, size))
	container.Add(circle)
	circle.Move(fyne.NewPos(x, y))

	// Add a label
	label := widget.NewLabel("EXIT")
	label.TextStyle.Bold = true
	container.Add(label)
	label.Move(fyne.NewPos(x, y+size))
}
