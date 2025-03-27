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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Editor represents the level editor
type Editor struct {
	// Main application components
	window  fyne.Window
	content *fyne.Container

	// UI elements
	statusLabel  *widget.Label
	canvasArea   *container.Scroll
	editorCanvas fyne.CanvasObject

	// Editor state
	currentTool  ToolType
	currentLevel *Level
	gridSize     int

	// Platform settings
	platformWidth  int
	platformHeight int
	platformType   string

	// File operations
	fileOps *FileOperations

	// Sprite management
	selectedSprite string
	sprites        map[string]string // Maps object IDs to sprite paths

	// Renderer
	renderer *GameObjectRenderer
}

// ToolType represents different editing tools
type ToolType int

const (
	ToolSelect ToolType = iota
	ToolPlatform
	ToolEnemy
	ToolCollectible
	ToolStart
	ToolExit
	ToolEraser
)

// SpriteCategory represents different sprite categories
type SpriteCategory string

const (
	SpriteCategoryBackground SpriteCategory = "Backgrounds"
	SpriteCategoryFurniture  SpriteCategory = "Furniture"
	SpriteCategoryObjects    SpriteCategory = "Objects"
	SpriteCategoryNPCs       SpriteCategory = "Pre-made NPCs"

	// Furniture subcategories
	SpriteCategoryLivingRoom SpriteCategory = "Furniture/Living Room"
	SpriteCategoryKitchen    SpriteCategory = "Furniture/Kitchen"
	SpriteCategoryBathroom   SpriteCategory = "Furniture/Bathroom"
	SpriteCategoryBedroom    SpriteCategory = "Furniture/Bedroom"

	// Objects subcategories
	SpriteCategoryObjLivingRoom SpriteCategory = "Objects/Living Room"
	SpriteCategoryObjKitchen    SpriteCategory = "Objects/Kitchen"
	SpriteCategoryObjBathroom   SpriteCategory = "Objects/Bathroom"
	SpriteCategoryObjBedroom    SpriteCategory = "Objects/Bedroom"
)

// Asset paths
const (
	PixelSpacesPath = "mazeGame/assets/PixelSpaces Free Pack"
)

// NewEditor creates a new level editor
func NewEditor(window fyne.Window) *Editor {
	editor := &Editor{
		window:         window,
		currentTool:    ToolSelect,
		gridSize:       32, // Default grid size
		sprites:        make(map[string]string),
		platformWidth:  3, // Default platform width
		platformHeight: 1, // Default platform height
		platformType:   "solid",
	}

	// Create a default level
	editor.currentLevel = NewLevel("New Level", 60, 40)

	// Initialize file operations
	editor.fileOps = NewFileOperations(editor)

	// Initialize renderer
	editor.renderer = NewGameObjectRenderer(editor)

	return editor
}

// BuildUI creates the editor interface
func (e *Editor) BuildUI() {
	// Create the editor canvas with background
	e.updateEditorCanvas()

	// Create a scrollable container for the canvas
	e.canvasArea = container.NewScroll(e.editorCanvas)

	// Create the status bar
	e.statusLabel = widget.NewLabel("Ready")
	statusBar := container.NewHBox(e.statusLabel)

	// Create the main layout
	e.content = container.NewBorder(
		e.createToolbar(),    // Top toolbar
		statusBar,            // Bottom status bar
		e.createTools(),      // Left toolbar
		e.createProperties(), // Right properties panel
		e.canvasArea,         // Center canvas area
	)

	// Set the content to the window
	e.window.SetContent(e.content)

	// Add mouse listeners directly to the content
	canvas := e.window.Canvas()
	canvas.SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if ke.Name == fyne.KeyEscape {
			e.statusLabel.SetText("Tool: " + e.getToolName())
		}
	})

	// Create a clickable canvas area using a custom widget
	clickableArea := NewClickableArea(e)

	// Create a container with both the editor canvas and the clickable area
	combinedContainer := container.NewStack(e.editorCanvas, clickableArea)

	// Update the scroll container with our combined content
	e.canvasArea.Content = combinedContainer
	e.canvasArea.Refresh()
}

// updateEditorCanvas updates the canvas display based on current level data
func (e *Editor) updateEditorCanvas() {
	// Use the renderer to create the game object display
	levelView := e.renderer.RenderLevel()

	// Set as the editor canvas
	e.editorCanvas = levelView

	// If we're in a container, update it
	if e.canvasArea != nil {
		e.canvasArea.Content = e.editorCanvas
		e.canvasArea.Refresh()
	}
}

