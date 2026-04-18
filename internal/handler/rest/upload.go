package rest

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	file, header, err := c.Request.FormFile("file")
	if err != nil {
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
	dst, err := os.Create(filepath.Join(h.dir, filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + filename})
}
