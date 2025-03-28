package main

import (
	"image/color"
	"image/gif"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 840
	screenHeight = 840
	cellSize     = 20
)

type Game struct {
	manager *GameManager
}

func NewGame() *Game {
	g := &Game{
		manager: NewGameManager(),
	}
	g.loadGifFrames("assets/run.gif") // Load player animations
	return g
}

func (g *Game) loadGifFrames(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: Could not load animation %s: %v", path, err)
		g.createPlaceholderAnimations()
		return
	}
	defer f.Close()

	gifImage, err := gif.DecodeAll(f)
	if err != nil {
		log.Printf("Warning: Could not decode gif %s: %v", path, err)
		g.createPlaceholderAnimations()
		return
	}

	frames := make([]*ebiten.Image, 0, len(gifImage.Image))
	for _, frame := range gifImage.Image {
		img := ebiten.NewImageFromImage(frame)
		frames = append(frames, img)
	}

	g.manager.PlayerFrames = frames
}

func (g *Game) createPlaceholderAnimations() {
	// Create 4 simple colored frames for player animation
	frames := make([]*ebiten.Image, 4)
	colors := []color.RGBA{
		{50, 100, 220, 255},
		{70, 120, 230, 255},
		{50, 100, 220, 255},
		{30, 80, 210, 255},
	}

	for i, c := range colors {
		frames[i] = ebiten.NewImage(cellSize, cellSize*2)
		frames[i].Fill(c)
	}

	g.manager.PlayerFrames = frames
}

func (g *Game) Update() error {
	return g.manager.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.manager.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.manager.Layout(outsideWidth, outsideHeight)
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gopher Platform Runner")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
