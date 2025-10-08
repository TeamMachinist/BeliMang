package image

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"belimang/internal/middleware"
	logger "belimang/internal/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageHandler struct{}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

type ImageUploadResponse struct {
	Message string    `json:"message"`
	Data    ImageData `json:"data"`
}

type ImageData struct {
	ImageURL string `json:"imageUrl"`
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	ctx := c.Request.Context()
	logger.InfoCtx(ctx, "Starting image upload")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to get uploaded file", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get uploaded file",
		})
		return
	}
	defer file.Close()

	if err := h.validateFile(header); err != nil {
		logger.ErrorCtx(ctx, "File validation failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	filename := h.generateFilename(header.Filename)

	imageURL := fmt.Sprintf("https://awss3.%s", filename)

	logger.InfoCtx(ctx, "Image uploaded successfully", "filename", filename, "imageUrl", imageURL)

	c.JSON(http.StatusOK, ImageUploadResponse{
		Message: "File uploaded sucessfully",
		Data: ImageData{
			ImageURL: imageURL,
		},
	})
}

func (h *ImageHandler) validateFile(header *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" {
		return fmt.Errorf("file must be in *.jpg or *.jpeg format")
	}

	if header.Size < 10*1024 {
		return fmt.Errorf("file size must be at least 10KB")
	}
	if header.Size > 2*1024*1024 {
		return fmt.Errorf("file size must not exceed 2MB")
	}

	return nil
}

func (h *ImageHandler) generateFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	filename := uuid.New().String() + ext
	return filename
}

func RegisterRoutes(router *gin.Engine, handler *ImageHandler) {
	imageGroup := router.Group("/")
	imageGroup.Use(middleware.AdminAuthMiddleware())
	{
		imageGroup.POST("/image", handler.UploadImage)
	}
}
