package domain

type Game struct {
	r Renderer
	i Input
	s Storage

	state *GameSession
}

func NewGame(s Storage, r Renderer, i Input) *Game {
	return &Game{r: r, i: i, s: s}
}

func (g *Game) Run() {
	// Пытаемся загрузить
	if sess, err := g.s.LoadProgress(); err == nil && sess != nil {
		g.state = sess
	} else {
		g.state = NewSession(1) // уровень 1
	}

	g.r.Message("Добро пожаловать в Rogue-Go")

	for !g.state.Over {
		g.tick() // один ход игрока + ход ИИ
	}
	// По завершению — записываем рекорд
	_ = g.s.AppendScore(g.state.ToScore())
}

func (g *Game) tick() {
	// 1) отрисовка
	g.r.Draw(BuildRenderState(g.state))
	// 2) действие игрока
	act := g.i.NextAction()
	applyPlayerAction(g.state, act) // перемещение / использование / меню
	// 3) перерасчёт ИИ
	runEnemiesAI(g.state)
	// 4) столкновения/бой
	resolveCombats(g.state)
	// 5) проверка: выход с уровня?
	if onExit(g.state) {
		NextLevel(g.state)            // генерируем новый Level
		_ = g.s.SaveProgress(g.state) // автосейв после уровня (ТЗ)
	}
}
