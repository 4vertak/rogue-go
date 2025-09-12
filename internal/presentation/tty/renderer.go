package tty

import (
	"fmt"
	"strings"

	"github.com/4vertak/rogue-go/internal/datalayer/repo"
	"github.com/4vertak/rogue-go/internal/domain"
	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/rthornton128/goncurses"
)

type Renderer struct {
	stdscr *goncurses.Window
	log    []string
}

func NewRenderer(stdscr *goncurses.Window) *Renderer {
	return &Renderer{stdscr: stdscr}
}

func (r *Renderer) Init() error {
	return nil
}

func (r *Renderer) Shutdown() {
	goncurses.End()
}

func (r *Renderer) Draw(rs *domain.RenderState) {
	r.stdscr.Clear()
	lvl := rs.Level

	maxY, maxX := r.stdscr.MaxYX()
	shiftX := (maxX - lvl.W) / 2
	shiftY := (maxY - lvl.H) / 2

	for y := 0; y < lvl.H; y++ {
		for x := 0; x < lvl.W; x++ {
			pos := entity.Pos{X: x, Y: y}
			var ch rune

			// Проверяем, видимость
			if rs.Visible[pos] {
				// Видимые тйлы отобража
				ch = tileChar(lvl.Tiles[y][x])

				// Отображаем обьекты только если они видимы
				for _, mob := range lvl.Mobs {
					if mob.Pos.X == x && mob.Pos.Y == y {
						ch = mob.Symbol
						break
					}
				}
				for _, item := range lvl.Items {
					if item.Pos.X == x && item.Pos.Y == y {
						ch = '*'
						break
					}
				}
				if rs.Player.Pos.X == x && rs.Player.Pos.Y == y {
					ch = '@'
				}
			} else if lvl.Explored[y][x] {
				// Прошли, внп зоны видимости будут затемненными потом цветом будум зтменять
				switch lvl.Tiles[y][x] {
				case entity.Wall:
					ch = '#'
				case entity.Floor:
					ch = '.'
				case entity.Exit:
					ch = '>'
				case entity.Door:
					ch = '+'
				default:
					ch = ' '
				}
			} else {
				// пусто
				ch = ' '
			}

			r.stdscr.MoveAddChar(shiftY+y, shiftX+x, goncurses.Char(ch))
		}
	}

	statsLine := shiftY + lvl.H + 1
	statsText := fmt.Sprintf("HP %d/%d  STR %d  DEX %d  GOLD %d  LVL %d",
		rs.Player.Stats.HP, rs.Player.Stats.MaxHP, rs.Player.Stats.STR, rs.Player.Stats.DEX, rs.Player.Gold, rs.Level.Index)
	r.stdscr.MovePrint(statsLine, shiftX, statsText)

	r.stdscr.Refresh()
}

func (r *Renderer) PromptChoice(title string, n int) int {
	maxY, maxX := r.stdscr.MaxYX()

	menuHeight := n + 4
	menuWidth := len(title) + 10
	shiftY := (maxY - menuHeight) / 2
	shiftX := (maxX - menuWidth) / 2

	r.stdscr.MovePrint(shiftY, shiftX, "+"+strings.Repeat("-", menuWidth-2)+"+")
	for i := 1; i <= menuHeight-2; i++ {
		r.stdscr.MovePrint(shiftY+i, shiftX, "|")
		r.stdscr.MovePrint(shiftY+i, shiftX+menuWidth-1, "|")
	}
	r.stdscr.MovePrint(shiftY+menuHeight-1, shiftX, "+"+strings.Repeat("-", menuWidth-2)+"+")

	r.stdscr.MovePrint(shiftY+1, shiftX+2, title)

	for i := 0; i < n; i++ {
		optionText := fmt.Sprintf("%d. Option %d", i+1, i+1)
		r.stdscr.MovePrint(shiftY+3+i, shiftX+2, optionText)
	}

	r.stdscr.Refresh()

	for {
		key := r.stdscr.GetChar()

		if key >= goncurses.Key('1') && key <= goncurses.Key('0')+goncurses.Key(n) {
			return int(key - goncurses.Key('0'))
		}
	}
}

func (r *Renderer) Message(text string) {
	r.log = append(r.log, text)
	if len(r.log) > 5 {
		r.log = r.log[1:]
	}
}

func tileChar(t entity.Tile) rune {
	switch t {
	case entity.Wall:
		return '#'
	case entity.Floor:
		return '.'
	case entity.Exit:
		return '>'
	case entity.Door:
		return '+'
	default:
		return ' '
	}
}

