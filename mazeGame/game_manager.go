package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

type GameState int

const (
	GameStateTitle GameState = iota
	GameStatePlaying
	GameStatePaused
	GameStateLevelComplete
	GameStateGameOver
)

type Projectile struct {
	X, Y       float64        // Position
	VelX, VelY float64        // Velocity
	Type       ProjectileType // Type of projectile
	Size       float64        // Size of projectile
	Active     bool           // Whether the projectile is active
	Damage     int            // Damage to player
}

type GameManager struct {
	State        GameState
	Rules        *GameRules
	Generator    *PlatformGenerator
	CurrentLevel *Level
	Player       *Player

	PlayerFrames      []*ebiten.Image
	BackgroundImages  []*ebiten.Image
	CurrentBackground *ebiten.Image
	FontFace          font.Face

	LevelNumber int
	TotalScore  int
	TimeLeft    float64
	StartTime   time.Time

	DebugMode bool
	CameraX   float64
	CameraY   float64
	Parallax  float64

	Projectiles []Projectile // Active projectiles in the game
}

func NewGameManager() *GameManager {
	manager := &GameManager{
		State:       GameStateTitle,
		Rules:       NewGameRules(),
		FontFace:    basicfont.Face7x13,
		LevelNumber: 1,
		DebugMode:   false,
		Projectiles: make([]Projectile, 0),
	}

	seed := time.Now().UnixNano()
	manager.Generator = NewPlatformGenerator(manager.Rules, seed)

	manager.loadPlayerGraphics()

	manager.loadBackgroundImages()

	return manager
}

func (g *GameManager) loadPlayerGraphics() {
	g.PlayerFrames = []*ebiten.Image{}

	// Try to load animated character
	animatedSprite, _, err := ebitenutil.NewImageFromFile("assets/run.gif")

	if err != nil {
		log.Printf("Warning: Could not load animation %s: %v", "assets/run.gif", err)

		// Fallback to a simple rectangle if animation can't be loaded
		fallbackImg := ebiten.NewImage(cellSize, cellSize*2)
		fallbackImg.Fill(color.RGBA{50, 50, 220, 255}) // Blue
		g.PlayerFrames = append(g.PlayerFrames, fallbackImg)
	} else {
		// Use the loaded image
		g.PlayerFrames = append(g.PlayerFrames, animatedSprite)
	}
}

func (g *GameManager) loadBackgroundImages() {
	backgrounds := []string{
		"assets/dark_night_world.jpg",
		"assets/sunset_flowers_world.jpg",
		"assets/city_skyline.jpg",
	}

	g.BackgroundImages = make([]*ebiten.Image, 0, len(backgrounds))

	// Try to load from files first
	for _, path := range backgrounds {
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			log.Printf("Failed to load background image %s: %v", path, err)
			continue
		}
		g.BackgroundImages = append(g.BackgroundImages, img)
	}

	// If no images were loaded, create placeholder backgrounds
	if len(g.BackgroundImages) == 0 {
		log.Printf("Creating placeholder backgrounds, no image files found")
		colors := []color.RGBA{
			{40, 40, 80, 255},    // Dark blue night
			{200, 150, 100, 255}, // Sunset orange
			{100, 150, 200, 255}, // Sky blue city
		}

		for _, bgColor := range colors {
			img := ebiten.NewImage(g.Rules.ScreenWidth, g.Rules.ScreenHeight)
			img.Fill(bgColor)
			g.BackgroundImages = append(g.BackgroundImages, img)
		}
	}

	if len(g.BackgroundImages) > 0 {
		g.CurrentBackground = g.BackgroundImages[0]
	}

	g.Parallax = 0.5
}

func (g *GameManager) StartNewGame() {
	g.LevelNumber = 1
	g.TotalScore = 0
	g.StartLevel(g.LevelNumber)
	g.State = GameStatePlaying
}

