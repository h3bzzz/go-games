package gui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"github.com/h3bzzz/go-chess/core/game"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme" // Import fyne theme for fallback icons
)

// PieceManager handles loading and managing chess piece images
type PieceManager struct {
	pieces     map[int]fyne.Resource
	pieceSizes map[int]fyne.Size
	theme      string
}

// NewPieceManager creates a new piece manager
func NewPieceManager(theme string) *PieceManager {
	return &PieceManager{
		pieces:     make(map[int]fyne.Resource),
		pieceSizes: make(map[int]fyne.Size),
		theme:      theme,
	}
}

// LoadPieces loads all piece images for the current theme
func (p *PieceManager) LoadPieces() error {
	// The theme string should already correspond to the folder name (chess, chess_green, chess_pink)
	folder := p.theme

	fmt.Printf("Loading pieces from folder: %s\n", folder)

	// Check if the folder exists
	assetDir := filepath.Join("assets", folder)
	if _, err := os.Stat(assetDir); os.IsNotExist(err) {
		fmt.Printf("ERROR: Asset directory not found: %s\n", assetDir)
		fmt.Printf("Current working directory: %s\n", getCwd())
		return fmt.Errorf("asset directory not found: %s", assetDir)
	} else {
		fmt.Printf("Asset directory found: %s\n", assetDir)
	}

	// Empty square has no image
	p.pieces[game.Empty] = nil

	// White pieces
	var err error
	p.pieces[game.WhitePawn], err = p.loadPieceResource(folder, "white_pawn.png")
	if err != nil {
		fmt.Printf("Error loading white pawn: %v\n", err)
		// Create a colored rectangle as fallback
		p.pieces[game.WhitePawn] = createColoredResource(color.RGBA{255, 255, 255, 255})
	}

	p.pieces[game.WhiteKnight], err = p.loadPieceResource(folder, "white_knight.png")
	if err != nil {
		fmt.Printf("Error loading white knight: %v\n", err)
		p.pieces[game.WhiteKnight] = createColoredResource(color.RGBA{255, 255, 255, 255})
	}

	p.pieces[game.WhiteBishop], err = p.loadPieceResource(folder, "white_bishop.png")
	if err != nil {
		fmt.Printf("Error loading white bishop: %v\n", err)
		p.pieces[game.WhiteBishop] = createColoredResource(color.RGBA{255, 255, 255, 255})
	}

	p.pieces[game.WhiteRook], err = p.loadPieceResource(folder, "white_rook.png")
	if err != nil {
		fmt.Printf("Error loading white rook: %v\n", err)
		p.pieces[game.WhiteRook] = createColoredResource(color.RGBA{255, 255, 255, 255})
	}

	p.pieces[game.WhiteQueen], err = p.loadPieceResource(folder, "white_queen.png")
	if err != nil {
		fmt.Printf("Error loading white queen: %v\n", err)
		p.pieces[game.WhiteQueen] = createColoredResource(color.RGBA{255, 255, 255, 255})
	}

	p.pieces[game.WhiteKing], err = p.loadPieceResource(folder, "white_king.png")
	if err != nil {
		fmt.Printf("Error loading white king: %v\n", err)
		p.pieces[game.WhiteKing] = createColoredResource(color.RGBA{255, 255, 255, 255})
	}

	// Black pieces
	p.pieces[game.BlackPawn], err = p.loadPieceResource(folder, "black_pawn.png")
	if err != nil {
		fmt.Printf("Error loading black pawn: %v\n", err)
		p.pieces[game.BlackPawn] = createColoredResource(color.RGBA{0, 0, 0, 255})
	}

	p.pieces[game.BlackKnight], err = p.loadPieceResource(folder, "black_knight.png")
	if err != nil {
		fmt.Printf("Error loading black knight: %v\n", err)
		p.pieces[game.BlackKnight] = createColoredResource(color.RGBA{0, 0, 0, 255})
	}

	p.pieces[game.BlackBishop], err = p.loadPieceResource(folder, "black_bishop.png")
	if err != nil {
		fmt.Printf("Error loading black bishop: %v\n", err)
		p.pieces[game.BlackBishop] = createColoredResource(color.RGBA{0, 0, 0, 255})
	}

	p.pieces[game.BlackRook], err = p.loadPieceResource(folder, "black_rook.png")
	if err != nil {
		fmt.Printf("Error loading black rook: %v\n", err)
		p.pieces[game.BlackRook] = createColoredResource(color.RGBA{0, 0, 0, 255})
	}

	p.pieces[game.BlackQueen], err = p.loadPieceResource(folder, "black_queen.png")
	if err != nil {
		fmt.Printf("Error loading black queen: %v\n", err)
		p.pieces[game.BlackQueen] = createColoredResource(color.RGBA{0, 0, 0, 255})
	}

	p.pieces[game.BlackKing], err = p.loadPieceResource(folder, "black_king.png")
	if err != nil {
		fmt.Printf("Error loading black king: %v\n", err)
		p.pieces[game.BlackKing] = createColoredResource(color.RGBA{0, 0, 0, 255})
	}

	return nil
}

