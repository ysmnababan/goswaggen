package model

type StructField struct {
	Name      string
	VarType   string
	Tag       map[string]string
	IsPointer bool
}

type PayloadInfo struct {
	ParamTypes string // <context>.Bind(<param> ParamTypes)
	BasicLit   string // for queryparam and param args, e.g. <context>.Param("this")
	BindMethod string

	// For storing all the field from a struct when
	// calling the `Bind()` binding function.
	// Depends on the BindMethod and HTTP method
	FieldLists []*StructField
}

func (i *PayloadInfo) GetAcceptTag() string {
	if i.BindMethod != "Bind" {
		return ""
	}
	// Default value is `json`
	// The `accept` tag depends on the Content-Type of the request.
	// So it is not possible to know the exact Content-Type.
	return "json"
}
