package main

import (
	"log"

	"rogue-go/internal/data/repo"
	"rogue-go/internal/domain"
	"rogue-go/internal/presentation"
)

func main() {
	storage := repo.NewJSONRepo()
	game := domain.NewGame(storage)

	ui := presentation.NewGameInterface(game)
	if err := ui.Initialize(); err != nil {
		log.Fatalf("Interface initialization error: %v", err)
	}
	// Загрузка статистики
	_ = game.LoadStats("statistics.json")

	presentation.ShowMainMenu(ui, game)
}
