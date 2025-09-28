package presentation

import "github.com/rthornton128/goncurses"

func (gi *GameInterface) initColors() {
	goncurses.StartColor()

	goncurses.InitPair(1, goncurses.C_WHITE, goncurses.C_BLACK) // стены блидние
	goncurses.InitPair(2, goncurses.C_CYAN, goncurses.C_BLACK)  // стены средние
	goncurses.InitPair(3, goncurses.C_BLUE, goncurses.C_BLACK)  // стеныдальние
	goncurses.InitPair(4, goncurses.C_WHITE, goncurses.C_BLACK) // пол
	goncurses.InitPair(5, goncurses.C_BLACK, goncurses.C_BLACK) // потолок

	goncurses.InitPair(6, goncurses.C_BLUE, goncurses.C_BLACK)

	goncurses.InitPair(7, goncurses.C_YELLOW, goncurses.C_BLACK)  // оружие
	goncurses.InitPair(8, goncurses.C_CYAN, goncurses.C_BLACK)    // зелье
	goncurses.InitPair(9, goncurses.C_WHITE, goncurses.C_BLACK)   // мимик
	goncurses.InitPair(10, goncurses.C_RED, goncurses.C_BLACK)    // Ключ и дверь
	goncurses.InitPair(11, goncurses.C_BLUE, goncurses.C_BLACK)   //  Ключ и дверь
	goncurses.InitPair(12, goncurses.C_GREEN, goncurses.C_BLACK)  //  Ключ и дверь
	goncurses.InitPair(13, goncurses.C_YELLOW, goncurses.C_BLACK) //  Ключ и дверь
	goncurses.InitPair(14, goncurses.C_WHITE, goncurses.C_BLACK)  // деволтный белый на всякий

	// двери с градиентом
	// Красная дверь (15–17)
	goncurses.InitPair(15, goncurses.C_RED, goncurses.C_BLACK)
	goncurses.InitPair(16, goncurses.C_RED, goncurses.C_BLACK)
	goncurses.InitPair(17, goncurses.C_RED, goncurses.C_BLACK)

	// Синяя дверь (18–20)
	goncurses.InitPair(18, goncurses.C_BLUE, goncurses.C_BLACK)
	goncurses.InitPair(19, goncurses.C_BLUE, goncurses.C_BLACK)
	goncurses.InitPair(20, goncurses.C_BLUE, goncurses.C_BLACK)
	// Зелёная дверь (21–23)
	goncurses.InitPair(21, goncurses.C_GREEN, goncurses.C_BLACK)
	goncurses.InitPair(22, goncurses.C_GREEN, goncurses.C_BLACK)
	goncurses.InitPair(23, goncurses.C_GREEN, goncurses.C_BLACK)

	// Жёлтая дверь (24–26)
	goncurses.InitPair(24, goncurses.C_YELLOW, goncurses.C_BLACK)
	goncurses.InitPair(25, goncurses.C_YELLOW, goncurses.C_BLACK)
	goncurses.InitPair(26, goncurses.C_YELLOW, goncurses.C_BLACK)

	goncurses.InitPair(27, goncurses.C_GREEN, goncurses.C_BLACK)  // Зомби
	goncurses.InitPair(28, goncurses.C_RED, goncurses.C_BLACK)    // Вампир
	goncurses.InitPair(29, goncurses.C_WHITE, goncurses.C_BLACK)  // Привидение
	goncurses.InitPair(30, goncurses.C_YELLOW, goncurses.C_BLACK) // Огр
	goncurses.InitPair(31, goncurses.C_WHITE, goncurses.C_BLACK)  // Змей
}
