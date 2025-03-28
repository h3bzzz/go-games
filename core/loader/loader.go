package loader

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
)

// AssetLoader handles loading assets for the game
type AssetLoader struct {
	BasePath string
}

// NewAssetLoader creates a new asset loader
func NewAssetLoader(basePath string) *AssetLoader {
	return &AssetLoader{
		BasePath: basePath,
	}
}

// LoadImage loads an image from the given path
func (l *AssetLoader) LoadImage(path string) (image.Image, error) {
	fullPath := filepath.Join(l.BasePath, path)
	
	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()
	
	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	
	return img, nil
}

// LoadImageFromTheme loads an image from the given theme and image name
func (l *AssetLoader) LoadImageFromTheme(theme, name string) (image.Image, error) {
	return l.LoadImage(filepath.Join("chess_"+theme, name+".png"))
}

// FileExists checks if a file exists
func (l *AssetLoader) FileExists(path string) bool {
	fullPath := filepath.Join(l.BasePath, path)
	_, err := os.Stat(fullPath)
	return err == nil
}