// fileExists checks if a file exists at the given path
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// placeObject places an object on the canvas at the given position
func (e *Editor) placeObject(x, y int) {
	// Calculate grid position
	gridX := x / e.gridSize
	gridY := y / e.gridSize

	// Determine the type of object to place based on the current tool
	switch e.currentTool {
	case ToolPlatform:
		// For platforms, we'll handle it specially to allow different sizes
		e.createPlatform(gridX, gridY)

	case ToolEnemy:
		// Create an enemy at the position
		enemy := Enemy{
			X:               gridX,
			Y:               gridY,
			Type:            "basic",
			Behavior:        "patrol",
			Speed:           1.0,
			DetectionRadius: 5,
		}

		// If a sprite is selected, assign it
		if e.selectedSprite != "" {
			enemy.SpritePath = e.selectedSprite
		}

		// Add to level
		e.currentLevel.Enemies = append(e.currentLevel.Enemies, enemy)
		e.statusLabel.SetText(fmt.Sprintf("Enemy added at %d,%d", gridX, gridY))

	case ToolCollectible:
		// Create an item at the position
		item := Item{
			X:     gridX,
			Y:     gridY,
			Type:  "coin",
			Value: 1,
		}

		// If a sprite is selected, assign it
		if e.selectedSprite != "" {
			item.SpritePath = e.selectedSprite
		}

		// Add to level
		e.currentLevel.Collectibles = append(e.currentLevel.Collectibles, item)
		e.statusLabel.SetText(fmt.Sprintf("Item added at %d,%d", gridX, gridY))

	case ToolStart:
		// Set start position
		e.currentLevel.StartX = gridX
		e.currentLevel.StartY = gridY
		e.statusLabel.SetText(fmt.Sprintf("Start position set to %d,%d", gridX, gridY))

	case ToolExit:
		// Set exit position
		e.currentLevel.ExitX = gridX
		e.currentLevel.ExitY = gridY
		e.statusLabel.SetText(fmt.Sprintf("Exit position set to %d,%d", gridX, gridY))

	case ToolEraser:
		// Remove any objects at this position
		e.eraseAt(gridX, gridY)
	}

	// Update the canvas to show the changes
	e.renderLevelObjects()
}

// createPlatform creates a new platform using the current platform settings
func (e *Editor) createPlatform(gridX, gridY int) {
	// Create a platform at the position
	platform := Platform{
		X:      gridX,
		Y:      gridY,
		Width:  e.platformWidth,
		Height: e.platformHeight,
		Type:   e.platformType,
	}

	// If a sprite is selected, assign it
	if e.selectedSprite != "" {
		platform.SpritePath = e.selectedSprite
	}

	// Add to level
	e.currentLevel.Platforms = append(e.currentLevel.Platforms, platform)
	e.statusLabel.SetText(fmt.Sprintf("Platform added at %d,%d", gridX, gridY))
}

// eraseAt removes objects at the given grid position
func (e *Editor) eraseAt(gridX, gridY int) {
	removed := false

	// Check platforms
	for i := len(e.currentLevel.Platforms) - 1; i >= 0; i-- {
		p := e.currentLevel.Platforms[i]
		if gridX >= p.X && gridX < p.X+p.Width &&
			gridY >= p.Y && gridY < p.Y+p.Height {
			// Remove this platform
			e.currentLevel.Platforms = append(e.currentLevel.Platforms[:i], e.currentLevel.Platforms[i+1:]...)
			removed = true
		}
	}

	// Check enemies
	for i := len(e.currentLevel.Enemies) - 1; i >= 0; i-- {
		if e.currentLevel.Enemies[i].X == gridX && e.currentLevel.Enemies[i].Y == gridY {
			// Remove this enemy
			e.currentLevel.Enemies = append(e.currentLevel.Enemies[:i], e.currentLevel.Enemies[i+1:]...)
			removed = true
		}
	}

	// Check collectibles
	for i := len(e.currentLevel.Collectibles) - 1; i >= 0; i-- {
		if e.currentLevel.Collectibles[i].X == gridX && e.currentLevel.Collectibles[i].Y == gridY {
			// Remove this collectible
			e.currentLevel.Collectibles = append(e.currentLevel.Collectibles[:i], e.currentLevel.Collectibles[i+1:]...)
			removed = true
		}
	}

	if removed {
		e.statusLabel.SetText(fmt.Sprintf("Objects erased at %d,%d", gridX, gridY))
	} else {
		e.statusLabel.SetText(fmt.Sprintf("No objects found at %d,%d", gridX, gridY))
	}
}

