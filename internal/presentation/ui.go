package presentation

import (
	"rogue-go/internal/domain"

	"github.com/rthornton128/goncurses"
)

// Добавим константы для текстур
const (
	TextureBrick = iota
	TextureStone
	TextureWood
	TextureMarble
	NumTextures
)

// Добавим структуру для текстуры
type Texture struct {
	Symbols []string
	Colors  []int16
}

type GameInterface struct {
	stdscr          *goncurses.Window
	gameWindow      *goncurses.Window
	statusWindow    *goncurses.Window
	messageWindow   *goncurses.Window
	inventoryWindow *goncurses.Window
	showMinimap     bool
	game            *domain.Game
	renderCache     map[string]interface{} // Кэш для рендеринга
}

func NewGameInterface(game *domain.Game) *GameInterface {
	return &GameInterface{game: game}
}

func (gi *GameInterface) Initialize() error {
	stdscr, err := goncurses.Init()
	if err != nil {
		return err
	}

	goncurses.Raw(true)
	goncurses.Echo(false)
	goncurses.Cursor(0)
	stdscr.Keypad(true)

	gi.stdscr = stdscr
	gi.createWindows()
	gi.initColors()

	gi.showMinimap = true
	gi.renderCache = make(map[string]interface{})
	return nil
}

func (gi *GameInterface) Run() {
	defer goncurses.End()

	gi.renderInitialScreen()
	gi.stdscr.Timeout(-1) // Неблокирующий ввод - переделал под блокирующий ввод (для пошагового режима)

	for gi.game.State.IsRunning {
		gi.renderGame()
		gi.handleInput()
		if !gi.game.State.GameOver {
			gi.game.UpdateGameState()
		}
		if gi.game.State.GameOver {
			gi.showGameOver()
			break
		}
	}
}
