package context

import "golang.org/x/tools/go/packages"

type HandlerContext interface {
	GetPackage() *packages.Package
}
