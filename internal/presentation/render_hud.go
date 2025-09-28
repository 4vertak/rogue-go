package presentation

import (
	"rogue-go/internal/domain/rules"
	"strconv"
)

func (gi *GameInterface) renderStatus() {
	s := rules.BuildStatus(gi.game.State)

	gi.statusWindow.Erase()
	gi.statusWindow.Box(0, 0)

	gi.statusWindow.MovePrint(1, 1, "Level:"+strconv.Itoa(s.Level))
	gi.statusWindow.MovePrint(2, 1, "Health:"+s.Health)
	gi.statusWindow.MovePrint(3, 1, "Str:"+s.Strength)
	gi.statusWindow.MovePrint(4, 1, "Agi:"+s.Agility)
	gi.statusWindow.MovePrint(5, 1, "Gold:"+s.Gold)
	gi.statusWindow.MovePrint(6, 1, s.Weapon)

	row := 8
	for _, eff := range s.Effects {
		gi.statusWindow.MovePrint(row, 1, eff)
		row++
	}

	row = 12
	gi.statusWindow.MovePrint(row, 1, "Recent items:")
	row++
	for _, msg := range s.StatusLines {
		gi.statusWindow.MovePrint(row, 1, msg)
		row++
	}

	if s.SleepStatus != "" {
		gi.statusWindow.MovePrint(15, 1, s.SleepStatus)
	}

	gi.statusWindow.Refresh()
}

func (gi *GameInterface) renderMessageBar() {
	m := rules.BuildMessageBar(gi.game.State)
	gi.messageWindow.Erase()
	gi.messageWindow.Box(0, 0)
	if m.LastMessage != "" {
		gi.messageWindow.MovePrint(1, 1, m.LastMessage)
	}
	gi.messageWindow.Refresh()
}

func (gi *GameInterface) renderInventory() {
	inv := rules.BuildInventory(gi.game.State)

	gi.inventoryWindow.Erase()
	gi.inventoryWindow.Box(0, 0)

	gi.inventoryWindow.MovePrint(1, 1, inv.Title)
	row := 2
	for _, line := range inv.Lines {
		gi.inventoryWindow.MovePrint(row, 1, line)
		row++
	}

	gi.inventoryWindow.Refresh()
}

func (gi *GameInterface) render3DHUD() {
	hud := rules.BuildHUD(gi.game.State)
	maxY, maxX := gi.gameWindow.MaxYX()

	gi.gameWindow.ColorOn(7)
	for x := 0; x < maxX; x++ {
		gi.gameWindow.MovePrint(maxY-3, x, " ")
	}
	gi.gameWindow.ColorOff(7)
	gi.gameWindow.MovePrint(maxY-2, 1, "HP: "+hud.HealthBar)
	gi.gameWindow.MovePrint(maxY-2, 25, hud.HealthText)
	gi.gameWindow.MovePrint(maxY-1, 1, hud.WeaponText)
	gi.gameWindow.MovePrint(maxY-1, maxX-len(hud.Controls)-1, hud.Controls)
}

func (gi *GameInterface) showGameOver() {
	eg := rules.BuildGameOver(gi.game.State)

	gi.stdscr.Erase()
	gi.stdscr.Print(eg.Title + "\n\n")
	gi.stdscr.Print(eg.Subtitle + "\n\n")
	gi.stdscr.Printf("Level reached: %d\n", eg.Level)
	gi.stdscr.Printf("Treasures collected: %d\n", eg.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", eg.Enemies)
	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
}

func (gi *GameInterface) showVictory() {
	eg := rules.BuildVictory(gi.game.State)

	gi.stdscr.Erase()
	gi.stdscr.Print(eg.Title + "\n\n")
	gi.stdscr.Print(eg.Subtitle + "\n\n")
	gi.stdscr.Printf("Final level: %d\n", eg.Level)
	gi.stdscr.Printf("Treasures collected: %d\n", eg.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", eg.Enemies)
	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
	gi.stdscr.GetChar()
}
