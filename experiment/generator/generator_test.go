package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmnababan/goswaggen/experiment/generator/model"
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

func TestSetAccept(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		BindMethod string

		want string
	}{
		{
			name:       "get",
			method:     "GET",
			BindMethod: "Bind",

			want: "",
		},
		{
			name:       "delete",
			method:     "DELETE",
			BindMethod: "Bind",

			want: "",
		},
		{
			name:       "post",
			method:     "POST",
			BindMethod: "Bind,Query,Param",

			want: "json",
		},
		{
			name:       "PUT",
			method:     "PUT",
			BindMethod: "Bind,Bind,Bind",

			want: "json",
		},
		{
			name:       "PATCH",
			method:     "PATCH",
			BindMethod: "Bind",

			want: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := strings.Split(tt.BindMethod, ",")
			payloads := []*model.PayloadInfo{}
			for _, m := range methods {
				payloads = append(payloads, &model.PayloadInfo{
					BindMethod: m,
				})
			}
			g := &generator{
				method:   tt.method,
				payloads: payloads,
			}

			// execute
			g.commentBlock = &model.CommentBlock{}
			g.setAccept()

			// assert
			if tt.method == "GET" || tt.method == "DELETE" {
				assert.Equal(t, "", g.commentBlock.Accept)
			} else {
				assert.Contains(t, g.commentBlock.Accept, tt.want)
			}
		})
	}
}
