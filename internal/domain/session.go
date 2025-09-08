package domain

import (
	"time"

	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/4vertak/rogue-go/internal/domain/gen"
)

type GameSession struct {
	Seed   int64
	Level  entity.Level
	Player entity.Player
	Over   bool
	Log    []string
}

type ScoreEntry struct {
	Name  string
	Gold  int
	Level int
	Time  int64
}

func NewSession() *GameSession {
	s := &GameSession{Seed: time.Now().UnixNano()}
	s.Player = entity.DefaultPlayer()
	// стартовый левел
	s.Level = gen.BuildLevel(gen.RNG(s.Seed), 1, gen.DefaultConfig())

	//ставим игрока в центр первой комнаты
	if len(s.Level.Rooms) > 0 {
		start := s.Level.Rooms[0]
		s.Player.Pos = entity.Pos{
			X: start.X + start.W/2,
			Y: start.Y + start.H/2,
		}
	}

	s.Log = append(s.Log, "Начало игры")
	return s
}

func (s *GameSession) NextLevel() {
	newIndex := s.Level.Index + 1
	s.Level = gen.BuildLevel(gen.RNG(time.Now().UnixNano()), newIndex, gen.DefaultConfig())
	s.Log = append(s.Log, "Спуск на уровень ", string(rune(newIndex)))
}

func (s *GameSession) LogTail(n int) []string {
	if len(s.Log) <= n {
		return s.Log
	}
	return s.Log[len(s.Log)-n:]
}

func (s *GameSession) ToScore() ScoreEntry {
	return ScoreEntry{
		Name:  "Hero",
		Gold:  s.Player.Gold,
		Level: s.Level.Index,
		Time:  time.Now().Unix(),
	}
}
