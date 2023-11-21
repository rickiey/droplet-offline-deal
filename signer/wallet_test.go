package signer

import (
	"encoding/hex"
	"fmt"

	// "sync_chain_data/utils/crypto"
	"testing"

	"github.com/filecoin-project/go-address"
)

func Test_001(t *testing.T) {
	pv := "7b2254797065223a22736563703235366b31222c22507269766174654b6579223a223535793342615865786a39363375574e4b32636a576f4c5850747a765471464a765238675671596c2b65513d227d"
	k, err := DecodePricateKey(pv)

	if err != nil {
		panic(err)
	}

	fmt.Println(hex.EncodeToString(k.PrivateKey))

	s := NewSecp256k1Singer()

	es, err := s.Sign(k.PrivateKey, k.PrivateKey)
	if err != nil {
		panic(err)
	}

	println(es)

	p, _ := s.ToPublic(k.PrivateKey)
	fmt.Println(address.NewSecp256k1Address(p))

}

// 盐值必须是 16,24,32 位
const salt = "TDSAOFIYGVBHFKHLPOUUIPMLMAODYSZX"

// func Test_DecPvk(t *testing.T) {
// 	s := "397a3778365872514f3430784672475a554e3866576369683565487730373741745872666c656b6d75304e79734b56422b2b74446c742b447a6f597366654637"
// 	pv := crypto.AES_Decrypt(s, []byte(salt))

// 	k := KeyInfo{
// 		Type:       KTSecp256k1,
// 		PrivateKey: pv,
// 	}

// 	b, _ := json.Marshal(k)

// 	fmt.Println(hex.EncodeToString(b))
// }
