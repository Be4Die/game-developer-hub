// Package filesystem implements archive extraction and verification for project-manager.
package filesystem

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

const (
	maxUnpackedSize = 500 * 1024 * 1024 // 500 МБ макс. размер распаковки
	maxFileCount    = 50000             // макс. количество файлов в архиве
)

// allowedExtensions содержит разрешенные расширения статических файлов для веб-сборок игр (HTML5/WASM/Unity/Godot).
var allowedExtensions = map[string]bool{
	".html":       true,
	".htm":        true,
	".js":         true,
	".mjs":        true,
	".css":        true,
	".wasm":       true,
	".png":        true,
	".jpg":        true,
	".jpeg":       true,
	".webp":       true,
	".gif":        true,
	".svg":        true,
	".ico":        true,
	".mp3":        true,
	".wav":        true,
	".ogg":        true,
	".ogv":        true,
	".flac":       true,
	".aac":        true,
	".opus":       true,
	".mp4":        true,
	".webm":       true,
	".json":       true,
	".txt":        true,
	".xml":        true,
	".ttf":        true,
	".woff":       true,
	".woff2":      true,
	".eot":        true,
	".otf":        true,
	".data":       true,
	".pck":        true,
	".bin":        true,
	".mem":        true,
	".symbols":    true,
	".map":        true,
	".atlas":      true,
	".fnt":        true,
	".tga":        true,
	".bmp":        true,
	".ktx":        true,
	".basis":      true,
	".dds":        true,
	".hdr":        true,
	".br":         true, // Brotli compressed assets (Unity WebGL, etc.)
	".gz":         true, // Gzip compressed assets
	".unityweb":   true, // Unity WebGL bundles
	".unity3d":    true,
	".bundle":     true,
	".glb":        true, // 3D assets
	".gltf":       true,
	".fbx":        true,
	".obj":        true,
	".mtl":        true,
	".csv":        true,
	".tsv":        true,
	".yaml":       true,
	".yml":        true,
	".plist":      true,
	".properties": true,
	".ini":        true,
}

func isAllowedFile(name string) bool {
	lowerName := strings.ToLower(name)
	ext := filepath.Ext(lowerName)
	if ext == "" {
		base := filepath.Base(lowerName)
		return base == "license" || base == "cname" || base == "readme"
	}
	if allowedExtensions[ext] {
		return true
	}
	// Проверка составных расширений типа .wasm.br, .data.br, .js.br, .symbols.json.br
	if ext == ".br" || ext == ".gz" {
		trimmed := strings.TrimSuffix(lowerName, ext)
		innerExt := filepath.Ext(trimmed)
		if innerExt != "" && allowedExtensions[innerExt] {
			return true
		}
	}
	return false
}

// ExtractArchive распаковывает zip или tar.gz архив в указанную целевую директорию targetDir.
// Выполняет валидацию на Zip Slip, Zip Bomb, белый список расширений и обязательное наличие index.html.
func ExtractArchive(archivePath, targetDir string) error {
	cleanArchive := filepath.Clean(archivePath)
	cleanTarget := filepath.Clean(targetDir)

	ext := strings.ToLower(filepath.Ext(cleanArchive))
	if strings.HasSuffix(strings.ToLower(cleanArchive), ".tar.gz") || strings.HasSuffix(strings.ToLower(cleanArchive), ".tgz") {
		return extractTarGz(cleanArchive, cleanTarget)
	}
	if ext == ".zip" {
		return extractZip(cleanArchive, cleanTarget)
	}

	// Попробуем определить по магическим байтам
	f, err := os.Open(cleanArchive) //nolint:gosec
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	header := make([]byte, 4)
	if _, err := io.ReadFull(f, header); err != nil {
		return domain.ErrInvalidArchive
	}

	if bytes.Equal(header[:2], []byte{0x1f, 0x8b}) {
		return extractTarGz(cleanArchive, cleanTarget)
	}
	if bytes.Equal(header, []byte{0x50, 0x4b, 0x03, 0x04}) {
		return extractZip(cleanArchive, cleanTarget)
	}

	return domain.ErrInvalidArchive
}

