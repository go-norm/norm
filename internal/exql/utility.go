// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"regexp"
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

func separateByAS(s string) []string {
	// The minimum expression with the AS keyword is "x AS y", 6 chars.
	if len(s) < 6 {
		return []string{s}
	}

	out := make([]string, 0, 2)
	start, lim := 0, len(s)-1
	for start <= lim {
		var end int
		for end = start; end <= lim; end++ {
			if end > 3 && isBlankSymbol(s[end]) && isBlankSymbol(s[end-3]) {
				if (s[end-1] == 's' || s[end-1] == 'S') && (s[end-2] == 'a' || s[end-2] == 'A') {
					break
				}
			}
		}

		if end < lim {
			out = append(out, trimString(s[start:end-3]))
		} else {
			out = append(out, trimString(s[start:end]))
		}

		start = end + 1
	}
	return out
}

func separateBySpace(s string) []string {
	if len(s) == 0 {
		return []string{""}
	}

	pre := strings.Split(s, " ")
	out := make([]string, 0, len(pre))
	for i := range pre {
		pre[i] = trimString(pre[i])
		if pre[i] != "" {
			out = append(out, pre[i])
		}
	}
	return out
}

var InvisibleCharsRegexp = regexp.MustCompile(`[\s\r\n\t]+`)

/*
// Separates by a comma, ignoring spaces too.
// This was slower than strings.Split.
func separateByComma(in string) (out []string) {

	out = []string{}

	start, lim := 0, len(in)-1

	for start < lim {
		var end int

		for end = start; end <= lim; end++ {
			// Is a comma?
			if in[end] == ',' {
				break
			}
		}

		out = append(out, trimString(in[start:end]))

		start = end + 1
	}

	return
}
*/

// Separates by a comma, ignoring spaces too.
func separateByComma(in string) []string {
	out := strings.Split(in, ",")
	for i := range out {
		out[i] = trimString(out[i])
	}
	return out
}
