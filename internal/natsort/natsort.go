package natsort

import (
	"unicode"
	"unicode/utf8"
)

// CompareIgnoreCase compares two strings in a natural order, ignoring case differences.
// This function uses Go inlining to optimize performance by avoiding function call overhead.
func CompareIgnoreCase(a, b string) int { // NOSONAR: Eliminate function call overhead.
	lenA, lenB := len(a), len(b)
	i, j := 0, 0

	for i < lenA && j < lenB {
		bA := a[i]
		bB := b[j]

		// Fast path for standard ASCII characters (< 128).
		if bA < 128 && bB < 128 {
			cA := rune(bA)
			if bA >= 'A' && bA <= 'Z' { cA |= 0x20 }
			cB := rune(bB)
			if bB >= 'A' && bB <= 'Z' { cB |= 0x20 }

			// Numeric mode activation.
			if cA >= '0' && cA <= '9' && cB >= '0' && cB <= '9' {
				startA := i
				startB := j

				// Determine if either numeric block starts with '0' (fractional rule).
				isFraction := bA == '0' || bB == '0'

				// Scan the entire consecutive numeric slice.
				for i < lenA && a[i] >= '0' && a[i] <= '9' { i++ }
				for j < lenB && b[j] >= '0' && b[j] <= '9' { j++ }

				numStrA := a[startA:i]
				numStrB := b[startB:j]

				if isFraction {
					// Lexicographical comparison for fractions.
					if numStrA < numStrB { return -1 }
					if numStrA > numStrB { return 1 }
				} else {
					// Standard numbers: compare length first, then content
					if len(numStrA) < len(numStrB) { return -1 }
					if len(numStrA) > len(numStrB) { return 1 }
					if numStrA < numStrB { return -1 }
					if numStrA > numStrB { return 1 }
				}
			} else {
				// Standard character comparison.
				if cA != cB {
					if cA < cB { return -1 }
					return 1
				}
				i++
				j++
			}
		} else {
			// Fallback for non-ASCII UTF-8 characters.
			cA, sizeA := utf8.DecodeRuneInString(a[i:])
			cB, sizeB := utf8.DecodeRuneInString(b[j:])

			cA = unicode.ToLower(cA)
			cB = unicode.ToLower(cB)

			if cA >= '0' && cA <= '9' && cB >= '0' && cB <= '9' {
				startA := i
				startB := j
				isFraction := a[i] == '0' || b[j] == '0'
				for i < lenA && a[i] >= '0' && a[i] <= '9' { i++ }
				for j < lenB && b[j] >= '0' && b[j] <= '9' { j++ }
				numStrA := a[startA:i]
				numStrB := b[startB:j]

				if isFraction {
					if numStrA < numStrB { return -1 }
					if numStrA > numStrB { return 1 }
				} else {
					if len(numStrA) < len(numStrB) { return -1 }
					if len(numStrA) > len(numStrB) { return 1 }
					if numStrA < numStrB { return -1 }
					if numStrA > numStrB { return 1 }
				}
			} else {
				if cA != cB {
					if cA < cB { return -1 }
					return 1
				}
				i += sizeA
				j += sizeB
			}
		}
	}

	if lenA == lenB { return 0 }
	if lenA < lenB { return -1 }
	return 1
}
