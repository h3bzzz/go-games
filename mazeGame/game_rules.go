package main

type GameRules struct {
	ScreenWidth  int
	ScreenHeight int

	MaxLevels          int
	CurrentLevel       int
	PlatformDensity    []int
	PlatformSizeRange  []int
	GapSizeRange       []int
	VerticalVariance   []int
	ObstacleDensity    []int
	CollectibleDensity []int

	Gravity      float64
	JumpForce    float64
	PlayerSpeed  float64
	MaxFallSpeed float64

	CanDoubleJump bool
	HasWallJump   bool
	HasDash       bool

	ScoreThresholds []int
	TimeLimit       []int
}

func NewGameRules() *GameRules {
	return &GameRules{
		ScreenWidth:  840,
		ScreenHeight: 840,

		MaxLevels:    10,
		CurrentLevel: 1,

		PlatformDensity:    []int{12, 11, 10, 9, 8, 7, 6, 5, 4, 3},
		PlatformSizeRange:  []int{3, 10},
		GapSizeRange:       []int{2, 8},
		VerticalVariance:   []int{0, 5},
		ObstacleDensity:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		CollectibleDensity: []int{5, 5, 6, 6, 7, 7, 8, 8, 9, 10},

		Gravity:      0.4,
		JumpForce:    10.0,
		PlayerSpeed:  6.0,
		MaxFallSpeed: 12.0,

		CanDoubleJump: false,
		HasWallJump:   false,
		HasDash:       false,

		ScoreThresholds: []int{0, 100, 200, 300, 400, 500, 600, 700, 800, 900},
		TimeLimit:       []int{0, 0, 60, 60, 55, 55, 50, 50, 45, 40}, // 0 = no limit
	}
}

func (g *GameRules) GetLevelDifficulty() (platformDensity, obstacleDensity, collectibleDensity int) {
	level := g.CurrentLevel - 1
	if level < 0 {
		level = 0
	}
	if level >= g.MaxLevels {
		level = g.MaxLevels - 1
	}

	platformDensity = g.PlatformDensity[level]
	obstacleDensity = g.ObstacleDensity[level]
	collectibleDensity = g.CollectibleDensity[level]

	return
}

func (g *GameRules) UnlockFeatures() {
	if g.CurrentLevel >= 3 {
		g.CanDoubleJump = true
	}

	if g.CurrentLevel >= 5 {
		g.HasWallJump = true
	}

	if g.CurrentLevel >= 7 {
		g.HasDash = true
	}
}
