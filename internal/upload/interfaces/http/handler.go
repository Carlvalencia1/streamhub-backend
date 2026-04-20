package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const uploadDir = "./uploads"

var allowedTypes = map[string]string{
	"image/jpeg":  "images",
	"image/png":   "images",
	"image/gif":   "images",
	"image/webp":  "images",
	"video/mp4":   "videos",
	"video/quicktime": "videos",
	"video/webm":  "videos",
	"audio/mpeg":  "audio",
	"audio/mp4":   "audio",
	"audio/ogg":   "audio",
	"audio/wav":   "audio",
	"audio/aac":   "audio",
	"audio/webm":  "audio",
}

var extByMime = map[string]string{
	"image/jpeg":  ".jpg",
	"image/png":   ".png",
	"image/gif":   ".gif",
	"image/webp":  ".webp",
	"video/mp4":   ".mp4",
	"video/quicktime": ".mov",
	"video/webm":  ".webm",
	"audio/mpeg":  ".mp3",
	"audio/mp4":   ".m4a",
	"audio/ogg":   ".ogg",
	"audio/wav":   ".wav",
	"audio/aac":   ".aac",
	"audio/webm":  ".weba",
}

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectMIME(header.Filename)
	}
	contentType = strings.Split(contentType, ";")[0]

	subDir, ok := allowedTypes[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type: " + contentType})
		return
	}

	ext := extByMime[contentType]
	if ext == "" {
		ext = filepath.Ext(header.Filename)
	}

	dir := filepath.Join(uploadDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storage error"})
		return
	}

	filename := uuid.NewString() + ext
	dst := filepath.Join(dir, filename)

	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storage error"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storage error"})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/uploads/%s/%s", scheme, c.Request.Host, subDir, filename)
	c.JSON(http.StatusOK, gin.H{
		"url":  url,
		"type": subDir[:len(subDir)-1],
	})
}

func detectMIME(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	m := map[string]string{
		".jpg": "image/jpeg", ".jpeg": "image/jpeg",
		".png": "image/png", ".gif": "image/gif", ".webp": "image/webp",
		".mp4": "video/mp4", ".mov": "video/quicktime", ".webm": "video/webm",
		".mp3": "audio/mpeg", ".m4a": "audio/mp4", ".ogg": "audio/ogg",
		".wav": "audio/wav", ".aac": "audio/aac",
	}
	if t, ok := m[ext]; ok {
		return t
	}
	return ""
}
