package ai

import (
	"fmt"
	"time"

	"github.com/h3bzzz/go-chess/core/game"
)

type AIManager struct {
	ai             *ChessAI
	gameState      *game.GameState
	enabled        bool
	aiColor        int
	ticker         *time.Ticker
	stopChan       chan bool
	onMoveCallback func()
	aiDifficulty   int
}

func NewAIManager(gameState *game.GameState) *AIManager {
	return &AIManager{
		gameState: gameState,
		enabled:   false,
		aiColor:   game.BlackPlayer,
		stopChan:  make(chan bool),
	}
}

func (m *AIManager) SetEnabled(enabled bool) {
	if m.enabled == enabled {
		return
	}

	m.enabled = enabled
	if enabled {
		m.Start()
	} else {
		m.Stop()
	}
}

func (m *AIManager) SetAIColor(color int) {
	m.aiColor = color
	m.rebuildAI()
}

func (m *AIManager) SetDifficulty(difficulty int) {
	m.aiDifficulty = difficulty
	m.rebuildAI()
}

func (m *AIManager) rebuildAI() {
	playerColor := game.BlackPlayer
	if m.aiColor == game.BlackPlayer {
		playerColor = game.WhitePlayer
	}
	m.ai = NewChessAI(m.gameState, playerColor, 2)

	// Set the difficulty if it was previously configured
	if m.ai != nil && m.aiDifficulty > 0 {
		m.ai.difficulty = m.aiDifficulty
	}
}

func (m *AIManager) SetMoveCallback(callback func()) {
	m.onMoveCallback = callback
}

func (m *AIManager) Start() {
	if m.ticker != nil {
		return
	}

	m.rebuildAI()
	m.ticker = time.NewTicker(700 * time.Millisecond)

	go func() {
		for {
			select {
			case <-m.ticker.C:
				if m.enabled && m.gameState.CurrentTurn == m.aiColor && !m.gameState.IsGameOver() {
					fmt.Println("AI turn detected, making move...")
					moveMade := m.ai.MakeMove()
					if moveMade {
						fmt.Println("AI move completed")
						if m.onMoveCallback != nil {
							time.Sleep(300 * time.Millisecond)
							m.onMoveCallback()
						}
					}
				}
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *AIManager) Stop() {
	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
		m.stopChan <- true
	}
}

func (m *AIManager) IsEnabled() bool {
	return m.enabled
}

func (m *AIManager) GetAIColor() int {
	return m.aiColor
}
