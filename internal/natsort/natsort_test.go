package natsort

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var benchData = []string{
	"photo-001.jpg", "PHOTO-1.jpg", "photo-01.jpg", "photo-10.jpg",
	"photo-2.jpg", "photo-02.jpg", "photo-11.jpg", "photo-3.jpg",
}

func BenchmarkCompareIgnoreCase(b *testing.B) {
	for b.Loop() {
		for i := 0; i < len(benchData)-1; i++ {
			_ = CompareIgnoreCase(benchData[i], benchData[i+1])
		}
	}
}

func TestCompareIgnoreCase(t *testing.T) {
	// Arrange

	// The test cases for the CompareIgnoreCase function.
	tests := []struct {
		a, b     string
		want int
	}{
		// 1. Basic numeric comparisons.
		{"rfc1.txt", "rfc2.txt", -1},
		{"rfc2.txt", "rfc10.txt", -1},
		{"rfc10.txt", "rfc20.txt", -1},

		// 2. Case-insensitivity assertions.
		{"RFC2.txt", "rfc10.txt", -1},
		{"rfc2.txt", "RFC10.txt", -1},
		{"A", "a", 0}, // Should evaluate as equal under ignore-case.

		// 3. The leading zeroes paradox.
		// More leading zeroes = smaller/comes first (fractional approach).
		{"file1.txt", "file01.txt", 1},    // file01.txt comes first, so file1.txt > file01.txt
		{"file01.txt", "file001.txt", 1},  // file001.txt comes first, so file01.txt > file001.txt
		{"1.001", "1.01", -1},             // 1.001 comes first, so 1.001 < 1.01

		// 4. Fractional tokens vs literal decimals.
		// Handle dot-sequences as distinct text+digit sets, not floating mathematical decimals.
		{"version-1.2", "version-1.10", -1},

		// 5. Special symbols and string prefixes.
		{"abc", "abcdef", -1},
		{"file-2.txt", "file-10.txt", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+" vs "+tt.b, func(t *testing.T) {
			// Act
			got := CompareIgnoreCase(tt.a, tt.b)

			// Normalize outputs to -1, 1, or 0 for comparative precision.
			if got < 0 {
				got = -1
			} else if got > 0 {
				got = 1
			}

			// Assert
			assert.Equal(t, tt.want, got)
		})
	}
}
