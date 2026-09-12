package s3

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"
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

// Deployer реализует domain.Deployer для публикации веб-игр в SeaweedFS S3.
type Deployer struct {
	client       *Client
	gamesBucket  string
	buildsBucket string
	urlPrefix    string
}

// NewDeployer создаёт экземпляр S3 Deployer.
func NewDeployer(client *Client, gamesBucket, buildsBucket, urlPrefix string) *Deployer {
	if gamesBucket == "" {
		gamesBucket = "games"
	}
	if buildsBucket == "" {
		buildsBucket = "builds"
	}
	if urlPrefix == "" {
		urlPrefix = "/games"
	}
	return &Deployer{
		client:       client,
		gamesBucket:  gamesBucket,
		buildsBucket: buildsBucket,
		urlPrefix:    urlPrefix,
	}
}

// resolveMimeAndEncoding определяет Content-Type и Content-Encoding для файлов веб-сборок (Unity WebGL / WASM и др.).
func resolveMimeAndEncoding(path string) (contentType, contentEncoding string) {
	lower := strings.ToLower(path)

	if strings.HasSuffix(lower, ".br") {
		contentEncoding = "br"
		lower = strings.TrimSuffix(lower, ".br")
	} else if strings.HasSuffix(lower, ".gz") {
		contentEncoding = "gzip"
		lower = strings.TrimSuffix(lower, ".gz")
	}

	if strings.HasSuffix(lower, ".wasm") {
		return "application/wasm", contentEncoding
	}
	if strings.HasSuffix(lower, ".data") {
		return "application/octet-stream", contentEncoding
	}
	if strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".mjs") {
		return "application/javascript", contentEncoding
	}
	if strings.HasSuffix(lower, ".json") {
		return "application/json", contentEncoding
	}
	if strings.HasSuffix(lower, ".css") {
		return "text/css", contentEncoding
	}
	if strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".htm") {
		return "text/html; charset=utf-8", contentEncoding
	}

	ext := filepath.Ext(lower)
	if detected := mime.TypeByExtension(ext); detected != "" {
		return detected, contentEncoding
	}

	return "application/octet-stream", contentEncoding
}

// getArchiveBytes загружает байты архива из S3 или локального файла.
func (d *Deployer) getArchiveBytes(ctx context.Context, archivePath string) ([]byte, error) {
	if info, err := os.Stat(archivePath); err == nil && !info.IsDir() {
		f, err := os.Open(archivePath)
		if err != nil {
			return nil, fmt.Errorf("open local archive: %w", err)
		}
		defer func() { _ = f.Close() }()
		return io.ReadAll(io.LimitReader(f, maxUnpackedSize))
	}

	out, err := d.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.buildsBucket),
		Key:    aws.String(archivePath),
	})
	if err != nil {
		return nil, fmt.Errorf("get archive from s3 %s: %w", archivePath, err)
	}
	defer func() { _ = out.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(out.Body, maxUnpackedSize))
	if err != nil {
		return nil, fmt.Errorf("read archive body from s3: %w", err)
	}
	return data, nil
}

// extractAndUploadArchive распаковывает архив в памяти и заливает файлы прямо в S3.
func (d *Deployer) extractAndUploadArchive(ctx context.Context, archivePath, targetPrefix string) error {
	data, err := d.getArchiveBytes(ctx, archivePath)
	if err != nil {
		return err
	}

	if len(data) < 4 {
		return domain.ErrInvalidArchive
	}

	if bytes.Equal(data[:2], []byte{0x1f, 0x8b}) {
		return d.extractAndUploadTarGz(ctx, data, targetPrefix)
	}
	if bytes.Equal(data[:4], []byte{0x50, 0x4b, 0x03, 0x04}) {
		return d.extractAndUploadZip(ctx, data, targetPrefix)
	}

	// Попробуем по расширению
	lower := strings.ToLower(archivePath)
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return d.extractAndUploadTarGz(ctx, data, targetPrefix)
	}
	if strings.HasSuffix(lower, ".zip") {
		return d.extractAndUploadZip(ctx, data, targetPrefix)
	}

	return domain.ErrInvalidArchive
}