func extractZip(archivePath, targetDir string) error {
	cleanArchive := filepath.Clean(archivePath)
	cleanTarget := filepath.Clean(targetDir)

	r, err := zip.OpenReader(cleanArchive)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
	}
	defer func() {
		_ = r.Close()
	}()

	var (
		totalSize int64
		hasIndex  bool
	)

	// Первый проход: проверка наличия index.html
	for _, f := range r.File {
		cleanName := filepath.Clean(f.Name)
		if cleanName == "index.html" || strings.HasSuffix(cleanName, "/index.html") || strings.HasSuffix(cleanName, "\\index.html") {
			hasIndex = true
		}
	}

	if !hasIndex {
		return domain.ErrNoIndexHTML
	}

	if err := os.MkdirAll(cleanTarget, 0o750); err != nil {
		return fmt.Errorf("mkdir targetDir: %w", err)
	}

	// Второй проход: безопасная распаковка
	for i, f := range r.File {
		if i > maxFileCount {
			return fmt.Errorf("%w: file count exceeds limit", domain.ErrInvalidArchive)
		}

		cleanPath := filepath.Clean(filepath.Join(cleanTarget, f.Name)) //nolint:gosec // Zip Slip checked below
		if !strings.HasPrefix(cleanPath, cleanTarget+string(os.PathSeparator)) && cleanPath != cleanTarget {
			return fmt.Errorf("%w: illegal file path (zip slip detected)", domain.ErrInvalidArchive)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanPath, 0o750); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
			continue
		}

		if !isAllowedFile(f.Name) {
			return fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, f.Name)
		}

		if err := os.MkdirAll(filepath.Dir(cleanPath), 0o750); err != nil {
			return fmt.Errorf("mkdir parent: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry: %w", err)
		}

		outFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()&0o750|0o640) //nolint:gosec // path traversal sanitized above
		if err != nil {
			_ = rc.Close()
			return fmt.Errorf("create file: %w", err)
		}

		written, err := io.Copy(outFile, io.LimitReader(rc, maxUnpackedSize-totalSize))
		_ = rc.Close()
		_ = outFile.Close()

		if err != nil {
			return fmt.Errorf("write unpacked file: %w", err)
		}

		totalSize += written
		if totalSize > maxUnpackedSize {
			return fmt.Errorf("%w: unpacked size exceeds limit", domain.ErrInvalidArchive)
		}
	}

	return nil
}

func extractTarGz(archivePath, targetDir string) error {
	cleanArchive := filepath.Clean(archivePath)
	cleanTarget := filepath.Clean(targetDir)

	file, err := os.Open(cleanArchive) //nolint:gosec
	if err != nil {
		return fmt.Errorf("open tar.gz: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
	}
	defer func() {
		_ = gzr.Close()
	}()

	tarReader := tar.NewReader(gzr)
	var (
		totalSize int64
		hasIndex  bool
		fileCount int
	)

	if err := os.MkdirAll(cleanTarget, 0o750); err != nil {
		return fmt.Errorf("mkdir targetDir: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
		}

		fileCount++
		if fileCount > maxFileCount {
			return fmt.Errorf("%w: file count exceeds limit", domain.ErrInvalidArchive)
		}

		cleanName := filepath.Clean(header.Name)
		if cleanName == "index.html" || strings.HasSuffix(cleanName, "/index.html") {
			hasIndex = true
		}

		cleanPath := filepath.Clean(filepath.Join(cleanTarget, header.Name)) //nolint:gosec // Tar Slip checked below
		if !strings.HasPrefix(cleanPath, cleanTarget+string(os.PathSeparator)) && cleanPath != cleanTarget {
			return fmt.Errorf("%w: illegal file path (zip slip detected)", domain.ErrInvalidArchive)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanPath, 0o750); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
		case tar.TypeReg:
			if !isAllowedFile(header.Name) {
				return fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, header.Name)
			}

			if err := os.MkdirAll(filepath.Dir(cleanPath), 0o750); err != nil {
				return fmt.Errorf("mkdir parent: %w", err)
			}

			outFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode()&0o750|0o640) //nolint:gosec // path traversal sanitized above
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}

			written, err := io.Copy(outFile, io.LimitReader(tarReader, maxUnpackedSize-totalSize))
			_ = outFile.Close()
			if err != nil {
				return fmt.Errorf("write file: %w", err)
			}

			totalSize += written
			if totalSize > maxUnpackedSize {
				return fmt.Errorf("%w: unpacked size exceeds limit", domain.ErrInvalidArchive)
			}
		}
	}

	if !hasIndex {
		return domain.ErrNoIndexHTML
	}

	return nil
}
