package main

import (
	"math/rand"
)

// PlatformType defines the type of platform
type PlatformType int

const (
	PlatformNormal PlatformType = iota
	PlatformMoving
	PlatformBreaking
	PlatformBouncy
)

type Platform struct {
	X, Y          int          
	Width, Height int          
	Type          PlatformType 
	MoveSpeed     float64     
	MoveRange     int          
	CurrentMove   float64      
	Breaking      bool         
	BreakTimer    int          
}

type Obstacle struct {
	X, Y          int // Position
	Width, Height int
	Type          int // Type of obstacle
	Moving        bool
	Speed         float64 
}

type EnemyType int

const (
	EnemyDeveloper EnemyType = iota
	EnemyManager
	EnemyQATester
)

type EnemyBehavior int

const (
	EnemyPacing EnemyBehavior = iota
	EnemyStationary
	EnemyPatrolling
)

// Projectile types
type ProjectileType int

const (
	ProjectileBug ProjectileType = iota
	ProjectileScrumNote
	ProjectileErrorReport
)

// Enemy defines developer enemies that attack the player
type Enemy struct {
	X, Y             int            // Position
	Type             EnemyType      // Type of enemy
	Gender           string         // "male" or "female"
	Behavior         EnemyBehavior  // Movement behavior
	PatrolPoints     [][2]int       // Patrol path points if patrolling
	Speed            float64        // Movement speed
	DetectionRadius  int            // How far they can detect the player
	ProjectileType   ProjectileType // What they throw
	FireRate         float64        // How often they throw projectiles
	CurrentPatrolIdx int            // Current position in patrol
	PatrolDirection  int            // 1 = forward, -1 = backward
	LastFireTime     float64        // Time tracking for projectile throwing
	Active           bool           // Whether enemy is currently active
}

// Collectible defines items the player can collect
type Collectible struct {
	X, Y   int  // Position
	Type   int  // Type of collectible
	Value  int  // Point value
	Active bool // Is it still available to collect
}

// Level contains all elements of a game level
type Level struct {
	Number         int           // Level number
	Platforms      []Platform    // All platforms in the level
	Obstacles      []Obstacle    // All obstacles in the level
	Enemies        []Enemy       // All developer enemies in the level
	Collectibles   []Collectible // All collectibles
	StartX, StartY int           // Starting position
	ExitX, ExitY   int           // Level exit position
	Width, Height  int           // Level dimensions
}

// PlatformGenerator manages level generation
type PlatformGenerator struct {
	Rules   *GameRules // Game rules reference
	RandGen *rand.Rand // Random number generator
}

// NewPlatformGenerator creates a new platform generator
func NewPlatformGenerator(rules *GameRules, seed int64) *PlatformGenerator {
	if seed == 0 {
		seed = rand.Int63()
	}

	return &PlatformGenerator{
		Rules:   rules,
		RandGen: rand.New(rand.NewSource(seed)),
	}
}