func (g *GameManager) StartLevel(levelNum int) {
	g.CurrentLevel = g.Generator.GenerateLevel(levelNum)

	// Add a platform at the starting position to ensure the player doesn't fall
	startPlatform := Platform{
		X:      g.CurrentLevel.StartX - 1,
		Y:      g.CurrentLevel.StartY + 1,
		Width:  3,
		Height: 1,
		Type:   PlatformNormal,
	}
	g.CurrentLevel.Platforms = append(g.CurrentLevel.Platforms, startPlatform)

	g.Player = NewPlayer(
		g.CurrentLevel.StartX,
		g.CurrentLevel.StartY,
		g.Rules,
		g.CurrentLevel,
		g.PlayerFrames,
	)

	g.TimeLeft = float64(g.Rules.TimeLimit[levelNum-1])
	g.StartTime = time.Now()

	if len(g.BackgroundImages) > 0 {
		randomIndex := rand.Intn(len(g.BackgroundImages))
		g.CurrentBackground = g.BackgroundImages[randomIndex]
	}

	// Initialize camera position to center on player
	g.CameraX = g.Player.X - float64(g.Rules.ScreenWidth)/2 + float64(cellSize)/2
	g.CameraY = g.Player.Y - float64(g.Rules.ScreenHeight)/2 + float64(cellSize)

	// Keep camera within level bounds
	if g.CameraX < 0 {
		g.CameraX = 0
	}
	if g.CameraY < 0 {
		g.CameraY = 0
	}

	levelWidthPixels := float64(g.CurrentLevel.Width * cellSize)
	levelHeightPixels := float64(g.CurrentLevel.Height * cellSize)

	maxX := levelWidthPixels - float64(g.Rules.ScreenWidth)
	maxY := levelHeightPixels - float64(g.Rules.ScreenHeight)

	if maxX > 0 && g.CameraX > maxX {
		g.CameraX = maxX
	}
	if maxY > 0 && g.CameraY > maxY {
		g.CameraY = maxY
	}

	g.State = GameStatePlaying
}

func (g *GameManager) Update() error {
	switch g.State {
	case GameStateTitle:
		return g.updateTitleScreen()

	case GameStatePlaying:
		return g.updateGameplay()

	case GameStatePaused:
		return g.updatePauseScreen()

	case GameStateLevelComplete:
		return g.updateLevelComplete()

	case GameStateGameOver:
		return g.updateGameOver()
	}

	return nil
}

func (g *GameManager) updateTitleScreen() error {
	// Start game when space is pressed
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.StartNewGame()
	}

	return nil
}

func (g *GameManager) updateGameplay() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.State = GameStatePaused
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.DebugMode = !g.DebugMode
	}

	g.updatePlatforms()

	g.updateObstacles()

	g.updateEnemies()
	g.updateProjectiles()

	input := GetPlayerInput()
	g.Player.Update(input)

	g.updateCamera()

	if int(g.Player.X) == g.CurrentLevel.ExitX*cellSize &&
		int(g.Player.Y) == g.CurrentLevel.ExitY*cellSize {
		g.completedLevel()
	}

	if g.Player.Lives <= 0 {
		g.State = GameStateGameOver
	}

	if g.TimeLeft > 0 {
		g.TimeLeft -= 1.0 / 60.0
		if g.TimeLeft <= 0 {
			g.Player.Lives--
			g.Player.respawn()
			g.TimeLeft = float64(g.Rules.TimeLimit[g.LevelNumber-1])
		}
	}

	return nil
}

func (g *GameManager) updatePauseScreen() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.State = GameStatePlaying
	}

	return nil
}

func (g *GameManager) updateLevelComplete() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.LevelNumber < g.Rules.MaxLevels {
			g.LevelNumber++
			g.StartLevel(g.LevelNumber)
		} else {
			g.State = GameStateTitle
		}
	}

	return nil
}

func (g *GameManager) updateGameOver() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.State = GameStateTitle
	}

	return nil
}

func (g *GameManager) updatePlatforms() {
	for i := range g.CurrentLevel.Platforms {
		platform := &g.CurrentLevel.Platforms[i]

		if platform.Type == PlatformMoving {
			platform.CurrentMove += platform.MoveSpeed
			if platform.CurrentMove >= float64(platform.MoveRange) || platform.CurrentMove <= 0 {
				platform.MoveSpeed = -platform.MoveSpeed
			}
		}

		if platform.Type == PlatformBreaking && platform.Breaking {
			platform.BreakTimer++
			if platform.BreakTimer > 30 {
				platform.Width = 0
			}
		}
	}
}

