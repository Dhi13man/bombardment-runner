package parsing

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// OpenFileFromPathOrContent opens a file from either a file path or base64 encoded content.
// If fileContentB64 is provided, it will decode it, save to a temporary file in the data directory,
// and return the file handle and the path to the temporary file.
// If fileContentB64 is empty, it will try to open the file at filePath.
func OpenFileFromPathOrContent(filePath, fileContentB64 string) (*os.File, string, error) {
	// If base64 content is provided, decode and save to temp file
	if fileContentB64 != "" {
		// Create uploads directory if it doesn't exist
		uploadsDir := "./data"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			zap.L().Error("Failed to create data directory", zap.Error(err))
			return nil, "", err
		}

		// Decode base64 content
		data, err := base64.StdEncoding.DecodeString(fileContentB64)
		if err != nil {
			zap.L().Error("Failed to decode base64 content", zap.Error(err))
			return nil, "", err
		}

		// Use filename from filePath or generate a temp filename
		var fileName string
		if filePath != "" {
			fileName = filepath.Base(filePath)
		} else {
			fileName = "temp_upload_" + strings.ReplaceAll(filepath.Base(filePath), " ", "_")
		}

		// Create a temporary file with the content
		tempFilePath := filepath.Join(uploadsDir, fileName)
		tempFile, err := os.Create(tempFilePath)
		if err != nil {
			zap.L().Error("Failed to create temporary file", zap.Error(err))
			return nil, "", err
		}

		// Write decoded content to file
		if _, err := tempFile.Write(data); err != nil {
			tempFile.Close()
			zap.L().Error("Failed to write content to file", zap.Error(err))
			return nil, "", err
		}

		// Reset file pointer to beginning
		if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
			tempFile.Close()
			zap.L().Error("Failed to reset file pointer", zap.Error(err))
			return nil, "", err
		}

		return tempFile, tempFilePath, nil
	}

	// If no content provided, just open the file at path
	file, err := os.Open(filePath)
	if err != nil {
		zap.L().Error("Error opening file", zap.Error(err))
		return nil, "", err
	}

	return file, filePath, nil
}
