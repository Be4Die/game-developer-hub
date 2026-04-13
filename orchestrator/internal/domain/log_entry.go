package domain

import "time"

// LogEntry описывает одну журнальную запись инстанса.
type LogEntry struct {
	Timestamp time.Time
	Source    LogSource
	Message   string
}