func (g *GameManager) updateObstacles() {
	for i := range g.CurrentLevel.Obstacles {
		obstacle := &g.CurrentLevel.Obstacles[i]

		if obstacle.Moving {
			obstacle.X += int(obstacle.Speed)

			onPlatform := false
			for _, platform := range g.CurrentLevel.Platforms {
				if obstacle.X >= platform.X && obstacle.X+obstacle.Width <= platform.X+platform.Width &&
					obstacle.Y == platform.Y-1 {
					onPlatform = true

					if obstacle.X <= platform.X || obstacle.X+obstacle.Width >= platform.X+platform.Width {
						obstacle.Speed = -obstacle.Speed
					}
				}
			}

			if !onPlatform {
				obstacle.Speed = -obstacle.Speed
			}
		}
	}
}

func (g *GameManager) updateEnemies() {
	for i := range g.CurrentLevel.Enemies {
		enemy := &g.CurrentLevel.Enemies[i]

		// Check if enemy should be activated (player is within detection radius)
		playerDistX := g.Player.X - float64(enemy.X*cellSize)
		playerDistY := g.Player.Y - float64(enemy.Y*cellSize)
		distanceSq := playerDistX*playerDistX + playerDistY*playerDistY
		detectionRadiusSq := float64(enemy.DetectionRadius * enemy.DetectionRadius * cellSize * cellSize)

		// Activate enemy if player is nearby
		if distanceSq <= detectionRadiusSq {
			enemy.Active = true
		} else if distanceSq > detectionRadiusSq*4 { // Deactivate if player is far away
			enemy.Active = false
		}

		// Only update active enemies
		if enemy.Active {
			// Handle movement based on behavior
			switch enemy.Behavior {
			case EnemyPacing:
				g.updatePacingEnemy(enemy)
			case EnemyPatrolling:
				g.updatePatrollingEnemy(enemy)
			case EnemyStationary:
				// Stationary enemies don't move, they just fire
			}

			// Handle projectile firing
			g.updateEnemyFiring(enemy)
		}
	}
}

func (g *GameManager) updatePacingEnemy(enemy *Enemy) {
	if len(enemy.PatrolPoints) < 2 {
		return // Need at least 2 points to pace between
	}

	// Current target point
	targetIndex := 0
	if enemy.PatrolDirection > 0 {
		targetIndex = 1
	}

	// Get current and target positions
	currentX := float64(enemy.X)
	currentY := float64(enemy.Y)
	targetX := float64(enemy.PatrolPoints[targetIndex][0])
	targetY := float64(enemy.PatrolPoints[targetIndex][1])

	// Calculate direction to target
	dirX := targetX - currentX
	dirY := targetY - currentY
	length := math.Sqrt(dirX*dirX + dirY*dirY)

	// If we're close to the target, switch direction
	if length < enemy.Speed {
		enemy.PatrolDirection *= -1
	} else {
		// Normalize direction and move
		dirX /= length
		dirY /= length

		enemy.X += int(dirX * enemy.Speed)
		enemy.Y += int(dirY * enemy.Speed)
	}
}

func (g *GameManager) updatePatrollingEnemy(enemy *Enemy) {
	if len(enemy.PatrolPoints) < 2 {
		return // Need at least 2 points for a patrol path
	}

	// Get current target point
	targetIndex := enemy.CurrentPatrolIdx

	// Get current and target positions
	currentX := float64(enemy.X)
	currentY := float64(enemy.Y)
	targetX := float64(enemy.PatrolPoints[targetIndex][0])
	targetY := float64(enemy.PatrolPoints[targetIndex][1])

	// Calculate direction to target
	dirX := targetX - currentX
	dirY := targetY - currentY
	length := math.Sqrt(dirX*dirX + dirY*dirY)

	// If we're close to the target, move to next point
	if length < enemy.Speed {
		// Update patrol index
		if enemy.PatrolDirection > 0 {
			enemy.CurrentPatrolIdx++
			if enemy.CurrentPatrolIdx >= len(enemy.PatrolPoints) {
				enemy.CurrentPatrolIdx = 0
			}
		} else {
			enemy.CurrentPatrolIdx--
			if enemy.CurrentPatrolIdx < 0 {
				enemy.CurrentPatrolIdx = len(enemy.PatrolPoints) - 1
			}
		}
	} else {
		// Normalize direction and move
		dirX /= length
		dirY /= length

		enemy.X += int(dirX * enemy.Speed)
		enemy.Y += int(dirY * enemy.Speed)
	}
}

