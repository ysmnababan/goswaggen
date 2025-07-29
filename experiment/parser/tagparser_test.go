package main

import (
	"reflect"
	"testing"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "Basic two tags",
			input: `json:"name" validate:"required"`,
			expected: map[string]string{
				"json":     "name",
				"validate": "required",
			},
		},
		{
			name:  "Single tag",
			input: `xml:"Name"`,
			expected: map[string]string{
				"xml": "Name",
			},
		},
		{
			name:     "Empty string",
			input:    ``,
			expected: map[string]string{},
		},
		{
			name:  "Escaped quotes in value",
			input: `desc:"A \"quoted\" value"`,
			expected: map[string]string{
				"desc": `A \"quoted\" value`,
			},
		},
		{
			name:  "Tag with spaces",
			input: `json:"full name" validate:"not empty"`,
			expected: map[string]string{
				"json":     "full name",
				"validate": "not empty",
			},
		},
		{
			name:  "Tag with escaped backslash",
			input: `note:"C:\\path\\to\\file"`,
			expected: map[string]string{
				"note": `C:\\path\\to\\file`,
			},
		},
		{
			name:  "Tag with colon inside value",
			input: `meta:"type:v1"`,
			expected: map[string]string{
				"meta": `type:v1`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTag(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseTag(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
