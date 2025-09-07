package repo

import (
	"encoding/json"
	"os"
	"sort"
	"sync"

	"github.com/4vertak/rogue-go/internal/domain"
)

type JSONRepo struct {
	savePath   string
	scoresPath string
	mu         sync.Mutex
}

func NewJSON(save, scores string) *JSONRepo { return &JSONRepo{savePath: save, scoresPath: scores} }

func (r *JSONRepo) SaveProgress(s *domain.GameSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := os.Create(r.savePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(s)
}

func (r *JSONRepo) LoadProgress() (*domain.GameSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := os.Open(r.savePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var s domain.GameSession
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *JSONRepo) AppendScore(se domain.ScoreEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	list, _ := r.TopScores(1<<31 - 1)
	list = append(list, se)
	sort.Slice(list, func(i, j int) bool { return list[i].Gold > list[j].Gold })
	f, err := os.Create(r.scoresPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(list)
}

func (r *JSONRepo) TopScores(limit int) ([]domain.ScoreEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := os.Open(r.scoresPath)
	if err != nil {
		return []domain.ScoreEntry{}, nil
	}
	defer f.Close()
	var list []domain.ScoreEntry
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return nil, err
	}
	if limit < len(list) {
		list = list[:limit]
	}
	return list, nil
}