func (g *GameManager) updateEnemyFiring(enemy *Enemy) {
	// Check if it's time to fire based on fire rate
	currentTime := float64(time.Now().UnixNano()) / 1e9
	if currentTime-enemy.LastFireTime < 1.0/enemy.FireRate {
		return // Not time to fire yet
	}

	// Update last fire time
	enemy.LastFireTime = currentTime

	// Calculate direction to player
	enemyX := float64(enemy.X * cellSize)
	enemyY := float64(enemy.Y * cellSize)

	dirX := g.Player.X - enemyX
	dirY := g.Player.Y - enemyY
	length := math.Sqrt(dirX*dirX + dirY*dirY)

	// Skip if too far away
	if length > float64(enemy.DetectionRadius*cellSize) {
		return
	}

	// Normalize direction
	dirX /= length
	dirY /= length

	// Create projectile
	projectile := Projectile{
		X:      enemyX,
		Y:      enemyY,
		VelX:   dirX * 3.0, // Projectile speed
		VelY:   dirY * 3.0,
		Type:   enemy.ProjectileType,
		Size:   float64(cellSize) / 2,
		Active: true,
		Damage: 1,
	}

	// Add projectile to game
	g.Projectiles = append(g.Projectiles, projectile)
}

func (g *GameManager) updateProjectiles() {
	activeProjectiles := make([]Projectile, 0, len(g.Projectiles))

	for _, proj := range g.Projectiles {
		if !proj.Active {
			continue
		}

		// Update position
		proj.X += proj.VelX
		proj.Y += proj.VelY

		// Check if projectile is out of bounds
		if proj.X < 0 || proj.X > float64(g.CurrentLevel.Width*cellSize) ||
			proj.Y < 0 || proj.Y > float64(g.CurrentLevel.Height*cellSize) {
			continue // Don't keep this projectile
		}

		// Check collision with platforms
		hitPlatform := false
		for _, platform := range g.CurrentLevel.Platforms {
			platX := float64(platform.X * cellSize)
			platY := float64(platform.Y * cellSize)
			platWidth := float64(platform.Width * cellSize)
			platHeight := float64(platform.Height * cellSize)

			if proj.X+proj.Size > platX && proj.X-proj.Size < platX+platWidth &&
				proj.Y+proj.Size > platY && proj.Y-proj.Size < platY+platHeight {
				hitPlatform = true
				break
			}
		}

		if hitPlatform {
			continue // Don't keep this projectile
		}

		// Check collision with player
		playerHitbox := 0.7 * float64(cellSize) // Slightly smaller than player size
		if !g.Player.isDashing() &&             // Player is immune while dashing
			proj.X+proj.Size > g.Player.X-playerHitbox && proj.X-proj.Size < g.Player.X+playerHitbox &&
			proj.Y+proj.Size > g.Player.Y-playerHitbox && proj.Y-proj.Size < g.Player.Y+playerHitbox {
			// Player hit by projectile
			g.Player.Lives -= proj.Damage
			continue // Don't keep this projectile
		}

		// Keep active projectile
		activeProjectiles = append(activeProjectiles, proj)
	}

	// Update projectiles list
	g.Projectiles = activeProjectiles
}

func (g *GameManager) completedLevel() {
	g.TotalScore += g.Player.Score

	g.State = GameStateLevelComplete
}

func (g *GameManager) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{90, 150, 220, 255}) // Sky blue

	switch g.State {
	case GameStateTitle:
		g.drawTitleScreen(screen)

	case GameStatePlaying:
		g.drawBackground(screen)
		g.drawGame(screen)

	case GameStatePaused:
		g.drawBackground(screen)
		g.drawGame(screen)
		g.drawPauseScreen(screen)

	case GameStateLevelComplete:
		g.drawBackground(screen)
		g.drawLevelComplete(screen)

	case GameStateGameOver:
		g.drawBackground(screen)
		g.drawGameOver(screen)
	}
}

