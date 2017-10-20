package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GuiltyMorishita/myBlockchain/blockchain"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateTransaction(t *testing.T) {

	var (
		txJSON = `{
		 "sender": "d4ee26eee15148ee92c6cd394edd974e",
		 "recipient": "someone-other-address",
		 "amount": 5
		}`
		resp struct {
			Message string `json:"message"`
		}
	)

	Convey("Transaction creation success", t, func() {
		e := echo.New()
		req := httptest.NewRequest(echo.POST, "/transactions/new", strings.NewReader(txJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &Handler{Bc: blockchain.NewBlockchain()}

		So(h.CreateTransaction(c), ShouldBeNil)
		So(rec.Code, ShouldEqual, http.StatusCreated)
		json.Unmarshal(rec.Body.Bytes(), &resp)
		So(resp.Message, ShouldNotBeBlank)
	})
}

func TestMine(t *testing.T) {

	var (
		resp struct {
			Message      string                   `json:"message"`
			Index        int64                    `json:"index"`
			Transactions []blockchain.Transaction `json:"transactions"`
		}
	)

	Convey("Mining success", t, func() {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/mine", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &Handler{
			Bc:             blockchain.NewBlockchain(),
			NodeIdentifire: uuid.NewV4().String(),
		}

		So(h.Mine(c), ShouldBeNil)
		So(rec.Code, ShouldEqual, http.StatusOK)
		json.Unmarshal(rec.Body.Bytes(), &resp)
		So(resp.Message, ShouldNotBeBlank)
		So(resp.Index, ShouldBeGreaterThan, 1)
		So(len(resp.Transactions), ShouldBeGreaterThan, 0)
	})
}

func TestFullChain(t *testing.T) {

	var (
		resp struct {
			Length int64              `json:"length"`
			Chain  []blockchain.Block `json:"chain"`
		}
	)

	Convey("Chain acquisition success", t, func() {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/chain", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &Handler{Bc: blockchain.NewBlockchain()}

		So(h.FullChain(c), ShouldBeNil)
		So(rec.Code, ShouldEqual, http.StatusOK)
		json.Unmarshal(rec.Body.Bytes(), &resp)
		So(resp.Length, ShouldBeGreaterThan, 0)
		So(len(resp.Chain), ShouldBeGreaterThan, 0)
	})
}

func TestRegisterNode(t *testing.T) {

	var (
		nodesJSON = `{ "nodes": ["http://localhost:5001"] }`
		resp      struct {
			Message string   `json:"message"`
			Nodes   []string `json:"nodes"`
		}
	)

	Convey("Chain acquisition success", t, func() {
		e := echo.New()
		req := httptest.NewRequest(echo.POST, "/nodes/register", strings.NewReader(nodesJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := &Handler{Bc: blockchain.NewBlockchain()}

		So(h.RegisterNode(c), ShouldBeNil)
		So(rec.Code, ShouldEqual, http.StatusCreated)
		json.Unmarshal(rec.Body.Bytes(), &resp)
		So(resp.Message, ShouldNotBeBlank)
		So(len(resp.Nodes), ShouldBeGreaterThan, 0)
	})
}
