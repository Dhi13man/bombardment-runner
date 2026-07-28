package parsing

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- OpenFileFromPathOrContent ---

func TestOpenFileFromPathOrContent_WhenBase64Content_ThenCreatesFileAndReturnsHandle(t *testing.T) {
	t.Parallel()

	// Arrange
	content := "hello,world\nfoo,bar"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))

	// Act
	file, path, err := OpenFileFromPathOrContent("upload.csv", encoded)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(path)
	}()

	// filepath.Join("./data", ...) resolves to "data/..." so check for both forms
	if !strings.HasPrefix(path, allowedDataDir) && !strings.HasPrefix(path, "data") {
		t.Errorf("expected path to contain data dir prefix, got %q", path)
	}

	// Verify file content is readable from the returned handle (seek was reset)
	buf := make([]byte, len(content))
	n, readErr := file.Read(buf)
	if readErr != nil {
		t.Fatalf("failed to read file: %v", readErr)
	}
	if string(buf[:n]) != content {
		t.Errorf("file content: got %q, want %q", string(buf[:n]), content)
	}
}

func TestOpenFileFromPathOrContent_WhenValidFilePath_ThenOpensExistingFile(t *testing.T) {
	t.Parallel()

	// Arrange
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	expected := "test data"
	if _, wErr := tmpFile.WriteString(expected); wErr != nil {
		t.Fatalf("failed to write test data: %v", wErr)
	}
	_ = tmpFile.Close()

	// Act
	file, absPath, err := OpenFileFromPathOrContent(tmpFile.Name(), "")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = file.Close() }()

	if absPath == "" {
		t.Error("expected non-empty absolute path")
	}

	buf := make([]byte, len(expected))
	n, _ := file.Read(buf)
	if string(buf[:n]) != expected {
		t.Errorf("file content: got %q, want %q", string(buf[:n]), expected)
	}
}

func TestOpenFileFromPathOrContent_WhenEmptyPathAndNoContent_ThenReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	file, _, err := OpenFileFromPathOrContent("", "")

	// Assert
	if err == nil {
		_ = file.Close()
		t.Fatal("expected error for empty path and no content, got nil")
	}
}

func TestOpenFileFromPathOrContent_WhenInvalidBase64_ThenReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	file, _, err := OpenFileFromPathOrContent("upload.csv", "not-valid-base64!!!")

	// Assert
	if err == nil {
		_ = file.Close()
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestOpenFileFromPathOrContent_WhenPathContainsTraversal_ThenReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{"parent directory", "../etc/passwd"},
		{"nested traversal", "foo/../../etc/shadow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			file, _, err := OpenFileFromPathOrContent(tt.path, "")

			// Assert
			if err == nil {
				_ = file.Close()
				t.Fatalf("expected error for path %q containing traversal, got nil", tt.path)
			}
		})
	}
}

func TestOpenFileFromPathOrContent_WhenFileNotFound_ThenReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	file, _, err := OpenFileFromPathOrContent("/nonexistent/path/file.csv", "")

	// Assert
	if err == nil {
		_ = file.Close()
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestOpenFileFromPathOrContent_WhenBase64WithEmptyFilePath_ThenGeneratesTemporaryFilename(t *testing.T) {
	t.Parallel()

	// Arrange
	encoded := base64.StdEncoding.EncodeToString([]byte("data"))

	// Act
	file, path, err := OpenFileFromPathOrContent("", encoded)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(path)
	}()

	baseName := filepath.Base(path)
	if !strings.HasPrefix(baseName, "upload-") {
		t.Errorf("expected temporary filename starting with 'upload-', got %q", baseName)
	}
}

func TestOpenFileFromPathOrContent_whenUploadedFilenameEscapesDirectory_thenUsesManagedTempPath(t *testing.T) {
	t.Parallel()

	// Arrange
	encoded := base64.StdEncoding.EncodeToString([]byte("safe data"))

	// Act
	file, path, err := OpenFileFromPathOrContent("../../outside.csv", encoded)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = file.Close()
		removeTempFile(path)
	}()
	if filepath.Clean(filepath.Dir(path)) != filepath.Clean(allowedDataDir) {
		t.Errorf("expected managed data directory, got %q", path)
	}
	if !strings.HasPrefix(filepath.Base(path), "upload-") {
		t.Errorf("expected OS-managed upload filename, got %q", filepath.Base(path))
	}
}

// --- ContainsPathTraversal ---

func TestContainsPathTraversal_WhenVariousInputs_ThenDetectsCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"parent directory", "../etc/passwd", true},
		{"nested traversal", "foo/../../etc", true},
		{"trailing double dot resolves to cwd", "path/..", false},
		{"bare double dot", "..", true},
		{"resolvable mid-path cleans away", "some/../path", false},
		{"clean relative path", "data/file.csv", false},
		{"absolute path", "/usr/local/bin", false},
		{"single dot", "./file.csv", false},
		{"empty string", "", false},
		{"dot in filename", "file.name.csv", false},
		{"consecutive dots in filename", "archive..tar.gz", false},
		{"double dot prefix in filename", "..hidden", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got := ContainsPathTraversal(tt.path)

			// Assert
			if got != tt.want {
				t.Errorf("ContainsPathTraversal(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// --- removeTempFile ---

func TestRemoveTempFile_WhenEmptyPath_ThenNoOp(t *testing.T) {
	t.Parallel()

	// Act - should not panic
	removeTempFile("")
}

func TestRemoveTempFile_WhenFileExists_ThenRemovesFile(t *testing.T) {
	t.Parallel()

	// Arrange
	if err := os.MkdirAll(allowedDataDir, 0o750); err != nil {
		t.Fatalf("failed to create managed data directory: %v", err)
	}
	tmpFile, err := os.CreateTemp(allowedDataDir, "remove-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	path := tmpFile.Name()
	_ = tmpFile.Close()
	t.Cleanup(func() { _ = os.Remove(path) })

	// Act
	removeTempFile(path)

	// Assert
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("expected file %q to be removed, but it still exists", path)
	}
}

func TestRemoveTempFile_WhenFileNotExists_ThenNoError(t *testing.T) {
	t.Parallel()

	// Act - should not panic on non-existent path
	removeTempFile("/nonexistent/path/that/does/not/exist.tmp")
}

func TestRemoveTempFile_whenPathIsOutsideManagedDirectory_thenPreservesFile(t *testing.T) {
	t.Parallel()

	// Arrange
	path := filepath.Join(t.TempDir(), "keep.txt")
	if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Act
	removeTempFile(path)

	// Assert
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected unmanaged file to remain, got %v", err)
	}
}

// --- MaxUploadSize enforcement ---

func TestOpenFileFromPathOrContent_WhenBase64ExceedsMaxSize_ThenReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large-allocation test in short mode")
	}
	t.Parallel()

	// Arrange - the pre-decode guard rejects payloads whose encoded length
	// exceeds the threshold, so we only need a valid base64 string longer
	// than base64.StdEncoding.EncodedLen(MaxUploadSize). Allocating ~100MB
	// of zeros and encoding them is the cheapest way to hit this path.
	oversized := make([]byte, MaxUploadSize+1)
	encoded := base64.StdEncoding.EncodeToString(oversized)

	// Act
	file, _, err := OpenFileFromPathOrContent("large.bin", encoded)

	// Assert
	if err == nil {
		_ = file.Close()
		t.Fatal("expected error for oversized upload, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "too large") && !strings.Contains(errMsg, "exceeds maximum") {
		t.Errorf("expected size-limit error, got: %v", err)
	}
}
