package ai

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileProcessor handles file processing for AI context
type FileProcessor struct {
	maxFileSize int64 // in bytes
	uploadDir   string
}

// NewFileProcessor creates a new file processor
func NewFileProcessor(uploadDir string, maxFileSizeMB int) *FileProcessor {
	return &FileProcessor{
		maxFileSize: int64(maxFileSizeMB) * 1024 * 1024,
		uploadDir:   uploadDir,
	}
}

// ProcessedFile represents a processed file with extracted content
type ProcessedFile struct {
	OriginalName string
	FilePath     string
	PublicURL    string
	ContentType  string
	Content      string // Extracted text content
	Size         int64
}

// Supported file types for text extraction
var supportedTextTypes = map[string]bool{
	".txt":        true,
	".md":         true,
	".markdown":   true,
	".json":       true,
	".xml":        true,
	".html":       true,
	".htm":        true,
	".csv":        true,
	".yaml":       true,
	".yml":        true,
	".log":        true,
	".go":         true,
	".js":         true,
	".ts":         true,
	".py":         true,
	".java":       true,
	".c":          true,
	".cpp":        true,
	".h":          true,
	".rs":         true,
	".rb":         true,
	".php":        true,
	".css":        true,
	".scss":       true,
	".sql":        true,
	".sh":         true,
	".bash":       true,
}

// Supported image types
var supportedImageTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// IsTextFile checks if a file extension is supported for text extraction
func IsTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedTextTypes[ext]
}

// IsImageFile checks if a file extension is a supported image
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedImageTypes[ext]
}

// UploadFile uploads a file and returns the processed result
func (fp *FileProcessor) UploadFile(fileHeader *multipart.FileHeader) (*ProcessedFile, error) {
	// Check file size
	if fileHeader.Size > fp.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of %d MB", fp.maxFileSize/(1024*1024))
	}

	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	isText := IsTextFile(fileHeader.Filename)
	isImage := IsImageFile(fileHeader.Filename)

	if !isText && !isImage {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Create upload directory
	if err := os.MkdirAll(fp.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("ai_file_%d%s", fileHeader.Size, ext)
	uploadPath := filepath.Join(fp.uploadDir, filename)

	// Create destination file
	dst, err := os.Create(uploadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	result := &ProcessedFile{
		OriginalName: fileHeader.Filename,
		FilePath:     uploadPath,
		PublicURL:    fmt.Sprintf("/uploads/ai/%s", filename),
		ContentType:  http.DetectContentType([]byte{}),
		Size:         size,
	}

	// Extract text content if it's a text file
	if isText {
		content, err := fp.extractText(uploadPath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text: %w", err)
		}
		result.Content = content
		result.ContentType = "text/plain"
	}

	return result, nil
}

// extractText extracts text content from a file
func (fp *FileProcessor) extractText(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Try to detect encoding and convert to UTF-8 if needed
	// For simplicity, we'll assume UTF-8 for now
	content := string(data)

	// Truncate if too large (limit to 50KB of text for context)
	const maxTextSize = 50 * 1024
	if len(content) > maxTextSize {
		content = content[:maxTextSize] + "\n\n... (content truncated due to size)"
	}

	return content, nil
}

// ReadFileFromURL reads a file from a URL and returns the content
func (fp *FileProcessor) ReadFileFromURL(url string) (string, error) {
	// For local files, read directly
	if strings.HasPrefix(url, "/uploads/") || strings.HasPrefix(url, "/") {
		filePath := "." + url
		return fp.extractText(filePath)
	}

	// For remote URLs, fetch the content
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	content := string(data)
	const maxTextSize = 50 * 1024
	if len(content) > maxTextSize {
		content = content[:maxTextSize] + "\n\n... (content truncated due to size)"
	}

	return content, nil
}

// FormatFileAsContext formats a processed file as context for the AI
func FormatFileAsContext(pf *ProcessedFile) string {
	if pf.Content == "" {
		return ""
	}

	var builder bytes.Buffer
	builder.WriteString(fmt.Sprintf("[文件: %s]\n", pf.OriginalName))

	if IsImageFile(pf.OriginalName) {
		builder.WriteString(fmt.Sprintf("图片 URL: %s\n", pf.PublicURL))
	} else {
		builder.WriteString("```\n")
		builder.WriteString(pf.Content)
		builder.WriteString("\n```\n")
	}

	return builder.String()
}