// drawBackground renders the current background image with parallax scrolling
func (g *GameManager) drawBackground(screen *ebiten.Image) {
	if g.CurrentBackground == nil {
		return
	}

	// Draw the background image with parallax scrolling effect
	op := &ebiten.DrawImageOptions{}

	// Calculate scaling to fill the screen
	bgWidth := g.CurrentBackground.Bounds().Dx()
	bgHeight := g.CurrentBackground.Bounds().Dy()
	scaleX := float64(g.Rules.ScreenWidth) / float64(bgWidth)
	scaleY := float64(g.Rules.ScreenHeight) / float64(bgHeight)

	// Choose scale that ensures the background covers the entire screen
	// while maintaining aspect ratio
	scale := math.Max(scaleX, scaleY) * 1.2 // Scale up slightly to ensure coverage
	op.GeoM.Scale(scale, scale)

	// Apply parallax effect - background moves more slowly than foreground
	parallaxX := g.CameraX * g.Parallax
	parallaxY := g.CameraY * g.Parallax

	// Center the background and offset by parallax amount
	centerOffsetX := (float64(g.Rules.ScreenWidth) - float64(bgWidth)*scale) / 2
	centerOffsetY := (float64(g.Rules.ScreenHeight) - float64(bgHeight)*scale) / 2

	op.GeoM.Translate(centerOffsetX-parallaxX, centerOffsetY-parallaxY)

	screen.DrawImage(g.CurrentBackground, op)
}

func (g *GameManager) drawTitleScreen(screen *ebiten.Image) {
	if len(g.BackgroundImages) > 0 {
		randomIndex := 0
		if len(g.BackgroundImages) > 1 {
			randomIndex = rand.Intn(len(g.BackgroundImages))
		}

		op := &ebiten.DrawImageOptions{}

		bgWidth := g.BackgroundImages[randomIndex].Bounds().Dx()
		bgHeight := g.BackgroundImages[randomIndex].Bounds().Dy()
		scaleX := float64(g.Rules.ScreenWidth) / float64(bgWidth)
		scaleY := float64(g.Rules.ScreenHeight) / float64(bgHeight)

		op.GeoM.Scale(scaleX, scaleY)

		op.ColorM.Scale(0.7, 0.7, 0.7, 1.0)

		screen.DrawImage(g.BackgroundImages[randomIndex], op)
	} else {
		screenWidth := g.Rules.ScreenWidth
		screenHeight := g.Rules.ScreenHeight

		for y := 0; y < screenHeight; y++ {
			factor := float64(y) / float64(screenHeight)
			r := uint8(50 + 150*factor)
			gg := uint8(50 + 100*factor)
			b := uint8(150 + 50*factor)
			c := color.RGBA{r, gg, b, 255}
			rect := ebiten.NewImage(screenWidth, 1)
			rect.Fill(c)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(0, float64(y))
			screen.DrawImage(rect, op)
		}
	}

	// Draw a semi-transparent panel for the instructions
	panel := ebiten.NewImage(g.Rules.ScreenWidth-100, 500)
	panel.Fill(color.RGBA{0, 0, 0, 200})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(50, 100)
	screen.DrawImage(panel, op)

	// Draw title
	title := "GOPHER PLATFORM RUNNER"
	titleX := (g.Rules.ScreenWidth - len(title)*10) / 2
	titleY := 150
	text.Draw(screen, title, g.FontFace, titleX, titleY, color.White)

	// Draw instructions
	textColor := color.RGBA{255, 255, 255, 255}

	instructions := []string{
		"HOW TO PLAY:",
		"",
		"• Use ARROW KEYS or WASD to move",
		"• Press SPACE to jump",
		"• Press SPACE in mid-air for double jump (Level 3+)",
		"• Slide against walls to wall-jump (Level 5+)",
		"• Press SHIFT or E to dash (Level 7+)",
		"• Collect items for points",
		"• Avoid obstacles",
		"• Reach the green exit to complete each level",
		"• Press F1 for debug mode",
		"• Press ESC to pause",
		"",
		"PLATFORM TYPES:",
		"• GRAY: Normal platforms",
		"• GREEN: Moving platforms",
		"• RED: Breaking platforms",
		"• YELLOW: Bouncy platforms",
		"",
		"PRESS SPACE TO START!",
	}

	y := 190
	for _, line := range instructions {
		// Center align all text
		x := (g.Rules.ScreenWidth - len(line)*7) / 2
		if line == "" {
			y += 15 // Add extra space between sections
		} else {
			text.Draw(screen, line, g.FontFace, x, y, textColor)
			y += 22 // Standard line spacing
		}
	}
}