func (d *Deployer) extractAndUploadZip(ctx context.Context, data []byte, targetPrefix string) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidArchive, err)
	}

	var (
		totalSize    int64
		hasIndexHTML bool
		fileCount    int
	)

	var validNames []string
	for _, f := range zr.File {
		if isIgnoredFile(f.Name) {
			continue
		}
		validNames = append(validNames, f.Name)
	}

	stripPrefix := determineRootPrefix(validNames)

	for _, f := range zr.File {
		if isIgnoredFile(f.Name) {
			continue
		}
		cleanName := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(f.Name, "./")))
		cleanName = strings.TrimPrefix(cleanName, "/")
		relPath := strings.TrimPrefix(cleanName, stripPrefix)
		if relPath == "" || relPath == "." {
			continue
		}

		totalSize += int64(f.UncompressedSize64)
		fileCount++

		if strings.EqualFold(relPath, "index.html") {
			hasIndexHTML = true
		}
	}

	if totalSize > maxUnpackedSize {
		return fmt.Errorf("%w: size %d exceeds limit %d", domain.ErrInvalidArchive, totalSize, maxUnpackedSize)
	}
	if fileCount > maxFileCount {
		return fmt.Errorf("%w: file count %d exceeds limit %d", domain.ErrInvalidArchive, fileCount, maxFileCount)
	}
	if !hasIndexHTML {
		return domain.ErrNoIndexHTML
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(16)

	for _, f := range zr.File {
		if isIgnoredFile(f.Name) || f.FileInfo().IsDir() {
			continue
		}

		cleanName := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(f.Name, "./")))
		cleanName = strings.TrimPrefix(cleanName, "/")
		relPath := strings.TrimPrefix(cleanName, stripPrefix)
		if relPath == "" || relPath == "." {
			continue
		}

		if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
			return fmt.Errorf("%w: illegal file path (zip slip detected)", domain.ErrInvalidArchive)
		}

		if !isAllowedFile(relPath) {
			return fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, f.Name)
		}

		fileEntry := f
		entryPath := relPath
		g.Go(func() error {
			rc, err := fileEntry.Open()
			if err != nil {
				return fmt.Errorf("open zip entry %s: %w", fileEntry.Name, err)
			}
			defer func() { _ = rc.Close() }()

			fileData, err := io.ReadAll(rc)
			if err != nil {
				return fmt.Errorf("read zip entry %s: %w", fileEntry.Name, err)
			}

			s3Key := fmt.Sprintf("%s/%s", strings.TrimSuffix(targetPrefix, "/"), filepath.ToSlash(entryPath))
			contentType, contentEncoding := resolveMimeAndEncoding(entryPath)

			putInput := &s3.PutObjectInput{
				Bucket:        aws.String(d.gamesBucket),
				Key:           aws.String(s3Key),
				Body:          bytes.NewReader(fileData),
				ContentLength: aws.Int64(int64(len(fileData))),
				ContentType:   aws.String(contentType),
			}
			if contentEncoding != "" {
				putInput.ContentEncoding = aws.String(contentEncoding)
			}

			if _, err := d.client.PutObject(gctx, putInput); err != nil {
				return fmt.Errorf("s3 put object %s: %w", s3Key, err)
			}
			return nil
		})
	}

	return g.Wait()
}

