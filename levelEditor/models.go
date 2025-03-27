package main

// Level represents a game level
type Level struct {
	Name         string     // Level name
	Width        int        // Level width in grid units
	Height       int        // Level height in grid units
	StartX       int        // Player start position X
	StartY       int        // Player start position Y
	ExitX        int        // Level exit position X
	ExitY        int        // Level exit position Y
	Platforms    []Platform // Platforms in the level
	Enemies      []Enemy    // Enemies in the level
	Collectibles []Item     // Collectible items
	Background   string     // Background image file
}

// Platform defines a platform in the level
type Platform struct {
	X          int     // Position X
	Y          int     // Position Y
	Width      int     // Width in grid units
	Height     int     // Height in grid units
	Type       string  // Platform type
	MovementX  int     // Movement distance X (0 for static)
	MovementY  int     // Movement distance Y (0 for static)
	Speed      float64 // Movement speed
	SpritePath string  // Path to sprite image
}

// Enemy defines an enemy in the level
type Enemy struct {
	X               int      // Position X
	Y               int      // Position Y
	Type            string   // Enemy type
	Gender          string   // Gender (male/female)
	Behavior        string   // Behavior type (patrol, chase, stationary)
	PatrolPoints    [][2]int // Patrol waypoints
	Speed           float64  // Movement speed
	DetectionRadius int      // Player detection radius
	ProjectileType  string   // Type of projectile enemy fires
	FireRate        float64  // Projectile fire rate
	SpritePath      string   // Path to sprite image
}

// Item defines a collectible item
type Item struct {
	X          int    // Position X
	Y          int    // Position Y
	Type       string // Item type
	Value      int    // Item value
	SpritePath string // Path to sprite image
}

// NewLevel creates a new level with default settings
func NewLevel(name string, width, height int) *Level {
	// Create a default level with a floor platform
	level := &Level{
		Name:         name,
		Width:        width,
		Height:       height,
		StartX:       2,
		StartY:       2,
		ExitX:        width - 3,
		ExitY:        2,
		Platforms:    make([]Platform, 0),
		Enemies:      make([]Enemy, 0),
		Collectibles: make([]Item, 0),
	}

	// Add a default floor platform
	floorPlatform := Platform{
		X:      0,
		Y:      height - 2,
		Width:  width,
		Height: 2,
		Type:   "solid",
	}
	level.Platforms = append(level.Platforms, floorPlatform)

	return level
}