// GenerateLevel creates a new level layout based on the current level number
func (p *PlatformGenerator) GenerateLevel(levelNum int) *Level {
	// Update rules for the current level
	p.Rules.CurrentLevel = levelNum
	p.Rules.UnlockFeatures()

	// Get difficulty parameters
	_, obstacleDensity, collectibleDensity := p.Rules.GetLevelDifficulty()

	// Set up the level
	width := p.Rules.ScreenWidth / cellSize
	height := p.Rules.ScreenHeight / cellSize

	level := &Level{
		Number:       levelNum,
		Width:        width,
		Height:       height,
		Platforms:    make([]Platform, 0),
		Obstacles:    make([]Obstacle, 0),
		Enemies:      make([]Enemy, 0),
		Collectibles: make([]Collectible, 0),
	}

	// Create starting platform - larger and more stable
	startX := 1
	startY := height - 5 // Lower position, closer to the bottom for better visibility
	startWidth := 7      // Wider platform to start on

	// Player starts directly on the platform
	level.StartX = startX + startWidth/2
	level.StartY = startY - 2 // Just above the platform

	// Add starting platform
	level.Platforms = append(level.Platforms, Platform{
		X:      startX,
		Y:      startY,
		Width:  startWidth,
		Height: 1,
		Type:   PlatformNormal,
	})

	// Add a floor platform that spans the entire level width
	floorY := height - 2
	level.Platforms = append(level.Platforms, Platform{
		X:      0,
		Y:      floorY,
		Width:  width,
		Height: 1,
		Type:   PlatformNormal,
	})

	// Create a Donkey Kong style layout with ladder-like platforms
	// Number of rows (levels)
	numRows := 10 + (levelNum / 2)
	if numRows > 15 {
		numRows = 15 // Cap at 15 rows maximum
	}

	// Vertical spacing between rows
	rowSpacing := (startY - 5) / numRows
	if rowSpacing < 3 {
		rowSpacing = 3 // Minimum spacing
	}

	// Current position tracking
	currentY := startY - rowSpacing

	// Direction flag (alternate between left and right)
	goingRight := true

	// Margin from screen edges
	leftMargin := 3
	rightMargin := width - 10

	// Create ladder-like platforms row by row
	for row := 0; row < numRows; row++ {
		// Alternate between left-to-right and right-to-left for each row
		platformWidth := 10 - (levelNum / 3)
		if platformWidth < 5 {
			platformWidth = 5 // Minimum platform width
		}

		// Difficulty scaling - later levels have narrower platforms
		if levelNum > 5 {
			widthReduction := 1 + (levelNum / 5)
			platformWidth -= widthReduction
			if platformWidth < 4 {
				platformWidth = 4
			}
		}

		// Platform positioning
		var platformX int
		if goingRight {
			platformX = leftMargin
		} else {
			platformX = rightMargin - platformWidth
		}

		// Add the main platform for this row
		mainPlatform := Platform{
			X:      platformX,
			Y:      currentY,
			Width:  platformWidth,
			Height: 1,
			Type:   PlatformNormal,
		}

		// Configure special platform types for variety
		if levelNum > 2 && p.RandGen.Intn(10) < levelNum-1 {
			platformTypeRoll := p.RandGen.Intn(10)
			if platformTypeRoll < 2 && levelNum > 3 { // 20% chance in later levels
				mainPlatform.Type = PlatformMoving
			} else if platformTypeRoll < 4 { // 20% chance
				mainPlatform.Type = PlatformBouncy
			} else if platformTypeRoll < 5 && levelNum > 4 { // 10% chance in higher levels
				mainPlatform.Type = PlatformBreaking
			}
		}

		// Configure moving platforms
		if mainPlatform.Type == PlatformMoving {
			mainPlatform.MoveSpeed = 0.2 + float64(levelNum)/25.0
			mainPlatform.MoveRange = 3 + levelNum/3

			// For ladder-like levels, make platforms move horizontally
			if goingRight {
				// Moving right from left edge
				mainPlatform.MoveRange = (width / 5) - platformWidth/2
			} else {
				// Moving left from right edge
				mainPlatform.MoveRange = (width / 5) - platformWidth/2
				mainPlatform.MoveSpeed = -mainPlatform.MoveSpeed           // Start moving left
				mainPlatform.CurrentMove = float64(mainPlatform.MoveRange) // Start at max position
			}
		}

		level.Platforms = append(level.Platforms, mainPlatform)

		// Add collectibles above this platform
		for i := 1; i < platformWidth; i += 2 {
			if p.RandGen.Intn(10) < collectibleDensity {
				collectibleX := platformX + i
				collectibleY := currentY - 1

				collectible := Collectible{
					X:      collectibleX,
					Y:      collectibleY,
					Type:   p.RandGen.Intn(3),
					Value:  10 + p.RandGen.Intn(levelNum*5),
					Active: true,
				}

				level.Collectibles = append(level.Collectibles, collectible)
			}
		}

		// Add obstacles on platforms (from level 3 onward)
		if levelNum >= 3 && p.RandGen.Intn(10) < obstacleDensity {
			obstacleX := platformX + platformWidth/2
			obstacleY := currentY - 1

			obstacle := Obstacle{
				X:      obstacleX,
				Y:      obstacleY,
				Width:  1,
				Height: 1,
				Type:   p.RandGen.Intn(3),
			}

			level.Obstacles = append(level.Obstacles, obstacle)
		}

		// Add connecting platforms to make jumping easier
		// These will be placed in the gap between the current platform and the next one
		if row < numRows-1 {
			bridgeOffset := 4 + p.RandGen.Intn(3) // How far from the edge to place it
			bridgeWidth := 2 + p.RandGen.Intn(2)  // Small platforms
			bridgeY := currentY - rowSpacing/2    // Halfway to the next row

			var bridgeX int
			if goingRight {
				// Coming from left side, place bridge on right side
				bridgeX = platformX + platformWidth + bridgeOffset

				// Ensure the bridge doesn't go too far right
				if bridgeX+bridgeWidth > rightMargin {
					bridgeX = rightMargin - bridgeWidth - 1
				}
			} else {
				// Coming from right side, place bridge on left side
				bridgeX = platformX - bridgeOffset - bridgeWidth

				// Ensure the bridge doesn't go too far left
				if bridgeX < leftMargin {
					bridgeX = leftMargin + 1
				}
			}

			// Create a helper platform to make jumps easier
			if bridgeX > 0 && bridgeX+bridgeWidth < width {
				bridgePlatform := Platform{
					X:      bridgeX,
					Y:      bridgeY,
					Width:  bridgeWidth,
					Height: 1,
					Type:   PlatformNormal,
				}

				// Higher chance of special platforms for bridges
				if p.RandGen.Intn(10) < 5 {
					bridgePlatform.Type = PlatformBouncy // Bouncy platforms help with vertical movement
				}

				level.Platforms = append(level.Platforms, bridgePlatform)

				// Add collectibles above the bridge sometimes
				if p.RandGen.Intn(10) < collectibleDensity {
					collectible := Collectible{
						X:      bridgeX + bridgeWidth/2,
						Y:      bridgeY - 1,
						Type:   p.RandGen.Intn(3),
						Value:  15 + p.RandGen.Intn(levelNum*5),
						Active: true,
					}

					level.Collectibles = append(level.Collectibles, collectible)
				}
			}

			// Add a second bridge in higher levels for more movement options
			if levelNum > 3 && p.RandGen.Intn(10) < 7 {
				bridge2Width := 2 + p.RandGen.Intn(2)
				bridge2Y := currentY - (rowSpacing*2)/3 // 2/3 way to the next row

				// Place it at a different horizontal position
				var bridge2X int
				middleX := width / 2

				if p.RandGen.Intn(2) == 0 {
					// Left half of screen
					bridge2X = leftMargin + p.RandGen.Intn(middleX-leftMargin-bridge2Width)
				} else {
					// Right half of screen
					bridge2X = middleX + p.RandGen.Intn(rightMargin-middleX-bridge2Width)
				}

				// Create the second bridge
				if bridge2X > 0 && bridge2X+bridge2Width < width {
					bridge2Platform := Platform{
						X:      bridge2X,
						Y:      bridge2Y,
						Width:  bridge2Width,
						Height: 1,
						Type:   PlatformNormal,
					}

					level.Platforms = append(level.Platforms, bridge2Platform)
				}
			}
		}

		// After updating for next row in the generateLevel function
		// Add developer enemies on some platforms
		if row > 0 && row < numRows-1 && p.RandGen.Intn(10) < 3+levelNum/2 {
			// 30% chance + level bonus to add an enemy on this row's platform
			p.addDeveloperEnemy(level, platformX, currentY, platformWidth, levelNum)
		}

		// Update for next row
		currentY -= rowSpacing
		goingRight = !goingRight // Alternate direction
	}

	// Create the exit platform at the top
	exitPlatformWidth := 5
	exitPlatformX := width/2 - exitPlatformWidth/2
	exitY := currentY + 1

	// Exit platform
	exitPlatform := Platform{
		X:      exitPlatformX,
		Y:      exitY,
		Width:  exitPlatformWidth,
		Height: 1,
		Type:   PlatformNormal,
	}

	level.Platforms = append(level.Platforms, exitPlatform)

	// Set the exit position
	level.ExitX = exitPlatformX + exitPlatformWidth/2
	level.ExitY = exitY - 1

	// Add bonus collectibles near exit
	for i := 0; i < 3; i++ {
		collectible := Collectible{
			X:      exitPlatformX + 1 + i,
			Y:      exitY - 1,
			Type:   p.RandGen.Intn(3),
			Value:  25 + p.RandGen.Intn(levelNum*10),
			Active: true,
		}
		level.Collectibles = append(level.Collectibles, collectible)
	}

	return level
}

