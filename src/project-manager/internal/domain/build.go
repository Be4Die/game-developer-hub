package domain

import "time"

// Build представляет неизменяемый артефакт загруженной клиентской сборки веб-игры.
type Build struct {
	ID           int64
	ProjectID    int64
	Version      string
	FilePath     string
	FileSize     int64
	IsUnpacked   bool
	UnpackedPath string
	CreatedAt    time.Time
}
