package presentation

import (
	"os"

	"rogue-go/internal/domain"
	"rogue-go/internal/domain/entity"

	"github.com/rthornton128/goncurses"
)

func ShowMainMenu(ui *GameInterface, game *domain.Game) {

	for {
		ui.stdscr.Erase()
		ui.stdscr.Print("=== GO ROGUELIKE ===\n\n")
		ui.stdscr.Print("1. New Game\n")
		ui.stdscr.Print("2. Load Game\n")
		ui.stdscr.Print("3. High Scores\n")
		ui.stdscr.Print("4. Exit\n\n")
		ui.stdscr.Print("Select option: ")
		ui.stdscr.Refresh()

		char := ui.stdscr.GetChar()
		switch char {
		case '1':
			startNewGame(ui, game)
		case '2':
			loadGameMenu(ui, game)
		case '3':
			ui.ShowHighScores()
		case '4':
			goncurses.End()
			os.Exit(0)
		}
	}
}

func startNewGame(ui *GameInterface, game *domain.Game) {
	state := game.State
	game.InitializeNewGame()
	ui.Run()

	// После завершения игры добавляем результат в таблицу рекордов
	if state.GameOver || state.GameWon {
		addToHighScores(game)
	}

	// Возвращаемся в главное меню
	ShowMainMenu(ui, game)
}

func loadGameMenu(ui *GameInterface, game *domain.Game) {
	state := game.State
	ui.stdscr.Erase()
	ui.stdscr.Print("Enter filename to load: ")
	ui.stdscr.Refresh()

	filename := ui.getInput()
	if filename != "" {
		err := game.LoadGame(filename)
		if err != nil {
			ui.showMessage("Load error: " + err.Error())
		} else {
			ui.showMessage("Game loaded successfully!")
			ui.Run()
			if state.GameOver || state.GameWon {
				addToHighScores(game)
			}
		}
	}
}

func addToHighScores(game *domain.Game) {
	state := game.State
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

	_ = game.SaveStats("statistics.json")
}