// drawGame renders the main gameplay
func (g *GameManager) drawGame(screen *ebiten.Image) {
	// Draw platforms
	for _, platform := range g.CurrentLevel.Platforms {
		// Skip if broken
		if platform.Width == 0 {
			continue
		}

		// Choose color based on platform type
		var platformColor color.RGBA
		switch platform.Type {
		case PlatformNormal:
			platformColor = color.RGBA{100, 100, 100, 255} // Gray
		case PlatformMoving:
			platformColor = color.RGBA{100, 200, 100, 255} // Green
		case PlatformBreaking:
			platformColor = color.RGBA{200, 100, 100, 255} // Red
		case PlatformBouncy:
			platformColor = color.RGBA{200, 200, 100, 255} // Yellow
		}

		// Draw platform
		x := float64(platform.X*cellSize) - g.CameraX
		y := float64(platform.Y*cellSize) - g.CameraY
		ebitenutil.DrawRect(
			screen,
			x,
			y,
			float64(platform.Width*cellSize),
			float64(platform.Height*cellSize),
			platformColor,
		)
	}

	// Draw obstacles
	for _, obstacle := range g.CurrentLevel.Obstacles {
		obstacleColor := color.RGBA{220, 50, 50, 255} // Red
		x := float64(obstacle.X*cellSize) - g.CameraX
		y := float64(obstacle.Y*cellSize) - g.CameraY
		ebitenutil.DrawRect(
			screen,
			x,
			y,
			float64(obstacle.Width*cellSize),
			float64(obstacle.Height*cellSize),
			obstacleColor,
		)
	}

	// Draw enemies
	g.drawEnemies(screen)

	// Draw projectiles
	g.drawProjectiles(screen)

	// Draw collectibles
	for _, collectible := range g.CurrentLevel.Collectibles {
		if !collectible.Active {
			continue
		}

		collectibleColor := color.RGBA{220, 220, 50, 255} // Yellow
		x := float64(collectible.X*cellSize) - g.CameraX
		y := float64(collectible.Y*cellSize) - g.CameraY
		ebitenutil.DrawRect(
			screen,
			x,
			y,
			float64(cellSize),
			float64(cellSize),
			collectibleColor,
		)
	}

	// Draw exit
	exitColor := color.RGBA{50, 220, 50, 255} // Green
	x := float64(g.CurrentLevel.ExitX*cellSize) - g.CameraX
	y := float64(g.CurrentLevel.ExitY*cellSize) - g.CameraY
	ebitenutil.DrawRect(
		screen,
		x,
		y,
		float64(cellSize),
		float64(cellSize*2),
		exitColor,
	)

	// Draw player
	g.drawPlayer(screen)

	// Draw HUD
	g.drawHUD(screen)

	// Draw debug info if enabled
	if g.DebugMode {
		g.drawDebugInfo(screen)
	}
}

// drawPlayer renders the player character
func (g *GameManager) drawPlayer(screen *ebiten.Image) {
	if len(g.PlayerFrames) == 0 || g.Player.AnimFrame >= len(g.PlayerFrames) {
		// Fallback if no frames are loaded
		playerColor := color.RGBA{50, 50, 220, 255} // Blue
		x := g.Player.X - g.CameraX
		y := g.Player.Y - g.CameraY
		ebitenutil.DrawRect(
			screen,
			x,
			y,
			float64(cellSize),
			float64(cellSize*2),
			playerColor,
		)
		return
	}

	// Get current frame
	frame := g.PlayerFrames[g.Player.AnimFrame]

	// Draw with appropriate scaling and facing direction
	op := &ebiten.DrawImageOptions{}

	// Make the character larger on screen - increase scale by 50%
	scale := float64(cellSize*1.5) / float64(frame.Bounds().Dx())
	op.GeoM.Scale(scale, scale)

	// Handle facing direction
	if g.Player.FacingLeft {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(cellSize*1.5), 0) // Adjusted for larger size
	}

	// Apply camera offset - center the player properly with the new size
	op.GeoM.Translate(
		g.Player.X-g.CameraX-(float64(cellSize*0.25)), // Center horizontally
		g.Player.Y-g.CameraY-(float64(cellSize*0.5)),  // Center vertically
	)

	screen.DrawImage(frame, op)
}

