package types

import (
	"fmt"
	"math/big"
)

const BigIntMaxSerializedLen = 128 // is this big enough? or too big?

// var TotalFilecoinInt = FromFil(build.TotalFilecoin)
var TotalFilecoinInt = FromFil(uint64(2_000_000_000))

type BigInt = Int

func BigZero() BigInt {
	return NewBigInt(0)
}

func NewBigInt(i uint64) BigInt {
	return BigInt{Int: NewInt(0).SetUint64(i)}
}

func FromFil(i uint64) BigInt {
	return BigMul(NewBigInt(i), NewBigInt(FilecoinPrecision))
}

func BigFromBytes(b []byte) BigInt {
	i := big.NewInt(0).SetBytes(b)
	return BigInt{Int: i}
}

func BigFromString(s string) (BigInt, error) {

	if s == "" {
		return NewBigInt(0), nil
	}

	v, ok := big.NewInt(0).SetString(s, 10)
	if !ok {
		return NewBigInt(0), fmt.Errorf("failed to parse string as a big int")
	}

	return BigInt{Int: v}, nil
}

func BigMul(a, b BigInt) BigInt {
	return BigInt{Int: big.NewInt(0).Mul(a.Int, b.Int)}
}

func BigDiv(a, b BigInt) BigInt {
	return BigInt{Int: big.NewInt(0).Div(a.Int, b.Int)}
}

func (bi *BigInt) BigDiv(b BigInt) BigInt {
	bi.Int = big.NewInt(0).Div(bi.Int, b.Int)
	return *bi
}

func (bi *BigInt) BigMul(b BigInt) BigInt {
	bi.Int = big.NewInt(0).Mul(bi.Int, b.Int)
	return *bi
}

func (bi *BigInt) BigAdd(b BigInt) BigInt {
	bi.Int = big.NewInt(0).Add(bi.Int, b.Int)
	return *bi
}

func (bi *BigInt) BigSub(b BigInt) BigInt {
	bi.Int = big.NewInt(0).Sub(bi.Int, b.Int)
	return *bi
}

func (bi *BigInt) BigCmp(b BigInt) int {
	return bi.Int.Cmp(b.Int)
}

func BigDivRat(a, b BigInt) float64 {
	div, _ := new(big.Rat).SetFrac(a.Int, b.Int).Float64()
	return div
}

func BigMod(a, b BigInt) BigInt {
	return BigInt{Int: big.NewInt(0).Mod(a.Int, b.Int)}
}

func BigAdd(a, b BigInt) BigInt {
	if a.Int == nil {
		a = NewBigInt(0)
	}
	if b.Int == nil {
		b = NewBigInt(0)
	}
	return BigInt{Int: big.NewInt(0).Add(a.Int, b.Int)}
}

func BigSub(a, b BigInt) BigInt {
	return BigInt{Int: big.NewInt(0).Sub(a.Int, b.Int)}
}

func BigCmp(a, b BigInt) int {
	return a.Int.Cmp(b.Int)
}

var byteSizeUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"}

func SizeStr(bi BigInt) string {
	r := new(big.Rat).SetInt(bi.Int)
	den := big.NewRat(1, 1024)

	var i int
	for f, _ := r.Float64(); f >= 1024 && i+1 < len(byteSizeUnits); f, _ = r.Float64() {
		i++
		r = r.Mul(r, den)
	}

	f, _ := r.Float64()
	return fmt.Sprintf("%.4g %s", f, byteSizeUnits[i])
}

var deciUnits = []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"}

func DeciStr(bi BigInt) string {
	r := new(big.Rat).SetInt(bi.Int)
	den := big.NewRat(1, 1024)

	var i int
	for f, _ := r.Float64(); f >= 1024 && i+1 < len(deciUnits); f, _ = r.Float64() {
		i++
		r = r.Mul(r, den)
	}
	f, _ := r.Float64()
	return fmt.Sprintf("%.3g %s", f, deciUnits[i])
}

func BigDivFlaot(a, b BigInt) float64 {
	c := big.Rat{}
	if b.IsZero() {
		return 0
	}

	f, _ := c.SetFrac(a.Int, b.Int).Float64()
	return f
}

func (bi *BigInt) ToFloatFIL() float64 {
	rat := big.NewRat(1, 1)
	if bi == nil || bi.Int == nil {
		*bi = NewBigInt(0)
	}
	rat.SetFrac(bi.Int, big.NewInt(int64(FilecoinPrecision)))
	f, _ := rat.Float64()
	return f
}

// ParseFloatFILInt 小数 解析成 FIL
func ParseFloatFILInt(s float64) (BigInt, error) {

	r := new(big.Rat).SetFloat64(s)

	r = r.Mul(r, big.NewRat(int64(FilecoinPrecision), 1))
	if !r.IsInt() {
		return BigInt{}, fmt.Errorf("invalid FIL value: %v", s)
	}

	return BigInt{r.Num()}, nil
}

func (bi BigInt) ToFloat64() float64 {
	rat := big.NewRat(1, 1)
	if bi.Int == nil {
		return 0
	}
	rat.SetFrac(bi.Int, big.NewInt(int64(FilecoinPrecision)))
	f, _ := rat.Float64()
	return f
}

func (bi *BigInt) DivToFloat(d BigInt) float64 {
	rat := big.NewRat(1, 1)
	if bi == nil || bi.Int == nil {
		*bi = NewBigInt(0)
	}

	if d.IsZero() || d.Nil() {
		return 0
	}

	rat.SetFrac(bi.Int, d.Int)
	f, _ := rat.Float64()
	return f
}

//
//
//func (bi *BigInt) MarshalBSONValue() (bsontype.Type, []byte, error) {
//	s := []byte("0")
//
//	if bi.Int == nil {
//		return bsontype.String, s, nil
//	}
//
//	if bi.Sign() == 0 {
//		return bsontype.String, s, nil
//	}
//
//	return bsontype.String, []byte(bi.String()), nil
//}
//
//
//func (bi *BigInt) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
//	println(t)
//	if t != bsontype.String {
//		return errors.New("unknow bson type")
//	}
//	if data == nil || len(data) == 0{
//		*bi = NewInt(0)
//		return nil
//	}
//
//
//	bint, err := FromString(string(data))
//	if err != nil {
//		return err
//	}
//	*bi = bint
//	return nil
//}

//
//func (bi *BigInt) MarshalBSON() ([]byte, error) {
//
//	s := []byte("0")
//	if bi.Int == nil {
//		return s,nil
//	}
//
//	if bi.Sign() == 0 {
//		return s,nil
//	}
//
//	return []byte(bi.String()), nil
//}
//
//
//
//// Scan implements the sql.Scanner interface
//func (bi *BigInt) UnmarshalBSON(data []byte) error {
//	if data == nil {
//		*bi = NewInt(0)
//		return nil
//	}
//
//	ss := []byte{}
//	for _, v := range data {
//		if strconv.IsPrint(rune(v)) {
//			ss = append(ss, v)
//		}
//	}
//
//	if len(data) == 0 {
//		*bi = NewInt(0)
//		return nil
//	}
//
//	fmt.Println(ss)
//
//	v, ok := big.NewInt(0).SetString(string(ss), 10)
//	if !ok {
//		return fmt.Errorf("failed to parse string as a big int")
//	}
//
//	bi.Int = v
//	return nil
//}
