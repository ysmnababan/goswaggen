package generator

var DEFAULT_PARAM_DESCRIPTION = "\"change this description\""
var DEFAULT_QUERY_PARAM_DESCRIPTION = "\"change this description\""
var DEFAULT_BODY_DESCRIPTION = "\"change this description\""
var DEFAULT_RESPONSE_SCHEME_TYPE = "object"
var DEFAULT_FAILURE_RESPONSE_DESCRIPTION = "\"error\""
var DEFAULT_SUCCESS_RESPONSE_DESCRIPTION = "\"success\""
var RESPONSE_BLOCK_TEMPLATE = "// @%s %d {%s} %s %s"
var GO_TO_SWAGGO_SCHEME_TYPES_MAP = map[string]string{
	"bool":      "boolean",
	"string":    "string",
	"int":       "integer",
	"int8":      "integer",
	"int16":     "integer",
	"int32":     "integer",
	"int64":     "integer",
	"uint":      "integer",
	"uint8":     "integer",
	"uint16":    "integer",
	"uint32":    "integer",
	"uint64":    "integer",
	"float32":   "number",
	"float64":   "number",
	"[]byte":    "string", // Usually encoded as base64 strings
	"time.Time": "string", // Formatted datetime string
	"file":      "file",   // For file uploads/downloads
	// And for structured types:
	"struct": "object",
	"map":    "object",
	"slice":  "array",
	"json":   "object",
	// Add your custom types as needed:
	// "MyModel": "object",
}

var ROUTER_COMMENT_BLOCK_PREFIX = "@Router "
var ROUTER_TEMPLATE = "// @Router %s [%s]"
var PRODUCE_TEMPLATE = "// @Produce %s"
