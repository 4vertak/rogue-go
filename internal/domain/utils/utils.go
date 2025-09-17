package utils

import (
	"rogue-go/internal/domain/entity"
	"strings"
)

func AddMessage(state *entity.GameState, message string) {

	// Основные сообщения (для нижней панели)
	if state.Messages == nil {
		state.Messages = make([]string, 0)
	}

	if len(state.Messages) >= 3 {
		state.Messages = state.Messages[1:]
	}
	state.Messages = append(state.Messages, message)

	// Сообщения о предметах - в статус (правую панель)
	if isItemMessage(message) {
		if state.StatusMessages == nil {
			state.StatusMessages = make([]string, 0)
		}

		if len(state.StatusMessages) >= 5 {
			state.StatusMessages = state.StatusMessages[1:]
		}
		state.StatusMessages = append(state.StatusMessages, message)
	}

}

func isItemMessage(message string) bool {
	itemKeywords := []string{
		"Picked up", "potion", "scroll", "weapon", "food", "treasure", "gold",
	}
	for _, keyword := range itemKeywords {
		if strings.Contains(strings.ToLower(message), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func GetKeyName(color entity.ItemSubType) string {
	switch color {
	case entity.KeyRed:
		return "Red Key(1)"
	case entity.KeyBlue:
		return "Blue Key(2)"
	case entity.KeyGreen:
		return "Green Key(3)"
	case entity.KeyYellow:
		return "Yellow Key(4)"
	default:
		return "Key(0)"
	}
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func GetEnemySymbol(state *entity.GameState, enemy *entity.Enemy) string {

	switch enemy.Type {
	case entity.Zombie:
		return "Z"
	case entity.Vampire:
		return "V"
	case entity.Ghost:
		return "G"
	case entity.Ogre:
		return "O"
	case entity.SnakeMage:
		return "S"
	case entity.Mimic:
		if enemy.IsDisguised {
			itemTypes := []string{"!", "%", "&", "?", "$"}
			if state.CurrentLevel < 5 {
				return itemTypes[0]
			} else if state.CurrentLevel < 9 {
				return itemTypes[1]
			} else if state.CurrentLevel < 13 {
				return itemTypes[2]
			} else if state.CurrentLevel < 17 {
				return itemTypes[3]
			} else {
				return itemTypes[4]
			}
		} else {
			return "M"
		}

	default:
		return "E"
	}
}

func GetEnemyColor(enemy *entity.Enemy) int16 {
	switch enemy.Type {
	case entity.Zombie:
		return 2
	case entity.Vampire:
		return 1
	case entity.Ghost:
		return 7
	case entity.Ogre:
		return 4
	case entity.SnakeMage:
		return 5
	case entity.Mimic:
		return 13 // Используем новую пару
	default:
		return 7
	}
}
