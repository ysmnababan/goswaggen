package inspector

import (
	"go/ast"
)

type Inspector interface {
	Inspect(ast.Node)
}