// addCollectiblesAndObstacles adds collectibles and obstacles to a platform
func (p *PlatformGenerator) addCollectiblesAndObstacles(level *Level, x, y, width, levelNum, obstacleDensity, collectibleDensity int) {
	// Add obstacles based on density
	if obstacleDensity > 0 && p.RandGen.Intn(10) < obstacleDensity && width > 3 {
		obstacleX := x + 1 + p.RandGen.Intn(width-2)
		obstacle := Obstacle{
			X:      obstacleX,
			Y:      y - 1, // Place on top of platform
			Width:  1,
			Height: 1,
			Type:   p.RandGen.Intn(3), // Different types of obstacles
		}

		if levelNum > 5 && p.RandGen.Intn(10) < 5 {
			obstacle.Moving = true
			obstacle.Speed = 0.3 + float64(levelNum)/25.0
		}

		level.Obstacles = append(level.Obstacles, obstacle)
	}

	// Add collectibles based on density
	if p.RandGen.Intn(10) < collectibleDensity && width > 2 {
		collectible := Collectible{
			X:      x + p.RandGen.Intn(width),
			Y:      y - 2, // Float above platform
			Type:   p.RandGen.Intn(3),
			Value:  10 + p.RandGen.Intn(levelNum*5),
			Active: true,
		}

		level.Collectibles = append(level.Collectibles, collectible)
	}
}

