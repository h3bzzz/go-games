package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PlayerState tracks the current state of the player
type PlayerState int

const (
	PlayerStateIdle PlayerState = iota
	PlayerStateRunning
	PlayerStateJumping
	PlayerStateFalling
	PlayerStateWallSliding
	PlayerStateDashing
)

// Player represents the player character with physics
type Player struct {
	// Position and movement
	X, Y       float64     // Position in the world
	VelX, VelY float64     // Velocity
	State      PlayerState // Current state
	FacingLeft bool        // Direction facing

	// Physics parameters (taken from GameRules)
	Gravity      float64 // Downward acceleration
	JumpForce    float64 // Upward velocity when jumping
	Speed        float64 // Horizontal movement speed
	MaxFallSpeed float64 // Terminal velocity

	// Jumping mechanics
	OnGround      bool // Whether the player is on a platform
	JumpCount     int  // Number of jumps performed (for double jump)
	CanDoubleJump bool // Whether double jump is enabled
	HasWallJump   bool // Whether wall jump is enabled
	AgainstWall   bool // Whether player is against a wall

	// Special abilities
	HasDash       bool // Whether dash is enabled
	DashTimer     int  // Cooldown timer for dash
	DashDirection int  // Direction of dash (-1 left, 1 right)

	// Animation
	AnimFrame   int             // Current animation frame
	AnimCounter int             // Counter for animation timing
	Frames      []*ebiten.Image // Animation frames

	// Game metrics
	Score int // Current score
	Lives int // Remaining lives

	// Level reference (for collision detection)
	CurrentLevel *Level
}

// NewPlayer creates a new player at the specified position
func NewPlayer(x, y int, rules *GameRules, level *Level, frames []*ebiten.Image) *Player {
	return &Player{
		X:             float64(x * cellSize),
		Y:             float64(y * cellSize),
		State:         PlayerStateIdle,
		FacingLeft:    false,
		Gravity:       rules.Gravity,
		JumpForce:     rules.JumpForce,
		Speed:         rules.PlayerSpeed,
		MaxFallSpeed:  rules.MaxFallSpeed,
		OnGround:      false,
		JumpCount:     0,
		CanDoubleJump: rules.CanDoubleJump,
		HasWallJump:   rules.HasWallJump,
		HasDash:       rules.HasDash,
		DashTimer:     0,
		Lives:         3,
		CurrentLevel:  level,
		Frames:        frames,
	}
}