func (r *Renderer) StartScreen() {
	strings := []string{
		"          _____                   _______                   _____                    _____                    _____          ",
		"         /\\    \\                 /::\\    \\                 /\\    \\                  /\\    \\                  /\\    \\         ",
		"        /::\\    \\               /::::\\    \\               /::\\    \\                /::\\____\\                /::\\    \\        ",
		"       /::::\\    \\             /::::::\\    \\             /::::\\    \\              /:::/    /               /::::\\    \\       ",
		"      /::::::\\    \\           /::::::::\\    \\           /::::::\\    \\            /:::/    /               /::::::\\    \\      ",
		"     /:::/\\:::\\    \\         /:::/~~\\:::\\    \\         /:::/\\:::\\    \\          /:::/    /               /:::/\\:::\\    \\     ",
		"    /:::/__\\:::\\    \\       /:::/    \\:::\\    \\       /:::/  \\:::\\    \\        /:::/    /               /:::/__\\:::\\    \\    ",
		"   /::::\\   \\:::\\    \\     /:::/    / \\:::\\    \\     /:::/    \\:::\\    \\      /:::/    /               /::::\\   \\:::\\    \\   ",
		"  /::::::\\   \\:::\\    \\   /:::/____/   \\:::\\____\\   /:::/    / \\:::\\    \\    /:::/    /      _____    /::::::\\   \\:::\\    \\  ",
		" /:::/\\:::\\   \\:::\\____\\ |:::|    |     |:::|    | /:::/    /   \\:::\\ ___\\  /:::/____/      /\\    \\  /:::/\\:::\\   \\:::\\    \\ ",
		"/:::/  \\:::\\   \\:::|    ||:::|____|     |:::|    |/:::/____/  ___\\:::|    ||:::|    /      /::\\____\\/:::/__\\:::\\   \\:::\\____\\",
		"\\::/   |::::\\  /:::|____| \\:::\\    \\   /:::/    / \\:::\\    \\ /\\  /:::|____||:::|____\\     /:::/    /\\:::\\   \\:::\\   \\::/    /",
		" \\/____|:::::\\/:::/    /   \\:::\\    \\ /:::/    /   \\:::\\    /::\\ \\::/    /  \\:::\\    \\   /:::/    /  \\:::\\   \\:::\\   \\/____/ ",
		"       |:::::::::/    /     \\:::\\    /:::/    /     \\:::\\   \\:::\\ \\/____/    \\:::\\    \\ /:::/    /    \\:::\\   \\:::\\    \\     ",
		"       |::|\\::::/    /       \\:::\\__/:::/    /       \\:::\\   \\:::\\____\\       \\:::\\    /:::/    /      \\:::\\   \\:::\\____\\    ",
		"       |::| \\::/____/         \\::::::::/    /         \\:::\\  /:::/    /        \\:::\\__/:::/    /        \\:::\\   \\::/    /    ",
		"       |::|  ~|                \\::::::/    /           \\:::\\/:::/    /          \\::::::::/    /          \\:::\\   \\/____/     ",
		"       |::|   |                 \\::::/    /             \\::::::/    /            \\::::::/    /            \\:::\\    \\         ",
		"       \\::|   |                  \\::/____/               \\::::/    /              \\::::/    /              \\:::\\____\\        ",
		"        \\:|   |                   ~~                      \\::/____/                \\::/____/                \\::/    /        ",
		"         \\|___|                                                                     ~~                       \\/____/         ",
	}

	maxY, maxX := r.stdscr.MaxYX()
	height := len(strings)
	width := len(strings[0])

	shiftX := (maxX - width) / 2
	shiftY := (maxY - height) / 2

	r.stdscr.Erase()

	for i, s := range strings {
		r.stdscr.MovePrint(shiftY+i, shiftX, s)
	}

	continueText := "Press any key to continue..."
	shiftX = (maxX - len(continueText)) / 2
	r.stdscr.MovePrint(shiftY+height+1, shiftX, continueText)

	r.stdscr.Refresh()
	r.stdscr.GetChar()
	r.stdscr.Erase()
}

