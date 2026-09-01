// Package filesystem provides archive extraction utilities.
package filesystem

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/domain"
)

// Default limits
const (
	defaultMaxUnpackedBytes = 500 * 1024 * 1024 // 500 MB
	defaultMaxFileCount     = 50000
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

// ExtractArchive распаковывает архив (ZIP или TAR.GZ) в целевую директорию targetDir.
// Производит проверки безопасности: Zip-Slip, Zip-Bomb, белый список расширений и наличие index.html.
func ExtractArchive(archivePath, targetDir string, maxUnpackedBytes int64, maxFiles int) error {
	if maxUnpackedBytes <= 0 {
		maxUnpackedBytes = defaultMaxUnpackedBytes
	}
	if maxFiles <= 0 {
		maxFiles = defaultMaxFileCount
	}

	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return fmt.Errorf("create target dir %s: %w", targetDir, err)
	}

	var hasIndexHTML bool
	var err error

	if strings.HasSuffix(archivePath, ".zip") {
		hasIndexHTML, err = extractZip(archivePath, targetDir, maxUnpackedBytes, maxFiles)
	} else if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		hasIndexHTML, err = extractTarGz(archivePath, targetDir, maxUnpackedBytes, maxFiles)
	} else {
		hasIndexHTML, err = extractZip(archivePath, targetDir, maxUnpackedBytes, maxFiles)
	}

	if err != nil {
		_ = os.RemoveAll(targetDir)
		return err
	}

	if !hasIndexHTML {
		_ = os.RemoveAll(targetDir)
		return domain.ErrNoIndexHTML
	}

	return nil
}

func extractZip(archivePath, targetDir string, maxBytes int64, maxFiles int) (bool, error) {
	cleanArchive := filepath.Clean(archivePath)
	r, err := zip.OpenReader(cleanArchive)
	if err != nil {
		return false, fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
	}
	defer func() {
		_ = r.Close()
	}()

	var totalBytes int64
	var fileCount int
	var hasIndexHTML bool

	cleanTarget := filepath.Clean(targetDir)

	for _, f := range r.File {
		fileCount++
		if fileCount > maxFiles {
			return false, domain.ErrZipBomb
		}

		cleanPath := filepath.Clean(filepath.Join(cleanTarget, f.Name)) //nolint:gosec // Zip Slip checked below
		if !strings.HasPrefix(cleanPath, cleanTarget+string(os.PathSeparator)) && cleanPath != cleanTarget {
			return false, domain.ErrZipSlip
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanPath, 0o750); err != nil {
				return false, fmt.Errorf("mkdir %s: %w", cleanPath, err)
			}
			continue
		}

		// Проверка белого списка расширений
		if !isAllowedFile(f.Name) {
			return false, fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, f.Name)
		}

		if strings.EqualFold(filepath.Clean(f.Name), "index.html") || strings.EqualFold(f.Name, "index.html") {
			hasIndexHTML = true
		}

		if err := os.MkdirAll(filepath.Dir(cleanPath), 0o750); err != nil {
			return false, fmt.Errorf("mkdir parent %s: %w", cleanPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			return false, fmt.Errorf("open zip item %s: %w", f.Name, err)
		}

		outFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()&0o750|0o640) //nolint:gosec // path traversal sanitized above
		if err != nil {
			_ = rc.Close()
			return false, fmt.Errorf("create file %s: %w", cleanPath, err)
		}

		written, err := io.Copy(outFile, io.LimitReader(rc, maxBytes-totalBytes+1))
		_ = rc.Close()
		_ = outFile.Close()

		if err != nil {
			return false, fmt.Errorf("copy file %s: %w", cleanPath, err)
		}

		totalBytes += written
		if totalBytes > maxBytes {
			return false, domain.ErrZipBomb
		}
	}

	return hasIndexHTML, nil
}

func extractTarGz(archivePath, targetDir string, maxBytes int64, maxFiles int) (bool, error) {
	cleanArchive := filepath.Clean(archivePath)
	file, err := os.Open(cleanArchive) //nolint:gosec
	if err != nil {
		return false, fmt.Errorf("%w: %w", domain.ErrInvalidArchive, err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return false, fmt.Errorf("%w: gzip error: %w", domain.ErrInvalidArchive, err)
	}
	defer func() {
		_ = gzr.Close()
	}()

	tr := tar.NewReader(gzr)

	var totalBytes int64
	var fileCount int
	var hasIndexHTML bool

	cleanTarget := filepath.Clean(targetDir)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return false, fmt.Errorf("tar next: %w", err)
		}

		fileCount++
		if fileCount > maxFiles {
			return false, domain.ErrZipBomb
		}

		cleanPath := filepath.Clean(filepath.Join(cleanTarget, header.Name)) //nolint:gosec // Tar Slip checked below
		if !strings.HasPrefix(cleanPath, cleanTarget+string(os.PathSeparator)) && cleanPath != cleanTarget {
			return false, domain.ErrZipSlip
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanPath, 0o750); err != nil {
				return false, fmt.Errorf("mkdir %s: %w", cleanPath, err)
			}
		case tar.TypeReg:
			if !isAllowedFile(header.Name) {
				return false, fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, header.Name)
			}

			if strings.EqualFold(filepath.Clean(header.Name), "index.html") || strings.EqualFold(header.Name, "index.html") {
				hasIndexHTML = true
			}

			if err := os.MkdirAll(filepath.Dir(cleanPath), 0o750); err != nil {
				return false, fmt.Errorf("mkdir parent %s: %w", cleanPath, err)
			}

			outFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode()&0o750|0o640) //nolint:gosec // path traversal sanitized above
			if err != nil {
				return false, fmt.Errorf("create file %s: %w", cleanPath, err)
			}

			written, err := io.Copy(outFile, io.LimitReader(tr, maxBytes-totalBytes+1))
			_ = outFile.Close()
			if err != nil {
				return false, fmt.Errorf("copy file %s: %w", cleanPath, err)
			}

			totalBytes += written
			if totalBytes > maxBytes {
				return false, domain.ErrZipBomb
			}
		}
	}

	return hasIndexHTML, nil
}
