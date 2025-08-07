package model

type Param struct {
	Name        string
	BindMethod  string
	ParamType   string
	IsRequired  string
	Description string
}
type ResponseBlock struct {
	StructType string
	SchemaType string // {object}, string, int, etc
	StatusCode int
	IsSuccess  bool
	AcceptType string // json, xml, string
}

type CommentBlock struct {
	Summary        string          // same as function name
	Description    string          // same as function name with better formating
	Tags           string          // `___` by default
	Accept         string          // info from the binding param
	Produce        []string        // info from response
	Params         []Param         // info from the binding
	ResponseBlocks []ResponseBlock // info from the  response
	Router         string          // info from the registration
}
