package domain

import (
	"context"
	"time"
)

// Locker предоставляет интерфейс распределенной блокировки для защиты критических секций от состояния гонки.
type Locker interface {
	// Acquire пытается захватить блокировку по ключу. Возвращает функцию освобождения (unlock) или ошибку.
	Acquire(ctx context.Context, key string, ttl time.Duration) (unlock func(), err error)
}
