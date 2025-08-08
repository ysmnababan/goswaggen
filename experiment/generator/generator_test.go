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

func TestProcessPayload(t *testing.T) {
	qp := model.PayloadInfo{
		BasicLit:   "name",
		BindMethod: "QueryParam",
	}
	p := model.PayloadInfo{
		BasicLit:   "id",
		BindMethod: "Param",
	}
	body := model.PayloadInfo{
		BindMethod: "Bind",
		ParamTypes: "myPkg.MyStruct",
		BasicLit:   "request",
	}
	tests := []struct {
		name   string
		method string
		pi     model.PayloadInfo

		wants []model.Param
	}{
		{
			name:   "query post",
			method: "POST",
			pi:     qp,
			wants: []model.Param{
				{
					Name:        "name",
					BindMethod:  "query",
					ParamType:   "string",
					IsRequired:  true,
					Description: DEFAULT_QUERY_PARAM_DESCRIPTION,
				},
			},
		},
		{
			name:   "query get",
			method: "GET",
			pi:     qp,
			wants: []model.Param{
				{
					Name:        "name",
					BindMethod:  "query",
					ParamType:   "string",
					IsRequired:  true,
					Description: DEFAULT_QUERY_PARAM_DESCRIPTION,
				},
			},
		},
		{
			name:   "param put",
			method: "PUT",
			pi:     p,
			wants: []model.Param{
				{
					Name:        "id",
					BindMethod:  "path",
					ParamType:   "string",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
		},
		{
			name:   "param delete",
			method: "DELETE",
			pi:     p,
			wants: []model.Param{
				{
					Name:        "id",
					BindMethod:  "path",
					ParamType:   "string",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
		},
		{
			name:   "param PATCH",
			method: "PATCH",
			pi:     body,
			wants: []model.Param{
				{
					Name:        "request",
					BindMethod:  "body",
					ParamType:   "myPkg.MyStruct",
					IsRequired:  true,
					Description: DEFAULT_BODY_DESCRIPTION,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processPayload(&tt.pi, tt.method)
			for i, want := range tt.wants {
				assert.Equal(t, want.Name, got[i].Name)
				assert.Equal(t, want.BindMethod, got[i].BindMethod)
				assert.Equal(t, want.ParamType, got[i].ParamType)
				assert.Equal(t, want.IsRequired, got[i].IsRequired)
				assert.Equal(t, want.Description, got[i].Description)
			}
		})
	}
}

func TestProcessPayload_WhenUsingBind(t *testing.T) {
	noTag := map[string]string{}
	tagWithParam := map[string]string{
		"param": "id",
	}
	tagWithQuery := map[string]string{
		"query": "email",
	}
	tagQueryWithRequired := map[string]string{
		"query":    "role",
		"validate": "required",
	}
	tagWithRequired := map[string]string{
		"validate": "required",
	}
	multipleTag := map[string]string{
		"query":    "query-email",
		"param":    "id-param",
		"validate": "required",
		"json":     "email",
		"gorm":     "email",
		"xml":      "email",
	}
	tests := []struct {
		name   string
		method string
		fields []*model.StructField

		wants   []model.Param
		wantLen int
	}{
		{
			name:    "GET with no field",
			method:  "GET",
			fields:  []*model.StructField{},
			wantLen: 0,
		},
		{
			name:   "GET with no tag",
			method: "GET",
			fields: []*model.StructField{
				{
					Name:      "field1",
					VarType:   "string",
					IsPointer: false,
					Tag:       noTag,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
				},
			},
			wantLen: 0,
		},
		{
			name:   "GET with param tag",
			method: "GET",
			fields: []*model.StructField{
				{
					Name:      "field",
					VarType:   "int",
					IsPointer: false,
					Tag:       tagWithParam,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
					Tag:       noTag,
				},
			},
			wants: []model.Param{
				{
					Name:        "id",
					BindMethod:  "path",
					ParamType:   "int",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
			wantLen: 1,
		},
		{
			name:   "GET with query tag",
			method: "GET",
			fields: []*model.StructField{
				{
					Name:      "field",
					VarType:   "string",
					IsPointer: false,
					Tag:       tagWithQuery,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
					Tag:       noTag,
				},
			},
			wants: []model.Param{
				{
					Name:        "email",
					BindMethod:  "query",
					ParamType:   "string",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
			wantLen: 1,
		},
		{
			name:   "GET with multiple tag",
			method: "GET",
			fields: []*model.StructField{
				{
					Name:      "field",
					VarType:   "float32",
					IsPointer: false,
					Tag:       multipleTag,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
					Tag:       noTag,
				},
			},
			wants: []model.Param{
				{
					Name:        "query-email",
					BindMethod:  "query",
					ParamType:   "float32",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
			wantLen: 1,
		},
		{
			name:   "DELETE with required tag, is pointer false",
			method: "DELETE",
			fields: []*model.StructField{
				{
					Name:      "field",
					VarType:   "int",
					IsPointer: false,
					Tag:       tagQueryWithRequired,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
					Tag:       tagWithRequired,
				},
			},
			wants: []model.Param{
				{
					Name:        "role",
					BindMethod:  "query",
					ParamType:   "int",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
			wantLen: 1,
		},
		{
			name:   "DELETE with required tag, is pointer true",
			method: "DELETE",
			fields: []*model.StructField{
				{
					Name:      "field",
					VarType:   "int",
					IsPointer: true,
					Tag:       tagQueryWithRequired,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
					Tag:       tagWithRequired,
				},
			},
			wants: []model.Param{
				{
					Name:        "role",
					BindMethod:  "query",
					ParamType:   "int",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
			wantLen: 1,
		},
		{
			name:   "DELETE with no required tag, is pointer false",
			method: "DELETE",
			fields: []*model.StructField{
				{
					Name:      "field",
					VarType:   "int",
					IsPointer: false,
					Tag:       tagWithQuery,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
					Tag:       tagWithRequired,
				},
			},
			wants: []model.Param{
				{
					Name:        "email",
					BindMethod:  "query",
					ParamType:   "int",
					IsRequired:  true,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
			wantLen: 1,
		},
		{
			name:   "DELETE with no required tag, is pointer true",
			method: "DELETE",
			fields: []*model.StructField{
				{
					Name:      "field",
					VarType:   "int",
					IsPointer: true,
					Tag:       tagWithQuery,
				},
				{
					Name:      "field2",
					VarType:   "string",
					IsPointer: false,
					Tag:       tagWithRequired,
				},
			},
			wants: []model.Param{
				{
					Name:        "email",
					BindMethod:  "query",
					ParamType:   "int",
					IsRequired:  false,
					Description: DEFAULT_PARAM_DESCRIPTION,
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := model.PayloadInfo{
				BindMethod: "Bind",
				ParamTypes: "myPkg.MyStruct",
				BasicLit:   "request",
				FieldLists: tt.fields,
			}

			got := processPayload(&body, tt.method)
			for i, want := range tt.wants {
				if tt.wantLen == 0 {
					assert.Equal(t, 0, len(got))
					continue
				}
				assert.Equal(t, want.Name, got[i].Name)
				assert.Equal(t, want.BindMethod, got[i].BindMethod)
				assert.Equal(t, want.ParamType, got[i].ParamType)
				assert.Equal(t, want.IsRequired, got[i].IsRequired)
				assert.Equal(t, want.Description, got[i].Description)
			}
		})
	}
}
