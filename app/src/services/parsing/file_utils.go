package parsing

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// allowedDataDir is the base directory for uploaded file storage.
const allowedDataDir = "./data"

// MaxUploadSize is the maximum decoded size for base64 file uploads (100MB).
const MaxUploadSize = 100 * 1024 * 1024

// ContainsPathTraversal checks if the given path contains directory traversal sequences.
func ContainsPathTraversal(path string) bool {
	return strings.Contains(path, "..")
}

// OpenFileFromPathOrContent opens a file from either a file path or base64 encoded content.
// If fileContentB64 is provided, it will decode it, save to a temporary file in the data directory,
// and return the file handle and the path to the temporary file.
// If fileContentB64 is empty, it will try to open the file at filePath.
func OpenFileFromPathOrContent(filePath, fileContentB64 string) (*os.File, string, error) {
	// If base64 content is provided, decode and save to a temp file
	if fileContentB64 != "" {
		file, path, err := openFromBase64Content(filePath, fileContentB64)
		if err != nil {
			return nil, "", err
		}
		zap.L().Info("Opened file from uploaded content", zap.String("path", path))
		return file, path, nil
	}

	// If no content provided, open the file at a validated path
	file, path, err := openFromPath(filePath)
	if err != nil {
		return nil, "", err
	}
	zap.L().Info("Opened file from path", zap.String("path", path))
	return file, path, nil
}

func openFromBase64Content(filePath, fileContentB64 string) (*os.File, string, error) {
	// Create an uploads directory if it doesn't exist
	if err := os.MkdirAll(allowedDataDir, 0755); err != nil {
		zap.L().Error("Failed to create data directory", zap.Error(err))
		return nil, "", err
	}

	// Decode base64 content
	data, err := base64.StdEncoding.DecodeString(fileContentB64)
	if err != nil {
		zap.L().Error("Failed to decode base64 content", zap.Error(err))
		return nil, "", err
	}

	if len(data) > MaxUploadSize {
		return nil, "", fmt.Errorf("decoded file size %d bytes exceeds maximum %d bytes", len(data), MaxUploadSize)
	}

	// Generate a safe filename: use sanitized base name from filePath, or UUID
	fileName := generateSafeFilename(filePath)

	// Create a temporary file with the content
	tempFilePath := filepath.Join(allowedDataDir, fileName)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		zap.L().Error("Failed to create temporary file", zap.Error(err))
		return nil, "", err
	}

	// Write decoded content to a file
	if _, writeErr := tempFile.Write(data); writeErr != nil {
		closeErr := tempFile.Close()
		if closeErr != nil {
			zap.L().Error("Failed to close temporary file after write error", zap.Error(closeErr))
		}
		zap.L().Error("Failed to write content to file", zap.Error(writeErr))
		return nil, "", writeErr
	}

	// Reset a file pointer to the beginning
	if _, seekErr := tempFile.Seek(0, io.SeekStart); seekErr != nil {
		closeErr := tempFile.Close()
		if closeErr != nil {
			zap.L().Error("Failed to close temporary file after seek error", zap.Error(closeErr))
		}
		zap.L().Error("Failed to reset file pointer", zap.Error(seekErr))
		return nil, "", seekErr
	}

	return tempFile, tempFilePath, nil
}

func openFromPath(filePath string) (*os.File, string, error) {
	if filePath == "" {
		return nil, "", fmt.Errorf("file path is required when no base64 content is provided")
	}

	// Ensure the path doesn't contain traversal sequences
	if ContainsPathTraversal(filePath) {
		return nil, "", fmt.Errorf("file path must not contain directory traversal sequences")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("invalid file path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		zap.L().Error("Error opening file", zap.Error(err))
		return nil, "", err
	}

	return file, absPath, nil
}

// removeTempFile removes a temporary file at the given path if non-empty.
// Used by parser constructors to clean up base64-uploaded temp files on error.
func removeTempFile(path string) {
	if path == "" {
		return
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		zap.L().Error("Failed to remove temp file", zap.String("path", path), zap.Error(err))
	}
}

// generateSafeFilename creates a sanitized filename from the given path,
// or generates a UUID-based filename if the path is empty.
func generateSafeFilename(filePath string) string {
	if filePath == "" {
		return "upload_" + uuid.New().String()
	}

	base := filepath.Base(filePath)
	// Replace any problematic characters
	safe := strings.ReplaceAll(base, " ", "_")
	// Prefix with a UUID to avoid collisions
	return uuid.New().String() + "_" + safe
}
