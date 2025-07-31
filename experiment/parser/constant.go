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

// 0 : no status code
// 1 : status code in 1st param
var ECHO_FRAMEWORK_STANDARD_RESPONSE = map[string]int{
	"String":     1,
	"HTML":       1,
	"HTMLBlob":   1,
	"JSON":       1,
	"JSONPretty": 1,
	"JSONBlob":   1,
	"JSONP":      1,
	"XML":        1,
	"XMLPretty":  1,
	"XMLBlob":    1,
	"File":       1,
	"Attachment": 0,
	"Inline":     0,
	"Blob":       1,
	"Stream":     1,
	"NoContent":  1,
	"Redirect":   1,
}
