package pathsafe

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	// Arrange

	// Shared base directory used across all cases below.
	// t.TempDir() gives a real, absolute, OS-appropriate path
	// without requiring the files themselves to exist.
	base := t.TempDir()

	// The test cases for the Contains function.
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{
			name:   "direct child file is contained",
			target: filepath.Join(base, "file.txt"),
			want:   true,
		},
		{
			name:   "nested subdirectory file is contained",
			target: filepath.Join(base, "sub", "dir", "file.txt"),
			want:   true,
		},
		{
			name:   "path that lexically resolves back inside base is contained",
			target: filepath.Join(base, "sub", "..", "file.txt"),
			want:   true,
		},
		{
			name:   "parent directory traversal is rejected",
			target: filepath.Join(base, "..", "escape.txt"),
			want:   false,
		},
		{
			name:   "deep parent traversal is rejected",
			target: filepath.Join(base, "sub", "..", "..", "..", "escape.txt"),
			want:   false,
		},
		{
			name:   "sibling directory with colliding name prefix is rejected",
			target: filepath.Join(filepath.Dir(base), filepath.Base(base)+"-sibling", "file.txt"),
			want:   false,
		},
		{
			name:   "file inside base literally named with leading dots is contained",
			target: filepath.Join(base, "..config", "file.txt"),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := Contains(base, tt.target)

			// Assert
			assert.Equal(t, tt.want, got)
		})
	}
}
