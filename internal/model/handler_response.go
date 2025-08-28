package model

type ReturnResponse struct {
	// ReturnStmt     *ast.ReturnStmt
	ReturnDataType string
	SchemaType     string // {object}, {string}, {number}, etc
	StatusCode     int
	IsSuccess      bool
	ProduceType    string // json, xml, string
}