// Helper function to get current working directory
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return cwd
}

// createColoredResource creates a fallback resource using built-in theme icons
func createColoredResource(c color.RGBA) fyne.Resource {
	// Use built-in theme resources as fallbacks for pieces
	// This provides at least some visible icon for each piece

	var res fyne.Resource

	if c.R == 255 && c.G == 255 && c.B == 255 {
		// White pieces - use light colored icons
		res = theme.HomeIcon() // Generic icon for pieces
	} else {
		// Black pieces - use dark colored icons
		res = theme.InfoIcon() // Different generic icon for black pieces
	}

	return res
}

// loadPieceResource loads a single piece resource
func (p *PieceManager) loadPieceResource(folder, filename string) (fyne.Resource, error) {
	path := filepath.Join("assets", folder, filename)

	// Debug output to identify paths
	fmt.Printf("Loading image from: %s\n", path)

	// Check if file exists before loading
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("ERROR: Image file not found: %s\n", path)
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Load resource
	res, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		fmt.Printf("ERROR loading resource: %v\n", err)
		return nil, err
	}

	return res, nil
}

// ChangeTheme changes the theme and reloads all piece images
func (p *PieceManager) ChangeTheme(theme string) error {
	if p == nil {
		return fmt.Errorf("PieceManager is nil")
	}

	fmt.Printf("PieceManager: Changing theme from '%s' to '%s'\n", p.theme, theme)

	// Check if theme folder exists before changing
	assetDir := filepath.Join("assets", theme)
	if _, err := os.Stat(assetDir); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Theme folder does not exist: %s", assetDir)
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}

	p.theme = theme
	fmt.Printf("PieceManager: Loading pieces for new theme '%s'\n", theme)
	err := p.LoadPieces()
	if err != nil {
		fmt.Printf("PieceManager: Error loading pieces for theme '%s': %v\n", theme, err)
		return err
	}

	fmt.Printf("PieceManager: Successfully changed to theme '%s'\n", theme)
	return nil
}

// GetResource returns the resource for a specific piece
func (p *PieceManager) GetResource(piece int) fyne.Resource {
	if piece == game.Empty {
		return nil
	}

	resource := p.pieces[piece]
	if resource == nil {
		// If the resource is still nil, return a fallback icon based on piece type
		return getFallbackIcon(piece)
	}

	return resource
}

// getFallbackIcon returns an appropriate fallback icon for a chess piece
func getFallbackIcon(piece int) fyne.Resource {
	isWhite := IsPieceWhite(piece)

	// Choose appropriate icon based on piece type
	switch piece {
	case game.WhitePawn, game.BlackPawn:
		if isWhite {
			return theme.RadioButtonIcon()
		}
		return theme.RadioButtonCheckedIcon()

	case game.WhiteKnight, game.BlackKnight:
		if isWhite {
			return theme.NavigateNextIcon()
		}
		return theme.NavigateBackIcon()

	case game.WhiteBishop, game.BlackBishop:
		if isWhite {
			return theme.ContentCutIcon()
		}
		return theme.ContentPasteIcon()

	case game.WhiteRook, game.BlackRook:
		if isWhite {
			return theme.MenuIcon()
		}
		return theme.MenuDropDownIcon()

	case game.WhiteQueen, game.BlackQueen:
		if isWhite {
			return theme.SettingsIcon()
		}
		return theme.DocumentIcon()

	case game.WhiteKing, game.BlackKing:
		if isWhite {
			return theme.HomeIcon()
		}
		return theme.InfoIcon()
	}

	// Default fallback
	if isWhite {
		return theme.VisibilityIcon()
	}
	return theme.VisibilityOffIcon()
}

// IsPieceWhite checks if a piece is white
func IsPieceWhite(piece int) bool {
	return piece >= game.WhitePawn && piece <= game.WhiteKing
}
