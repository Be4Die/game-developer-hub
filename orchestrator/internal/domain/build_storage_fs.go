package domain

import "io"

// BuildStorageFS управляет хранением файлов серверных билдов на файловой системе.
type BuildStorageFS interface {
	// Save сохраняет tar-архив билда. gameID и version формируют путь хранения.
	// Возвращает абсолютный путь к сохранённому файлу.
	Save(gameID int64, version string, reader io.Reader, size int64) (path string, err error)

	// Get возвращает reader для чтения файла билда.
	Get(gameID int64, version string) (io.ReadCloser, error)

	// Delete удаляет файл билда. Возвращает ErrNotFound при отсутствии.
	Delete(gameID int64, version string) error

	// Exists проверяет существование билда в файловом хранилище.
	Exists(gameID int64, version string) bool
}
