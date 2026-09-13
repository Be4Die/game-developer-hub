// Package domain defines core deployment entities and error types.
package domain

import "errors"

var (
	// ErrInvalidArchive возникает, если переданный файл не является валидным ZIP или TAR.GZ архивом.
	ErrInvalidArchive = errors.New("invalid or corrupted archive")

	// ErrZipSlip возникает при попытке распаковать файл за пределы целевой директории.
	ErrZipSlip = errors.New("zip slip vulnerability detected")

	// ErrZipBomb возникает при превышении лимита размера распакованных данных или количества файлов.
	ErrZipBomb = errors.New("archive exceeds decompression limit (zip bomb protection)")

	// ErrNoIndexHTML возникает, если в корне веб-сборки отсутствует файл index.html.
	ErrNoIndexHTML = errors.New("archive must contain an index.html file in the root directory")

	// ErrDisallowedFileType возникает, если архив содержит запрещенные типы файлов (скрипты/бинарники).
	ErrDisallowedFileType = errors.New("archive contains disallowed file type")

	// ErrUnauthorized возникает при передаче неверного или отсутствующего API-ключа.
	ErrUnauthorized = errors.New("unauthorized: invalid or missing api key")
)

// DeploymentResult описывает результат развертывания веб-сборки на агенте.
type DeploymentResult struct {
	URL          string
	UnpackedPath string
	Success      bool
}

// CSPManifest определяет структуру манифеста сетевой безопасности игры.
type CSPManifest struct {
	IsOnline   bool     `json:"is_online"`
	ConnectSrc []string `json:"connect_src"`
}
