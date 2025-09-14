package repo

import (
	"encoding/json"
	"os"

	"rogue-go/internal/domain/entity"
)

// JSONRepo — адаптер для сохранений и статистики в JSON.
type JSONRepo struct{}

func NewJSONRepo() *JSONRepo {
	return &JSONRepo{}
}

// SaveGame сохраняет GameState в файл
func (r *JSONRepo) SaveGame(filename string, state *entity.GameState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadGame загружает GameState из файла
func (r *JSONRepo) LoadGame(filename string) (*entity.GameState, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var state entity.GameState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveStats сохраняет статистику в файл
func (r *JSONRepo) SaveStats(filename string, stats *entity.Statistics) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadStats загружает статистику из файла
func (r *JSONRepo) LoadStats(filename string) (*entity.Statistics, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// если файла нет — возвращаем пустую статистику
		return &entity.Statistics{BestScores: make([]*entity.GameSession, 0)}, nil
	}
	var stats entity.Statistics
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}
