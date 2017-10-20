package handler

import (
	"fmt"
	"net/http"

	"github.com/GuiltyMorishita/myBlockchain/blockchain"
	"github.com/labstack/echo"
)

type (
	Handler struct {
		Bc             *(blockchain.BlockChain)
		NodeIdentifire string
	}
	Response map[string]interface{}
)

func (h *Handler) CreateTransaction(c echo.Context) (err error) {
	t := new(blockchain.Transaction)
	if err = c.Bind(t); err != nil {
		return
	}
	index := h.Bc.NewTransaction(t.Sender, t.Recipient, t.Amount)
	return c.JSON(http.StatusCreated, Response{
		"message": fmt.Sprintf("Transaction will be added to Block %d", index),
	})
}

func (h *Handler) Mine(c echo.Context) (err error) {
	lastBlock := h.Bc.LastBlock()
	lastProof := lastBlock.Proof
	proof := h.Bc.ProofOfWork(lastProof)

	h.Bc.NewTransaction("0", h.NodeIdentifire, 1)

	newBlock := h.Bc.NewBlock(proof, "")

	return c.JSON(http.StatusOK, Response{
		"message":      "New Block Mined",
		"index":        newBlock.Index,
		"transactions": newBlock.Transactions,
	})
}

func (h *Handler) FullChain(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, Response{
		"chain":  h.Bc.Chain,
		"length": len(h.Bc.Chain),
	})
}

func (h *Handler) RegisterNode(c echo.Context) (err error) {
	type body struct {
		Nodes []string `json:"nodes"`
	}
	b := new(body)
	if err = c.Bind(b); err != nil {
		return
	}
	for _, node := range b.Nodes {
		h.Bc.RegisterNode(node)
	}

	return c.JSON(http.StatusCreated, Response{
		"message": "New nodes have been added",
		"nodes":   h.Bc.Nodes,
	})
}

func (h *Handler) Consensus(c echo.Context) (err error) {
	message := "Our chain is authoritative"
	if h.Bc.ResolveConflicts() {
		message = "Our chain was replaced"
	}
	return c.JSON(http.StatusOK, Response{
		"message": message,
		"chain":   h.Bc.Chain,
	})
}
