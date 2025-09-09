package main

import (
	"encoding/json"
	"os"

	"github.com/rthornton128/goncurses"
)

func main() {
	// Инициализация состояния игры
	initializeGame()
	
	// Инициализация интерфейса
	ui := &GameInterface{}
	if err := ui.Initialize(); err != nil {
		panic("Interface initialization error: " + err.Error())
	}

	// Загрузка статистики
	loadStatistics()

	// Запуск главного меню
	showMainMenu(ui)
}

func initializeGame() {
	// Создаем начальное состояние игры
	CurrentGameState = &GameState{
		IsRunning: true,
		Player: &Character{
			Health:          100,
			MaxHealth:       100,
			Agility:         10,
			Strength:        10,
			Position:        Position{X: 0, Y: 0},
			Treasures:       0,
			DefeatedEnemies: 0,
			ActiveEffects:   make([]*Effect, 0),
		},
		Backpack: &Backpack{
			Weapons:  make([]*Item, 0),
			Food:     make([]*Item, 0),
			Potions:  make([]*Item, 0),
			Scrolls:  make([]*Item, 0),
			Treasure: 0,
		},
		Enemies:     make([]*Enemy, 0),
		Items:       make([]*Item, 0),
		GameSession: &GameSession{},
		Statistics: &Statistics{
			BestScores: make([]*GameSession, 0),
		},
		Camera: &Camera{
			Width:  80,
			Height: 24,
		},
		CurrentLevel: 1,
		GameWon:      false,
		GameOver:     false,
	}

	// Генерация первого уровня
	GameActionsImpl.GenerateLevel(1)
}

func showMainMenu(ui *GameInterface) {
	ui.stdscr.Print("=== YES ===\n")
	for {
		ui.stdscr.Clear()
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
			startNewGame(ui)
		case '2':
			loadGame(ui)
		case '3':
			showHighScores(ui)
		case '4':
			goncurses.End()
			os.Exit(0)
		}
	}
}

func startNewGame(ui *GameInterface) {
	// Переинициализируем игру
	initializeGame()
	
	// Запускаем игровой цикл
	ui.Run()
	
	// После завершения игры добавляем результат в таблицу рекордов
	if CurrentGameState.GameOver || CurrentGameState.GameWon {
		addToHighScores()
	}
	
	// Возвращаемся в главное меню
	showMainMenu(ui)
}

func loadGame(ui *GameInterface) {
	ui.stdscr.Clear()
	ui.stdscr.Print("Enter filename to load: ")
	ui.stdscr.Refresh()

	filename := ui.getInput()
	if filename != "" {
		err := GameActionsImpl.LoadGame(filename)
		if err != nil {
			ui.showMessage("Load error: " + err.Error())
		} else {
			ui.showMessage("Game loaded successfully!")
			ui.Run()
			
			if CurrentGameState.GameOver || CurrentGameState.GameWon {
				addToHighScores()
			}
		}
	}
}

func showHighScores(ui *GameInterface) {
	ui.stdscr.Clear()
	ui.stdscr.Print("=== HIGH SCORES ===\n\n")

	if CurrentGameState.Statistics == nil || len(CurrentGameState.Statistics.BestScores) == 0 {
		ui.stdscr.Print("No high scores yet!\n")
	} else {
		// Сортируем результаты по количеству сокровищ (по убыванию)
		sortedScores := make([]*GameSession, len(CurrentGameState.Statistics.BestScores))
		copy(sortedScores, CurrentGameState.Statistics.BestScores)
		
		for i := 0; i < len(sortedScores)-1; i++ {
			for j := i + 1; j < len(sortedScores); j++ {
				if sortedScores[i].Treasures < sortedScores[j].Treasures {
					sortedScores[i], sortedScores[j] = sortedScores[j], sortedScores[i]
				}
			}
		}

		ui.stdscr.Print("Rank | Level | Treasures | Enemies\n")
		ui.stdscr.Print("-----+-------+-----------+--------\n")
		
		for i, score := range sortedScores {
			if i >= 10 { // Показываем топ-10
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

func addToHighScores() {
	if CurrentGameState.Statistics == nil {
		CurrentGameState.Statistics = &Statistics{
			BestScores: make([]*GameSession, 0),
		}
	}

	// Создаем копию текущей игровой сессии для таблицы рекордов
	score := &GameSession{
		LevelReached:    CurrentGameState.CurrentLevel,
		Treasures:       CurrentGameState.Player.Treasures,
		EnemiesDefeated: CurrentGameState.Player.DefeatedEnemies,
		FoodConsumed:    CurrentGameState.GameSession.FoodConsumed,
		PotionsUsed:     CurrentGameState.GameSession.PotionsUsed,
		ScrollsUsed:     CurrentGameState.GameSession.ScrollsUsed,
		HitsLanded:      CurrentGameState.GameSession.HitsLanded,
		HitsTaken:       CurrentGameState.GameSession.HitsTaken,
		CellsWalked:     CurrentGameState.GameSession.CellsWalked,
	}

	// Добавляем в таблицу рекордов
	CurrentGameState.Statistics.BestScores = append(CurrentGameState.Statistics.BestScores, score)

	// Сохраняем статистику в файл
	saveStatistics()
}

func saveStatistics() {
	// Сохраняем статистику в файл
	if CurrentGameState.Statistics == nil {
		return
	}
	
	data, err := json.MarshalIndent(CurrentGameState.Statistics, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile("statistics.json", data, 0644)
}

func loadStatistics() {
	// Загружаем статистику из файла
	data, err := os.ReadFile("statistics.json")
	if err != nil {
		// Если файла нет, создаем пустую статистику
		CurrentGameState.Statistics = &Statistics{
			BestScores: make([]*GameSession, 0),
		}
		return
	}

	var stats Statistics
	err = json.Unmarshal(data, &stats)
	if err == nil {
		CurrentGameState.Statistics = &stats
	}
}