// renderLevelObjects renders all the objects in the level
func (e *Editor) renderLevelObjects() {
	// Use our renderer to update the canvas
	e.updateEditorCanvas()
}

// registerEventHandlers sets up event handlers
func (e *Editor) registerEventHandlers() {
	// Key handlers
	e.window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if ke.Name == fyne.KeyEscape {
			e.statusLabel.SetText("Tool: " + e.getToolName())
		}
	})

	// We'll add a tap handler to the canvas area after it's set up
	// This is done in BuildUI after the canvas area is created
}

// Create a custom clickable widget
type ClickableArea struct {
	widget.BaseWidget
	editor *Editor
}

func NewClickableArea(editor *Editor) *ClickableArea {
	area := &ClickableArea{
		editor: editor,
	}
	area.ExtendBaseWidget(area)
	return area
}

func (c *ClickableArea) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.RGBA{0, 0, 0, 0}) // Transparent background
	return widget.NewSimpleRenderer(r)
}

func (c *ClickableArea) MinSize() fyne.Size {
	// Match the editor canvas size
	return c.editor.editorCanvas.Size()
}

func (c *ClickableArea) Tapped(ev *fyne.PointEvent) {
	// Calculate grid position including scroll offset
	gridX := int(ev.Position.X + c.editor.canvasArea.Offset.X)
	gridY := int(ev.Position.Y + c.editor.canvasArea.Offset.Y)

	// Check if within level bounds
	if gridX >= 0 && gridY >= 0 &&
		gridX < c.editor.currentLevel.Width*c.editor.gridSize &&
		gridY < c.editor.currentLevel.Height*c.editor.gridSize {

		// Snap to grid
		gridX = (gridX / c.editor.gridSize) * c.editor.gridSize
		gridY = (gridY / c.editor.gridSize) * c.editor.gridSize

		// Place object at grid position
		c.editor.placeObject(gridX, gridY)

		c.editor.statusLabel.SetText(fmt.Sprintf("Placed at %d,%d", gridX/c.editor.gridSize, gridY/c.editor.gridSize))
	} else {
		c.editor.statusLabel.SetText("Click outside level bounds")
	}
}

// createToolbar creates the top menu bar and toolbar
func (e *Editor) createToolbar() fyne.CanvasObject {
	// Create file menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New", func() { e.fileOps.NewLevel() }),
		fyne.NewMenuItem("Open...", func() { e.fileOps.OpenLevel() }),
		fyne.NewMenuItem("Save", func() { e.fileOps.SaveLevel() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Export to Game", func() { e.fileOps.ExportLevel() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() { e.window.Close() }),
	)

	// Create view menu
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Zoom In", func() { e.statusLabel.SetText("Zoom In (not implemented)") }),
		fyne.NewMenuItem("Zoom Out", func() { e.statusLabel.SetText("Zoom Out (not implemented)") }),
		fyne.NewMenuItem("Reset Zoom", func() { e.statusLabel.SetText("Reset Zoom (not implemented)") }),
	)

	// Create furniture submenu
	furnitureSubmenu := fyne.NewMenu("Furniture",
		fyne.NewMenuItem("All Furniture", func() { e.showSpriteSelector(SpriteCategoryFurniture) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Living Room", func() { e.showSpriteSelector(SpriteCategoryLivingRoom) }),
		fyne.NewMenuItem("Kitchen", func() { e.showSpriteSelector(SpriteCategoryKitchen) }),
		fyne.NewMenuItem("Bathroom", func() { e.showSpriteSelector(SpriteCategoryBathroom) }),
		fyne.NewMenuItem("Bedroom", func() { e.showSpriteSelector(SpriteCategoryBedroom) }),
	)

	// Create objects submenu
	objectsSubmenu := fyne.NewMenu("Objects",
		fyne.NewMenuItem("All Objects", func() { e.showSpriteSelector(SpriteCategoryObjects) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Living Room", func() { e.showSpriteSelector(SpriteCategoryObjLivingRoom) }),
		fyne.NewMenuItem("Kitchen", func() { e.showSpriteSelector(SpriteCategoryObjKitchen) }),
		fyne.NewMenuItem("Bathroom", func() { e.showSpriteSelector(SpriteCategoryObjBathroom) }),
		fyne.NewMenuItem("Bedroom", func() { e.showSpriteSelector(SpriteCategoryObjBedroom) }),
	)

	// Create sprites menu
	spritesMenu := fyne.NewMenu("Sprites",
		fyne.NewMenuItem("Backgrounds", func() { e.showSpriteSelector(SpriteCategoryBackground) }),
		fyne.NewMenuItem("NPCs", func() { e.showSpriteSelector(SpriteCategoryNPCs) }),
		fyne.NewMenuItemSeparator(),
	)

	// Add furniture submenu
	for _, item := range furnitureSubmenu.Items {
		spritesMenu.Items = append(spritesMenu.Items, item)
	}

	// Add objects submenu
	for _, item := range objectsSubmenu.Items {
		spritesMenu.Items = append(spritesMenu.Items, item)
	}

	// Add browse all option
	spritesMenu.Items = append(spritesMenu.Items,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Browse All...", func() { e.showAllSpriteCategories() }),
	)

	// Set the main menu
	mainMenu := fyne.NewMainMenu(fileMenu, viewMenu, spritesMenu)
	e.window.SetMainMenu(mainMenu)

	// Simple toolbar with basic actions
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			e.fileOps.NewLevel()
		}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			e.fileOps.OpenLevel()
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			e.fileOps.SaveLevel()
		}),
	)

	return toolbar
}