// Update handles player physics and movement
func (p *Player) Update(input *PlayerInput) {
	prevX, prevY := p.X, p.Y

	// Handle horizontal movement with acceleration and deceleration
	const (
		accelerationGround = 1.0  // How quickly to accelerate on ground
		accelerationAir    = 0.7  // How quickly to accelerate in air
		friction           = 0.85 // Friction when on ground (higher = less friction)
		airResistance      = 0.95 // Resistance in air (higher = less resistance)
	)

	if !p.isDashing() {
		// Apply friction/air resistance
		if p.OnGround {
			p.VelX *= friction
		} else {
			p.VelX *= airResistance
		}

		// Apply acceleration based on input
		acceleration := accelerationGround
		if !p.OnGround {
			acceleration = accelerationAir
		}

		if input.Left {
			p.VelX -= acceleration
			p.FacingLeft = false
			if p.OnGround {
				p.State = PlayerStateRunning
			}
		} else if input.Right {
			p.VelX += acceleration
			p.FacingLeft = true
			if p.OnGround {
				p.State = PlayerStateRunning
			}
		} else if p.OnGround && math.Abs(p.VelX) < 0.5 {
			p.VelX = 0
			p.State = PlayerStateIdle
		}

		// Clamp horizontal velocity
		if p.VelX > p.Speed {
			p.VelX = p.Speed
		} else if p.VelX < -p.Speed {
			p.VelX = -p.Speed
		}
	}

	// Add a slight grace period for jumping after leaving a platform
	const groundedGracePeriod = 5 // Frames
	var groundedGraceCounter int  // Changed from static to var

	if p.OnGround {
		groundedGraceCounter = groundedGracePeriod
	} else if groundedGraceCounter > 0 {
		groundedGraceCounter--
	}

	// Allow jump if within grace period or actually on ground
	canGroundJump := p.OnGround || groundedGraceCounter > 0

	// Handle jumping with variable height based on button hold
	if input.JumpPressed && canGroundJump {
		p.VelY = -p.JumpForce
		p.OnGround = false
		groundedGraceCounter = 0
		p.JumpCount = 1
		p.State = PlayerStateJumping
	} else if input.JumpPressed && p.JumpCount < 2 && !canGroundJump {
		// Double jump - always enabled for better gameplay
		p.VelY = -p.JumpForce * 0.9 // Slightly stronger double jump
		p.JumpCount = 2
		p.State = PlayerStateJumping
	}

	// Variable jump height based on early button release
	if !input.Jump && p.VelY < 0 && p.State == PlayerStateJumping {
		p.VelY *= 0.5 // Cut jump short if button released
	}

	// Handle wall jump
	if p.HasWallJump && input.JumpPressed && !p.OnGround && p.AgainstWall && p.JumpCount < 2 {
		// Jump away from wall
		p.VelY = -p.JumpForce * 0.9
		if p.FacingLeft {
			p.VelX = p.Speed * 1.2 // Jump right with some boost
			p.FacingLeft = true    // Keep facing left when jumping right
		} else {
			p.VelX = -p.Speed * 1.2 // Jump left with some boost
			p.FacingLeft = false    // Keep facing right when jumping left
		}
		p.JumpCount = 1 // Reset to allow double jump after wall jump
		p.State = PlayerStateJumping
	}

	// Handle dash
	if p.HasDash && input.DashPressed && p.DashTimer <= 0 {
		direction := 1
		if p.FacingLeft {
			direction = -1 // Left-facing dashes left
		} else {
			direction = 1 // Right-facing dashes right
		}
		p.DashDirection = direction
		p.DashTimer = 15 // Dash duration
		p.State = PlayerStateDashing

		// Add slight upward boost to dash so it doesn't drop too much
		if !p.OnGround {
			p.VelY = -1.0
		}
	}

	// Apply dash
	if p.isDashing() {
		p.VelX = float64(p.DashDirection) * p.Speed * 2.5

		// Gradually restore gravity toward end of dash
		if p.DashTimer < 5 {
			p.VelY += p.Gravity * 0.5
		} else {
			p.VelY = 0 // No gravity during dash
		}

		p.DashTimer--
	} else {
		// Apply gravity if not dashing
		p.VelY += p.Gravity

		// Wall sliding
		if p.HasWallJump && p.AgainstWall && p.VelY > 0 && !p.OnGround {
			p.VelY = math.Min(p.VelY, p.MaxFallSpeed/3) // Slower fall when against wall
			p.State = PlayerStateWallSliding
		}

		// Limit fall speed
		if p.VelY > p.MaxFallSpeed {
			p.VelY = p.MaxFallSpeed
		}
	}

	// Update position
	p.X += p.VelX
	p.Y += p.VelY

	// Update state based on velocity
	if p.VelY > 0.1 && !p.OnGround && !p.isDashing() && p.State != PlayerStateWallSliding {
		p.State = PlayerStateFalling
	}

	// Reset variables for this frame
	wasOnGround := p.OnGround
	p.OnGround = false
	p.AgainstWall = false

	// Check platform collisions
	p.handleCollisions(prevX, prevY)

	// Reset jump count when landing
	if p.OnGround && !wasOnGround {
		p.JumpCount = 0
	}

	// Handle collectibles
	p.checkCollectibles()

	// Check if player has fallen off the level
	if p.Y > float64(p.CurrentLevel.Height*cellSize) {
		p.respawn()
	}

	// Handle animation
	p.updateAnimation()
}

