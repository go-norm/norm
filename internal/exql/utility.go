// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
)

// isBlankSymbol returns true if the given byte is either space, tab, carriage
// return or newline.
func isBlankSymbol(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// trimString returns a slice of s with a leading and trailing blank symbols
// (as defined by isBlankSymbol) removed.
func trimString(s string) string {
	start, end := 0, len(s)-1
	if end < start {
		return ""
	}

	for isBlankSymbol(s[start]) {
		start++
		if start >= end {
			return ""
		}
	}

	for isBlankSymbol(s[end]) {
		end--
	}
	return s[start : end+1]
}

func separateByAS(in string) []string {
	if len(in) < 6 {
		// The minimum expression with the AS keyword is "x AS y", 6 chars.
		return []string{in}
	}

	out := make([]string, 0, 2)
	start, lim := 0, len(in)-1
	for start <= lim {
		var end int
		for end = start; end <= lim; end++ {
			if end > 3 && isBlankSymbol(in[end]) && isBlankSymbol(in[end-3]) {
				if (in[end-1] == 's' || in[end-1] == 'S') && (in[end-2] == 'a' || in[end-2] == 'A') {
					break
				}
			}
		}

		if end < lim {
			out = append(out, trimString(in[start:end-3]))
		} else {
			out = append(out, trimString(in[start:end]))
		}

		start = end + 1
	}
	return out
}

// Separates by spaces, ignoring spaces too.
func separateBySpace(in string) []string {
	if len(in) == 0 {
		return []string{""}
	}

	pre := strings.Split(in, " ")
	out := make([]string, 0, len(pre))
	for i := range pre {
		pre[i] = trimString(pre[i])
		if pre[i] != "" {
			out = append(out, pre[i])
		}
	}
	return out
}
