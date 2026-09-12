package service

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateMediaFootprint(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test_icon.png")

	// Создаем тестовое изображение 512x512
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 100, A: 255})
		}
	}
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("create test image: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode test image: %v", err)
	}
	_ = f.Close()

	// Тестируем генерацию отпечатка
	footprint, err := GenerateMediaFootprint(imgPath, 240)
	if err != nil {
		t.Fatalf("GenerateMediaFootprint failed: %v", err)
	}

	if footprint.FileName != "test_icon.png" {
		t.Errorf("expected filename test_icon.png, got %s", footprint.FileName)
	}
	if footprint.MimeType != "image/png" {
		t.Errorf("expected mime image/png, got %s", footprint.MimeType)
	}
	if footprint.Sha256 == "" {
		t.Errorf("expected non-empty sha256")
	}
	if !strings.HasPrefix(footprint.ThumbnailData, "data:image/jpeg;base64,") {
		t.Errorf("expected base64 jpeg data-uri, got prefix %s", footprint.ThumbnailData[:30])
	}
	// Проверяем, что размер миниатюры ультра-компактный (меньше 15 КБ)
	if len(footprint.ThumbnailData) > 20000 {
		t.Errorf("thumbnail too large: %d bytes", len(footprint.ThumbnailData))
	}
}
