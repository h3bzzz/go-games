package main

import (
	"image/color"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 640
	screenHeight = 480
	gridSize     = 20
	speed        = 5 // Control how many frames between each snake move
)

type Game struct {
	snake     []Position
	direction Position
	food      Position
	score     int
	gameOver  bool
	tick      int        // Frame counter to control speed
	randGen   *rand.Rand // Custom random generator instance
}

type Position struct {
	X, Y int
}

func (g *Game) spawnFood() {
	g.food = Position{
		X: g.randGen.Intn(screenWidth / gridSize),
		Y: g.randGen.Intn(screenHeight / gridSize),
	}
}

func (g *Game) collidesWithSelf(head Position) bool {
	for _, part := range g.snake[1:] {
		if head == part {
			return true
		}
	}
	return false
}

func (g *Game) Update() error {
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.Restart()
		}
		return nil
	}

	g.tick++
	if g.tick%speed != 0 {
		return nil
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) && g.direction.Y == 0 {
		g.direction = Position{X: 0, Y: -1}
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) && g.direction.Y == 0 {
		g.direction = Position{X: 0, Y: 1}
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.direction.X == 0 {
		g.direction = Position{X: -1, Y: 0}
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) && g.direction.X == 0 {
		g.direction = Position{X: 1, Y: 0}
	}

	head := g.snake[0]
	newHead := Position{X: head.X + g.direction.X, Y: head.Y + g.direction.Y}
	g.snake = append([]Position{newHead}, g.snake[:len(g.snake)-1]...)

	if newHead == g.food {
		g.snake = append(g.snake, Position{})
		g.score++
		g.spawnFood()
	}

	if newHead.X < 0 || newHead.X >= screenWidth/gridSize || newHead.Y < 0 || newHead.Y >= screenHeight/gridSize || g.collidesWithSelf(newHead) {
		g.gameOver = true
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

	for _, pos := range g.snake {
		x := pos.X * gridSize
		y := pos.Y * gridSize

		segment := ebiten.NewImage(gridSize, gridSize)
		segment.Fill(color.RGBA{0x00, 0x80, 0x00, 0xff})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(segment, op)
	}

	fx := g.food.X * gridSize
	fy := g.food.Y * gridSize

	food := ebiten.NewImage(gridSize, gridSize)
	food.Fill(color.RGBA{0x80, 0x00, 0x00, 0xff})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(fx), float64(fy))
	screen.DrawImage(food, op)

	ebitenutil.DebugPrint(screen, "Score: "+strconv.Itoa(g.score))

	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "Game Over! Press R to Restart", screenWidth/2-60, screenHeight/2)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Restart() {
	g.snake = []Position{{X: 5, Y: 5}, {X: 4, Y: 5}, {X: 3, Y: 5}, {X: 2, Y: 5}, {X: 1, Y: 5}}
	g.direction = Position{X: 1, Y: 0}
	g.score = 0
	g.gameOver = false
	g.tick = 0
	g.spawnFood()
}

func Run() {
	game := &Game{
		randGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	game.Restart()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake Game in Ebiten")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func main() {
	Run()
}
