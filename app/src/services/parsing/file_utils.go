package parsing

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// safeFilenameRe keeps only alphanumeric, dash, underscore, and dot characters.
var safeFilenameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// allowedDataDir is the base directory for uploaded file storage.
const allowedDataDir = "./data"

// MaxUploadSize is the maximum decoded size for base64 file uploads (100MB).
const MaxUploadSize = 100 * 1024 * 1024

// ContainsPathTraversal checks if the cleaned path contains a ".." component,
// indicating an attempt to escape the current directory. It uses filepath.Clean
// first so that benign substrings like "file..txt" are not falsely rejected.
func ContainsPathTraversal(path string) bool {
	cleaned := filepath.Clean(path)
	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == ".." {
			return true
		}
	}
	return false
}

// OpenFileFromPathOrContent opens a file from either a file path or base64 encoded content.
// If fileContentB64 is provided, it will decode it, save to a temporary file in the data directory,
// and return the file handle and the path to the temporary file.
// If fileContentB64 is empty, it will try to open the file at filePath.
func OpenFileFromPathOrContent(filePath, fileContentB64 string) (*os.File, string, error) {
	// Base64 content provided: decode and save to a temp file
	if fileContentB64 != "" {
		file, path, err := openFromBase64Content(filePath, fileContentB64)
		if err != nil {
			return nil, "", err
		}
		zap.L().Info("Opened file from uploaded content", zap.String("path", path))
		return file, path, nil
	}

	// No content provided: open from validated path
	file, path, err := openFromPath(filePath)
	if err != nil {
		return nil, "", err
	}
	zap.L().Info("Opened file from path", zap.String("path", path))
	return file, path, nil
}

func openFromBase64Content(filePath, fileContentB64 string) (*os.File, string, error) {
	if err := os.MkdirAll(allowedDataDir, 0755); err != nil {
		zap.L().Error("Failed to create data directory", zap.Error(err))
		return nil, "", err
	}

	data, err := base64.StdEncoding.DecodeString(fileContentB64)
	if err != nil {
		zap.L().Error("Failed to decode base64 content", zap.Error(err))
		return nil, "", err
	}

	if len(data) > MaxUploadSize {
		return nil, "", fmt.Errorf("decoded file size %d bytes exceeds maximum %d bytes", len(data), MaxUploadSize)
	}

	fileName := generateSafeFilename(filePath)

	tempFilePath := filepath.Join(allowedDataDir, fileName)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		zap.L().Error("Failed to create temporary file", zap.Error(err))
		return nil, "", err
	}

	if _, writeErr := tempFile.Write(data); writeErr != nil {
		_ = tempFile.Close()
		removeTempFile(tempFilePath)
		zap.L().Error("Failed to write content to file", zap.Error(writeErr))
		return nil, "", writeErr
	}

	if _, seekErr := tempFile.Seek(0, io.SeekStart); seekErr != nil {
		_ = tempFile.Close()
		removeTempFile(tempFilePath)
		zap.L().Error("Failed to reset file pointer", zap.Error(seekErr))
		return nil, "", seekErr
	}

	return tempFile, tempFilePath, nil
}

func openFromPath(filePath string) (*os.File, string, error) {
	if filePath == "" {
		return nil, "", fmt.Errorf("file path is required when no base64 content is provided")
	}

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
// Uses an allowlist (alphanumeric, dash, underscore, dot) to strip
// potentially dangerous characters like null bytes or unicode overrides.
func generateSafeFilename(filePath string) string {
	if filePath == "" {
		return "upload_" + uuid.New().String()
	}

	base := filepath.Base(filePath)
	safe := safeFilenameRe.ReplaceAllString(base, "_")
	// Prefix with a UUID to avoid collisions
	return uuid.New().String() + "_" + safe
}
