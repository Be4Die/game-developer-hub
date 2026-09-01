package domain

import "errors"

// ErrNoAvailablePort возвращается когда все порты в диапазоне заняты.
var ErrNoAvailablePort = errors.New("no available port in range")

// PortStrategy определяет стратегию выделения порта на хосте.
// Ровно одно из полей Any, Exact или Range должно быть установлено.
type PortStrategy struct {
	Any   bool
	Exact uint32
	Range *PortRange
}

// PortRange задаёт диапазон портов для поиска свободного.
type PortRange struct {
	Min uint32
	Max uint32
}
