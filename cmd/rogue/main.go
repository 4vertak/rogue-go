package main

import (
	"log"

	"github.com/4vertak/rogue-go/internal/datalayer/repo"
	"github.com/4vertak/rogue-go/internal/domain"
	"github.com/4vertak/rogue-go/internal/presentation/tty"
)

func main() {
	r := tty.NewRenderer()
	in := tty.NewInput()
	store := repo.NewJSON("save.json", "scores.json")

	if err := r.Init(); err != nil {
		log.Fatal(err)
	}
	defer r.Shutdown()

	game := domain.NewGame(store, r, in)
	game.Run()
}
