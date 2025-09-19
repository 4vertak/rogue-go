package presentation

import (
	"rogue-go/internal/domain/entity"

	"github.com/rthornton128/goncurses"
)

func (ui *GameInterface) showQuitMenu() {

	title := []string{
		"           QUIT  MENU           ",
		"+------------------------------+",
		"|                              |",
	}

	footer := []string{
		"|                              |",
		"+------------------------------+",
	}

	menuItems := []string{
		"SAVE & EXIT",
		"EXIT GAME",
		"CANCEL",
	}

	currentLine := 0

	for {
		choice := ui.patternRenderMenu(title, footer, menuItems, currentLine)

		switch choice {
		case 0:
			ui.showSaveMenu()
			ui.game.State.IsRunning = false
			return
		case 1:
			ui.game.State.IsRunning = false
			return
		case 2:
			return
		}
	}
}

func (ui *GameInterface) showSaveMenu() {
	ui.stdscr.Erase()
	ui.stdscr.Print("Enter filename to save: ")
	ui.stdscr.Refresh()

	filename := ui.getInput()
	if filename != "" {
		err := ui.game.SaveGame(filename)
		if err != nil {
			ui.showMessage("Save error: " + err.Error())
		} else {
			ui.showMessage("Game saved successfully!")
		}
	}
}

func (ui *GameInterface) showLoadMenu() {
	ui.stdscr.Erase()
	ui.stdscr.Print("Enter filename to load: ")
	ui.stdscr.Refresh()

	filename := ui.getInput()
	if filename != "" {
		err := ui.game.LoadGame(filename)
		if err != nil {
			ui.showMessage("Load error: " + err.Error())
		} else {
			ui.showMessage("Game loaded successfully!")
		}
	}
}

func (ui *GameInterface) showUseItemMenu(itemType entity.ItemType) {
	state := ui.game.State
	ui.stdscr.Erase()

	bp := state.Backpack
	if bp == nil {
		ui.showMessage("Backpack empty")
		return
	}

	// Временно переключаемся на блокирующий ввод
	//	ui.stdscr.Timeout(-1)
	//	defer ui.stdscr.Timeout(100) // Возвращаем неблокирующий ввод в конце

	switch itemType {
	case entity.Weapon:
		ui.stdscr.Print("=== WEAPONS (0 - unequip) ===\n\n")
		ui.stdscr.Print("0. Unequip current weapon\n")
		if len(bp.Weapons) == 0 {
			ui.stdscr.Print("No weapons in backpack\n")
		} else {
			for i, w := range bp.Weapons {
				ui.stdscr.Printf("%d. Weapon (Str+%d)\n", i+1, w.StrengthBoost)
			}
		}
		ui.stdscr.Print("\nChoose 0-9 or any other key to cancel: ")
		ui.stdscr.Refresh()

		ch := ui.stdscr.GetChar()
		if ch == '0' {
			ui.game.UseItem(entity.Weapon, 0)
			return
		}
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '0')
			if idx-1 < len(bp.Weapons) {
				ui.game.UseItem(entity.Weapon, idx)
				return
			}
		}

	case entity.Food:
		if len(bp.Food) == 0 {
			ui.showMessage("No food in backpack")
			return
		}
		ui.stdscr.Print("=== FOOD ===\n\n")
		for i, f := range bp.Food {
			ui.stdscr.Printf("%d. Food (HP+%d)\n", i+1, f.HealthBoost)
		}
		ui.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		ui.stdscr.Refresh()

		ch := ui.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Food) {
				ui.game.UseItem(entity.Food, idx)
				return
			}
		}

	case entity.Potion:
		if len(bp.Potions) == 0 {
			ui.showMessage("No potions in backpack")
			return
		}
		ui.stdscr.Print("=== POTIONS ===\n\n")
		for i := range bp.Potions {
			ui.stdscr.Printf("%d. Potion\n", i+1)
		}
		ui.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		ui.stdscr.Refresh()

		ch := ui.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Potions) {
				ui.game.UseItem(entity.Potion, idx)
				return
			}
		}

	case entity.Scroll:
		if len(bp.Scrolls) == 0 {
			ui.showMessage("No scrolls in backpack")
			return
		}
		ui.stdscr.Print("=== SCROLLS & KEYS ===\n\n")
		for i, scroll := range bp.Scrolls {
			// ОТОБРАЖАЕМ НАЗВАНИЯ КЛЮЧЕВ вместо "Scroll"
			if scroll.SubType == entity.KeyRed || scroll.SubType == entity.KeyBlue ||
				scroll.SubType == entity.KeyGreen || scroll.SubType == entity.KeyYellow {
				ui.stdscr.Printf("%d. %s\n", i+1, ui.game.GetKeyName(scroll.SubType))
			} else {
				ui.stdscr.Printf("%d. Scroll\n", i+1)
			}
		}
		ui.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		ui.stdscr.Refresh()

		ch := ui.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Scrolls) {
				ui.game.UseItem(entity.Scroll, idx)
				return
			}
		}
	}

	// Если нажата любая другая клавиша - просто возвращаемся без сообщения
}

