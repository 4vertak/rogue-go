package repo

import (
	"encoding/json"
	"os"
	"sort"

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

	scores := &entity.Statistics{BestScores: make([]*entity.GameSession, 0)}
	if data, err := os.ReadFile(filename); err == nil {
		_ = json.Unmarshal(data, scores)
	}

	if scores.BestScores == nil {
		scores.BestScores = make([]*entity.GameSession, 0)
	}

	scores.BestScores = append(scores.BestScores, stats.BestScores...)

	if len(scores.BestScores) > 1 {

		sort.Slice(scores.BestScores, func(i, j int) bool {
			if scores.BestScores[i].Treasures == scores.BestScores[j].Treasures {
				return scores.BestScores[i].LevelReached > scores.BestScores[j].LevelReached
			}
			return scores.BestScores[i].Treasures > scores.BestScores[j].Treasures
		})
	}

	data, err := json.MarshalIndent(scores, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)

}

// LoadStats загружает статистику из файла
func (r *JSONRepo) LoadStats(filename string) (*entity.Statistics, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return &entity.Statistics{BestScores: make([]*entity.GameSession, 0)}, nil
	}
	var stats entity.Statistics
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}

	if stats.BestScores == nil {
		stats.BestScores = make([]*entity.GameSession, 0)
	}
	return &stats, nil
}