// showSpriteSelector shows a dialog to select sprites from a category
func (e *Editor) showSpriteSelector(category SpriteCategory) {
	// Create and show the sprite browser
	browser := NewSpriteBrowser(e)
	browser.BrowseCategory(category)

	e.statusLabel.SetText("Browsing sprite category: " + string(category))
}

// createTools creates the left toolbar with editing tools
func (e *Editor) createTools() fyne.CanvasObject {
	// Create the tools buttons
	selectButton := widget.NewButton("Select", func() {
		e.setTool(ToolSelect)
	})

	platformButton := widget.NewButton("Platform", func() {
		e.setTool(ToolPlatform)
	})

	enemyButton := widget.NewButton("Enemy", func() {
		e.setTool(ToolEnemy)
	})

	collectibleButton := widget.NewButton("Item", func() {
		e.setTool(ToolCollectible)
	})

	startButton := widget.NewButton("Start", func() {
		e.setTool(ToolStart)
	})

	exitButton := widget.NewButton("Exit", func() {
		e.setTool(ToolExit)
	})

	eraseButton := widget.NewButton("Erase", func() {
		e.setTool(ToolEraser)
	})

	// Sprite selector button
	spriteButton := widget.NewButton("Choose Sprite", func() {
		e.showAllSpriteCategories()
	})

	// Arrange buttons in a vertical container
	return container.NewVBox(
		widget.NewLabel("Tools:"),
		selectButton,
		platformButton,
		enemyButton,
		collectibleButton,
		startButton,
		exitButton,
		eraseButton,
		widget.NewSeparator(),
		spriteButton,
	)
}

