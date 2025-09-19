package domain

import (
	"log"
	"math/rand"
	"time"

	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/gen"
	"rogue-go/internal/domain/rules"
	"rogue-go/internal/domain/utils"
)

type Game struct {
	State   *entity.GameState
	storage Storage
}

func NewGame(storage Storage) *Game {
	rand.Seed(time.Now().UnixNano())
	return &Game{
		State:   &entity.GameState{},
		storage: storage,
	}
}

func (g *Game) InitializeNewGame() {
	scoresStats := g.State.Statistics
	g.State = &entity.GameState{
		IsRunning: true,
		Player: &entity.Character{
			Health:          100,
			MaxHealth:       100,
			Agility:         10,
			Strength:        10,
			Position:        entity.Position{X: 0, Y: 0},
			Treasures:       0,
			DefeatedEnemies: 0,
			ActiveEffects:   make([]*entity.Effect, 0),
		},
		Backpack: &entity.Backpack{
			Weapons:  make([]*entity.Item, 0),
			Food:     make([]*entity.Item, 0),
			Potions:  make([]*entity.Item, 0),
			Scrolls:  make([]*entity.Item, 0),
			Treasure: 0,
		},
		Enemies:     make([]*entity.Enemy, 0),
		Items:       make([]*entity.Item, 0),
		GameSession: &entity.GameSession{},
		Statistics:  scoresStats,
		Camera: &entity.Camera{
			Width:  80,
			Height: 24,
		},
		PlayerProgress: &entity.ProgressTracker{ // ДОБАВЛЯЕМ ТРЕКЕР ПРОГРЕССА
			DifficultyLevel: 3, // Средняя сложность
		},
		CurrentLevel:   1,
		GameWon:        false,
		GameOver:       false,
		Messages:       make([]string, 0),
		StatusMessages: make([]string, 0),
		SleepTurns:     0,
		RenderMode:     entity.Render2D,
		PlayerAngle:    0,
		PlayerDir:      entity.PositionF{X: 1, Y: 0},
		PlayerPlane:    entity.PositionF{X: 0, Y: 0.66},
	}
	g.GenerateLevel(1)
	g.AddMessage("Welcome to Go Rogue! Move with WASD")
}

func (g *Game) MovePlayer(dx, dy int) { rules.MovePlayer(g.State, dx, dy) }

func (g *Game) MoveIn3D(distance int) { rules.MoveIn3D(g.State, distance) }

func (g *Game) Strafe(distance int) { rules.Strafe(g.State, distance) }

func (g *Game) Rotate(angle float64) {
	rules.Rotate(g.State, angle)
}

func (g *Game) InteractWithDoor() { rules.InteractWithDoor(g.State) }

func (g *Game) UseItem(t entity.ItemType, index int) { rules.UseItem(g.State, t, index) }

func (g *Game) GetItemName(t *entity.Item) string { return rules.GetItemName(t) }

func (g *Game) GetItemSymbol(t *entity.Item) string { return rules.GetItemSymbol(t) }

func (g *Game) GetItemColor(t *entity.Item) int16 { return rules.GetItemColor(t) }

func (g *Game) GetKeySymbol(t entity.ItemSubType) string { return rules.GetKeySymbol(t) }

func (g *Game) GetKeyColor(t entity.ItemSubType) int16 { return rules.GetKeyColor(t) }

func (g *Game) GetEnemySymbol(t *entity.Enemy) string { return utils.GetEnemySymbol(g.State, t) }

func (g *Game) GetEnemyColor(t *entity.Enemy) int16 { return utils.GetEnemyColor(t) }

func (g *Game) PickUpItem() { rules.PickUpItem(g.State) }

func (g *Game) Attack(enemyIndex int) { rules.Attack(g.State, enemyIndex) }

func (g *Game) GenerateLevel(levelNum int) {
	level := gen.GenerateLevel(levelNum, g.State)
	g.State.Level = level
	g.State.CurrentLevel = levelNum
	if g.State.Level != nil && g.State.Level.StartRoom != nil {
		r := g.State.Level.StartRoom
		g.State.Player.Position = entity.Position{X: r.X + r.Width/2, Y: r.Y + r.Height/2}
	}

}

func (g *Game) UpdateGameState() {
	rules.UpdateVisibility(g.State)
	rules.MoveEnemies(g.State)
	rules.CheckAutomaticExit(g.State, func() {
		g.GenerateLevel(g.State.CurrentLevel + 1)
		// автосохраненение -при каждом достижении уровня то автосейв
		if err := g.SaveGame("last_session.json"); err != nil {
			log.Println("Error autosave:", err)
		}
	})
	rules.CheckExitProximity(g.State)
	rules.AdjustDifficulty(g.State)
	rules.UpdateEffects(g.State)
	g.CheckGameConditions()
	rules.UpdateCamera(g.State)
}

func (g *Game) CheckGameConditions() {
	if g.State.Player != nil && g.State.Player.Health <= 0 {
		g.State.GameOver = true
		g.AddMessage("You have died! Game Over.")
	}
}

func (g *Game) AddMessage(msg string) {
	utils.AddMessage(g.State, msg)
}

func (g *Game) SaveGame(filename string) error {
	if g.storage == nil {
		return nil
	}
	return g.storage.SaveGame(filename, g.State)
}

func (g *Game) LoadGame(filename string) error {
	if g.storage == nil {
		return nil
	}
	state, err := g.storage.LoadGame(filename)
	if err != nil {
		return err
	}
	g.State = state
	return nil
}

func (g *Game) SaveStats(filename string) error {
	if g.storage == nil {
		return nil
	}
	return g.storage.SaveStats(filename, g.State.Statistics)
}

func (g *Game) LoadStats(filename string) error {
	if g.storage == nil {
		return nil
	}
	stats, err := g.storage.LoadStats(filename)
	if err != nil {
		return err
	}
	g.State.Statistics = stats
	return nil
}

func (g *Game) GetKeyName(color entity.ItemSubType) string {
	return utils.GetKeyName(color)
}

func (g *Game) GetDoorSymbol(door *entity.Door) string {
	return rules.GetDoorSymbol(door)
}

func (g *Game) GetDoorColor(door *entity.Door) int16 {
	return rules.GetDoorColor(door)
}

func (g *Game) InteractWithNearest() {
	rules.InteractWithNearest(g.State)
}

func (g *Game) GetNearestInteractable() (interface{}, float64) {
	return rules.GetNearestInteractable(g.State)
}
