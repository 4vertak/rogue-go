package domain

type Renderer interface {
	Init() error
	Shutdown()
	Draw(state *RenderState)              // кадр: карта, туман, акторы, UI
	PromptChoice(title string, n int) int // запрос 0..n, блокирующий
	Message(text string)
}

type Input interface {
	NextAction() Action // блокирующий: перемещение/использование/выход
}

type Storage interface {
	SaveProgress(*GameSession) error
	LoadProgress() (*GameSession, error)
	AppendScore(ScoreEntry) error
	TopScores(limit int) ([]ScoreEntry, error)
}
