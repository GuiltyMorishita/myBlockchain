package main

import (
	"github.com/GuiltyMorishita/myBlockchain/blockchain"
	"github.com/GuiltyMorishita/myBlockchain/handler"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/satori/go.uuid"
)

func main() {
	e := echo.New()

	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	h := handler.Handler{
		Bc:             blockchain.NewBlockchain(),
		NodeIdentifire: uuid.NewV4().String(),
	}

	e.POST("/transactions/new", h.CreateTransaction)
	e.GET("/mine", h.Mine)
	e.GET("/chain", h.FullChain)
	e.POST("/nodes/register", h.RegisterNode)
	e.GET("/nodes/resolve", h.Consensus)
	e.Logger.Fatal(e.Start(":5000"))
}
