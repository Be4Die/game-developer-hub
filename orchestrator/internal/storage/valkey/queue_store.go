package valkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/redis/go-redis/v9"
)

// Префиксы ключей очереди.
const (
	keyQueue      = "queue:"
	keyQueueMeta  = "queue:meta:"
	keyQueueResv  = "queue:resv:"
)

// QueueStore реализует domain.QueueStore поверх Valkey.
type QueueStore struct {
	client *redis.Client
}

// NewQueueStore создаёт хранилище очередей.
func NewQueueStore(client *redis.Client) *QueueStore {
	return &QueueStore{client: client}
}

// queueKey возвращает ключ SortedSet для игры.
func queueKey(gameID int64) string {
	return keyQueue + strconv.FormatInt(gameID, 10)
}

// metaKey возвращает ключ Hash для мета-данных игрока.
func metaKey(gameID int64, playerID string) string {
	return keyQueueMeta + strconv.FormatInt(gameID, 10) + ":" + playerID
}

// resvKey возвращает ключ резервации для игрока.
func resvKey(gameID int64, playerID string) string {
	return keyQueueResv + strconv.FormatInt(gameID, 10) + ":" + playerID
}

// Join добавляет игрока в очередь. Если уже есть — обновляет heartbeat.
func (s *QueueStore) Join(ctx context.Context, gameID int64, playerID, mode string) error {
	qKey := queueKey(gameID)
	mKey := metaKey(gameID, playerID)
	now := time.Now()
	score := float64(now.UnixMilli())

	// Добавляем/обновляем в SortedSet
	if err := s.client.ZAdd(ctx, qKey, redis.Z{Score: score, Member: playerID}).Err(); err != nil {
		return fmt.Errorf("valkey.QueueStore.Join: zadd: %w", err)
	}

	// Сохраняем мета-данные
	meta := map[string]interface{}{
		"join_time":       now.Unix(),
		"last_heartbeat":  now.Unix(),
		"mode":            mode,
		"reserved_inst":   0,
		"reserved_until":  0,
	}
	if err := s.client.HSet(ctx, mKey, meta).Err(); err != nil {
		return fmt.Errorf("valkey.QueueStore.Join: hset: %w", err)
	}

	return nil
}

// Leave удаляет игрока из очереди.
func (s *QueueStore) Leave(ctx context.Context, gameID int64, playerID string) error {
	qKey := queueKey(gameID)
	mKey := metaKey(gameID, playerID)
	rKey := resvKey(gameID, playerID)

	pipe := s.client.Pipeline()
	pipe.ZRem(ctx, qKey, playerID)
	pipe.Del(ctx, mKey)
	pipe.Del(ctx, rKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("valkey.QueueStore.Leave: %w", err)
	}
	return nil
}

// Heartbeat обновляет last_heartbeat игрока.
func (s *QueueStore) Heartbeat(ctx context.Context, gameID int64, playerID string) error {
	mKey := metaKey(gameID, playerID)

	exists, err := s.client.Exists(ctx, mKey).Result()
	if err != nil {
		return fmt.Errorf("valkey.QueueStore.Heartbeat: exists: %w", err)
	}
	if exists == 0 {
		return domain.ErrNotFound
	}

	if err := s.client.HSet(ctx, mKey, "last_heartbeat", time.Now().Unix()).Err(); err != nil {
		return fmt.Errorf("valkey.QueueStore.Heartbeat: hset: %w", err)
	}

	return nil
}

// GetPosition возвращает позицию (1-based) и общее количество.
func (s *QueueStore) GetPosition(ctx context.Context, gameID int64, playerID string) (position, total int64, err error) {
	qKey := queueKey(gameID)

	// Проверяем, есть ли игрок в очереди
	score, err := s.client.ZScore(ctx, qKey, playerID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, domain.ErrNotFound
		}
		return 0, 0, fmt.Errorf("valkey.QueueStore.GetPosition: zscore: %w", err)
	}

	// Позиция = количество элементов с меньшим score + 1
	pos, err := s.client.ZCount(ctx, qKey, "-inf", fmt.Sprintf("(%f", score)).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("valkey.QueueStore.GetPosition: zcount: %w", err)
	}

	tot, err := s.client.ZCard(ctx, qKey).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("valkey.QueueStore.GetPosition: zcard: %w", err)
	}

	return pos + 1, tot, nil
}

