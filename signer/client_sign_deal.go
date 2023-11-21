package signer

import (
	"bytes"
	"fmt"
	"reflect"

	cborutil "github.com/filecoin-project/go-cbor-util"
	"github.com/filecoin-project/go-state-types/builtin/v8/market"
	"github.com/filecoin-project/go-state-types/cbor"
	c "github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus/venus-shared/types"
)

func ClientSignDeal(pvk Key, data []byte, meta types.MsgMeta) (*c.Signature, error) {
	// parse msg
	_, toSign, err := GetSignBytesAndObj(data, meta)
	if err != nil {
		return nil, fmt.Errorf("get sign bytes: %w", err)
	}

	sgtp := ActSigType(pvk.Type)

	signature, signErr := Sign(sgtp, pvk.PrivateKey, toSign)
	return signature, signErr
}

type (
	FGetSignBytes func(signObj interface{}) ([]byte, error)
	FParseObj     func(toSign []byte, meta types.MsgMeta) (interface{}, error)
)

func CborDecodeInto(r []byte, v interface{}) error {
	unmarshaler, isOk := v.(cbor.Unmarshaler)
	if !isOk {
		return fmt.Errorf("not an 'unmarhsaler'")
	}
	if err := unmarshaler.UnmarshalCBOR(bytes.NewReader(r)); err != nil {
		return fmt.Errorf("cbor unmarshal:%w", err)
	}
	return nil
}

var defaultPaseObjFunc = func(t reflect.Type) FParseObj {
	return func(b []byte, meta types.MsgMeta) (interface{}, error) {
		obj := reflect.New(t).Interface()
		if err := CborDecodeInto(b, obj); err != nil {
			return nil, err
		}
		return obj, nil
	}
}

// GetSignBytesAndObj Matches the type and returns the data that needs to be signed
func GetSignBytesAndObj(toSign []byte, meta types.MsgMeta) (interface{}, []byte, error) {
	// t := wallet.Types{
	// 	Type: reflect.TypeOf(market.DealProposal{}),
	// 	SignBytes: func(i interface{}) ([]byte, error) {
	// 		return cborutil.Dump(i)
	// 	},
	// 	ParseObj: defaultPaseObjFunc(reflect.TypeOf(market.DealProposal{})),
	// }

	// ParseObj may be nil registered through RegisterSupportedMsgTypes func.
	var (
		in  interface{}
		err error
	)

	in, err = defaultPaseObjFunc(reflect.TypeOf(market.DealProposal{}))(toSign, meta)
	if err != nil {
		return nil, nil, fmt.Errorf("parseObj failed:%w", err)
	}

	var data []byte
	data, err = cborutil.Dump(in)
	return in, data, err
}
