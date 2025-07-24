package dummyhandler

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func JustDummyHandler(c echo.Context) error {
	fmt.Println("do nothing")
	return nil
}
