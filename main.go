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

	nodeIdentifire := uuid.NewV4().String()
	blockchain := blockchain.NewBlockchain()

	e.POST("/transactions/new", handler.CreateTransaction(blockchain))
	e.GET("/mine", handler.Mine(blockchain, nodeIdentifire))
	e.GET("/chain", handler.FullChain(blockchain))
	e.POST("/nodes/register", handler.RegisterNode(blockchain))
	e.GET("/nodes/resolve", handler.Consensus(blockchain))
	e.Logger.Fatal(e.Start(":5000"))
}
