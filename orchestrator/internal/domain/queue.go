package domain

import (
	"context"
	"time"
)

// QueueEntry описывает позицию игрока в очереди.
type QueueEntry struct {
	GameID             int64
	PlayerID           string
	Mode               string
	JoinTime           time.Time
	LastHeartbeat      time.Time
	ReservedInstanceID int64
	ReservedUntil      time.Time
}

// QueueStore хранит активные очереди игроков в KV (Valkey/Redis).
// Использует SortedSet (score = join_time) для FIFO + Hash для мета-данных.
type QueueStore interface {
	// Join добавляет игрока в очередь игры.
	// Если игрок уже в очереди — обновляет heartbeat.
	Join(ctx context.Context, gameID int64, playerID, mode string) error

	// Leave удаляет игрока из очереди.
	Leave(ctx context.Context, gameID int64, playerID string) error

	// Heartbeat обновляет last_heartbeat игрока.
	// Возвращает ErrNotFound если игрока нет в очереди.
	Heartbeat(ctx context.Context, gameID int64, playerID string) error

	// GetPosition возвращает позицию игрока в очереди (1-based) и общее количество.
	// Возвращает ErrNotFound если игрока нет в очереди.
	GetPosition(ctx context.Context, gameID int64, playerID string) (position, total int64, err error)

	// GetReservation возвращает зарезервированный эндпоинт для игрока.
	// Возвращает ErrNotFound если нет активной резервации.
	GetReservation(ctx context.Context, gameID int64, playerID string) (*ServerEndpoint, time.Time, error)

	// Reserve резервирует слот для первого игрока в очереди.
	// Возвращает player_id которому зарезервировали, или ErrNotFound если очередь пуста.
	Reserve(ctx context.Context, gameID int64, endpoint *ServerEndpoint, reservationTimeout time.Duration) (string, error)

	// PopFirst возвращает и удаляет первого игрока из очереди (FIFO).
	// Используется при cleanup или forced dequeue.
	PopFirst(ctx context.Context, gameID int64) (*QueueEntry, error)

	// CleanupExpired удаляет игроков с просроченным heartbeat.
	// Возвращает список удалённых player_id.
	CleanupExpired(ctx context.Context, gameID int64, heartbeatTimeout time.Duration) ([]string, error)

	// ListQueue возвращает всех игроков в очереди (для мониторинга).
	ListQueue(ctx context.Context, gameID int64) ([]*QueueEntry, error)

	// Count возвращает количество игроков в очереди.
	Count(ctx context.Context, gameID int64) (int64, error)

	// DeleteAll удаляет всю очередь игры (например, при отключении оркестрации).
	DeleteAll(ctx context.Context, gameID int64) error
}

// QueueEventRepo хранит аудит-лог событий очереди в PostgreSQL.
type QueueEventRepo interface {
	// Log записывает событие в аудит-лог.
	Log(ctx context.Context, gameID int64, playerID string, eventType QueueEventType, instanceID int64, waitSeconds int) error

	// ListEvents возвращает события очереди игры.
	ListEvents(ctx context.Context, gameID int64, limit int) ([]*QueueEvent, error)

	// GetStats возвращает агрегированную статистику очереди за период.
	GetStats(ctx context.Context, gameID int64, since time.Time) (*QueueStats, error)
}

// QueueEvent — запись аудит-лога очереди.
type QueueEvent struct {
	ID          int64
	GameID      int64
	PlayerID    string
	EventType   QueueEventType
	InstanceID  int64
	WaitSeconds int
	CreatedAt   time.Time
}

// QueueStats — агрегированная статистика очереди.
type QueueStats struct {
	TotalJoined    int64
	TotalReserved  int64
	TotalConnected int64
	TotalTimeouts  int64
	AvgWaitSeconds int64
}
