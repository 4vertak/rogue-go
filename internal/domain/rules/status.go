package rules

import (
	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/utils"
	"strconv"
)

type StatusData struct {
	Level       int
	Health      string
	Strength    string
	Agility     string
	Gold        string
	Weapon      string
	Effects     []string
	StatusLines []string
	SleepStatus string
}

func BuildStatus(state *entity.GameState) StatusData {
	var status StatusData
	if state == nil || state.Player == nil {
		return status
	}
	player := state.Player

	status.Level = state.CurrentLevel
	status.Health = strconv.Itoa(player.Health) + "/" + strconv.Itoa(player.MaxHealth)

	if player.CurrentWeapon != nil {
		status.Strength = strconv.Itoa(player.Strength) +
			" (" + strconv.Itoa(player.Strength+player.CurrentWeapon.StrengthBoost) + ")"
	} else {
		status.Strength = strconv.Itoa(player.Strength)
	}

	status.Agility = strconv.Itoa(player.Agility)

	status.Gold = strconv.Itoa(player.Treasures)

	// оружие
	if player.CurrentWeapon != nil {
		status.Weapon = "Weapon: Equipped"
	} else {
		status.Weapon = "Weapon: None"
	}

	// эффект
	for _, eff := range player.ActiveEffects {
		status.Effects = append(status.Effects,
			getEffectName(eff.Type)+" ("+strconv.Itoa(eff.Duration)+")")
	}

	// сбщения
	if len(state.StatusMessages) > 0 {
		startIdx := utils.Max(0, len(state.StatusMessages)-5)
		for i := startIdx; i < len(state.StatusMessages); i++ {
			status.StatusLines = append(status.StatusLines, state.StatusMessages[i])
		}
	}

	// сон
	if state.SleepTurns > 0 {
		status.SleepStatus = "Sleep: " + strconv.Itoa(state.SleepTurns) + " turns"
	}

	return status
}

func getEffectName(effectType entity.EffectType) string {
	switch effectType {
	case entity.HealthEffect:
		return "Health"
	case entity.AgilityEffect:
		return "Agility"
	case entity.StrengthEffect:
		return "Strength"
	case entity.MaxHealthEffect:
		return "Max Health"
	default:
		return "Unknown"
	}
}

type MessageBarData struct {
	LastMessage string
}

func BuildMessageBar(state *entity.GameState) MessageBarData {
	var m MessageBarData
	if state != nil && len(state.Messages) > 0 {
		m.LastMessage = state.Messages[len(state.Messages)-1]
	}
	return m
}

type EndGameData struct {
	Title     string
	Subtitle  string
	Level     int
	Treasures int
	Enemies   int
	Victory   bool
}

func BuildGameOver(state *entity.GameState) EndGameData {
	var endGame EndGameData
	if state == nil || state.Player == nil {
		return endGame
	}
	endGame.Title = "=== GAME OVER ==="
	endGame.Subtitle = "You have been defeated!"
	endGame.Level = state.CurrentLevel
	endGame.Treasures = state.Player.Treasures
	endGame.Enemies = state.Player.DefeatedEnemies
	endGame.Victory = false
	return endGame
}

func BuildVictory(state *entity.GameState) EndGameData {
	var endGame EndGameData
	if state == nil || state.Player == nil {
		return endGame
	}
	endGame.Title = "=== VICTORY! ==="
	endGame.Subtitle = "You have completed the game!"
	endGame.Level = state.CurrentLevel
	endGame.Treasures = state.Player.Treasures
	endGame.Enemies = state.Player.DefeatedEnemies
	endGame.Victory = true
	return endGame
}
