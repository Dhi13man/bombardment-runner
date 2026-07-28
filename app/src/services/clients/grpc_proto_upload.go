package clients

import (
	"encoding/base64"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.dhi13man.com/bombardment-runner/src/services/parsing"
	"go.uber.org/zap"
)

const maxFilenameLogLen = 200

// sanitizeFilename truncates and strips control characters from a filename for safe use in error messages.
func sanitizeFilename(name string) string {
	if len(name) > maxFilenameLogLen {
		name = name[:maxFilenameLogLen] + "..."
	}
	return strings.Map(func(r rune) rune {
		if r == utf8.RuneError || r < 0x20 {
			return '?'
		}
		return r
	}, name)
}

const (
	// maxProtoFileSize is the maximum decoded size of a single uploaded proto file (10 MB).
	maxProtoFileSize = 10 << 20
	// maxProtoFileCount is the maximum number of proto files that can be uploaded in a single request.
	maxProtoFileCount = 100
)

// writeProtoContents decodes base64-encoded proto file contents to a temp directory.
// Returns the temp directory path and a slice of written file paths.
// Caller is responsible for cleaning up the temp directory.
func writeProtoContents(contents map[string]string) (string, []string, error) {
	if len(contents) == 0 {
		return "", nil, nil
	}
	if len(contents) > maxProtoFileCount {
		return "", nil, fmt.Errorf("too many proto files: %d (max %d)", len(contents), maxProtoFileCount)
	}

	tempDir, err := os.MkdirTemp("", "bombardment-proto-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Sort keys for deterministic file ordering in paths slice and error messages.
	keys := slices.Sorted(maps.Keys(contents))

	var paths []string
	for _, name := range keys {
		b64 := contents[name]

		// Sanitize filename: reject path traversal and non-.proto files.
		safeName := sanitizeFilename(name)
		if !parsing.IsPortableLocalPath(name) {
			cleanupTempDir(tempDir)
			return "", nil, fmt.Errorf("invalid proto filename: %s", safeName)
		}
		clean := filepath.Clean(name)
		if !strings.HasSuffix(clean, ".proto") {
			cleanupTempDir(tempDir)
			return "", nil, fmt.Errorf("file must end in .proto: %s", safeName)
		}

		// Reject oversized content before allocating memory for decoding.
		if len(b64) > base64.StdEncoding.EncodedLen(maxProtoFileSize) {
			cleanupTempDir(tempDir)
			return "", nil, fmt.Errorf("proto file %s too large (max %d bytes decoded)", name, maxProtoFileSize)
		}

		// Support subdirectories in filename (e.g., "google/protobuf/timestamp.proto")
		filePath := filepath.Join(tempDir, clean)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
			cleanupTempDir(tempDir)
			return "", nil, fmt.Errorf("create subdir for %s: %w", name, err)
		}

		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			cleanupTempDir(tempDir)
			return "", nil, fmt.Errorf("decode base64 for %s: %w", name, err)
		}

		if err := os.WriteFile(filePath, data, 0o600); err != nil {
			cleanupTempDir(tempDir)
			return "", nil, fmt.Errorf("write %s: %w", name, err)
		}

		paths = append(paths, clean)
	}

	return tempDir, paths, nil
}

// cleanupTempDir removes the temp directory, logging on failure.
func cleanupTempDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		zap.L().Warn("failed to clean up proto temp dir", zap.String("dir", dir), zap.Error(err))
	}
}
