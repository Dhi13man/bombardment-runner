package clients

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestWriteProtoContents(t *testing.T) {
	t.Parallel()

	validProto := base64.StdEncoding.EncodeToString([]byte(`syntax = "proto3"; package test;`))

	tests := []struct {
		name      string
		contents  map[string]string
		wantErr   string // substring match; empty = no error expected
		wantCount int    // expected number of returned paths
	}{
		{
			name:      "single valid file",
			contents:  map[string]string{"echo.proto": validProto},
			wantCount: 1,
		},
		{
			name: "multiple valid files",
			contents: map[string]string{
				"a.proto": validProto,
				"b.proto": validProto,
			},
			wantCount: 2,
		},
		{
			name:      "subdirectory in filename",
			contents:  map[string]string{"google/protobuf/timestamp.proto": validProto},
			wantCount: 1,
		},
		{
			name:     "path traversal with .. rejected",
			contents: map[string]string{"../../../etc/passwd.proto": validProto},
			wantErr:  "invalid proto filename",
		},
		{
			name:     "embedded traversal rejected",
			contents: map[string]string{"a/../../etc/shadow.proto": validProto},
			wantErr:  "invalid proto filename",
		},
		{
			name:     "absolute path rejected",
			contents: map[string]string{"/etc/shadow.proto": validProto},
			wantErr:  "invalid proto filename",
		},
		{
			name:     "non-proto extension rejected",
			contents: map[string]string{"malicious.sh": validProto},
			wantErr:  "file must end in .proto",
		},
		{
			name:     "invalid base64 content",
			contents: map[string]string{"test.proto": "not-valid-base64!!!"},
			wantErr:  "decode base64",
		},
		{
			name:      "empty map returns no paths",
			contents:  map[string]string{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir, paths, err := writeProtoContents(tt.contents)

			if tt.wantErr != "" {
				if err == nil {
					// Clean up if unexpectedly created
					if tempDir != "" {
						_ = os.RemoveAll(tempDir)
					}
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				// Verify temp dir is cleaned up on error
				if tempDir != "" {
					if _, statErr := os.Stat(tempDir); statErr == nil {
						t.Errorf("temp dir %q should have been removed on error", tempDir)
						_ = os.RemoveAll(tempDir)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Cleanup(func() {
				if tempDir != "" {
					_ = os.RemoveAll(tempDir)
				}
			})

			if len(paths) != tt.wantCount {
				t.Errorf("got %d paths, want %d", len(paths), tt.wantCount)
			}

			// Verify all returned paths are relative and the files exist in tempDir
			for _, p := range paths {
				if filepath.IsAbs(p) {
					t.Errorf("path %q should be relative, not absolute", p)
				}
				absPath := filepath.Join(tempDir, p)
				if _, err := os.Stat(absPath); err != nil {
					t.Errorf("file %q does not exist at %q: %v", p, absPath, err)
				}
			}
		})
	}
}

func TestWriteProtoContents_ContentIntegrity(t *testing.T) {
	t.Parallel()

	original := `syntax = "proto3"; package test; message Foo { string bar = 1; }`
	b64 := base64.StdEncoding.EncodeToString([]byte(original))

	tempDir, paths, err := writeProtoContents(map[string]string{"foo.proto": b64})
	if err != nil {
		t.Fatalf("writeProtoContents() error: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}

	data, err := os.ReadFile(filepath.Join(tempDir, paths[0]))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != original {
		t.Errorf("file content = %q, want %q", string(data), original)
	}
}

func TestWriteProtoContents_SubdirectoryCreated(t *testing.T) {
	t.Parallel()

	b64 := base64.StdEncoding.EncodeToString([]byte(`syntax = "proto3";`))
	tempDir, paths, err := writeProtoContents(map[string]string{
		"google/protobuf/any.proto": b64,
	})
	if err != nil {
		t.Fatalf("writeProtoContents() error: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}

	// Verify the relative path preserves subdirectory structure
	want := filepath.Join("google", "protobuf", "any.proto")
	if paths[0] != want {
		t.Errorf("path = %q, want %q", paths[0], want)
	}

	// Verify the file actually exists on disk
	absPath := filepath.Join(tempDir, paths[0])
	if _, err := os.Stat(absPath); err != nil {
		t.Errorf("file does not exist at %q: %v", absPath, err)
	}
}

func TestWriteProtoContents_CleanupOnPartialFailure(t *testing.T) {
	t.Parallel()

	validB64 := base64.StdEncoding.EncodeToString([]byte(`syntax = "proto3";`))

	// Use sorted keys: "a.proto" will succeed, "b.proto" (bad base64) will fail.
	tempDir, _, err := writeProtoContents(map[string]string{
		"a.proto": validB64,
		"b.proto": "not-valid-base64!!!",
	})
	if err == nil {
		_ = os.RemoveAll(tempDir)
		t.Fatal("expected error, got nil")
	}

	// Temp dir should be cleaned up despite first file being written successfully
	if tempDir != "" {
		if _, statErr := os.Stat(tempDir); statErr == nil {
			t.Error("temp dir should have been removed after partial failure")
			_ = os.RemoveAll(tempDir)
		}
	}
}

func TestWriteProtoContents_TooManyFiles(t *testing.T) {
	t.Parallel()

	contents := make(map[string]string, maxProtoFileCount+1)
	b64 := base64.StdEncoding.EncodeToString([]byte(`syntax = "proto3";`))
	for i := range maxProtoFileCount + 1 {
		contents["file_"+strconv.Itoa(i)+".proto"] = b64
	}

	_, _, err := writeProtoContents(contents)
	if err == nil {
		t.Fatal("expected error for too many files, got nil")
	}
	if !strings.Contains(err.Error(), "too many proto files") {
		t.Errorf("error = %q, want containing 'too many proto files'", err.Error())
	}
}

func TestWriteProtoContents_OversizedFile(t *testing.T) {
	t.Parallel()

	// Create base64 content that exceeds the size limit
	oversized := strings.Repeat("A", base64.StdEncoding.EncodedLen(maxProtoFileSize)+1)

	_, _, err := writeProtoContents(map[string]string{"big.proto": oversized})
	if err == nil {
		t.Fatal("expected error for oversized file, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error = %q, want containing 'too large'", err.Error())
	}
}
