package main

var IMPORT_PATH_VALUE = map[string]string{
	"echo": "github.com/labstack/echo/v4",
}

var OBJECT_TYPE_IMPORT = map[string]string{
	"ECHO":       "*github.com/labstack/echo/v4.Echo",
	"ECHO_GROUP": "*github.com/labstack/echo/v4.Group",
}
var ECHO_VARIABLE_TYPE = "*github.com/labstack/echo/v4.Echo"
var ECHO_GROUP_VARIABLE_TYPE = "*github.com/labstack/echo/v4.Group"
var ECHO_CONTEXT_TYPE = "github.com/labstack/echo/v4.Context"

var ECHO_REQUEST_DATA_METHOD = map[string]bool{
	"Bind":       true,
	"QueryParam": true,
	"Param":      true,
}
