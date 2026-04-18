package rest

import (
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"voronka/internal/platform/logger"
)

type uploadHandler struct {
	dir string
}

func newUploadHandler(dir string) *uploadHandler {
	_ = os.MkdirAll(dir, 0o755)
	return &uploadHandler{dir: dir}
}

var allowedImageMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func (h *uploadHandler) upload(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Info("upload: missing file", slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = mime.TypeByExtension(strings.ToLower(filepath.Ext(header.Filename)))
	}
	mediaType, _, _ := mime.ParseMediaType(ct)

	ext, ok := allowedImageMIME[mediaType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only image files are allowed (jpeg, png, gif, webp)"})
		return
	}

	filename := uuid.New().String() + ext
	path := filepath.Join(h.dir, filename)
	dst, err := os.Create(path)
	if err != nil {
		log.Error("upload: create destination file", slog.String("err", err.Error()), slog.String("path", path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		log.Error("upload: write file", slog.String("err", err.Error()), slog.String("path", path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + filename})
}
