package model

type Param struct {
	Name        string
	BindMethod  string
	ParamType   string
	IsRequired  bool
	Description string
}

type CommentBlock struct {
	Summary     string   // same as function name
	Description string   // same as function name with better formating
	Tags        string   // `___` by default
	Accept      string   // info from the binding param
	Produce     []string // info from response
	Params      []string // info from the binding
	Response    []string // info from the  response
	Router      string   // info from the registration
}
