# Design Decisions

## Why Use Go AST?
We chose Go's `go/ast` package over simpler regex-based parsing because:
- AST provides structured access to identifiers, function declarations, and comments.
- It ensures that only valid Go code is processed.
- Avoids brittle hacks or unsafe replacements.

## Why Add Comments Above the Function?
Swagger tooling (like Swaggo) parses function declarations and expects annotations to be directly above them. This is consistent with GoDoc and makes it easier for developers to read inline.

## Why Ignore Private Functions?
Private (unexported) functions are typically not public API endpoints, so we exclude them by default to avoid noise.

## Why No YAML/JSON Output?
This tool focuses on inline code annotations. Supporting OpenAPI YAML export would require a different approach (and possibly a separate tool).