// fillEmptyAreas adds extra platforms to fill larger empty spaces
func (p *PlatformGenerator) fillEmptyAreas(level *Level, targetDensity, levelNum int) {
	width := level.Width
	height := level.Height

	// Create a grid to mark where platforms exist
	grid := make([][]bool, width)
	for i := range grid {
		grid[i] = make([]bool, height)
	}

	// Mark existing platforms on the grid
	for _, platform := range level.Platforms {
		for x := platform.X; x < platform.X+platform.Width && x < width; x++ {
			if x >= 0 && platform.Y >= 0 && platform.Y < height {
				grid[x][platform.Y] = true
			}
		}
	}

	// Look for empty areas to place additional platforms
	for x := 5; x < width-10; x += 5 {
		for y := 5; y < height-5; y += 5 {
			// Check if this area is empty
			isEmpty := true
			for checkX := x - 3; checkX <= x+3 && isEmpty; checkX++ {
				for checkY := y - 3; checkY <= y+3 && isEmpty; checkY++ {
					if checkX >= 0 && checkX < width && checkY >= 0 && checkY < height {
						if grid[checkX][checkY] {
							isEmpty = false
						}
					}
				}
			}

			// Add a platform in this empty area
			if isEmpty && p.RandGen.Intn(10) < 7 {
				platformWidth := 2 + p.RandGen.Intn(4)
				platformType := PlatformType(p.RandGen.Intn(int(PlatformBouncy) + 1))

				platform := Platform{
					X:      x,
					Y:      y,
					Width:  platformWidth,
					Height: 1,
					Type:   platformType,
				}

				// Configure special platforms
				if platformType == PlatformMoving {
					platform.MoveSpeed = 0.3 + float64(levelNum)/20.0
					platform.MoveRange = 2 + levelNum/3
				}

				level.Platforms = append(level.Platforms, platform)

				// Mark this platform on the grid
				for px := x; px < x+platformWidth && px < width; px++ {
					if px >= 0 && y >= 0 && y < height {
						grid[px][y] = true
					}
				}

				// Add a collectible
				if p.RandGen.Intn(10) < 7 {
					collectible := Collectible{
						X:      x + platformWidth/2,
						Y:      y - 1,
						Type:   p.RandGen.Intn(3),
						Value:  15 + p.RandGen.Intn(levelNum*5),
						Active: true,
					}
					level.Collectibles = append(level.Collectibles, collectible)
				}

				// Stop once we've reached the target density
				if len(level.Platforms) >= targetDensity {
					return
				}
			}
		}
	}
}