func (d *Deployer) extractAndUploadTarGz(ctx context.Context, data []byte, targetPrefix string) error {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidArchive, err)
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	type tarFileEntry struct {
		name string
		data []byte
	}

	var (
		entries      []tarFileEntry
		validNames   []string
		totalSize    int64
		hasIndexHTML bool
		fileCount    int
	)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrInvalidArchive, err)
		}

		if isIgnoredFile(header.Name) {
			continue
		}
		validNames = append(validNames, header.Name)

		if header.Typeflag == tar.TypeReg {
			fileCount++
			totalSize += header.Size

			if totalSize > maxUnpackedSize {
				return fmt.Errorf("%w: archive size exceeds limit", domain.ErrInvalidArchive)
			}
			if fileCount > maxFileCount {
				return fmt.Errorf("%w: file count exceeds limit", domain.ErrInvalidArchive)
			}

			fileData, err := io.ReadAll(tr)
			if err != nil {
				return fmt.Errorf("read tar entry %s: %w", header.Name, err)
			}

			entries = append(entries, tarFileEntry{
				name: header.Name,
				data: fileData,
			})
		}
	}

	stripPrefix := determineRootPrefix(validNames)

	for _, entry := range entries {
		cleanName := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(entry.name, "./")))
		cleanName = strings.TrimPrefix(cleanName, "/")
		relPath := strings.TrimPrefix(cleanName, stripPrefix)
		if strings.EqualFold(relPath, "index.html") {
			hasIndexHTML = true
		}
	}

	if !hasIndexHTML {
		return domain.ErrNoIndexHTML
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(16)

	for _, entry := range entries {
		cleanName := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(entry.name, "./")))
		cleanName = strings.TrimPrefix(cleanName, "/")
		relPath := strings.TrimPrefix(cleanName, stripPrefix)
		if relPath == "" || relPath == "." {
			continue
		}

		if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
			return fmt.Errorf("%w: illegal file path (zip slip detected)", domain.ErrInvalidArchive)
		}

		if !isAllowedFile(relPath) {
			return fmt.Errorf("%w: %s", domain.ErrDisallowedFileType, entry.name)
		}

		ent := entry
		entryPath := relPath
		g.Go(func() error {
			s3Key := fmt.Sprintf("%s/%s", strings.TrimSuffix(targetPrefix, "/"), filepath.ToSlash(entryPath))
			contentType, contentEncoding := resolveMimeAndEncoding(entryPath)

			putInput := &s3.PutObjectInput{
				Bucket:        aws.String(d.gamesBucket),
				Key:           aws.String(s3Key),
				Body:          bytes.NewReader(ent.data),
				ContentLength: aws.Int64(int64(len(ent.data))),
				ContentType:   aws.String(contentType),
			}
			if contentEncoding != "" {
				putInput.ContentEncoding = aws.String(contentEncoding)
			}

			if _, err := d.client.PutObject(gctx, putInput); err != nil {
				return fmt.Errorf("s3 put object %s: %w", s3Key, err)
			}
			return nil
		})
	}

	return g.Wait()
}

// deletePrefix удаляет все объекты под заданным префиксом в бакете games.
func (d *Deployer) deletePrefix(ctx context.Context, prefix string) error {
	paginator := s3.NewListObjectsV2Paginator(d.client.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.gamesBucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list objects for delete prefix %s: %w", prefix, err)
		}
		if len(page.Contents) == 0 {
			continue
		}

		deleteObjects := make([]types.ObjectIdentifier, len(page.Contents))
		for i, obj := range page.Contents {
			deleteObjects[i] = types.ObjectIdentifier{Key: obj.Key}
		}

		_, err = d.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(d.gamesBucket),
			Delete: &types.Delete{
				Objects: deleteObjects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("delete objects in batch prefix %s: %w", prefix, err)
		}
	}
	return nil
}

