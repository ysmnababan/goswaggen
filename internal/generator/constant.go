package generator

var (
	DEFAULT_PARAM_DESCRIPTION            = "\"change this description\""
	DEFAULT_QUERY_PARAM_DESCRIPTION      = "\"change this description\""
	DEFAULT_BODY_DESCRIPTION             = "\"change this description\""
	DEFAULT_RESPONSE_SCHEME_TYPE         = "object"
	DEFAULT_FAILURE_RESPONSE_DESCRIPTION = "\"error\""
	DEFAULT_SUCCESS_RESPONSE_DESCRIPTION = "\"success\""
	RESPONSE_BLOCK_TEMPLATE              = "// @%s %d %s %s %s"
	GO_TO_SWAGGO_SCHEME_TYPES_MAP        = map[string]string{
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
)

var (
	ROUTER_TEMPLATE                  = "// @Router %s [%s]"
	ROUTER_COMMENT_BLOCK_PREFIX      = "@Router "
	PRODUCE_TEMPLATE                 = "// @Produce %s"
	SUMMARY_COMMENT_BLOCK_PREFIX     = "@Summary"
	DESCRIPTION_COMMENT_BLOCK_PREFIX = "@Description"
	TAGS_COMMENT_BLOCK_PREFIX        = "@Tags"
	DOCS_TEMPLATE                    = "// %s handles %s %s"
)
