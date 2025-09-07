package domain

type Renderer interface {
	Init() error
	Shutdown()
	Draw(state *RenderState)
	PromptChoice(title string, n int) int
	Message(text string)
}

type Input interface {
	NextAction() Action
}

type Storage interface {
	SaveProgress(*GameSession) error
	LoadProgress() (*GameSession, error)
	AppendScore(ScoreEntry) error
	TopScores(limit int) ([]ScoreEntry, error)
}
