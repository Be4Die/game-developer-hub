package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MediaFootprint содержит визуальный отпечаток и метаданные медиафайла.
type MediaFootprint struct {
	FileName      string
	FileSize      int64
	MimeType      string
	Sha256        string
	ThumbnailData string // data:image/jpeg;base64,...
	Width         int
	Height        int
}

// GenerateMediaFootprint анализирует файл на диске, вычисляет SHA-256 и создает ультра-компактную миниатюру.
func GenerateMediaFootprint(filePath string, maxDim int) (*MediaFootprint, error) {
	cleanPath := filepath.Clean(filePath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// 1. Вычисление SHA-256 хэша
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("compute sha256: %w", err)
	}
	sha256Hex := hex.EncodeToString(hasher.Sum(nil))

	fileName := filepath.Base(cleanPath)
	ext := strings.ToLower(filepath.Ext(fileName))
	mimeType := detectMimeByExt(ext)

	res := &MediaFootprint{
		FileName: fileName,
		FileSize: info.Size(),
		MimeType: mimeType,
		Sha256:   sha256Hex,
	}

	// Если это не изображение (например видео mp4) — возвращаем только метаданные и sha256
	if !strings.HasPrefix(mimeType, "image/") {
		return res, nil
	}

	// 2. Декодирование изображения для генерации миниатюры
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return res, nil
	}

	srcImg, _, err := image.Decode(file)
	if err != nil {
		// Не удалось декодировать — возвращаем метаданные без миниатюры
		return res, nil
	}

	bounds := srcImg.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	res.Width = origW
	res.Height = origH

	if maxDim <= 0 {
		maxDim = 260
	}

	// Вычисление целевых пропорций
	var targetW, targetH int
	if origW > origH {
		if origW > maxDim {
			targetW = maxDim
			targetH = int(float64(origH) * float64(maxDim) / float64(origW))
		} else {
			targetW = origW
			targetH = origH
		}
	} else {
		if origH > maxDim {
			targetH = maxDim
			targetW = int(float64(origW) * float64(maxDim) / float64(origH))
		} else {
			targetW = origW
			targetH = origH
		}
	}

	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}

	thumbImg := scaleImage(srcImg, targetW, targetH)

	// 3. Кодирование в компактный JPEG (quality 70)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbImg, &jpeg.Options{Quality: 70}); err != nil {
		return res, nil
	}

	res.ThumbnailData = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	return res, nil
}

func scaleImage(src image.Image, targetW, targetH int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	bounds := src.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	if origW <= 0 || origH <= 0 {
		return dst
	}

	for y := 0; y < targetH; y++ {
		srcY := bounds.Min.Y + (y*origH)/targetH
		for x := 0; x < targetW; x++ {
			srcX := bounds.Min.X + (x*origW)/targetW
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func detectMimeByExt(ext string) string {
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	default:
		return "application/octet-stream"
	}
}
