package main

import (
	"image/color"
	"image/gif"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 840
	screenHeight = 840
	cellSize     = 20
	cols         = screenWidth / cellSize
	rows         = screenHeight / cellSize
	animSpeed    = 10
)

type Game struct {
	maze        [][]bool
	player      Position
	exit        Position
	randGen     *rand.Rand
	frames      []*ebiten.Image // Store extracted frames from the GIF
	animFrame   int
	animCounter int
	facingLeft  bool // Track if player is facing left
}

type Position struct {
	X, Y int
}

func NewGame() *Game {
	g := &Game{
		randGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.loadGifFrames("assets/run.gif") // Load GIF frames for animation
	g.generateMaze()
	g.player = Position{X: 0, Y: 0}
	g.exit = Position{X: cols - 2, Y: rows - 2}
	return g
}

// Load and extract frames from a GIF file
func (g *Game) loadGifFrames(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	gifImage, err := gif.DecodeAll(f)
	if err != nil {
		log.Fatal(err)
	}

	// Convert each frame in the GIF to an ebiten.Image
	for _, frame := range gifImage.Image {
		img := ebiten.NewImageFromImage(frame)
		g.frames = append(g.frames, img)
	}
}

func (g *Game) generateMaze() {
	g.maze = make([][]bool, cols)
	for x := range g.maze {
		g.maze[x] = make([]bool, rows)
	}

	stack := []Position{{X: 0, Y: 0}}
	g.maze[0][0] = true

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		neighbors := g.getUnvisitedNeighbors(current)

		if len(neighbors) > 0 {
			next := neighbors[g.randGen.Intn(len(neighbors))]
			g.maze[next.X][next.Y] = true
			between := Position{(current.X + next.X) / 2, (current.Y + next.Y) / 2}
			g.maze[between.X][between.Y] = true
			stack = append(stack, next)
		} else {
			stack = stack[:len(stack)-1]
		}
	}
}

func (g *Game) getUnvisitedNeighbors(pos Position) []Position {
	neighbors := []Position{}
	directions := []Position{{X: 2, Y: 0}, {X: -2, Y: 0}, {X: 0, Y: 2}, {X: 0, Y: -2}}
	for _, dir := range directions {
		nx, ny := pos.X+dir.X, pos.Y+dir.Y
		if nx >= 0 && ny >= 0 && nx < cols && ny < rows && !g.maze[nx][ny] {
			neighbors = append(neighbors, Position{X: nx, Y: ny})
		}
	}
	return neighbors
}

func (g *Game) Update() error {
	moved := false
	if ebiten.IsKeyPressed(ebiten.KeyUp) && g.canMoveTo(g.player.X, g.player.Y-1) {
		g.player.Y--
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) && g.canMoveTo(g.player.X, g.player.Y+1) {
		g.player.Y++
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.canMoveTo(g.player.X-1, g.player.Y) {
		g.player.X--
		g.facingLeft = true // Set facing left
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) && g.canMoveTo(g.player.X+1, g.player.Y) {
		g.player.X++
		g.facingLeft = false // Set facing right
		moved = true
	}

	// Update animation frame if player moved
	if moved {
		g.animCounter++
		if g.animCounter >= animSpeed {
			g.animFrame = (g.animFrame + 1) % len(g.frames)
			g.animCounter = 0
		}
	}

	// Check if player reached the exit
	if g.player == g.exit {
		g.generateMaze()
		g.player = Position{X: 0, Y: 0}
	}
	return nil
}

func (g *Game) canMoveTo(x, y int) bool {
	return x >= 0 && y >= 0 && x < cols && y < rows && g.maze[x][y]
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xf0, 0xf0, 0xf0, 0xff})

	wallColor := color.RGBA{0x80, 0x80, 0x80, 0xff}
	cellImage := ebiten.NewImage(cellSize, cellSize)
	cellImage.Fill(wallColor)

	for x := 0; x < cols; x++ {
		for y := 0; y < rows; y++ {
			if g.maze[x][y] {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x*cellSize), float64(y*cellSize))
				screen.DrawImage(cellImage, op)
			}
		}
	}

	exitImage := ebiten.NewImage(cellSize, cellSize)
	exitImage.Fill(color.RGBA{0xff, 0x00, 0x00, 0xff})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.exit.X*cellSize), float64(g.exit.Y*cellSize))
	screen.DrawImage(exitImage, op)

	g.drawPlayer(screen)

	ebitenutil.DebugPrint(screen, "Maze Game (Press arrow keys to move)")
}

// Draw the player with flipping based on direction
func (g *Game) drawPlayer(screen *ebiten.Image) {
	playerFrame := g.frames[g.animFrame]

	// Set up options for drawing with scaling
	op := &ebiten.DrawImageOptions{}
	scale := float64(cellSize) / float64(playerFrame.Bounds().Dx())
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(g.player.X*cellSize), float64(g.player.Y*cellSize))

	// Flip horizontally if the player is facing left
	if g.facingLeft {
		op.GeoM.Scale(-1, 1)                    // Horizontal flip
		op.GeoM.Translate(float64(cellSize), 0) // Adjust for flip position
	}

	screen.DrawImage(playerFrame, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Maze Game with Animated Gopher GIF")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