// copyPrefix копирует все объекты с одного префикса на другой в бакете games.
func (d *Deployer) copyPrefix(ctx context.Context, srcPrefix, destPrefix string) error {
	srcPrefix = strings.TrimSuffix(srcPrefix, "/") + "/"
	destPrefix = strings.TrimSuffix(destPrefix, "/") + "/"

	paginator := s3.NewListObjectsV2Paginator(d.client.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.gamesBucket),
		Prefix: aws.String(srcPrefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list objects for copy prefix %s: %w", srcPrefix, err)
		}

		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(16)

		for _, obj := range page.Contents {
			if obj.Key == nil || *obj.Key == "" {
				continue
			}

			srcKey := *obj.Key
			rel := strings.TrimPrefix(srcKey, srcPrefix)
			destKey := destPrefix + rel
			copySource := url.PathEscape(fmt.Sprintf("%s/%s", d.gamesBucket, srcKey))

			g.Go(func() error {
				_, err := d.client.CopyObject(gctx, &s3.CopyObjectInput{
					Bucket:     aws.String(d.gamesBucket),
					CopySource: aws.String(copySource),
					Key:        aws.String(destKey),
				})
				if err != nil {
					return fmt.Errorf("copy object %s -> %s: %w", srcKey, destKey, err)
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}
	}
	return nil
}

// DeployDev развёртывает сборку в Dev-окружение (распаковка и загрузка в S3).
func (d *Deployer) DeployDev(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	versionPrefix := fmt.Sprintf("%d/versions/%s", projectID, version)
	if err := d.extractAndUploadArchive(ctx, archivePath, versionPrefix); err != nil {
		return nil, fmt.Errorf("s3_deployer: extract and upload dev build: %w", err)
	}

	devPrefix := fmt.Sprintf("%d/dev", projectID)
	_ = d.deletePrefix(ctx, devPrefix+"/")
	if err := d.copyPrefix(ctx, versionPrefix, devPrefix); err != nil {
		return nil, fmt.Errorf("s3_deployer: copy to dev: %w", err)
	}

	devURL := fmt.Sprintf("%s/%d/dev/index.html", d.urlPrefix, projectID)

	return &domain.DeploymentResult{
		URL:          devURL,
		UnpackedPath: fmt.Sprintf("s3://%s/%s", d.gamesBucket, versionPrefix),
		Success:      true,
	}, nil
}

// DeployProd развёртывает одобренную версию в прод-окружение.
func (d *Deployer) DeployProd(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	versionPrefix := fmt.Sprintf("%d/versions/%s", projectID, version)
	checkKey := fmt.Sprintf("%s/index.html", versionPrefix)

	_, err := d.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.gamesBucket),
		Key:    aws.String(checkKey),
	})
	if err != nil {
		if err := d.extractAndUploadArchive(ctx, archivePath, versionPrefix); err != nil {
			return nil, fmt.Errorf("s3_deployer: extract and upload prod build: %w", err)
		}
	}

	prodPrefix := fmt.Sprintf("%d/prod", projectID)
	_ = d.deletePrefix(ctx, prodPrefix+"/")
	if err := d.copyPrefix(ctx, versionPrefix, prodPrefix); err != nil {
		return nil, fmt.Errorf("s3_deployer: copy to prod: %w", err)
	}

	prodURL := fmt.Sprintf("%s/%d/prod/index.html", d.urlPrefix, projectID)

	return &domain.DeploymentResult{
		URL:          prodURL,
		UnpackedPath: fmt.Sprintf("s3://%s/%s", d.gamesBucket, prodPrefix),
		Success:      true,
	}, nil
}

// UndeployProd снимает игру с публикации в прод-окружении.
func (d *Deployer) UndeployProd(ctx context.Context, projectID int64) error {
	prodPrefix := fmt.Sprintf("%d/prod/", projectID)
	return d.deletePrefix(ctx, prodPrefix)
}

// DeleteVersion удаляет версию сборки из S3.
func (d *Deployer) DeleteVersion(ctx context.Context, projectID int64, version string) error {
	versionPrefix := fmt.Sprintf("%d/versions/%s/", projectID, version)
	return d.deletePrefix(ctx, versionPrefix)
}

// DeleteProject удаляет все файлы проекта (версии, dev, prod) из бакета games.
func (d *Deployer) DeleteProject(ctx context.Context, projectID int64) error {
	projPrefix := fmt.Sprintf("%d/", projectID)
	return d.deletePrefix(ctx, projPrefix)
}
