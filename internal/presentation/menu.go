package presentation

import (
	"encoding/json"
	"os"

	"rogue-go/internal/domain"
	"rogue-go/internal/domain/entity"

	"github.com/rthornton128/goncurses"
)

func (gi *GameInterface) showQuitMenu() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== QUIT GAME ===\n\n")
	gi.stdscr.Print("Save game before quitting? (y/n/c): ")
	gi.stdscr.Refresh()

	gi.stdscr.Timeout(-1)
	char := gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)

	switch char {
	case 'y', 'Y':
		gi.showSaveMenu()
		gi.game.State.IsRunning = false
	case 'n', 'N':
		gi.game.State.IsRunning = false
	}
}

func (gi *GameInterface) showSaveMenu() {
	gi.stdscr.Erase()
	gi.stdscr.Print("Enter filename to save: ")
	gi.stdscr.Refresh()

	filename := gi.getInput()
	if filename != "" {
		err := gi.game.SaveGame(filename)
		if err != nil {
			gi.showMessage("Save error: " + err.Error())
		} else {
			gi.showMessage("Game saved successfully!")
		}
	}
}

func (gi *GameInterface) showLoadMenu() {
	gi.stdscr.Erase()
	gi.stdscr.Print("Enter filename to load: ")
	gi.stdscr.Refresh()

	filename := gi.getInput()
	if filename != "" {
		err := gi.game.LoadGame(filename)
		if err != nil {
			gi.showMessage("Load error: " + err.Error())
		} else {
			gi.showMessage("Game loaded successfully!")
		}
	}
}

func (gi *GameInterface) showUseItemMenu(itemType entity.ItemType) {
	state := gi.game.State
	gi.stdscr.Erase()

	bp := state.Backpack
	if bp == nil {
		gi.showMessage("Backpack empty")
		return
	}

	// Временно переключаемся на блокирующий ввод
	gi.stdscr.Timeout(-1)
	defer gi.stdscr.Timeout(100) // Возвращаем неблокирующий ввод в конце

	switch itemType {
	case entity.Weapon:
		gi.stdscr.Print("=== WEAPONS (0 - unequip) ===\n\n")
		gi.stdscr.Print("0. Unequip current weapon\n")
		if len(bp.Weapons) == 0 {
			gi.stdscr.Print("No weapons in backpack\n")
		} else {
			for i, w := range bp.Weapons {
				gi.stdscr.Printf("%d. Weapon (Str+%d)\n", i+1, w.StrengthBoost)
			}
		}
		gi.stdscr.Print("\nChoose 0-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch == '0' {
			gi.game.UseItem(entity.Weapon, 0)
			return
		}
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '0')
			if idx-1 < len(bp.Weapons) {
				gi.game.UseItem(entity.Weapon, idx)
				return
			}
		}

	case entity.Food:
		if len(bp.Food) == 0 {
			gi.showMessage("No food in backpack")
			return
		}
		gi.stdscr.Print("=== FOOD ===\n\n")
		for i, f := range bp.Food {
			gi.stdscr.Printf("%d. Food (HP+%d)\n", i+1, f.HealthBoost)
		}
		gi.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Food) {
				gi.game.UseItem(entity.Food, idx)
				return
			}
		}

	case entity.Potion:
		if len(bp.Potions) == 0 {
			gi.showMessage("No potions in backpack")
			return
		}
		gi.stdscr.Print("=== POTIONS ===\n\n")
		for i := range bp.Potions {
			gi.stdscr.Printf("%d. Potion\n", i+1)
		}
		gi.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Potions) {
				gi.game.UseItem(entity.Potion, idx)
				return
			}
		}

	case entity.Scroll:
		if len(bp.Scrolls) == 0 {
			gi.showMessage("No scrolls in backpack")
			return
		}
		gi.stdscr.Print("=== SCROLLS ===\n\n")
		for i := range bp.Scrolls {
			gi.stdscr.Printf("%d. Scroll\n", i+1)
		}
		gi.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Scrolls) {
				gi.game.UseItem(entity.Scroll, idx)
				return
			}
		}
	}

	// Если нажата любая другая клавиша - просто возвращаемся без сообщения
}

