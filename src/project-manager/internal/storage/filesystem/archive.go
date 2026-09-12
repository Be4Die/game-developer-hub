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

func isIgnoredFile(name string) bool {
	clean := filepath.ToSlash(filepath.Clean(name))
	if strings.HasPrefix(clean, "__MACOSX/") || strings.Contains(clean, "/__MACOSX/") || clean == "__MACOSX" {
		return true
	}
	base := filepath.Base(clean)
	if strings.HasPrefix(base, "._") || base == ".DS_Store" || strings.EqualFold(base, "thumbs.db") {
		return true
	}
	return false
}

// determineRootPrefix проверяет, находятся ли все файлы архива в единой папке-обёртке с index.html.
// Если да, возвращает префикс для удаления (например, "Archive/").
func determineRootPrefix(names []string) string {
	for _, name := range names {
		clean := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(name, "./")))
		clean = strings.TrimPrefix(clean, "/")
		if clean == "index.html" {
			return ""
		}
	}

	var candidateDir string
	var minDepth = -1

	for _, name := range names {
		clean := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(name, "./")))
		clean = strings.TrimPrefix(clean, "/")
		base := filepath.Base(clean)
		if strings.EqualFold(base, "index.html") {
			dir := filepath.Dir(clean)
			if dir != "." && dir != "" {
				depth := strings.Count(dir, "/") + 1
				if minDepth == -1 || depth < minDepth {
					minDepth = depth
					candidateDir = dir
				}
			}
		}
	}

	if candidateDir == "" {
		return ""
	}

	prefix := candidateDir + "/"
	for _, name := range names {
		clean := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(name, "./")))
		clean = strings.TrimPrefix(clean, "/")
		if clean == "." || clean == "" {
			continue
		}
		if clean != candidateDir && !strings.HasPrefix(clean, prefix) {
			topDir := strings.Split(candidateDir, "/")[0]
			if clean != topDir && !strings.HasPrefix(clean, topDir+"/") {
				return ""
			}
			prefix = topDir + "/"
		}
	}

	return prefix
}

func detectArchiveType(archivePath string) (string, error) {
	f, err := os.Open(archivePath) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("open archive: %w", err)
	}
	defer func() { _ = f.Close() }()

	header := make([]byte, 4)
	n, err := io.ReadFull(f, header)
	if err == nil || (errors.Is(err, io.ErrUnexpectedEOF) && n >= 2) {
		if n >= 2 && bytes.Equal(header[:2], []byte{0x1f, 0x8b}) {
			return "tar.gz", nil
		}
		if n >= 4 && bytes.Equal(header[:4], []byte{0x50, 0x4b, 0x03, 0x04}) {
			return "zip", nil
		}
	}

	lower := strings.ToLower(archivePath)
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return "tar.gz", nil
	}
	if strings.HasSuffix(lower, ".zip") {
		return "zip", nil
	}

	return "", domain.ErrInvalidArchive
}

