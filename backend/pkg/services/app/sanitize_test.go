// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"strings"
	"testing"
)

func TestSanitizeSVG(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "valid SVG",
			input: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
						<rect x="10" y="10" width="80" height="80" fill="#ff0000"/>
					</svg>`,
			expected: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
						<rect x="10" y="10" width="80" height="80" fill="#ff0000"/>
					</svg>`,
		},
		{
			name: "remove script elements",
			input: `<svg>
						<script>alert('XSS')</script>
						<circle cx="50" cy="50" r="40"/>
					</svg>`,
			expected: `<svg>
						
						<circle cx="50" cy="50" r="40"/>
					</svg>`,
		},
		{
			name: "strip dangerous attributes",
			input: `<svg>
						<path d="M10 10" onclick="malicious()" stroke="black"/>
					</svg>`,
			expected: `<svg>
						<path d="M10 10" stroke="black"/>
					</svg>`,
		},
		{
			name: "valid attribute with space-separated tokens",
			input: `<svg>
						<circle cx="50" cy="50" r="30" stroke="black" stroke-linecap="round"/>
					</svg>`,
			expected: `<svg>
						<circle cx="50" cy="50" r="30" stroke="black" stroke-linecap="round"/>
					</svg>`,
		},
		{
			name: "preserve namespace attributes",
			input: `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
						<image xlink:href="legal.png"/>
					</svg>`,
			expected: `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
						<image xlink:href="legal.png"/>
					</svg>`,
		},
		{
			name: "mixed content",
			input: `<svg>
						<rect x="10" y="10" width="80" height="80" fill="blue"/>
						<script>evil()</script>
						<circle onload="danger" r="30"/>
					</svg>`,
			expected: `<svg>
						<rect x="10" y="10" width="80" height="80" fill="blue"/>
						
						<circle r="30"/>
					</svg>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := sanitizeSVG([]byte(tt.input))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			got := normalizeWhitespace(string(output))
			want := normalizeWhitespace(tt.expected)

			if got != want {
				t.Errorf("Output mismatch\nGot:\n%s\nWant:\n%s", got, want)
			}
		})
	}
}

func normalizeWhitespace(s string) string {
	s = strings.TrimSpace(s)

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	return strings.Join(lines, "")
}
