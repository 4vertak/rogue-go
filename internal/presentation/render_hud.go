package presentation

import (
	"fmt"
	"rogue-go/internal/domain/entity"
	"strconv"
)

func (gi *GameInterface) renderStatus() {
	state := gi.game.State
	gi.statusWindow.Erase()
	gi.statusWindow.Box(0, 0)

	player := state.Player
	gi.statusWindow.MovePrint(1, 1, "Level:"+strconv.Itoa(state.CurrentLevel))
	gi.statusWindow.MovePrint(2, 1, "Health:"+strconv.Itoa(player.Health)+"/"+strconv.Itoa(player.MaxHealth))
	gi.statusWindow.MovePrint(3, 1, "Str:"+strconv.Itoa(player.Strength))
	if player.CurrentWeapon != nil {
		gi.statusWindow.MovePrint(3, 10, "("+strconv.Itoa(player.Strength+player.CurrentWeapon.StrengthBoost)+")")
	}
	gi.statusWindow.MovePrint(4, 1, "Agi:"+strconv.Itoa(player.Agility))
	gi.statusWindow.MovePrint(5, 1, "Gold:"+strconv.Itoa(player.Treasures))

	if player.CurrentWeapon != nil {
		gi.statusWindow.MovePrint(6, 1, "Weapon: Equipped")
	} else {
		gi.statusWindow.MovePrint(6, 1, "Weapon: None")
	}

	// Отображаем активные эффекты
	gi.renderEffects()

	// Отображаем сообщения о предметах в статусе
	gi.renderStatusMessages()

	// Отображаем статус сна
	if state.SleepTurns > 0 {
		gi.statusWindow.MovePrint(15, 1, "Sleep: "+strconv.Itoa(state.SleepTurns)+" turns")
	}

	gi.statusWindow.Refresh()
}

func (gi *GameInterface) renderStatusMessages() {
	state := gi.game.State

	if state.StatusMessages != nil && len(state.StatusMessages) > 0 {
		startRow := 12
		maxY, _ := gi.statusWindow.MaxYX() // Добавляем получение maxY

		gi.statusWindow.MovePrint(startRow, 1, "Recent items:")
		startRow++

		startIdx := max(0, len(state.StatusMessages)-5)
		for i := startIdx; i < len(state.StatusMessages) && startRow < maxY-2; i++ {
			gi.statusWindow.MovePrint(startRow, 1, state.StatusMessages[i])
			startRow++
		}
	}
}

func (gi *GameInterface) renderEffects() {
	state := gi.game.State
	player := state.Player
	if player.ActiveEffects != nil && len(player.ActiveEffects) > 0 {
		maxY, _ := gi.statusWindow.MaxYX()
		startRow := 8

		gi.statusWindow.MovePrint(startRow, 1, "Effects:")
		startRow++

		for _, effect := range player.ActiveEffects {
			if startRow >= maxY-2 {
				break
			}

			effectName := getEffectName(effect.Type)
			effectText := effectName + " (" + strconv.Itoa(effect.Duration) + ")"
			gi.statusWindow.MovePrint(startRow, 1, effectText)
			startRow++
		}
	}
}

func (gi *GameInterface) renderMessageBar() {
	state := gi.game.State
	gi.messageWindow.Erase()
	gi.messageWindow.Box(0, 0)

	if state.Messages != nil && len(state.Messages) > 0 {
		lastMessage := state.Messages[len(state.Messages)-1]
		gi.messageWindow.MovePrint(1, 1, lastMessage)
	}

	gi.messageWindow.Refresh()
}

func (gi *GameInterface) renderInventory() {
	state := gi.game.State
	gi.inventoryWindow.Erase()
	gi.inventoryWindow.Box(0, 0)
	gi.inventoryWindow.MovePrint(1, 1, "Inventory:")

	backpack := state.Backpack
	if backpack != nil {
		weaponCount := len(backpack.Weapons)
		foodCount := len(backpack.Food)
		potionCount := len(backpack.Potions)
		scrollCount := len(backpack.Scrolls)

		gi.inventoryWindow.MovePrint(2, 1, "W:"+strconv.Itoa(weaponCount))
		gi.inventoryWindow.MovePrint(2, 5, "F:"+strconv.Itoa(foodCount))
		gi.inventoryWindow.MovePrint(2, 9, "P:"+strconv.Itoa(potionCount))
		gi.inventoryWindow.MovePrint(2, 13, "S:"+strconv.Itoa(scrollCount))
	}

	gi.inventoryWindow.Refresh()
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

// Рендерим HUD поверх 3D сцены
func (gi *GameInterface) render3DHUD() {
	state := gi.game.State
	maxY, maxX := gi.gameWindow.MaxYX()

	// Полупрозрачный фон для HUD
	gi.gameWindow.ColorOn(7)
	for x := 0; x < maxX; x++ {
		gi.gameWindow.MovePrint(maxY-3, x, " ")
	}
	gi.gameWindow.ColorOff(7)

	// Информация о здоровье
	healthPercent := float64(state.Player.Health) / float64(state.Player.MaxHealth)
	healthBar := "["
	barWidth := 20
	filled := int(healthPercent * float64(barWidth))

	for i := 0; i < barWidth; i++ {
		if i < filled {
			healthBar += "="
		} else {
			healthBar += " "
		}
	}
	healthBar += "]"

	gi.gameWindow.MovePrint(maxY-2, 1, "HP: "+healthBar)
	gi.gameWindow.MovePrint(maxY-2, 25,
		fmt.Sprintf("%d/%d", state.Player.Health, state.Player.MaxHealth))

	// Текущее оружие
	if state.Player.CurrentWeapon != nil {
		gi.gameWindow.MovePrint(maxY-1, 1,
			"Weapon: "+gi.game.GetItemName(state.Player.CurrentWeapon))
	} else {
		gi.gameWindow.MovePrint(maxY-1, 1, "Weapon: None")
	}

	// Подсказки управления
	controls := "WASD:Move  Mouse:Look  E:Inventory  Space:Interact  T:2D/3D"
	gi.gameWindow.MovePrint(maxY-1, maxX-len(controls)-1, controls)
}

func (gi *GameInterface) showGameOver() {
	state := gi.game.State
	gi.stdscr.Erase()
	gi.stdscr.Print("=== GAME OVER ===\n\n")
	gi.stdscr.Print("You have been defeated!\n\n")

	player := state.Player
	gi.stdscr.Printf("Level reached: %d\n", state.CurrentLevel)
	gi.stdscr.Printf("Treasures collected: %d\n", player.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", player.DefeatedEnemies)

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
}

func (gi *GameInterface) showVictory() {
	state := gi.game.State
	gi.stdscr.Erase()
	gi.stdscr.Print("=== VICTORY! ===\n\n")
	gi.stdscr.Print("You have completed the game!\n\n")

	player := state.Player
	gi.stdscr.Printf("Final level: %d\n", state.CurrentLevel)
	gi.stdscr.Printf("Treasures collected: %d\n", player.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", player.DefeatedEnemies)

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
	gi.stdscr.GetChar()
}
