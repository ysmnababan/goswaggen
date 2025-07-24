package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	ef "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"learn-go/config"
	"learn-go/dummyhandler"

	// "learn-go/internals/app/example_feat"
	"learn-go/internals/factory"
	middleware "learn-go/internals/middleware"
	"learn-go/internals/pkg/database"
	"learn-go/internals/pkg/logger"
	httpserver "learn-go/internals/server"
	"learn-go/internals/utils/env"
)

func init() {
	selectedEnv := config.Env()
	env := env.NewEnv()
	env.Load(`.env`)
	logger.InitLogger()
	log.Info().Msg("Choosen environment " + selectedEnv)
}

func handlerTest(c ef.Context) error {
	return nil
}

// @title learn-go-Project
// @version 0.0.1
// @description This is a doc for learn-go-Project

// @securityDefinitions.apikey Authorization
// @in header
// @name Authorization
func main() {
	cfg := config.Get()

	port := cfg.App.Port

	logLevel, err := zerolog.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	database.Init("std")

	f := factory.NewFactory()
	_ = f
	e := ef.New()
	e.HideBanner = true
	e.IPExtractor = ef.ExtractIPDirect()
	middleware.Init(e, f.Redis)
	
	e.GET("/test", func(c ef.Context) error {
		return nil
	})
	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal("shutting down the server")
	}
	first_group := e.Group("/first")
	first_group.GET("/TEST1", HandlerForFirstGroup)
	
	e.GET("/TEST2", handlerTest)
	
	second_group := first_group.Group("/second")
	second_group.GET("/TEST3", HandlerForSecondGroup)
	
	e.POST("/TEST4", dummyhandler.JustDummyHandler)
	NotRegisterEcho("hehe")
	RegisterEcho(e, "ignore-this")
	RegisterEchoWithGroup(first_group)
	RegisterEchoSelectorAsParamGroup(first_group.Group("/second"))
	RegisterEchoGroupFromEcho(e.Group("/first"))
	httpserver.Init(e, f)
}

func NotRegisterEcho(str string) {
	fmt.Println(str)
}
func DummyHandlerInsideFunc(c ef.Context) error {
	return nil
}

func RegisterEcho(c *ef.Echo, somedata string) {
	c.PUT("/TEST5", DummyHandlerInsideFunc)

	third_group := c.Group("/third")
	third_group.GET("/TEST6", HandlerForThirdGroup)

	fourth_group := third_group.Group("/fourth")
	fourth_group.GET("/TEST7", HandlerForFourthGroup)
}

func HandlerForFirstGroup(c ef.Context) error {
	return nil
}

func HandlerForSecondGroup(c ef.Context) error {
	return nil
}

func HandlerForThirdGroup(c ef.Context) error {
	return nil
}

func HandlerForFourthGroup(c ef.Context) error {
	return nil
}

func HandlerFor5thGroup(c ef.Context) error {
	return nil
}
func Handler6(c ef.Context) error {
	return nil
}
func Handler7(c ef.Context) error {
	return nil
}
func Handler8(c ef.Context) error {
	return nil
}
func Handler9(c ef.Context) error {
	return nil
}
func Handler10(c ef.Context) error {
	return nil
}
func RegisterEchoWithGroup(g *echo.Group) {
	g.GET("/TEST8", HandlerFor5thGroup)
	fifth_Group := g.Group("/fifth") // should be /first/fifth
	fifth_Group.GET("/TEST9", Handler6)
}

func RegisterEchoSelectorAsParamGroup(g *echo.Group) {
	g.GET("/TEST10", Handler7)
	sixth_group := g.Group("/sixth") // should be /first/second/sixth
	sixth_group.GET("/TEST11", Handler8)
}

func RegisterEchoGroupFromEcho(g *echo.Group) {
	wrapper(g.Group("/wrapper"))
}

func wrapper(g *echo.Group) {
	g.GET("/TEST12", Handler9)
	seventh_group := g.Group("/seventh") // should be /first/wrapper/second/seventh
	seventh_group.GET("/TEST13", Handler10)
}
