package filesystem

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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

// allowedExtensions содержит разрешенные расширения статических файлов для веб-сборок игр (HTML5/WASM).
var allowedExtensions = map[string]bool{
	".html":    true,
	".htm":     true,
	".js":      true,
	".mjs":     true,
	".css":     true,
	".wasm":    true,
	".png":     true,
	".jpg":     true,
	".jpeg":    true,
	".webp":    true,
	".gif":     true,
	".svg":     true,
	".ico":     true,
	".mp3":     true,
	".wav":     true,
	".ogg":     true,
	".json":    true,
	".txt":     true,
	".xml":     true,
	".ttf":     true,
	".woff":    true,
	".woff2":   true,
	".eot":     true,
	".otf":     true,
	".data":    true,
	".pck":     true,
	".bin":     true,
	".mem":     true,
	".symbols": true,
	".map":     true,
	".atlas":   true,
	".fnt":     true,
	".tga":     true,
	".bmp":     true,
	".ktx":     true,
	".basis":   true,
	".dds":     true,
	".hdr":     true,
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

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target dir %s: %w", targetDir, err)
	}

	var hasIndexHtml bool
	var err error

	if strings.HasSuffix(archivePath, ".zip") {
		hasIndexHtml, err = extractZip(archivePath, targetDir, maxUnpackedBytes, maxFiles)
	} else if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		hasIndexHtml, err = extractTarGz(archivePath, targetDir, maxUnpackedBytes, maxFiles)
	} else {
		hasIndexHtml, err = extractZip(archivePath, targetDir, maxUnpackedBytes, maxFiles)
	}

	if err != nil {
		_ = os.RemoveAll(targetDir)
		return err
	}

	if !hasIndexHtml {
		_ = os.RemoveAll(targetDir)
		return domain.ErrNoIndexHtml
	}

	return nil
}

func isAllowedFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		// Разрешаем служебные файлы без расширения, если это не скрипты
		base := strings.ToLower(filepath.Base(name))
		return base == "license" || base == "cname" || base == "readme"
	}
	return allowedExtensions[ext]
}

func extractZip(archivePath, targetDir string, maxBytes int64, maxFiles int) (bool, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return false, fmt.Errorf("%w: %v", domain.ErrInvalidArchive, err)
	}
	defer r.Close()

	var totalBytes int64
	var fileCount int
	var hasIndexHtml bool

	cleanTarget := filepath.Clean(targetDir)

	for _, f := range r.File {
		fileCount++
		if fileCount > maxFiles {
			return false, domain.ErrZipBomb
		}

		targetPath := filepath.Join(cleanTarget, f.Name)
		if !strings.HasPrefix(filepath.Clean(targetPath), cleanTarget+string(os.PathSeparator)) && filepath.Clean(targetPath) != cleanTarget {
			return false, domain.ErrZipSlip
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return false, fmt.Errorf("mkdir %s: %w", targetPath, err)
			}
			continue
		}

		// Проверка белого списка расширений
		if !isAllowedFile(f.Name) {
			return false, fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, f.Name)
		}

		if strings.EqualFold(filepath.Clean(f.Name), "index.html") || strings.EqualFold(f.Name, "index.html") {
			hasIndexHtml = true
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return false, fmt.Errorf("mkdir parent %s: %w", targetPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			return false, fmt.Errorf("open zip item %s: %w", f.Name, err)
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()&0o755|0o644)
		if err != nil {
			rc.Close()
			return false, fmt.Errorf("create file %s: %w", targetPath, err)
		}

		written, err := io.Copy(outFile, io.LimitReader(rc, maxBytes-totalBytes+1))
		rc.Close()
		outFile.Close()

		if err != nil {
			return false, fmt.Errorf("copy file %s: %w", targetPath, err)
		}

		totalBytes += written
		if totalBytes > maxBytes {
			return false, domain.ErrZipBomb
		}
	}

	return hasIndexHtml, nil
}

func extractTarGz(archivePath, targetDir string, maxBytes int64, maxFiles int) (bool, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return false, fmt.Errorf("%w: %v", domain.ErrInvalidArchive, err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return false, fmt.Errorf("%w: gzip error: %v", domain.ErrInvalidArchive, err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var totalBytes int64
	var fileCount int
	var hasIndexHtml bool

	cleanTarget := filepath.Clean(targetDir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, fmt.Errorf("tar next: %w", err)
		}

		fileCount++
		if fileCount > maxFiles {
			return false, domain.ErrZipBomb
		}

		targetPath := filepath.Join(cleanTarget, header.Name)
		if !strings.HasPrefix(filepath.Clean(targetPath), cleanTarget+string(os.PathSeparator)) && filepath.Clean(targetPath) != cleanTarget {
			return false, domain.ErrZipSlip
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return false, fmt.Errorf("mkdir %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if !isAllowedFile(header.Name) {
				return false, fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, header.Name)
			}

			if strings.EqualFold(filepath.Clean(header.Name), "index.html") || strings.EqualFold(header.Name, "index.html") {
				hasIndexHtml = true
			}

			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return false, fmt.Errorf("mkdir parent %s: %w", targetPath, err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode()&0o755|0o644)
			if err != nil {
				return false, fmt.Errorf("create file %s: %w", targetPath, err)
			}

			written, err := io.Copy(outFile, io.LimitReader(tr, maxBytes-totalBytes+1))
			outFile.Close()
			if err != nil {
				return false, fmt.Errorf("copy file %s: %w", targetPath, err)
			}

			totalBytes += written
			if totalBytes > maxBytes {
				return false, domain.ErrZipBomb
			}
		}
	}

	return hasIndexHtml, nil
}
