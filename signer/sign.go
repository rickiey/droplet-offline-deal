package signer

import (
	"encoding/hex"
	"errors"
	"fmt"

	"droplet-offline-deal/types"

	"github.com/filecoin-project/go-address"
	crypto2 "github.com/filecoin-project/go-state-types/crypto"
)

// SigShim is used for introducing signature functions
type SigShim interface {
	GenPrivate() ([]byte, error)
	ToPublic(pk []byte) ([]byte, error)
	Sign(pk []byte, msg []byte) ([]byte, error)
	Verify(sig []byte, a address.Address, msg []byte) error
}

var sigs map[crypto2.SigType]SigShim

func init() {
	sigs = make(map[crypto2.SigType]SigShim, 2)
	sigs[crypto2.SigTypeBLS] = new(blsSigner)
	sigs[crypto2.SigTypeSecp256k1] = new(secpSigner)

}

func NewSecp256k1Singer() SigShim {
	return new(secpSigner)
}

func NewBLSSinger() SigShim {
	return new(blsSigner)
}

// Sign takes in signature type, private key and message. Returns a signature for that message.
// Valid sigTypes are: "secp256k1" and "bls"
func Sign(sigType crypto2.SigType, privkey []byte, msg []byte) (*crypto2.Signature, error) {
	sv, ok := sigs[sigType]
	if !ok {
		return nil, fmt.Errorf("cannot sign message with signature of unsupported type: %v", sigType)
	}

	sb, err := sv.Sign(privkey, msg)
	if err != nil {
		return nil, err
	}
	return &crypto2.Signature{
		Type: sigType,
		Data: sb,
	}, nil
}

func SignMessageV2(msg types.Message, priKey string) (*crypto2.Signature, error) {

	sigType := AddressSigType(msg.From)

	mb, err := msg.ToStorageBlock()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("serializing message: %v", err))
	}

	pv, err := hex.DecodeString(priKey)
	if nil != err {
		return nil, errors.New(fmt.Sprintf("hex.DecodeString private key failed: %v", err))
	}

	return Sign(sigType, pv, mb.Cid().Bytes())
}

func AddressSigType(addr address.Address) crypto2.SigType {
	if addr.Protocol() == address.SECP256K1 {
		return crypto2.SigTypeSecp256k1
	}

	if addr.Protocol() == address.BLS {
		return crypto2.SigTypeBLS
	}

	return crypto2.SigTypeUnknown
}
