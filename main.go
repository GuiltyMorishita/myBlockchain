package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/satori/go.uuid"
)

type BlockChain struct {
	Chain               []Block
	CurrentTransactions []Transaction
	Nodes               mapset.Set
}

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        int64         `json:"proof"`
	PreviousHash string        `json:"previousHash"`
}

type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
}

func NewBlockchain() *BlockChain {
	newBlockchain := &BlockChain{
		Chain:               make([]Block, 0),
		CurrentTransactions: make([]Transaction, 0),
		Nodes:               mapset.NewSet(),
	}
	newBlockchain.NewBlock(100, "1")
	return newBlockchain
}

func (bc *BlockChain) NewBlock(proof int64, previousHash string) Block {
	if previousHash == "" {
		previousBlock := bc.Chain[len(bc.Chain)-1]
		previousHash = Hash(previousBlock)
	}

	newBlock := Block{
		Index:        int64(len(bc.Chain)) + 1,
		Timestamp:    time.Now().UnixNano(),
		Transactions: bc.CurrentTransactions,
		Proof:        proof,
		PreviousHash: previousHash,
	}

	bc.CurrentTransactions = []Transaction{}
	bc.Chain = append(bc.Chain, newBlock)
	return newBlock
}

func (bc *BlockChain) NewTransaction(sender string, recipient string, amount int64) int64 {
	bc.CurrentTransactions = append(bc.CurrentTransactions,
		Transaction{
			Sender:    sender,
			Recipient: recipient,
			Amount:    amount,
		})
	return bc.LastBlock().Index + 1
}

func Hash(block Block) string {
	blockJSON, _ := json.Marshal(block)
	return Sha256(blockJSON)
}

func (bc *BlockChain) LastBlock() Block {
	return bc.Chain[len(bc.Chain)-1]
}

func (bc *BlockChain) ProofOfWork(lastProof int64) int64 {
	var proof int64 = 0
	for !bc.ValidProof(lastProof, proof) {
		proof += 1
	}
	return proof
}

func (bc *BlockChain) ValidProof(lastProof, proof int64) bool {
	guess := fmt.Sprintf("%d%d", lastProof, proof)
	guessHash := Sha256([]byte(guess))
	return guessHash[:4] == "0000"
}

func (bc *BlockChain) ValidChain(chain *[]Block) bool {
	lastBlock := (*chain)[0]
	currentIndex := 1

	for currentIndex < len(*chain) {
		block := (*chain)[currentIndex]
		fmt.Println(lastBlock)
		fmt.Println(block)
		fmt.Println("--------------")

		if block.PreviousHash != Hash(lastBlock) {
			return false
		}
		if !bc.ValidProof(lastBlock.Proof, block.Proof) {
			return false
		}
		lastBlock = block
		currentIndex++
	}
	return true
}

type NeighbourChain struct {
	Length int     `json:"length"`
	Chain  []Block `json:"chain"`
}

func (bc *BlockChain) ResolveConflicts() bool {
	neighbours := bc.Nodes
	newChain := make([]Block, 0)
	maxLength := len(bc.Chain)

	for node := range neighbours.Iter() {
		response, err := http.Get(fmt.Sprintf("http://%s/chain", node))
		if err == nil && response.StatusCode == http.StatusOK {
			nc := new(NeighbourChain)
			if err := json.NewDecoder(response.Body).Decode(&nc); err != nil {
				fmt.Println(err)
			}
			if nc.Length > maxLength && bc.ValidChain(&nc.Chain) {
				maxLength = nc.Length
				newChain = nc.Chain
			}
		}
	}

	if len(newChain) > 0 {
		bc.Chain = newChain
		return true
	}

	return false
}

func (bc *BlockChain) RegisterNode(address string) {
	parsedUrl, err := url.Parse(address)
	if err != nil {
		return
	}
	bc.Nodes.Add(parsedUrl.Host)
}

func Sha256(bytes []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(bytes))
}

type Response map[string]interface{}

func CreateTransaction(bc *BlockChain) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		t := new(Transaction)
		if err = c.Bind(t); err != nil {
			return
		}
		index := bc.NewTransaction(t.Sender, t.Recipient, t.Amount)
		return c.JSON(http.StatusCreated, Response{
			"message": fmt.Sprintf("Transaction will be added to Block %d", index),
		})
	}
}

func Mine(bc *BlockChain, nodeIdentifire string) echo.HandlerFunc {
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

func FullChain(bc *BlockChain) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		return c.JSON(http.StatusOK, Response{
			"chain":  bc.Chain,
			"length": len(bc.Chain),
		})
	}
}

func RegisterNode(bc *BlockChain) echo.HandlerFunc {
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

func Consensus(bc *BlockChain) echo.HandlerFunc {
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

func main() {
	e := echo.New()

	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	nodeIdentifire := uuid.NewV4().String()
	blockchain := NewBlockchain()

	e.POST("/transactions/new", CreateTransaction(blockchain))
	e.GET("/mine", Mine(blockchain, nodeIdentifire))
	e.GET("/chain", FullChain(blockchain))
	e.POST("/nodes/register", RegisterNode(blockchain))
	e.GET("/nodes/resolve", Consensus(blockchain))
	e.Logger.Fatal(e.Start(":5000"))
}