// handleCollisions checks and resolves collisions with platforms and obstacles
func (p *Player) handleCollisions(prevX, prevY float64) {
	// Player hitbox
	playerWidth := float64(cellSize)
	playerHeight := float64(cellSize * 2) // Taller than a single cell

	// Check platform collisions
	for i := range p.CurrentLevel.Platforms {
		platform := &p.CurrentLevel.Platforms[i]
		platformX := float64(platform.X * cellSize)
		platformY := float64(platform.Y * cellSize)
		platformWidth := float64(platform.Width * cellSize)
		platformHeight := float64(platform.Height * cellSize)

		// Check if player is horizontally aligned with platform
		if p.X+playerWidth > platformX && p.X < platformX+platformWidth {
			// Landing on platform from above
			if p.Y+playerHeight > platformY && prevY+playerHeight <= platformY && p.VelY >= 0 {
				// Handle different platform types
				switch platform.Type {
				case PlatformNormal:
					p.Y = platformY - playerHeight
					p.OnGround = true
					p.VelY = 0

				case PlatformBouncy:
					p.Y = platformY - playerHeight
					p.VelY = -p.JumpForce * 1.5 // Higher bounce
					p.JumpCount = 0             // Reset jump count for consecutive bounces

				case PlatformBreaking:
					p.Y = platformY - playerHeight
					p.OnGround = true
					p.VelY = 0
					platform.Breaking = true

				case PlatformMoving:
					p.Y = platformY - playerHeight
					p.OnGround = true
					p.VelY = 0
					// Move player with platform
					p.X += platform.MoveSpeed
				}
			}

			// Hitting head on platform from below
			if p.Y < platformY+platformHeight && prevY >= platformY+platformHeight && p.VelY < 0 {
				p.Y = platformY + platformHeight
				p.VelY = 0
			}
		}

		// Check for wall collision (side collision)
		if p.Y+playerHeight > platformY && p.Y < platformY+platformHeight {
			// Collision from the right
			if p.X < platformX+platformWidth && p.X+playerWidth > platformX+platformWidth && prevX >= platformX+platformWidth {
				p.X = platformX + platformWidth
				p.AgainstWall = true
				p.VelX = 0
			}

			// Collision from the left
			if p.X+playerWidth > platformX && p.X < platformX && prevX+playerWidth <= platformX {
				p.X = platformX - playerWidth
				p.AgainstWall = true
				p.VelX = 0
			}
		}
	}

	// Check obstacle collisions
	for _, obstacle := range p.CurrentLevel.Obstacles {
		obstacleX := float64(obstacle.X * cellSize)
		obstacleY := float64(obstacle.Y * cellSize)
		obstacleWidth := float64(obstacle.Width * cellSize)
		obstacleHeight := float64(obstacle.Height * cellSize)

		// Simple AABB collision detection
		if p.X+playerWidth > obstacleX && p.X < obstacleX+obstacleWidth &&
			p.Y+playerHeight > obstacleY && p.Y < obstacleY+obstacleHeight {
			// Player hit an obstacle
			p.respawn()
			break
		}
	}

	// Check level exit
	exitX := float64(p.CurrentLevel.ExitX * cellSize)
	exitY := float64(p.CurrentLevel.ExitY * cellSize)
	exitWidth := float64(cellSize)
	exitHeight := float64(cellSize * 2)

	if p.X+playerWidth > exitX && p.X < exitX+exitWidth &&
		p.Y+playerHeight > exitY && p.Y < exitY+exitHeight {
		// Player reached the exit - level complete!
		// This would normally trigger level completion
	}
}

// checkCollectibles checks for collisions with collectibles
func (p *Player) checkCollectibles() {
	playerWidth := float64(cellSize)
	playerHeight := float64(cellSize * 2)

	for i := range p.CurrentLevel.Collectibles {
		collectible := &p.CurrentLevel.Collectibles[i]

		if !collectible.Active {
			continue
		}

		collectibleX := float64(collectible.X * cellSize)
		collectibleY := float64(collectible.Y * cellSize)
		collectibleSize := float64(cellSize)

		// Check collision
		if p.X+playerWidth > collectibleX && p.X < collectibleX+collectibleSize &&
			p.Y+playerHeight > collectibleY && p.Y < collectibleY+collectibleSize {
			// Collect the item
			p.Score += collectible.Value
			collectible.Active = false
		}
	}
}

// canJump determines if the player can jump
func (p *Player) canJump() bool {
	if p.OnGround {
		return true
	}

	if p.CanDoubleJump && p.JumpCount < 2 {
		return true
	}

	return false
}

// isDashing checks if player is currently dashing
func (p *Player) isDashing() bool {
	return p.DashTimer > 0
}

// respawn resets the player to the starting position
func (p *Player) respawn() {
	p.X = float64(p.CurrentLevel.StartX * cellSize)
	p.Y = float64(p.CurrentLevel.StartY * cellSize)
	p.VelX = 0
	p.VelY = 0
	p.Lives--
	// Game over check would go here
}

// updateAnimation handles player animation
func (p *Player) updateAnimation() {
	// Animation speed depends on state
	animSpeed := 10
	if p.State == PlayerStateRunning {
		animSpeed = 5
	}

	p.AnimCounter++
	if p.AnimCounter >= animSpeed {
		p.AnimFrame = (p.AnimFrame + 1) % len(p.Frames)
		p.AnimCounter = 0
	}
}

// PlayerInput encapsulates input state
type PlayerInput struct {
	Left, Right, Jump, Dash  bool
	JumpPressed, DashPressed bool // New fields for single-frame detection
}

// GetInput reads keyboard input
func GetPlayerInput() *PlayerInput {
	return &PlayerInput{
		Left:  ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA),
		Right: ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD),
		Jump:  ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
		Dash:  ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyE),

		// Just-pressed detection for precise jumping and dashing
		JumpPressed: inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsKeyJustPressed(ebiten.KeyW) ||
			inpututil.IsKeyJustPressed(ebiten.KeyUp),

		DashPressed: inpututil.IsKeyJustPressed(ebiten.KeyShift) ||
			inpututil.IsKeyJustPressed(ebiten.KeyE),
	}
}
