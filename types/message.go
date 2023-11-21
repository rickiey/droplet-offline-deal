package types

import (
	"bytes"
	"fmt"
	"io"

	"github.com/filecoin-project/go-state-types/abi"
	block "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"

	cbg "github.com/whyrusleeping/cbor-gen"
)

const MessageVersion = 0

type ChainMsg interface {
	Cid() cid.Cid
	VMMessage() *Message
	ToStorageBlock() (block.Block, error)
	// FIXME: This is the *message* length, this name is misleading.
	ChainLength() int
}

type MessageSendSpec struct {
	MaxFee BigInt
}

var DefaultMessageSendSpec = MessageSendSpec{
	// MaxFee of 0.1FIL
	MaxFee: NewBigInt(FilecoinPrecision / 10),
}

type SimpleMessage struct {
	To   address.Address
	From address.Address

	Nonce uint64

	Method abi.MethodNum

	Value BigInt
}

type ParentMessage struct {
	Cid     cid.Cid
	Message *Message
}

type Message struct {
	Version int64

	To   address.Address
	From address.Address

	Nonce uint64

	Value TokenAmount

	GasLimit   int64
	GasFeeCap  TokenAmount
	GasPremium TokenAmount

	Method abi.MethodNum
	Params []byte
}

type MessageWithCid struct {
	MsgCid  cid.Cid `json:"cid"`
	Version int64

	To   address.Address
	From address.Address

	Nonce uint64

	Value TokenAmount

	GasLimit   int64
	GasFeeCap  TokenAmount
	GasPremium TokenAmount

	Method abi.MethodNum
	Params []byte
}

func (m *Message) Caller() address.Address {
	return m.From
}

func (m *Message) Receiver() address.Address {
	return m.To
}

func (m *Message) ValueReceived() TokenAmount {
	return m.Value
}

var lengthBufMessage = []byte{138}

func (t *Message) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufMessage); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Version (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Version)); err != nil {
		return err
	}

	// t.To (address.Address) (struct)
	if err := t.To.MarshalCBOR(w); err != nil {
		return err
	}

	// t.From (address.Address) (struct)
	if err := t.From.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Nonce (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Nonce)); err != nil {
		return err
	}

	// t.Value (big.Int) (struct)
	if err := t.Value.MarshalCBOR(w); err != nil {
		return err
	}

	// t.GasLimit (int64) (int64)
	if t.GasLimit >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.GasLimit)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.GasLimit-1)); err != nil {
			return err
		}
	}

	// t.GasFeeCap (big.Int) (struct)
	if err := t.GasFeeCap.MarshalCBOR(w); err != nil {
		return err
	}

	// t.GasPremium (big.Int) (struct)
	if err := t.GasPremium.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Method (abi.MethodNum) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Method)); err != nil {
		return err
	}

	// t.Params ([]uint8) (slice)
	if len(t.Params) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Params was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Params))); err != nil {
		return err
	}

	if _, err := w.Write(t.Params[:]); err != nil {
		return err
	}
	return nil
}

func (t *Message) UnmarshalCBOR(r io.Reader) error {
	*t = Message{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 10 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Version (uint64) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Version = int64(extra)

	}
	// t.To (address.Address) (struct)

	{

		if err := t.To.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.To: %w", err)
		}

	}
	// t.From (address.Address) (struct)

	{

		if err := t.From.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.From: %w", err)
		}

	}
	// t.Nonce (uint64) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Nonce = uint64(extra)

	}
	// t.Value (big.Int) (struct)

	{

		if err := t.Value.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Value: %w", err)
		}

	}
	// t.GasLimit (int64) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			return err
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			return fmt.Errorf("wrong type for int64 field: %d", maj)
		}

		t.GasLimit = int64(extraI)
	}
	// t.GasFeeCap (big.Int) (struct)

	{

		if err := t.GasFeeCap.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.GasFeeCap: %w", err)
		}

	}
	// t.GasPremium (big.Int) (struct)

	{

		if err := t.GasPremium.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.GasPremium: %w", err)
		}

	}
	// t.Method (abi.MethodNum) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Method = abi.MethodNum(extra)

	}
	// t.Params ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.Params: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.Params = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.Params[:]); err != nil {
		return err
	}
	return nil
}

func DecodeMessage(b []byte) (*Message, error) {
	var msg Message
	if err := msg.UnmarshalCBOR(bytes.NewReader(b)); err != nil {
		return nil, err
	}

	if msg.Version != MessageVersion {
		return nil, fmt.Errorf("decoded message had incorrect version (%d)", msg.Version)
	}

	return &msg, nil
}

func (m *Message) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *Message) ChainLength() int {
	ser, err := m.Serialize()
	if err != nil {
		panic(err)
	}
	return len(ser)
}

func (m *Message) ToStorageBlock() (block.Block, error) {
	data, err := m.Serialize()
	if err != nil {
		return nil, err
	}

	c, err := abi.CidBuilder.Sum(data)
	if err != nil {
		return nil, err
	}

	return block.NewBlockWithCid(data, c)
}

func (m *Message) Cid() cid.Cid {
	b, err := m.ToStorageBlock()
	if err != nil {
		panic(fmt.Sprintf("failed to marshal message: %s", err)) // I think this is maybe sketchy, what happens if we try to serialize a message with an undefined address in it?
	}

	return b.Cid()
}

func (m *Message) RequiredFunds() BigInt {
	return BigMul(BigInt(m.GasFeeCap), NewBigInt(uint64(m.GasLimit)))
}

func (m *Message) VMMessage() *Message {
	return m
}

func (m *Message) Equals(o *Message) bool {
	return m.Cid() == o.Cid()
}

func (m *Message) ValidForBlockInclusion(minGas int64) error {
	if m.Version != 0 {
		return xerrors.New("'Version' unsupported")
	}

	if m.To == address.Undef {
		return xerrors.New("'To' address cannot be empty")
	}

	if m.From == address.Undef {
		return xerrors.New("'From' address cannot be empty")
	}

	if m.Value.LessThan(Zero()) {
		return xerrors.New("'Value' field cannot be negative")
	}

	if m.Value.GreaterThan(TotalFilecoinInt) {
		return xerrors.New("'Value' field cannot be greater than total filecoin supply")
	}

	if m.GasFeeCap.LessThan(Zero()) {
		return xerrors.New("'GasFeeCap' field cannot be negative")
	}

	if m.GasPremium.LessThan(Zero()) {
		return xerrors.New("'GasPremium' field cannot be negative")
	}

	if m.GasLimit > BlockGasLimit {
		return xerrors.New("'GasLimit' field cannot be greater than a block's gas limit")
	}

	// since prices might vary with time, this is technically semantic validation
	if m.GasLimit < minGas {
		return xerrors.New("'GasLimit' field cannot be less than the cost of storing a message on chain")
	}

	return nil
}

const TestGasLimit = 100e6