// меню
func (r *Renderer) MenuScreen(currentLine int, store *repo.JSONRepo) int {
	menuItems := []string{
		"NEW GAME",
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

	maxY, maxX := r.stdscr.MaxYX()
	height := len(title) + len(menuItems) + len(footer)
	width := len(title[0])

	shiftX := (maxX - width) / 2
	shiftY := (maxY - height) / 2

	r.stdscr.Erase()

	for i, line := range title {
		r.stdscr.MovePrint(shiftY+i, shiftX, line)
	}

	for i, item := range menuItems {
		line := fmt.Sprintf("|          %-13s       |", item)
		r.stdscr.MovePrint(shiftY+len(title)+i, shiftX, line)

		if i == currentLine {
			r.stdscr.MovePrint(shiftY+len(title)+i, shiftX+5, "<<<")
			r.stdscr.MovePrint(shiftY+len(title)+i, shiftX+width-8, ">>>")
		}
	}

	for i, line := range footer {
		r.stdscr.MovePrint(shiftY+len(title)+len(menuItems)+i, shiftX, line)
	}

	r.stdscr.Refresh()

	for {
		key := r.stdscr.GetChar()
		switch key {
		case goncurses.KEY_UP:
			if currentLine > 0 {
				currentLine--
			} else {
				currentLine = len(menuItems) - 1
			}
			return r.MenuScreen(currentLine, store)
		case goncurses.KEY_DOWN:
			if currentLine < len(menuItems)-1 {
				currentLine++
			} else {
				currentLine = 0
			}
			return r.MenuScreen(currentLine, store)
		case goncurses.KEY_ENTER, '\n', '\r':

			if currentLine == 2 {
				scores, err := store.TopScores(10)
				if err != nil {
					r.Message("Error loading scores: " + err.Error())
				} else {
					r.DisplayScoreboard(scores)
				}

				return r.MenuScreen(currentLine, store)
			}
			return currentLine
		case goncurses.KEY_ESC:
			return len(menuItems) - 1
		}
	}
}

// экран смерти
func (r *Renderer) DeadScreen() {
	strings := []string{
		"                  ___           ___                                  ___           ___                   ",
		"                 /\\  \\         /\\  \\                  _____         /\\__\\         /\\  \\         _____    ",
		"      ___       /::\\  \\        \\:\\  \\                /::\\  \\       /:/ _/_       /::\\  \\       /::\\  \\   ",
		"     /|  |     /:/\\:\\  \\        \\:\\  \\              /:/\\:\\  \\     /:/ /\\__\\     /:/\\:\\  \\     /:/\\:\\  \\  ",
		"    |:|  |    /:/  \\:\\  \\   ___  \\:\\  \\            /:/  \\:\\__\\   /:/ /:/ _/_   /:/ /::\\  \\   /:/  \\:\\__\\ ",
		"    |:|  |   /:/__/ \\:\\__\\ /\\  \\  \\:\\__\\          /:/__/ \\:|__| /:/_/:/ /\\__\\ /:/_/:/\\:\\__\\ /:/__/ \\:|__|",
		"  __|:|__|   \\:\\  \\ /:/  / \\:\\  \\ /:/  /          \\:\\  \\ /:/  / \\:\\/:/ /:/  / \\:\\/:/  \\/__/ \\:\\  \\ /:/  /",
		" /::::\\  \\    \\:\\  /:/  /   \\:\\  /:/  /            \\:\\  /:/  /   \\::/_/:/  /   \\::/__/       \\:\\  /:/  / ",
		" ~~~~\\:\\  \\    \\:\\/:/  /     \\:\\/:/  /              \\:\\/:/  /     \\:\\/:/  /     \\:\\  \\        \\:\\/:/  /  ",
		"      \\:\\__\\    \\::/  /       \\::/  /                \\::/  /       \\::/  /       \\:\\__\\        \\::/  /   ",
		"       \\/__/     \\/__/         \\/__/                  \\/__/         \\/__/         \\/__/         \\/__/    ",
	}

	maxY, maxX := r.stdscr.MaxYX()
	height := len(strings)
	width := len(strings[0])

	shiftX := (maxX - width) / 2
	shiftY := (maxY - height) / 2

	r.stdscr.Erase()

	for i, s := range strings {
		r.stdscr.MovePrint(shiftY+i, shiftX, s)
	}

	continueText := "Press any key to continue..."
	shiftX = (maxX - len(continueText)) / 2
	r.stdscr.MovePrint(shiftY+height+1, shiftX, continueText)

	r.stdscr.Refresh()
	r.stdscr.GetChar()
	r.stdscr.Erase()
}

// завершения игры
func (r *Renderer) EndgameScreen() {
	strings := []string{
		"      ___           ___                         ___           ___           ___           ___     ",
		"     /\\__\\         /\\  \\         _____         /\\__\\         /\\  \\         /\\  \\         /\\__\\    ",
		"    /:/ _/_        \\:\\  \\       /::\\  \\       /:/ _/_       /::\\  \\       |::\\  \\       /:/ _/_   ",
		"   /:/ /\\__\\        \\:\\  \\     /:/\\:\\  \\     /:/ /\\  \\     /:/\\:\\  \\      |:|:\\  \\     /:/ /\\__\\  ",
		"  /:/ /:/ _/_   _____\\:\\  \\   /:/  \\:\\__\\   /:/ /::\\  \\   /:/ /::\\  \\   __|:|\\:\\  \\   /:/ /:/ _/_ ",
		" /:/_/:/ /\\__\\ /::::::::\\__\\ /:/__/ \\:|__| /:/__\\/\\:\\__\\ /:/_/:/\\:\\__\\ /::::|_\\:\\__\\ /:/_/:/ /\\__\\",
		" \\:\\/:/ /:/  / \\:\\~~\\~~\\/__/ \\:\\  \\ /:/  / \\:\\  \\ /:/  / \\:\\/:/  \\/__/ \\:\\~~\\  \\/__/ \\:\\/:/ /:/  /",
		"  \\::/_/:/  /   \\:\\  \\        \\:\\  /:/  /   \\:\\  /:/  /   \\::/__/       \\:\\  \\        \\::/_/:/  / ",
		"   \\:\\/:/  /     \\:\\  \\        \\:\\/:/  /     \\:\\/:/  /     \\:\\  \\        \\:\\  \\        \\:\\/:/  /  ",
		"    \\::/  /       \\:\\__\\        \\::/  /       \\::/  /       \\:\\__\\        \\:\\__\\        \\::/  /   ",
		"     \\/__/         \\/__/         \\/__/         \\/__/         \\/__/         \\/__/         \\/__/    ",
	}

	maxY, maxX := r.stdscr.MaxYX()
	height := len(strings)
	width := len(strings[0])

	shiftX := (maxX - width) / 2
	shiftY := (maxY - height) / 2

	r.stdscr.Erase()

	for i, s := range strings {
		r.stdscr.MovePrint(shiftY+i, shiftX, s)
	}

	continueText := "Press any key to continue..."
	shiftX = (maxX - len(continueText)) / 2
	r.stdscr.MovePrint(shiftY+height+1, shiftX, continueText)

	r.stdscr.Refresh()
	r.stdscr.GetChar()
	r.stdscr.Erase()
}

// Таблицу рекордов
func (r *Renderer) DisplayScoreboard(scores []domain.ScoreEntry) {
	maxY, maxX := r.stdscr.MaxYX()

	sizeArray := len(scores)
	if sizeArray > 10 {
		sizeArray = 10
	}

	nameWidth := 20
	goldWidth := 10
	levelWidth := 10
	timeWidth := 15

	totalWidth := nameWidth + goldWidth + levelWidth + timeWidth + 5

	shiftX := (maxX - totalWidth) / 2
	shiftY := (maxY - (sizeArray + 4)) / 2

	r.stdscr.Erase()

	title := "HIGH SCORES"
	r.stdscr.MovePrint(shiftY, (maxX-len(title))/2, title)

	topBorder := "+" + strings.Repeat("-", nameWidth) + "+" +
		strings.Repeat("-", goldWidth) + "+" +
		strings.Repeat("-", levelWidth) + "+" +
		strings.Repeat("-", timeWidth) + "+"
	r.stdscr.MovePrint(shiftY+1, shiftX, topBorder)

	header := fmt.Sprintf("|%-*s|%-*s|%-*s|%-*s|",
		nameWidth, "NAME",
		goldWidth, "GOLD",
		levelWidth, "LEVEL",
		timeWidth, "TIME")
	r.stdscr.MovePrint(shiftY+2, shiftX, header)

	r.stdscr.MovePrint(shiftY+3, shiftX, topBorder)

	for i := 0; i < sizeArray; i++ {

		minutes := scores[i].Time / 60
		seconds := scores[i].Time % 60
		timeFormatted := fmt.Sprintf("%02d:%02d", minutes, seconds)

		row := fmt.Sprintf("|%-*s|%*d|%*d|%*s|",
			nameWidth, scores[i].Name,
			goldWidth, scores[i].Gold,
			levelWidth, scores[i].Level,
			timeWidth, timeFormatted)
		r.stdscr.MovePrint(shiftY+4+i, shiftX, row)
	}

	r.stdscr.MovePrint(shiftY+4+sizeArray, shiftX, topBorder)

	exitText := "Press ESC to exit."
	r.stdscr.MovePrint(shiftY+5+sizeArray, (maxX-len(exitText))/2, exitText)

	r.stdscr.Refresh()

	for {
		key := r.stdscr.GetChar()
		if key == goncurses.KEY_ESC {
			break
		}
	}

	r.stdscr.Erase()
}