func (gi *GameInterface) showInventory() {
	state := gi.game.State
	gi.stdscr.Erase()
	gi.stdscr.Print("=== INVENTORY ===\n\n")

	backpack := state.Backpack

	if backpack != nil {
		gi.stdscr.Print("Weapons:\n")
		for i := range backpack.Weapons {
			gi.stdscr.Printf("%d. Weapon (Str+%d)\n", i+1, backpack.Weapons[i].StrengthBoost)
		}

		gi.stdscr.Print("\nFood:\n")
		for i := range backpack.Food {
			gi.stdscr.Printf("%d. Food (HP+%d)\n", i+1, backpack.Food[i].HealthBoost)
		}

		gi.stdscr.Print("\nPotions:\n")
		for i := range backpack.Potions {
			gi.stdscr.Printf("%d. Potion\n", i+1)
		}

		gi.stdscr.Print("\nScrolls and Keys:\n")
		for i, scroll := range backpack.Scrolls {
			if scroll.SubType == entity.KeyRed || scroll.SubType == entity.KeyBlue ||
				scroll.SubType == entity.KeyGreen || scroll.SubType == entity.KeyYellow {
				gi.stdscr.Printf("%d. %s\n", i+1, gi.game.GetKeyName(scroll.SubType))
			} else {
				gi.stdscr.Printf("%d. Scroll\n", i+1)
			}
		}
	}

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) showPickupMenu() {
	state := gi.game.State
	player := state.Player
	var itemsToPick []*entity.Item

	for _, item := range state.Items {
		if item.Position.X == player.Position.X &&
			item.Position.Y == player.Position.Y {
			itemsToPick = append(itemsToPick, item)
		}
	}

	if len(itemsToPick) == 0 {
		gi.game.AddMessage("No items here to pick up")
		return
	}

	if len(itemsToPick) == 1 {
		// Используем существующий метод PickUpItem вместо PickupItem
		gi.game.PickUpItem()
		return
	}

	gi.stdscr.Erase()
	gi.stdscr.Print("=== PICK UP ITEM ===\n\n")

	for i, item := range itemsToPick {
		gi.stdscr.Printf("%d. %s\n", i+1, gi.game.GetItemName(item))
	}

	gi.stdscr.Print("\nSelect item to pick up (number) or any key to cancel: ")
	gi.stdscr.Refresh()

	char := gi.stdscr.GetChar()
	if char >= '1' && char <= '9' {
		index := int(char - '1')
		if index < len(itemsToPick) {
			gi.game.PickUpItem()
		}
	}
}

func (gi *GameInterface) ShowHighScores() {
	state := gi.game.State
	gi.stdscr.Erase()
	gi.stdscr.Print("=== HIGH SCORES ===\n\n")

	if state.Statistics == nil || len(state.Statistics.BestScores) == 0 {
		gi.stdscr.Print("No high scores yet!\n")
	} else {
		// Сортируем результаты по количеству сокровищ (по убыванию)
		sortedScores := make([]*entity.GameSession, len(state.Statistics.BestScores))
		copy(sortedScores, state.Statistics.BestScores)

		for i := 0; i < len(sortedScores)-1; i++ {
			for j := i + 1; j < len(sortedScores); j++ {
				if sortedScores[i].Treasures < sortedScores[j].Treasures {
					sortedScores[i], sortedScores[j] = sortedScores[j], sortedScores[i]
				}
			}
		}

		gi.stdscr.Print("Rank | Level | Treasures | Enemies\n")
		gi.stdscr.Print("-----+-------+-----------+--------\n")

		for i, score := range sortedScores {
			if i >= 10 {
				break
			}
			gi.stdscr.Printf("%4d | %5d | %9d | %7d\n",
				i+1, score.LevelReached, score.Treasures, score.EnemiesDefeated)
		}
	}

	gi.stdscr.Print("\nPress any key to return...")
	gi.stdscr.Refresh()
	// Устанавливаем блокирующий ввод
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100) // Возвращаем неблокирующий ввод
}

func (gi *GameInterface) showMessage(message string) {
	gi.stdscr.Erase()
	gi.stdscr.Print(message)
	gi.stdscr.Print("\n\nPress any key to continue...")
	gi.stdscr.Refresh()

	// Устанавливаем блокирующий ввод
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100) // Возвращаем неблокирующий ввод
}

func (gi *GameInterface) getInput() string {
	// Явно отключаем echo чтобы избежать дублирования
	goncurses.Echo(false)

	gi.stdscr.Timeout(-1)
	defer gi.stdscr.Timeout(100)

	input := ""
	cursorPos := 0

	// Очищаем строку и показываем курсор
	y, x := gi.stdscr.CursorYX()
	gi.stdscr.Print("> ")
	cursorPos = x + 2
	gi.stdscr.Refresh()

	for {
		char := gi.stdscr.GetChar()

		if char == '\n' || char == '\r' {
			break
		} else if char == 127 || char == 8 { // Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
				cursorPos--
				gi.stdscr.Move(y, cursorPos)
				gi.stdscr.Print(" ")
				gi.stdscr.Move(y, cursorPos)
			}
		} else if char >= 32 && char <= 126 {
			input += string(byte(char))
			gi.stdscr.Print(string(byte(char)))
			cursorPos++
		}

		gi.stdscr.Refresh()
	}

	return input
}

func saveStatistics(game *domain.Game) {
	if game.State.Statistics == nil {
		return
	}
	data, err := json.MarshalIndent(game.State.Statistics, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile("statistics.json", data, 0644)
}

func loadStatistics(game *domain.Game) {
	data, err := os.ReadFile("statistics.json")
	if err != nil {
		game.State.Statistics = &entity.Statistics{BestScores: make([]*entity.GameSession, 0)}
		return
	}

	var stats entity.Statistics
	if json.Unmarshal(data, &stats) == nil {
		game.State.Statistics = &stats
	}
}