// drawHUD renders the heads-up display
func (g *GameManager) drawHUD(screen *ebiten.Image) {
	// Level info
	levelInfo := fmt.Sprintf("Level: %d/%d", g.LevelNumber, g.Rules.MaxLevels)
	text.Draw(screen, levelInfo, g.FontFace, 10, 20, color.White)

	// Score
	scoreInfo := fmt.Sprintf("Score: %d", g.Player.Score)
	text.Draw(screen, scoreInfo, g.FontFace, 10, 40, color.White)

	// Lives
	livesInfo := fmt.Sprintf("Lives: %d", g.Player.Lives)
	text.Draw(screen, livesInfo, g.FontFace, 10, 60, color.White)

	// Timer if applicable
	if g.TimeLeft > 0 {
		timeInfo := fmt.Sprintf("Time: %.1f", g.TimeLeft)
		text.Draw(screen, timeInfo, g.FontFace, g.Rules.ScreenWidth-100, 20, color.White)
	}

	// Unlocked abilities
	abilitiesY := 60
	if g.Rules.CanDoubleJump {
		text.Draw(screen, "Double Jump", g.FontFace, g.Rules.ScreenWidth-150, abilitiesY, color.White)
		abilitiesY += 20
	}
	if g.Rules.HasWallJump {
		text.Draw(screen, "Wall Jump", g.FontFace, g.Rules.ScreenWidth-150, abilitiesY, color.White)
		abilitiesY += 20
	}
	if g.Rules.HasDash {
		text.Draw(screen, "Dash", g.FontFace, g.Rules.ScreenWidth-150, abilitiesY, color.White)
	}
}

// drawDebugInfo renders debug information
func (g *GameManager) drawDebugInfo(screen *ebiten.Image) {
	debugInfo := fmt.Sprintf(
		"FPS: %.2f\nPlayer: (%.1f, %.1f)\nVel: (%.1f, %.1f)\nOnGround: %v\nJumpCount: %d",
		ebiten.ActualFPS(),
		g.Player.X, g.Player.Y,
		g.Player.VelX, g.Player.VelY,
		g.Player.OnGround,
		g.Player.JumpCount,
	)

	ebitenutil.DebugPrint(screen, debugInfo)
}

// drawPauseScreen renders the pause overlay
func (g *GameManager) drawPauseScreen(screen *ebiten.Image) {
	// Semi-transparent overlay
	overlay := ebiten.NewImage(g.Rules.ScreenWidth, g.Rules.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 128})
	screen.DrawImage(overlay, nil)

	// Pause text
	pauseText := "PAUSED"
	instructions := "Press SPACE or ESC to continue"

	x := (g.Rules.ScreenWidth - len(pauseText)*7) / 2
	y := g.Rules.ScreenHeight / 2
	text.Draw(screen, pauseText, g.FontFace, x, y, color.White)

	x = (g.Rules.ScreenWidth - len(instructions)*7) / 2
	y += 40
	text.Draw(screen, instructions, g.FontFace, x, y, color.White)
}

// drawLevelComplete renders the level complete screen
func (g *GameManager) drawLevelComplete(screen *ebiten.Image) {
	// Level complete text
	completeText := fmt.Sprintf("LEVEL %d COMPLETE!", g.LevelNumber)
	scoreText := fmt.Sprintf("Score: %d", g.Player.Score)

	var nextText string
	if g.LevelNumber < g.Rules.MaxLevels {
		nextText = "Press SPACE for next level"
	} else {
		nextText = "You beat the game! Press SPACE to return to title"
	}

	// Draw text
	x := (g.Rules.ScreenWidth - len(completeText)*7) / 2
	y := g.Rules.ScreenHeight / 3
	text.Draw(screen, completeText, g.FontFace, x, y, color.White)

	x = (g.Rules.ScreenWidth - len(scoreText)*7) / 2
	y += 40
	text.Draw(screen, scoreText, g.FontFace, x, y, color.White)

	x = (g.Rules.ScreenWidth - len(nextText)*7) / 2
	y += 60
	text.Draw(screen, nextText, g.FontFace, x, y, color.White)
}

