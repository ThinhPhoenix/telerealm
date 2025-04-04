package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"telerealm/models"
	"telerealm/services"
	"telerealm/utils"
)

type Handlers struct {
	service services.FileService
}

var (
	mu sync.Mutex // Giữ lại nếu bạn có biến dùng chung khác cần mutual exclusion
)

func NewHandlers(service services.FileService) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (h *Handlers) SendFile(c *gin.Context) {
	botToken := c.MustGet("bot_token").(string)
	chatID := c.PostForm("chat_id")

	var fileID string
	var fileURL string
	var fileSize int
	var err error
	var fileExt string

	fileHeader, err := c.FormFile("document")
	if err == nil {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer file.Close()

		fileID, err = h.service.SendFile(botToken, chatID, file, fileHeader.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send document: %v", err)})
			return
		}

		fileURL, fileSize, err = h.service.GetFileInfo(botToken, fileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get file info: %v", err)})
			return
		}

		fileExt = filepath.Ext(fileHeader.Filename)
		if fileExt != "" {
			fileExt = fileExt[1:] // Remove the leading dot
		}
	} else {
		fileURL = c.PostForm("document")
		fileSize = 0

		isFile, contentType, contentLength, err := isURLFile(fileURL)
		_ = contentLength

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check URL: %v", err)})
			return
		}

		if !isFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL does not point to a file"})
			return
		}

		fileExt = getExtensionFromContentType(contentType)
	}

	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}

	encryptedToken, err := utils.EncryptFileInfo(botToken, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secure URL"})
		return
	}

	secureURL := fmt.Sprintf("%s://%s/drive/%s", scheme, c.Request.Host, encryptedToken)

	response := models.Response{
		Success: true,
		Message: "Upload file successfully!",
		Data: models.FileData{
			ID:        fileID,
			URL:       fileURL,
			SecureURL: secureURL,
			Bytes:     fileSize,
			Format:    fileExt,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) GetFileURL(c *gin.Context) {
	botToken := c.MustGet("bot_token").(string)
	fileID := c.Query("file_id")

	fileURL, _, err := h.service.GetFileInfo(botToken, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get file info: %v", err)})
		return
	}

	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}

	// Sử dụng phương pháp mã hóa mới
	encryptedToken, err := utils.EncryptFileInfo(botToken, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secure URL"})
		return
	}

	secureURL := fmt.Sprintf("%s://%s/drive/%s", scheme, c.Request.Host, encryptedToken)

	response := models.Response{
		Success: true,
		Message: "File URL retrieved successfully!",
		Data: models.FileData{
			URL:       fileURL,
			SecureURL: secureURL,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) DownloadFile(c *gin.Context) {
	encryptedToken := c.Param("key")

	botToken, fileID, err := utils.DecryptFileInfo(encryptedToken)
	if err != nil {
		fmt.Printf("Decryption error: %v\n", err) // Add this logging
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	fileURL, _, err := h.service.GetFileInfo(botToken, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get file info: %v", err)})
		return
	}

	resp, err := http.Get(fileURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file"})
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Lấy phần mở rộng từ URL gốc
	fileExt := filepath.Ext(fileURL)
	if fileExt != "" {
		fileExt = fileExt[1:] // Bỏ dấu chấm ở đầu
	} else {
		// Nếu URL không có phần mở rộng, thử lấy từ content type
		fileExt = getExtensionFromContentType(contentType)
	}

	// Tạo tên file ngắn gọn
	// Lấy 8 ký tự đầu của fileID để đặt tên file
	shortID := fileID
	if len(fileID) > 8 {
		shortID = fileID[:8]
	}

	filename := fmt.Sprintf("file_%s", shortID)
	if fileExt != "" {
		filename += "." + fileExt
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func (h *Handlers) GetFileInfo(c *gin.Context) {
	botToken := c.MustGet("bot_token").(string)
	fileID := c.Query("file_id")

	fileURL, fileSize, err := h.service.GetFileInfo(botToken, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Failed to get file info: %v", err),
		})
		return
	}

	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}

	fileExt := filepath.Ext(fileURL)
	if fileExt != "" {
		fileExt = fileExt[1:] // Remove the leading dot
	}

	// Sử dụng phương pháp mã hóa mới
	encryptedToken, err := utils.EncryptFileInfo(botToken, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secure URL"})
		return
	}

	secureURL := fmt.Sprintf("%s://%s/drive/%s", scheme, c.Request.Host, encryptedToken)

	response := models.Response{
		Success: true,
		Message: "Get file information successfully!",
		Data: models.FileData{
			ID:        fileID,
			URL:       fileURL,
			SecureURL: secureURL,
			Bytes:     fileSize,
			Format:    fileExt,
		},
	}

	c.JSON(http.StatusOK, response)
}

func getExtensionFromContentType(contentType string) string {
	// Loại bỏ các tham số như charset
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	contentType = strings.TrimSpace(contentType)

	switch contentType {
	case "application/zip":
		return "zip"
	case "application/x-7z-compressed":
		return "7z"
	case "application/pdf":
		return "pdf"
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "text/plain":
		return "txt"
	case "text/html":
		return "html"
	case "application/json":
		return "json"
	case "application/xml", "text/xml":
		return "xml"
	case "application/msword":
		return "doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case "application/vnd.ms-excel":
		return "xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "xlsx"
	case "application/vnd.ms-powerpoint":
		return "ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return "pptx"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "audio/mpeg":
		return "mp3"
	case "audio/ogg":
		return "ogg"
	case "application/vnd.rar":
		return "rar"
	default:
		return ""
	}
}

func isURLFile(url string) (bool, string, int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return false, "", 0, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	contentLength := resp.ContentLength

	isFile := strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "application/") ||
		strings.HasPrefix(contentType, "video/") ||
		strings.HasPrefix(contentType, "audio/")

	return isFile, contentType, contentLength, nil
}

func (h *Handlers) CheckBotAndChat(c *gin.Context) {
	botToken := c.MustGet("bot_token").(string)
	chatID := c.Query("chat_id")

	botInfo, chatInfo, botInChat, botIsAdmin, err := h.service.CheckBotAndChat(botToken, chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Failed to check bot and chat info: %v", err),
		})
		return
	}

	response := models.Response{
		Success: true,
		Message: "Bot and chat information retrieved successfully!",
		Data: gin.H{
			"bot_info":     botInfo,
			"chat_info":    chatInfo,
			"bot_in_chat":  botInChat,
			"bot_is_admin": botIsAdmin,
		},
	}

	c.JSON(http.StatusOK, response)
}
