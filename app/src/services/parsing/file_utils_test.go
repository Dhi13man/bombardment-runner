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
		file.Close()
		os.Remove(path)
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
	tmpFile.Close()

	// Act
	file, absPath, err := OpenFileFromPathOrContent(tmpFile.Name(), "")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()

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
		file.Close()
		t.Fatal("expected error for empty path and no content, got nil")
	}
}

func TestOpenFileFromPathOrContent_WhenInvalidBase64_ThenReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	file, _, err := OpenFileFromPathOrContent("upload.csv", "not-valid-base64!!!")

	// Assert
	if err == nil {
		file.Close()
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
		{"double dot in middle", "some/../path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			file, _, err := OpenFileFromPathOrContent(tt.path, "")

			// Assert
			if err == nil {
				file.Close()
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
		file.Close()
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestOpenFileFromPathOrContent_WhenBase64WithEmptyFilePath_ThenGeneratesUUIDFilename(t *testing.T) {
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
		file.Close()
		os.Remove(path)
	}()

	baseName := filepath.Base(path)
	if !strings.HasPrefix(baseName, "upload_") {
		t.Errorf("expected UUID-based filename starting with 'upload_', got %q", baseName)
	}
}

// --- generateSafeFilename ---

func TestGenerateSafeFilename_WhenEmptyPath_ThenReturnsUploadPrefixedUUID(t *testing.T) {
	t.Parallel()

	// Act
	name := generateSafeFilename("")

	// Assert
	if !strings.HasPrefix(name, "upload_") {
		t.Errorf("expected prefix 'upload_', got %q", name)
	}
	// UUID is 36 chars; "upload_" is 7 chars; total at least 43
	if len(name) < 43 {
		t.Errorf("expected length >= 43 (upload_ + UUID), got %d", len(name))
	}
}

func TestGenerateSafeFilename_WhenPathProvided_ThenReturnsUUIDPrefixedBase(t *testing.T) {
	t.Parallel()

	// Act
	name := generateSafeFilename("/some/dir/myfile.csv")

	// Assert
	if !strings.HasSuffix(name, "_myfile.csv") {
		t.Errorf("expected suffix '_myfile.csv', got %q", name)
	}
	// UUID (36) + "_" (1) + "myfile.csv" (10) = 47
	if len(name) < 47 {
		t.Errorf("expected length >= 47, got %d", len(name))
	}
}

func TestGenerateSafeFilename_WhenPathHasSpaces_ThenReplacesWithUnderscores(t *testing.T) {
	t.Parallel()

	// Act
	name := generateSafeFilename("my file name.csv")

	// Assert
	if strings.Contains(name, " ") {
		t.Errorf("expected no spaces in filename, got %q", name)
	}
	if !strings.HasSuffix(name, "_my_file_name.csv") {
		t.Errorf("expected suffix '_my_file_name.csv', got %q", name)
	}
}

func TestGenerateSafeFilename_WhenCalledTwice_ThenReturnsDifferentNames(t *testing.T) {
	t.Parallel()

	// Act
	name1 := generateSafeFilename("file.csv")
	name2 := generateSafeFilename("file.csv")

	// Assert
	if name1 == name2 {
		t.Error("expected unique filenames on each call due to UUID, got identical")
	}
}