// drawGameOver renders the game over screen
func (g *GameManager) drawGameOver(screen *ebiten.Image) {
	// Game over text
	gameOverText := "GAME OVER"
	scoreText := fmt.Sprintf("Final Score: %d", g.TotalScore+g.Player.Score)
	instructions := "Press SPACE to return to title"

	// Draw text
	x := (g.Rules.ScreenWidth - len(gameOverText)*7) / 2
	y := g.Rules.ScreenHeight / 3
	text.Draw(screen, gameOverText, g.FontFace, x, y, color.White)

	x = (g.Rules.ScreenWidth - len(scoreText)*7) / 2
	y += 40
	text.Draw(screen, scoreText, g.FontFace, x, y, color.White)

	x = (g.Rules.ScreenWidth - len(instructions)*7) / 2
	y += 60
	text.Draw(screen, instructions, g.FontFace, x, y, color.White)
}

// drawEnemies renders all enemies in the level
func (g *GameManager) drawEnemies(screen *ebiten.Image) {
	for _, enemy := range g.CurrentLevel.Enemies {
		// Enemy position with camera offset
		x := float64(enemy.X*cellSize) - g.CameraX
		y := float64(enemy.Y*cellSize) - g.CameraY

		// Choose color based on enemy type
		var enemyColor color.RGBA
		switch enemy.Type {
		case EnemyDeveloper:
			enemyColor = color.RGBA{220, 50, 50, 255} // Red
		case EnemyManager:
			enemyColor = color.RGBA{50, 50, 220, 255} // Blue
		case EnemyQATester:
			enemyColor = color.RGBA{220, 220, 50, 255} // Yellow
		default:
			enemyColor = color.RGBA{220, 50, 50, 255} // Default red
		}

		// Draw enemy as rectangle for now (will be replaced with sprites later)
		ebitenutil.DrawRect(
			screen,
			x,
			y,
			float64(cellSize),
			float64(cellSize*2),
			enemyColor,
		)

		// Draw indicator if enemy is active
		if enemy.Active {
			ebitenutil.DrawCircle(
				screen,
				x+float64(cellSize)/2,
				y-5,
				3,
				color.RGBA{255, 255, 255, 255},
			)
		}
	}
}

// drawProjectiles renders all active projectiles
func (g *GameManager) drawProjectiles(screen *ebiten.Image) {
	for _, proj := range g.Projectiles {
		// Skip inactive projectiles
		if !proj.Active {
			continue
		}

		// Projectile position with camera offset
		x := proj.X - g.CameraX
		y := proj.Y - g.CameraY

		// Choose color based on projectile type
		var projColor color.RGBA
		switch proj.Type {
		case ProjectileBug:
			projColor = color.RGBA{255, 100, 100, 255} // Red bug
		case ProjectileScrumNote:
			projColor = color.RGBA{100, 255, 100, 255} // Green scrum note
		case ProjectileErrorReport:
			projColor = color.RGBA{255, 100, 255, 255} // Pink error report
		default:
			projColor = color.RGBA{255, 100, 100, 255} // Default red
		}

		// Draw projectile as circle
		ebitenutil.DrawCircle(
			screen,
			x,
			y,
			proj.Size,
			projColor,
		)
	}
}

// Layout implements the ebiten.Game interface
func (g *GameManager) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.Rules.ScreenWidth, g.Rules.ScreenHeight
}

// updateCamera adjusts the camera position to follow the player with some smoothing
func (g *GameManager) updateCamera() {
	// Target position - center the player on screen
	targetX := g.Player.X - float64(g.Rules.ScreenWidth)/2 + float64(cellSize)/2
	targetY := g.Player.Y - float64(g.Rules.ScreenHeight)/2 + float64(cellSize)

	// Smoothly move camera toward target (camera lag)
	smoothFactor := 0.1
	g.CameraX += (targetX - g.CameraX) * smoothFactor
	g.CameraY += (targetY - g.CameraY) * smoothFactor

	// Limit camera to level bounds
	levelWidthPixels := float64(g.CurrentLevel.Width * cellSize)
	levelHeightPixels := float64(g.CurrentLevel.Height * cellSize)

	// Don't go beyond left or top edge
	if g.CameraX < 0 {
		g.CameraX = 0
	}
	if g.CameraY < 0 {
		g.CameraY = 0
	}

	// Don't go beyond right or bottom edge
	maxX := levelWidthPixels - float64(g.Rules.ScreenWidth)
	maxY := levelHeightPixels - float64(g.Rules.ScreenHeight)

	if maxX > 0 && g.CameraX > maxX {
		g.CameraX = maxX
	}
	if maxY > 0 && g.CameraY > maxY {
		g.CameraY = maxY
	}
}
