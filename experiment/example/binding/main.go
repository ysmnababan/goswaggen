package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Request struct {
	Name     string
	Email    string
	Personal Nested
}

type Nested struct {
	Age   int
	Hobby string
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/test-post", postHandler)
	e.Logger.Fatal(e.Start(":1323"))
}

func postHandler(c echo.Context) error {
	req := new(Request)
	err := c.Bind(&req.Personal.Age)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Request", *req)
	fmt.Println("Nested", req.Personal)
	return nil
}
