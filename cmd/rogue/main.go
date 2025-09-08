package main

import (
	"log"

	"github.com/4vertak/rogue-go/internal/datalayer/repo"
	"github.com/4vertak/rogue-go/internal/domain"
	"github.com/4vertak/rogue-go/internal/presentation/tty"
	"github.com/rthornton128/goncurses"
)

func main() {

	stdscr, err := goncurses.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer goncurses.End()

	goncurses.Raw(true)
	goncurses.Echo(false)
	goncurses.Cursor(0)
	stdscr.Keypad(true)

	if goncurses.HasColors() {
		goncurses.StartColor()
		goncurses.InitPair(1, goncurses.C_WHITE, goncurses.C_BLACK)
		goncurses.InitPair(2, goncurses.C_RED, goncurses.C_BLACK)
		goncurses.InitPair(3, goncurses.C_GREEN, goncurses.C_BLACK)
		goncurses.InitPair(4, goncurses.C_BLUE, goncurses.C_BLACK)
		goncurses.InitPair(5, goncurses.C_YELLOW, goncurses.C_BLACK)
		goncurses.InitPair(6, goncurses.C_CYAN, goncurses.C_BLACK)
	}

	renderer := tty.NewRenderer(stdscr)
	in := tty.NewInput(stdscr)
	store := repo.NewJSON("save.json", "scores.json")

	renderer.StartScreen()

	currentLine := 0
	for {
		choice := renderer.MenuScreen(currentLine, store)
		switch choice {
		case 0: // новая игра
			game := domain.NewGame(store, renderer, in)
			game.Run()
		case 1: // загрузка игры
			// TODO: реализовать загрузку сохранения
			renderer.Message("Загрузка игры пока не реализована")

		case 2: // Рекорды
			scores, err := store.TopScores(10)
			if err != nil {
				renderer.Message("Error loading scores: " + err.Error())
			} else {
				renderer.DisplayScoreboard(scores)
			}
		case 3:
			return
		}

		currentLine = choice
	}
}
