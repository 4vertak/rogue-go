package domain

import (
	"errors"
	"rogue-go/internal/domain/entity"
)

// Renderer отвечает за рисование и диалоги
type Renderer interface {
	Init() error
	Close()
	RenderFrame(state *entity.GameState)
	ShowMessage(msg string)
	AskString(prompt string) (string, error)
}

// Input отвечает за поток ввода (key events) — в простом варианте можно
// встроить в Renderer, но лучше выделить для тестов.
type Input interface {
	Poll() (rune, error)
}

// Storage — интерфейс для сохранений/загрузок игры и статистики
type Storage interface {
	SaveGame(filename string, state *entity.GameState) error
	LoadGame(filename string) (*entity.GameState, error)
	SaveStats(filename string, stats *entity.Statistics) error
	LoadStats(filename string) (*entity.Statistics, error)
}

var ErrCancel = errors.New("cancel")
