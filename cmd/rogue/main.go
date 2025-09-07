package main

import (
	"log"

	"github.com/4vertak/rogue-go/internal/data/repo"
	"github.com/4vertak/rogue-go/internal/domain"
	"github.com/4vertak/rogue-go/internal/presentation/input"
	"github.com/4vertak/rogue-go/internal/presentation/ncui"
)

func main() {
	rend := ncui.New()
	in := input.New()
	store := repo.NewJSON("save.json", "scores.json")

	if err := rend.Init(); err != nil {
		log.Fatal(err)
	}
	defer rend.Shutdown()

	game := domain.NewGame(store, rend, in) // инверсия зависимостей
	game.Run()                              // блокирующий цикл тиков
}