func (ui *GameInterface) showInventory() {
	state := ui.game.State
	ui.stdscr.Erase()
	ui.stdscr.Print("=== INVENTORY ===\n\n")

	backpack := state.Backpack

	if backpack != nil {
		ui.stdscr.Print("Weapons:\n")
		for i := range backpack.Weapons {
			ui.stdscr.Printf("%d. Weapon (Str+%d)\n", i+1, backpack.Weapons[i].StrengthBoost)
		}

		ui.stdscr.Print("\nFood:\n")
		for i := range backpack.Food {
			ui.stdscr.Printf("%d. Food (HP+%d)\n", i+1, backpack.Food[i].HealthBoost)
		}

		ui.stdscr.Print("\nPotions:\n")
		for i := range backpack.Potions {
			ui.stdscr.Printf("%d. Potion\n", i+1)
		}

		ui.stdscr.Print("\nScrolls and Keys:\n")
		for i, scroll := range backpack.Scrolls {
			if scroll.SubType == entity.KeyRed || scroll.SubType == entity.KeyBlue ||
				scroll.SubType == entity.KeyGreen || scroll.SubType == entity.KeyYellow {
				ui.stdscr.Printf("%d. %s\n", i+1, ui.game.GetKeyName(scroll.SubType))
			} else {
				ui.stdscr.Printf("%d. Scroll\n", i+1)
			}
		}
	}

	ui.stdscr.Print("\nPress any key to continue...")
	ui.stdscr.Refresh()
	//	ui.stdscr.Timeout(-1)
	ui.stdscr.GetChar()
	// ui.stdscr.Timeout(100)
}

func (ui *GameInterface) showPickupMenu() {
	state := ui.game.State
	player := state.Player
	var itemsToPick []*entity.Item

	for _, item := range state.Items {
		if item.Position.X == player.Position.X &&
			item.Position.Y == player.Position.Y {
			itemsToPick = append(itemsToPick, item)
		}
	}

	if len(itemsToPick) == 0 {
		ui.game.AddMessage("No items here to pick up")
		return
	}

	if len(itemsToPick) == 1 {
		// Используем существующий метод PickUpItem вместо PickupItem
		ui.game.PickUpItem()
		return
	}

	ui.stdscr.Erase()
	ui.stdscr.Print("=== PICK UP ITEM ===\n\n")

	for i, item := range itemsToPick {
		ui.stdscr.Printf("%d. %s\n", i+1, ui.game.GetItemName(item))
	}

	ui.stdscr.Print("\nSelect item to pick up (number) or any key to cancel: ")
	ui.stdscr.Refresh()

	char := ui.stdscr.GetChar()
	if char >= '1' && char <= '9' {
		index := int(char - '1')
		if index < len(itemsToPick) {
			ui.game.PickUpItem()
		}
	}
}

func (ui *GameInterface) ShowHighScores() {
	_ = ui.game.LoadStats("statistics.json") // сохарнение было - надо загружать думаю собираемся показывать))
	state := ui.game.State
	ui.stdscr.Erase()
	ui.stdscr.Print("=== HIGH SCORES ===\n\n")

	if state.Statistics == nil || len(state.Statistics.BestScores) == 0 {
		ui.stdscr.Print("No high scores yet!\n")
	} else {
		// Сортируем результаты по количеству сокровищ (по убыванию)
		scores := state.Statistics.BestScores

		ui.stdscr.Print("Rank | Level | Treasures | Enemies\n")
		ui.stdscr.Print("-----+-------+-----------+--------\n")

		for i, score := range scores {
			if i >= 10 {
				break
			}
			ui.stdscr.Printf("%4d | %5d | %9d | %7d\n",
				i+1, score.LevelReached, score.Treasures, score.EnemiesDefeated)
		}
	}

	ui.stdscr.Print("\nPress any key to return...")
	ui.stdscr.Refresh()
	ui.stdscr.GetChar()
}

func (ui *GameInterface) showMessage(message string) {
	ui.stdscr.Erase()
	ui.stdscr.Print(message)
	ui.stdscr.Print("\n\nPress any key to continue...")
	ui.stdscr.Refresh()

	// Устанавливаем блокирующий ввод
	ui.stdscr.Timeout(-1)
	ui.stdscr.GetChar()
	// ui.stdscr.Timeout(100) // Возвращаем неблокирующий ввод
}

func (ui *GameInterface) getInput() string {
	// Явно отключаем echo чтобы избежать дублирования
	goncurses.Echo(false)

	//	ui.stdscr.Timeout(-1)
	//	defer ui.stdscr.Timeout(100)

	input := ""
	cursorPos := 0

	// Очищаем строку и показываем курсор
	y, x := ui.stdscr.CursorYX()
	ui.stdscr.Print("> ")
	cursorPos = x + 2
	ui.stdscr.Refresh()

	for {
		char := ui.stdscr.GetChar()

		if char == '\n' || char == '\r' {
			break
		} else if char == 127 || char == 8 { // Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
				cursorPos--
				ui.stdscr.Move(y, cursorPos)
				ui.stdscr.Print(" ")
				ui.stdscr.Move(y, cursorPos)
			}
		} else if char >= 32 && char <= 126 {
			input += string(byte(char))
			ui.stdscr.Print(string(byte(char)))
			cursorPos++
		}

		ui.stdscr.Refresh()
	}

	return input
}