// showAllSpriteCategories shows a dialog with all sprite categories
func (e *Editor) showAllSpriteCategories() {
	categories := []SpriteCategory{
		SpriteCategoryBackground,
		SpriteCategoryFurniture,
		SpriteCategoryLivingRoom,
		SpriteCategoryKitchen,
		SpriteCategoryBathroom,
		SpriteCategoryBedroom,
		SpriteCategoryObjects,
		SpriteCategoryObjLivingRoom,
		SpriteCategoryObjKitchen,
		SpriteCategoryObjBathroom,
		SpriteCategoryObjBedroom,
		SpriteCategoryNPCs,
	}

	categoryNames := []string{
		"Backgrounds - Mountains, clouds, trees, houses",
		"Furniture - All",
		"Furniture - Living Room",
		"Furniture - Kitchen",
		"Furniture - Bathroom",
		"Furniture - Bedroom",
		"Objects - All",
		"Objects - Living Room",
		"Objects - Kitchen",
		"Objects - Bathroom",
		"Objects - Bedroom",
		"NPCs - Characters",
	}

	list := widget.NewList(
		func() int { return len(categories) },
		func() fyne.CanvasObject { return widget.NewLabel("Category") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(categoryNames[id])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		e.showSpriteSelector(categories[id])
	}

	w := fyne.CurrentApp().NewWindow("Sprite Categories")
	w.SetContent(container.NewBorder(
		widget.NewLabel("Select Sprite Category:"),
		nil, nil, nil,
		list,
	))
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}

// createProperties creates the right properties panel
func (e *Editor) createProperties() fyne.CanvasObject {
	// Create level info display
	nameLabel := widget.NewLabel("Level: " + e.currentLevel.Name)
	sizeLabel := widget.NewLabel("Size: " + fmt.Sprintf("%dx%d", e.currentLevel.Width, e.currentLevel.Height))

	// Background selection
	backgroundButton := widget.NewButton("Select Background", func() {
		e.fileOps.SelectBackground()
	})

	// Create platform properties panel
	platformProps := e.createPlatformProperties()

	// Sprite section
	spriteSection := e.createSpritePreview()

	// Create properties panel
	return widget.NewCard("Level Properties", "",
		container.NewVBox(
			nameLabel,
			sizeLabel,
			widget.NewSeparator(),
			backgroundButton,
			widget.NewSeparator(),
			platformProps,
			widget.NewSeparator(),
			spriteSection,
		),
	)
}

// createSpritePreview creates the sprite preview section
func (e *Editor) createSpritePreview() fyne.CanvasObject {
	// Create a container for the sprite info
	container := container.NewVBox(widget.NewLabel("Current Sprite:"))

	if e.selectedSprite == "" {
		// No sprite selected
		container.Add(widget.NewLabel("No sprite selected"))
		placeholder := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
		placeholder.SetMinSize(fyne.NewSize(150, 150))
		container.Add(placeholder)
	} else {
		// Display sprite info
		container.Add(widget.NewLabel("Selected: " + filepath.Base(e.selectedSprite)))

		// Check if the file exists
		if _, err := os.Stat(e.selectedSprite); err != nil {
			// Display error
			container.Add(widget.NewLabel(fmt.Sprintf("Error: %v", err)))
			placeholder := canvas.NewRectangle(color.RGBA{255, 100, 100, 255})
			placeholder.SetMinSize(fyne.NewSize(150, 150))
			container.Add(placeholder)
		} else {
			// Create a preview of the sprite
			uri := storage.NewFileURI(e.selectedSprite)
			fmt.Printf("DEBUG: Creating preview image from URI: %s\n", uri.String())

			img := canvas.NewImageFromURI(uri)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(150, 150))
			container.Add(img)

			// Debug info
			infoLabel := widget.NewLabel("Path: " + e.selectedSprite)
			infoLabel.Wrapping = fyne.TextWrapWord
			container.Add(infoLabel)
		}
	}

	// Add button to clear the selected sprite
	clearButton := widget.NewButton("Clear Selected Sprite", func() {
		e.selectedSprite = ""
		e.updatePropertiesPanel()
	})
	container.Add(clearButton)

	return container
}

// updatePropertiesPanel refreshes the properties panel to show current editor state
func (e *Editor) updatePropertiesPanel() {
	if e.content != nil {
		// Replace the right component with an updated properties panel
		e.content.Objects[3] = e.createProperties()
		e.content.Refresh()
	}
}

// createPlatformProperties creates the platform settings panel
func (e *Editor) createPlatformProperties() fyne.CanvasObject {
	// Create width slider
	widthSlider := widget.NewSlider(1, 10)
	widthSlider.Value = float64(e.platformWidth)
	widthLabel := widget.NewLabel(fmt.Sprintf("Width: %d", e.platformWidth))

	// Width slider callback
	widthSlider.OnChanged = func(value float64) {
		e.platformWidth = int(value)
		widthLabel.SetText(fmt.Sprintf("Width: %d", e.platformWidth))
	}

	// Create height slider
	heightSlider := widget.NewSlider(1, 5)
	heightSlider.Value = float64(e.platformHeight)
	heightLabel := widget.NewLabel(fmt.Sprintf("Height: %d", e.platformHeight))

	// Height slider callback
	heightSlider.OnChanged = func(value float64) {
		e.platformHeight = int(value)
		heightLabel.SetText(fmt.Sprintf("Height: %d", e.platformHeight))
	}

	// Create platform type selector
	typeSelect := widget.NewSelect([]string{"solid", "moving", "breaking", "bouncy"}, func(value string) {
		e.platformType = value
	})
	typeSelect.Selected = e.platformType

	return widget.NewCard("Platform Properties", "",
		container.NewVBox(
			widthLabel,
			widthSlider,
			heightLabel,
			heightSlider,
			widget.NewLabel("Type:"),
			typeSelect,
		),
	)
}

// setTool changes the current editing tool
func (e *Editor) setTool(tool ToolType) {
	e.currentTool = tool
	e.statusLabel.SetText("Tool: " + e.getToolName())
}

// getToolName returns the name of the current tool
func (e *Editor) getToolName() string {
	switch e.currentTool {
	case ToolSelect:
		return "Select"
	case ToolPlatform:
		return "Platform"
	case ToolEnemy:
		return "Enemy"
	case ToolCollectible:
		return "Collectible"
	case ToolStart:
		return "Start Position"
	case ToolExit:
		return "Exit Position"
	case ToolEraser:
		return "Eraser"
	default:
		return "Unknown"
	}
}