// GetReservation возвращает зарезервированный эндпоинт.
func (s *QueueStore) GetReservation(ctx context.Context, gameID int64, playerID string) (*domain.ServerEndpoint, time.Time, error) {
	rKey := resvKey(gameID, playerID)

	data, err := s.client.Get(ctx, rKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, time.Time{}, domain.ErrNotFound
		}
		return nil, time.Time{}, fmt.Errorf("valkey.QueueStore.GetReservation: %w", err)
	}

	var payload struct {
		Endpoint  domain.ServerEndpoint `json:"endpoint"`
		ExpiresAt int64                 `json:"expires_at"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, time.Time{}, fmt.Errorf("valkey.QueueStore.GetReservation: unmarshal: %w", err)
	}

	return &payload.Endpoint, time.Unix(payload.ExpiresAt, 0), nil
}

// Reserve резервирует слот для первого игрока в очереди.
func (s *QueueStore) Reserve(ctx context.Context, gameID int64, endpoint *domain.ServerEndpoint, reservationTimeout time.Duration) (string, error) {
	qKey := queueKey(gameID)

	// Берём первого игрока
	result, err := s.client.ZRangeWithScores(ctx, qKey, 0, 0).Result()
	if err != nil {
		return "", fmt.Errorf("valkey.QueueStore.Reserve: zrange: %w", err)
	}
	if len(result) == 0 {
		return "", domain.ErrNotFound
	}

	playerID := result[0].Member.(string)
	mKey := metaKey(gameID, playerID)
	rKey := resvKey(gameID, playerID)

	now := time.Now()
	expiresAt := now.Add(reservationTimeout)

	payload, err := json.Marshal(struct {
		Endpoint  domain.ServerEndpoint `json:"endpoint"`
		ExpiresAt int64                 `json:"expires_at"`
	}{Endpoint: *endpoint, ExpiresAt: expiresAt.Unix()})
	if err != nil {
		return "", fmt.Errorf("valkey.QueueStore.Reserve: marshal: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.HSet(ctx, mKey, "reserved_inst", endpoint.InstanceID)
	pipe.HSet(ctx, mKey, "reserved_until", expiresAt.Unix())
	pipe.Set(ctx, rKey, payload, reservationTimeout)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("valkey.QueueStore.Reserve: pipeline: %w", err)
	}

	return playerID, nil
}

// PopFirst возвращает и удаляет первого игрока из очереди.
func (s *QueueStore) PopFirst(ctx context.Context, gameID int64) (*domain.QueueEntry, error) {
	qKey := queueKey(gameID)

	result, err := s.client.ZRangeWithScores(ctx, qKey, 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("valkey.QueueStore.PopFirst: zrange: %w", err)
	}
	if len(result) == 0 {
		return nil, domain.ErrNotFound
	}

	playerID := result[0].Member.(string)
	score := int64(result[0].Score)

	// Удаляем
	mKey := metaKey(gameID, playerID)
	rKey := resvKey(gameID, playerID)
	pipe := s.client.Pipeline()
	pipe.ZRem(ctx, qKey, playerID)
	pipe.Del(ctx, mKey)
	pipe.Del(ctx, rKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("valkey.QueueStore.PopFirst: del: %w", err)
	}

	return &domain.QueueEntry{
		GameID:   gameID,
		PlayerID: playerID,
		JoinTime: time.UnixMilli(score),
	}, nil
}

// CleanupExpired удаляет игроков с просроченным heartbeat.
func (s *QueueStore) CleanupExpired(ctx context.Context, gameID int64, heartbeatTimeout time.Duration) ([]string, error) {
	qKey := queueKey(gameID)
	cutoff := time.Now().Add(-heartbeatTimeout).Unix()

	// Получаем всех игроков
	members, err := s.client.ZRange(ctx, qKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("valkey.QueueStore.CleanupExpired: zrange: %w", err)
	}

	var expired []string
	pipe := s.client.Pipeline()
	for _, playerID := range members {
		mKey := metaKey(gameID, playerID)
		lastHB, err := s.client.HGet(ctx, mKey, "last_heartbeat").Int64()
		if err != nil {
			continue // ключ пропал или ошибка — пропускаем
		}
		if lastHB < cutoff {
			pipe.ZRem(ctx, qKey, playerID)
			pipe.Del(ctx, mKey)
			pipe.Del(ctx, resvKey(gameID, playerID))
			expired = append(expired, playerID)
		}
	}

	if len(expired) > 0 {
		_, err = pipe.Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("valkey.QueueStore.CleanupExpired: pipeline: %w", err)
		}
	}

	return expired, nil
}

// ListQueue возвращает всех игроков в очереди.
func (s *QueueStore) ListQueue(ctx context.Context, gameID int64) ([]*domain.QueueEntry, error) {
	qKey := queueKey(gameID)

	members, err := s.client.ZRangeWithScores(ctx, qKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("valkey.QueueStore.ListQueue: zrange: %w", err)
	}

	entries := make([]*domain.QueueEntry, 0, len(members))
	for _, m := range members {
		playerID := m.Member.(string)
		mKey := metaKey(gameID, playerID)

		meta, err := s.client.HGetAll(ctx, mKey).Result()
		if err != nil {
			continue
		}

		entry := &domain.QueueEntry{
			GameID:   gameID,
			PlayerID: playerID,
			JoinTime: time.UnixMilli(int64(m.Score)),
		}
		if v, ok := meta["mode"]; ok {
			entry.Mode = v
		}
		if v, ok := meta["last_heartbeat"]; ok {
			if ts, _ := strconv.ParseInt(v, 10, 64); ts > 0 {
				entry.LastHeartbeat = time.Unix(ts, 0)
			}
		}
		if v, ok := meta["reserved_inst"]; ok {
			if id, _ := strconv.ParseInt(v, 10, 64); id > 0 {
				entry.ReservedInstanceID = id
			}
		}
		if v, ok := meta["reserved_until"]; ok {
			if ts, _ := strconv.ParseInt(v, 10, 64); ts > 0 {
				entry.ReservedUntil = time.Unix(ts, 0)
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Count возвращает количество игроков в очереди.
func (s *QueueStore) Count(ctx context.Context, gameID int64) (int64, error) {
	count, err := s.client.ZCard(ctx, queueKey(gameID)).Result()
	if err != nil {
		return 0, fmt.Errorf("valkey.QueueStore.Count: %w", err)
	}
	return count, nil
}

// DeleteAll удаляет всю очередь игры.
func (s *QueueStore) DeleteAll(ctx context.Context, gameID int64) error {
	qKey := queueKey(gameID)

	members, err := s.client.ZRange(ctx, qKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("valkey.QueueStore.DeleteAll: zrange: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, qKey)
	for _, playerID := range members {
		pipe.Del(ctx, metaKey(gameID, playerID))
		pipe.Del(ctx, resvKey(gameID, playerID))
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("valkey.QueueStore.DeleteAll: pipeline: %w", err)
	}
	return nil
}
