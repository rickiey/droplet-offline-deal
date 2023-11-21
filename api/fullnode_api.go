package api

import (
	"droplet-offline-deal/types"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/proof"
	"github.com/go-resty/resty/v2"
	"github.com/ipfs/go-cid"
	"github.com/rickiey/loggo"
)

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
)

var (
	client   = resty.New().SetRetryCount(3)
	lotusURL = ""
)

func InitChainAPI(endpoint string) {
	lotusURL = endpoint
}

type RequestMetaDate struct {
	JSONRPC string        `json:"json_rpc"`
	Method  string        `json:"method"`
	ID      int           `json:"id"`
	Params  []interface{} `json:"params"`
}

func NewRequestMetaDate() RequestMetaDate {
	return RequestMetaDate{
		// ID 这个值传什么返回什么， 随便给个
		ID:      233,
		JSONRPC: "2.0",
		Params:  make([]interface{}, 0),
	}
}

type RespTipSet struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Error   types.Err `json:"error"`
	Result  ExpTipSet `json:"result"`
}
type ExpTipSet struct {
	Cids   []cid.Cid
	Blocks []*BlockHeader
	Height abi.ChainEpoch
}

type RespMinerInfo struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Error   types.Err       `json:"error"`
	Result  types.MinerInfo `json:"result"`
}

type BlockHeader struct {
	Miner address.Address // 0

	// Ticket *Ticket // 1

	// ElectionProof *ElectionProof // 2

	// BeaconEntries []BeaconEntry // 3

	WinPoStProof []proof.PoStProof // 4

	Parents []cid.Cid // 5

	// ParentWeight BigInt // 6

	Height abi.ChainEpoch // 7

	ParentStateRoot cid.Cid // 8

	ParentMessageReceipts cid.Cid // 8

	Messages cid.Cid // 10

	BLSAggregate *crypto.Signature // 11

	Timestamp int64 // 12

	BlockSig *crypto.Signature // 13

	ForkSignaling uint64 // 14

	// ParentBaseFee BigInt // 15

	// internal
	validated bool // true if the signature has been validated
}

func ChainHead() (*ExpTipSet, error) {
	if len(lotusURL) == 0 {
		err := errors.New("not init endpoint")
		loggo.Error(err)
		return nil, err
	}
	method := "Filecoin.ChainHead"
	rmd := NewRequestMetaDate()
	rmd.Method = method

	r, err := client.R().
		SetHeader(ContentType, ApplicationJSON).
		SetBody(rmd).
		Post(lotusURL)
	if err != nil {
		loggo.Warn("request lotus gateway failed : ", err)
		return nil, err
	}

	if r.StatusCode() != 200 {
		return nil, fmt.Errorf("StatusCode = %v function: %v", r.StatusCode(), rmd.Method)
	}

	resp := new(RespTipSet)

	err = json.Unmarshal(r.Body(), resp)
	if err != nil {
		loggo.Errorf("ChainHead Unmarshal Failed: %v body:", err, string(r.Body()))
		return nil, err
	}

	if resp.Error.Code != 0 {
		return nil, &resp.Error
	}

	return &resp.Result, nil
}

func StateMinerInfo(minerAddr address.Address) (*types.MinerInfo, error) {
	if len(lotusURL) == 0 {
		err := errors.New("not init endpoint")
		loggo.Error(err)
		return nil, err
	}
	method := "Filecoin.StateMinerInfo"
	rmd := NewRequestMetaDate()
	rmd.Method = method
	rmd.Params = []interface{}{
		minerAddr.String(),
		nil,
	}

	r, err := client.R().
		SetHeader(ContentType, ApplicationJSON).
		SetBody(rmd).
		Post(lotusURL)
	if err != nil {
		loggo.Warn("request lotus gateway failed : ", err)
		return nil, err
	}

	if r.StatusCode() != 200 {
		return nil, fmt.Errorf("StatusCode = %v function: %v", r.StatusCode(), rmd.Method)
	}

	resp := new(RespMinerInfo)

	err = json.Unmarshal(r.Body(), resp)
	if err != nil {
		loggo.Errorf("ChainHead Unmarshal Failed: %v body:", err, string(r.Body()))
		return nil, err
	}

	if resp.Error.Code != 0 {
		return nil, &resp.Error
	}

	return &resp.Result, nil
}
