package main

import (
	"github.com/famz/SetLocale"
	"log"

	"rogue-go/internal/data/repo"
	"rogue-go/internal/domain"
	"rogue-go/internal/presentation"
)

func main() {
	SetLocale.SetLocale(SetLocale.LC_ALL, "ru_RU.UTF-8")
	storage := repo.NewJSONRepo()
	game := domain.NewGame(storage)

	ui := presentation.NewGameInterface(game)
	if err := ui.Initialize(); err != nil {
		log.Fatalf("Interface initialization error: %v", err)
	}
	// Загрузка статистики
	_ = game.LoadStats("statistics.json")

	ui.ShowMainMenu()
}
