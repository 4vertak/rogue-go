package main

// GameState представляет все состояние игры
type GameState struct {
	// Игровые сущности
	Player      *Character
	Enemies     []*Enemy
	Items       []*Item
	Level       *Level
	Backpack    *Backpack
	GameSession *GameSession
	Statistics  *Statistics
	Camera      *Camera
	PlayerProgress *ProgressTracker // ДОБАВЛЯЕМ ТРЕКЕР ПРОГРЕССА
	
	// Флаги состояния
	IsRunning    bool
	CurrentLevel int
	GameWon      bool
	GameOver     bool
	SleepTurns   int
	Messages     []string
	StatusMessages []string
}

// Добавим ProgressTracker в GameState
type ProgressTracker struct {
    LevelsCompleted    int
    EnemiesDefeated    int
    DamageTaken        int
    DamageDealt        int
    ItemsUsed          int
    DifficultyLevel    int
    LastAdjustmentTime int
}

// Character представляет игрока
type Character struct {
	Health          int
	MaxHealth       int
	Agility         int
	Strength        int
	CurrentWeapon   *Item
	Position        Position
	Treasures       int
	DefeatedEnemies int
	ActiveEffects   []*Effect
}

// Enemy представляет противника
type Enemy struct {
	Type        EnemyType
	Health      int
	Agility     int
	Strength    int
	Hostility   int
	Position    Position
	IsVisible   bool
	Pattern     MovementPattern
	moveCooldown int // Для паттерна TwoStep
	hasBeenHit bool // Для вампиров
	attackCooldown int // Для огров после атаки
	teleportCooldown int // Для призраков
	invisibilityCooldown int // Для призраков
	IsDisguised bool
}

// Item представляет предмет
type Item struct {
	Type           ItemType
	SubType        ItemSubType
	HealthBoost    int
	MaxHealthBoost int
	AgilityBoost   int
	StrengthBoost  int
	Value          int
	Position       Position
	Name		   string
}

// Level представляет игровой уровень
type Level struct {
	Rooms      []*Room
	Corridors  []*Corridor
	StartRoom  *Room
	ExitRoom   *Room
	Grid       [][]Cell
	Discovered [][]bool
	Visible    [][]bool
	Doors      []*Door  // ДОБАВЛЯЕМ ДВЕРИ
}

// Room представляет комнату
type Room struct {
	X, Y, Width, Height int
	Enemies             []*Enemy
	Items               []*Item
	IsStart             bool
	IsExit              bool
}

// Corridor представляет коридор
type Corridor struct {
	Start, End Position
	Width      int
}

// Backpack представляет рюкзак
type Backpack struct {
	Weapons  []*Item
	Food     []*Item
	Potions  []*Item
	Scrolls  []*Item
	Treasure int
}

// GameSession представляет игровую сессию
type GameSession struct {
	LevelReached    int
	Treasures       int
	EnemiesDefeated int
	FoodConsumed    int
	PotionsUsed     int
	ScrollsUsed     int
	HitsLanded      int
	HitsTaken       int
	CellsWalked     int
}

// Statistics представляет статистику всех игр
type Statistics struct {
	BestScores []*GameSession
}

// Effect представляет временный эффект
type Effect struct {
	Type        EffectType
	Duration    int
	Value       int
	Description string
}

// Position представляет координаты
type Position struct {
	X, Y int
}

// Cell представляет клетку карты
type Cell struct {
	IsWall      bool
	IsCorridor  bool
	IsDiscovered bool
	IsVisible   bool
}

// Camera представляет камеру для отображения карты
type Camera struct {
	X, Y int
	Width, Height int
}

// EnemyType представляет тип противника
type EnemyType int
const (
	Zombie EnemyType = iota
	Vampire
	Ghost
	Ogre
	SnakeMage
	Mimic // ДОБАВЛЯЕМ MIMIC
)

// ItemType представляет тип предмета
type ItemType int
const (
	Treasure ItemType = iota
	Food
	Potion
	Scroll
	Weapon
)

// ItemSubType представляет подтип предмета
type ItemSubType int
const (
	HealthPotion ItemSubType = iota
	AgilityPotion
	StrengthPotion
	MaxHealthScroll
	AgilityScroll
	StrengthScroll
	Sword
	Axe
	KeyRed    // ДОБАВЛЯЕМ КЛЮЧИ
	KeyBlue
	KeyGreen
	KeyYellow
)

// Тип для дверей
type Door struct {
    Position  Position
    Color     ItemSubType
    IsLocked  bool
}

// MovementPattern представляет паттерн движения
type MovementPattern int
const (
	Random MovementPattern = iota
	Chase
	Teleport
	Diagonal
	TwoStep
)

// EffectType представляет тип эффекта
type EffectType int
const (
	HealthEffect EffectType = iota
	AgilityEffect
	StrengthEffect
	MaxHealthEffect
)

// GameActions представляет интерфейс для взаимодействия между слоями
type GameActions interface {
	// Действия игрока
	MovePlayer(dx, dy int)
	UseItem(itemType ItemType, index int)
	PickUpItem()
	Attack(enemyIndex int)
	
	// Генерация
	GenerateLevel(levelNum int)
	
	// Сохранение/загрузка
	SaveGame(filename string) error
	LoadGame(filename string) error
	
	// Обновление состояния
	UpdateGameState()
	CheckGameConditions()

	AddMessage(message string)
	TryExitLevel()
}

var CurrentGameState *GameState
var GameActionsImpl GameActions

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// GraphNode представляет узел в графе доступности
type GraphNode struct {
    Position  Position
    IsDoor    bool
    IsLocked  bool
    KeyType   ItemSubType
    Neighbors []*GraphNode
    Visited   bool
}

// AccessibilityGraph представляет граф доступности уровня
type AccessibilityGraph struct {
    Nodes []*GraphNode
}