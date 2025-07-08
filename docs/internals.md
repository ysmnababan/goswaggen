# How It Works: Internals of goswaggen

`goswaggen` parses your Go project to automatically generate Swagger annotations for your HTTP handler functions. Here's how it works under the hood:

## High-Level Flow

1. **Parse Go Files**
   - Uses the `go/parser` and `go/ast` packages to walk through each `.go` file recursively under the provided directory.

2. **Detect Target Functions**
   - Looks for methods that:
     - Are exported
     - Belong to a struct that matches the naming pattern (e.g., `*Controller`)
     - Take `echo.Context` or `gin.Context` as a parameter

3. **Generate Comment Blocks**
   - Uses Go's `token` package to determine the line above each function.
   - Constructs Swagger-compatible annotations like:
     ```go
     // @Summary Get user by ID
     // @Tags users
     // @Param id path int true "User ID"
     // @Success 200 {object} dto.UserResponse
     // @Failure 404 {object} dto.ErrorResponse
     ```

4. **Insert or Update Comments**
   - Rewrites the file using `go/printer` with the new comment block injected above the function.
   - Optional dry-run to preview changes.

## File Structure

- `ast/parser.go`: AST traversal and function detection
- `generator/comment.go`: Swagger comment template generator
- `injector/writer.go`: Safely writes updated file with annotations
