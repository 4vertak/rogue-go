package presentation

import (
	"fmt"
	"os"
	"rogue-go/internal/domain/entity"

	"github.com/rthornton128/goncurses"
)

func (ui *GameInterface) ShowMainMenu() {
	ui.StartScreen()
	ui.MenuScreen()
}

// шаблон отрисовки меню
func (ui *GameInterface) patternRenderMenu(title []string, footer []string, items []string, currentLine int) int {
	maxY, maxX := ui.stdscr.MaxYX()
	height := len(title) + len(items) + len(footer)
	width := len(title[0])

	shiftX := (maxX - width) / 2
	shiftY := (maxY - height) / 2

	ui.stdscr.Erase()

	for i, line := range title {
		ui.stdscr.MovePrint(shiftY+i, shiftX, line)
	}

	for i, item := range items {
		line := fmt.Sprintf("|          %-13s       |", item)
		ui.stdscr.MovePrint(shiftY+len(title)+i, shiftX, line)

		if i == currentLine {
			ui.stdscr.MovePrint(shiftY+len(title)+i, shiftX+5, "<<<")
			ui.stdscr.MovePrint(shiftY+len(title)+i, shiftX+width-8, ">>>")
		}
	}

	for i, line := range footer {
		ui.stdscr.MovePrint(shiftY+len(title)+len(items)+i, shiftX, line)
	}

	ui.stdscr.Refresh()

	for {
		key := ui.stdscr.GetChar()
		switch key {
		case goncurses.KEY_UP:
			if currentLine > 0 {
				currentLine--
			} else {
				currentLine = len(items) - 1
			}
			return ui.patternRenderMenu(title, footer, items, currentLine)

		case goncurses.KEY_DOWN:
			if currentLine < len(items)-1 {
				currentLine++
			} else {
				currentLine = 0
			}
			return ui.patternRenderMenu(title, footer, items, currentLine)

		case goncurses.KEY_ENTER, '\n', '\r':
			return currentLine
		}
	}
}

func (ui *GameInterface) MenuScreen() int {
	menuItems := []string{
		"NEW GAME",
		"CONTINUE",
		"LOAD GAME",
		"SCOREBOARD",
		"EXIT GAME",
	}

	title := []string{
		"           GAME  MENU           ",
		"+------------------------------+",
		"|                              |",
	}

	footer := []string{
		"|                              |",
		"+------------------------------+",
	}
	currentLine := 0

	for {
		choice := ui.patternRenderMenu(title, footer, menuItems, currentLine)
		switch choice {
		case 0:
			ui.StartNewGame()
		case 1:
			ui.LoadlastGameMenu()
		case 2:
			ui.LoadGameMenu()
		case 3:
			ui.ShowHighScores()
		case 4:
			goncurses.End()
			os.Exit(0)
		}
	}
}

func (ui *GameInterface) StartNewGame() {
	ui.game.InitializeNewGame()
	ui.Run()

	// После завершения игры добавляем результат в таблицу рекордов
	if ui.game.State.GameOver || ui.game.State.GameWon {
		ui.addToHighScores()
	}

	// Возвращаемся в главное меню
	ui.ShowMainMenu()
}

func (ui *GameInterface) LoadlastGameMenu() {
	ui.stdscr.Erase()
	if err := ui.game.LoadGame("last_session.json"); err != nil {
		ui.showMessage("No last session found.\n")
	} else {
		ui.showMessage("Last game loaded successfully!")
		ui.Run()
	}

}

func (ui *GameInterface) LoadGameMenu() {
	// state := game.State
	ui.stdscr.Erase()
	ui.stdscr.Print("Enter filename to load: ")
	ui.stdscr.Refresh()

	filename := ui.getInput()
	if filename != "" {
		err := ui.game.LoadGame(filename)
		if err != nil {
			ui.showMessage("Load error: " + err.Error())
		} else {
			_ = ui.game.LoadStats("statistics.json")
			ui.showMessage("Game loaded successfully!")
			ui.Run()
			if ui.game.State.GameOver || ui.game.State.GameWon {
				ui.addToHighScores()
			}
		}
	}
}

func (ui *GameInterface) addToHighScores() {
	state := ui.game.State
	if state.Statistics == nil {
		state.Statistics = &entity.Statistics{
			BestScores: make([]*entity.GameSession, 0),
		}
		return
	}
	// Создаем копию текущей игровой сессии для таблицы рекордов
	score := &entity.GameSession{
		LevelReached:    state.CurrentLevel,
		Treasures:       state.Player.Treasures,
		EnemiesDefeated: state.Player.DefeatedEnemies,
		FoodConsumed:    state.GameSession.FoodConsumed,
		PotionsUsed:     state.GameSession.PotionsUsed,
		ScrollsUsed:     state.GameSession.ScrollsUsed,
		HitsLanded:      state.GameSession.HitsLanded,
		HitsTaken:       state.GameSession.HitsTaken,
		CellsWalked:     state.GameSession.CellsWalked,
	}
	// Добавляем в таблицу рекордов
	state.Statistics.BestScores = append(state.Statistics.BestScores, score)

	// Сохраняем статистику в файл
	_ = ui.game.SaveStats("statistics.json")

}