// randomInRange returns a random number within the given range
func (p *PlatformGenerator) randomInRange(rangeValues []int) int {
	min, max := rangeValues[0], rangeValues[1]
	return p.RandGen.Intn(max-min+1) + min
}

// addDeveloperEnemy places an angry developer enemy on a platform
func (p *PlatformGenerator) addDeveloperEnemy(level *Level, platformX, platformY, platformWidth int, levelNum int) {
	// Determine position on platform
	posX := platformX + p.RandGen.Intn(platformWidth)
	posY := platformY - 1 // Place on top of platform

	// Determine enemy type - weighted based on level
	enemyTypeRoll := p.RandGen.Intn(10)
	var enemyType EnemyType

	if enemyTypeRoll < 6 {
		enemyType = EnemyDeveloper // 60% chance of developer
	} else if enemyTypeRoll < 9 {
		enemyType = EnemyQATester // 30% chance of QA tester
	} else {
		enemyType = EnemyManager // 10% chance of manager
	}

	// Determine gender
	gender := "male"
	if p.RandGen.Intn(2) == 1 {
		gender = "female"
	}

	// Determine behavior
	var behavior EnemyBehavior
	behaviorRoll := p.RandGen.Intn(10)

	if behaviorRoll < 5 {
		behavior = EnemyPacing // 50% pacing
	} else if behaviorRoll < 8 {
		behavior = EnemyStationary // 30% stationary
	} else {
		behavior = EnemyPatrolling // 20% patrolling
	}

	// Set up patrol points if needed
	var patrolPoints [][2]int

	if behavior == EnemyPacing {
		// Pacing behavior - move across platform
		patrolPoints = make([][2]int, 2)
		patrolPoints[0] = [2]int{platformX, posY}
		patrolPoints[1] = [2]int{platformX + platformWidth - 1, posY}
	} else if behavior == EnemyPatrolling {
		// Patrolling behavior - create more complex path
		numPoints := 2 + p.RandGen.Intn(3) // 2-4 points
		patrolPoints = make([][2]int, numPoints)

		// First point is current position
		patrolPoints[0] = [2]int{posX, posY}

		// Create additional points
		for i := 1; i < numPoints; i++ {
			// Find a nearby platform to patrol to
			found := false
			attempts := 0
			var nextX, nextY int

			for !found && attempts < 10 {
				attempts++

				// Try to find a nearby platform
				searchX := posX + p.RandGen.Intn(15) - 7 // Look 7 cells in either direction
				searchY := posY + p.RandGen.Intn(7) - 3  // Look a few cells up/down

				// Check if there's a platform at this position
				for _, plat := range level.Platforms {
					if searchX >= plat.X && searchX < plat.X+plat.Width &&
						searchY == plat.Y-1 { // On top of platform
						nextX = searchX
						nextY = searchY
						found = true
						break
					}
				}
			}

			if found {
				patrolPoints[i] = [2]int{nextX, nextY}
			} else {
				// If we couldn't find a good platform, just use a point on the current platform
				offsetX := p.RandGen.Intn(platformWidth)
				patrolPoints[i] = [2]int{platformX + offsetX, posY}
			}
		}
	}

	// Determine projectile type
	var projectileType ProjectileType
	projRoll := p.RandGen.Intn(10)

	if projRoll < 5 {
		projectileType = ProjectileBug // 50% bugs
	} else if projRoll < 8 {
		projectileType = ProjectileScrumNote // 30% scrum notes
	} else {
		projectileType = ProjectileErrorReport // 20% error reports
	}

	// Create the enemy
	enemy := Enemy{
		X:               posX,
		Y:               posY,
		Type:            enemyType,
		Gender:          gender,
		Behavior:        behavior,
		PatrolPoints:    patrolPoints,
		Speed:           0.5 + float64(levelNum)/10.0, // Speed increases with level
		DetectionRadius: 5 + levelNum/2,               // Detection radius increases with level
		ProjectileType:  projectileType,
		FireRate:        0.5 + float64(levelNum)/10.0, // Fire rate increases with level
		PatrolDirection: 1,                            // Start moving forward
		Active:          false,                        // Start inactive
	}

	level.Enemies = append(level.Enemies, enemy)
}
