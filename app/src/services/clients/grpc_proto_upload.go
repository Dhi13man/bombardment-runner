package clients

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeProtoContents decodes base64-encoded proto file contents to a temp directory.
// Returns the temp directory path and a slice of written file paths.
// Caller is responsible for cleaning up the temp directory.
func writeProtoContents(contents map[string]string) (string, []string, error) {
	tempDir, err := os.MkdirTemp("", "bombardment-proto-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp dir: %w", err)
	}

	var paths []string
	for name, b64 := range contents {
		// Sanitize filename: reject path traversal and non-.proto files
		clean := filepath.Clean(name)
		if strings.Contains(clean, "..") || filepath.IsAbs(clean) {
			os.RemoveAll(tempDir)
			return "", nil, fmt.Errorf("invalid proto filename: %s", name)
		}
		if !strings.HasSuffix(clean, ".proto") {
			os.RemoveAll(tempDir)
			return "", nil, fmt.Errorf("file must end in .proto: %s", name)
		}

		// Support subdirectories in filename (e.g., "google/protobuf/timestamp.proto")
		filePath := filepath.Join(tempDir, clean)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
			os.RemoveAll(tempDir)
			return "", nil, fmt.Errorf("create subdir for %s: %w", name, err)
		}

		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", nil, fmt.Errorf("decode base64 for %s: %w", name, err)
		}

		if err := os.WriteFile(filePath, data, 0o600); err != nil {
			os.RemoveAll(tempDir)
			return "", nil, fmt.Errorf("write %s: %w", name, err)
		}

		paths = append(paths, filePath)
	}

	return tempDir, paths, nil
}
