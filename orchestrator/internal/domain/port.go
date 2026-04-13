package domain

import "errors"

// ErrNoAvailablePort возвращается когда все порты в диапазоне заняты.
var ErrNoAvailablePort = errors.New("no available port in range")

// PortAllocation определяет стратегию выделения хостового порта.
// Ровно одно из полей Any, Exact или Range должно быть установлено.
type PortAllocation struct {
	Any   bool
	Exact uint32
	Range *PortRange
}

// PortRange задаёт диапазон портов для поиска свободного.
type PortRange struct {
	Min uint32
	Max uint32
}
