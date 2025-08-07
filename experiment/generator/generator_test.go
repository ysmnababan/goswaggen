package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelCaseToTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "MultipleWords",
			input: "CreateCommentBlock",
			want:  "Create comment block",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single word",
			input: "Login",
			want:  "Login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, camelCaseToTitle(tt.input))
		})
	}
}
