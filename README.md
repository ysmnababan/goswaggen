## 🛠️ goswaggen

**goswaggen** is a command-line tool that helps you automatically generate Swagger (OpenAPI) comment annotations for your Go HTTP handlers.

It reduces the manual effort of writing documentation by analyzing your code and generating structured comments that tools like [`swaggo/swag`](https://github.com/swaggo/swag) can parse to build your API docs.

### 💡 Key Features

- Parses your Go handler functions
- Generates Swagger-compatible comment blocks above each handler
- Compatible with popular frameworks like `net/http`, `Echo`, or `Gin`
- Simple CLI usage — integrates into your workflow
