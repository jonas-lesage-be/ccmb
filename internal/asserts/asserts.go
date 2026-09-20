package asserts

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

const epsilonF64 = 1.0 / (1 << 52)

// RelativeEq asserts that two float64 values are approximately equal within a relative tolerance.
func RelativeEq(t *testing.T, actual, expected float64) {
	t.Helper()
	maxVal := math.Max(math.Abs(actual), math.Abs(expected))
	if maxVal < 1.0 {
		maxVal = 1.0
	}

	delta := epsilonF64 * maxVal
	assert.InDelta(t, expected, actual, delta)
}
