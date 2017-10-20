package handler

import (
	"fmt"
	"net/http"

	"github.com/GuiltyMorishita/myBlockchain/blockchain"
	"github.com/labstack/echo"
)

type Response map[string]interface{}

func CreateTransaction(bc *blockchain.BlockChain) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		t := new(blockchain.Transaction)
		if err = c.Bind(t); err != nil {
			return
		}
		index := bc.NewTransaction(t.Sender, t.Recipient, t.Amount)
		return c.JSON(http.StatusCreated, Response{
			"message": fmt.Sprintf("Transaction will be added to Block %d", index),
		})
	}
}

func Mine(bc *blockchain.BlockChain, nodeIdentifire string) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		lastBlock := bc.LastBlock()
		lastProof := lastBlock.Proof
		proof := bc.ProofOfWork(lastProof)

		bc.NewTransaction("0", nodeIdentifire, 1)

		newBlock := bc.NewBlock(proof, "")

		return c.JSON(http.StatusOK, Response{
			"message":      "New Block Mined",
			"index":        newBlock.Index,
			"transactions": newBlock.Transactions,
		})
	}
}

func FullChain(bc *blockchain.BlockChain) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		return c.JSON(http.StatusOK, Response{
			"chain":  bc.Chain,
			"length": len(bc.Chain),
		})
	}
}

func RegisterNode(bc *blockchain.BlockChain) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		type body struct {
			Nodes []string `json:"nodes"`
		}
		b := new(body)
		if err = c.Bind(b); err != nil {
			return
		}
		for _, node := range b.Nodes {
			bc.RegisterNode(node)
		}

		return c.JSON(http.StatusCreated, Response{
			"message": "New nodes have been added",
			"nodes":   bc.Nodes,
		})
	}
}

func Consensus(bc *blockchain.BlockChain) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		message := "Our chain is authoritative"
		if bc.ResolveConflicts() {
			message = "Our chain was replaced"
		}
		return c.JSON(http.StatusOK, Response{
			"message": message,
			"chain":   bc.Chain,
		})
	}
}
