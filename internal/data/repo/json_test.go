package repo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"rogue-go/internal/domain/entity"
)

func TestGameSaveRoundTrip(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "session.json")
	want := &entity.GameState{
		Player: &entity.Character{
			Health: 73, MaxHealth: 100, Position: entity.Position{X: 3, Y: 4},
			ActiveEffects: []*entity.Effect{{Type: entity.StrengthEffect, Duration: 5, Value: 2}},
		},
		Backpack: &entity.Backpack{
			Scrolls: []*entity.Item{{Type: entity.Scroll, SubType: entity.KeyBlue, Name: "Синий ключ"}},
		},
		Level: &entity.Level{
			Grid:    [][]entity.Cell{{{IsWall: true}, {IsDiscovered: true}}},
			Visible: [][]bool{{false, true}},
			Doors:   []*entity.Door{{Position: entity.Position{X: 1, Y: 2}, Color: entity.KeyBlue, IsLocked: true}},
		},
		GameSession:    &entity.GameSession{CellsWalked: 17, Treasures: 12},
		PlayerProgress: &entity.ProgressTracker{DifficultyLevel: 3},
		RenderMode:     entity.Render3D,
		PlayerDir:      entity.PositionF{X: 0, Y: 1},
		PlayerPlane:    entity.PositionF{X: -0.66, Y: 0},
		CurrentLevel:   4,
		IsRunning:      true,
		Messages:       []string{"Сохранение игры"},
	}
	r := NewJSONRepo()
	if err := r.SaveGame(filename, want); err != nil {
		t.Fatal(err)
	}
	got, err := r.LoadGame(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("saved game did not survive JSON round trip: got %+v, want %+v", got, want)
	}
}

func TestLoadGameErrors(t *testing.T) {
	r := NewJSONRepo()
	dir := t.TempDir()
	if _, err := r.LoadGame(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("loading a missing save must return an error")
	}
	filename := filepath.Join(dir, "broken.json")
	if err := os.WriteFile(filename, []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := r.LoadGame(filename); err == nil {
		t.Fatal("loading malformed JSON must return an error")
	}
}

func TestStatsAppendAndRanking(t *testing.T) {
	r := NewJSONRepo()
	filename := filepath.Join(t.TempDir(), "statistics.json")
	empty, err := r.LoadStats(filename)
	if err != nil || empty == nil || len(empty.BestScores) != 0 {
		t.Fatalf("missing statistics should start empty: stats=%+v, err=%v", empty, err)
	}
	first := &entity.GameSession{Treasures: 10, LevelReached: 2}
	second := &entity.GameSession{Treasures: 10, LevelReached: 5}
	third := &entity.GameSession{Treasures: 20, LevelReached: 1}
	for _, session := range []*entity.GameSession{first, second, third} {
		if err := r.SaveStats(filename, &entity.Statistics{BestScores: []*entity.GameSession{session}}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := r.LoadStats(filename)
	if err != nil {
		t.Fatal(err)
	}
	want := []*entity.GameSession{third, second, first}
	if !reflect.DeepEqual(got.BestScores, want) {
		t.Fatalf("scores should be retained and ranked by treasure, then level: got %+v", got.BestScores)
	}
}