// ExtractArchive распаковывает zip или tar.gz архив в указанную целевую директорию targetDir.
// Выполняет валидацию на Zip Slip, Zip Bomb, белый список расширений и обязательное наличие index.html.
func ExtractArchive(archivePath, targetDir string) error {
	cleanArchive := filepath.Clean(archivePath)
	cleanTarget := filepath.Clean(targetDir)

	if err := os.MkdirAll(cleanTarget, 0o750); err != nil {
		return fmt.Errorf("mkdir targetDir: %w", err)
	}

	archType, err := detectArchiveType(cleanArchive)
	if err != nil {
		_ = os.RemoveAll(cleanTarget)
		return err
	}

	if archType == "tar.gz" {
		err = extractTarGz(cleanArchive, cleanTarget)
	} else {
		err = extractZip(cleanArchive, cleanTarget)
	}

	if err != nil {
		_ = os.RemoveAll(cleanTarget)
		return err
	}

	return nil
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
		totalSize    int64
		hasIndexHTML bool
		fileCount    int
	)

	// Первый проход: сбор имен для определения stripPrefix
	var validNames []string
	for _, f := range r.File {
		if isIgnoredFile(f.Name) {
			continue
		}
		validNames = append(validNames, f.Name)
	}

	stripPrefix := determineRootPrefix(validNames)

	// Второй проход: безопасная распаковка
	for _, f := range r.File {
		if isIgnoredFile(f.Name) {
			continue
		}

		cleanName := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(f.Name, "./")))
		cleanName = strings.TrimPrefix(cleanName, "/")
		relPath := strings.TrimPrefix(cleanName, stripPrefix)
		if relPath == "" || relPath == "." {
			continue
		}

		fileCount++
		if fileCount > maxFileCount {
			return fmt.Errorf("%w: file count exceeds limit", domain.ErrInvalidArchive)
		}

		cleanPath := filepath.Clean(filepath.Join(cleanTarget, filepath.FromSlash(relPath))) //nolint:gosec // Zip Slip checked below
		if !strings.HasPrefix(cleanPath, cleanTarget+string(os.PathSeparator)) && cleanPath != cleanTarget {
			return fmt.Errorf("%w: illegal file path (zip slip detected)", domain.ErrInvalidArchive)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanPath, 0o750); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
			continue
		}

		if !isAllowedFile(relPath) {
			return fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, f.Name)
		}

		if strings.EqualFold(relPath, "index.html") {
			hasIndexHTML = true
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

	if !hasIndexHTML {
		return domain.ErrNoIndexHTML
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

	// Первый проход: определение stripPrefix
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
	}
	tarReader := tar.NewReader(gzr)

	var validNames []string
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_ = gzr.Close()
			return fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
		}
		if isIgnoredFile(header.Name) {
			continue
		}
		validNames = append(validNames, header.Name)
	}
	_ = gzr.Close()

	stripPrefix := determineRootPrefix(validNames)

	// Второй проход: распаковка
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek archive: %w", err)
	}
	gzr2, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
	}
	defer func() {
		_ = gzr2.Close()
	}()
	tarReader2 := tar.NewReader(gzr2)

	var (
		totalSize    int64
		hasIndexHTML bool
		fileCount    int
	)

	for {
		header, err := tarReader2.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
		}
		if isIgnoredFile(header.Name) {
			continue
		}

		cleanName := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(header.Name, "./")))
		cleanName = strings.TrimPrefix(cleanName, "/")
		relPath := strings.TrimPrefix(cleanName, stripPrefix)
		if relPath == "" || relPath == "." {
			continue
		}

		fileCount++
		if fileCount > maxFileCount {
			return fmt.Errorf("%w: file count exceeds limit", domain.ErrInvalidArchive)
		}

		cleanPath := filepath.Clean(filepath.Join(cleanTarget, filepath.FromSlash(relPath))) //nolint:gosec // Tar Slip checked below
		if !strings.HasPrefix(cleanPath, cleanTarget+string(os.PathSeparator)) && cleanPath != cleanTarget {
			return fmt.Errorf("%w: illegal file path (zip slip detected)", domain.ErrInvalidArchive)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanPath, 0o750); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
		case tar.TypeReg:
			if !isAllowedFile(relPath) {
				return fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, header.Name)
			}

			if strings.EqualFold(relPath, "index.html") {
				hasIndexHTML = true
			}

			if err := os.MkdirAll(filepath.Dir(cleanPath), 0o750); err != nil {
				return fmt.Errorf("mkdir parent: %w", err)
			}

			outFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode()&0o750|0o640) //nolint:gosec // path traversal sanitized above
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}

			written, err := io.Copy(outFile, io.LimitReader(tarReader2, maxUnpackedSize-totalSize))
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

	if !hasIndexHTML {
		return domain.ErrNoIndexHTML
	}

	return nil
}
