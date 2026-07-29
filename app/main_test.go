package main

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestIsActionableLoggerSyncError_WhenClassifyingSyncFailures_ThenOnlyEINVALIsIgnored(t *testing.T) {
	t.Parallel()

	// Arrange
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "no error", err: nil, want: false},
		{
			name: "terminal sync EINVAL",
			err:  &os.PathError{Op: "sync", Path: "/dev/stderr", Err: syscall.EINVAL},
			want: false,
		},
		{name: "actionable sync error", err: errors.New("sync failed"), want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got := isActionableLoggerSyncError(test.err)

			// Assert
			if got != test.want {
				t.Errorf("isActionableLoggerSyncError() = %t, want %t", got, test.want)
			}
		})
	}
}
