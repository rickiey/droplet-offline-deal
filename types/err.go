package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Err struct {
	// 错误码,
	Code int `json:"code"`
	// 信息
	Msg string `json:"message"`
	// 错误
	Data interface{} `json:"data"`
}

func (er *Err) Set(code int, msg string, v interface{}) {
	er.Data = v
	er.Code = code
	er.Msg = msg
}

func (er *Err) Error() string {
	return fmt.Sprintf(`{"code": %v, "msg"": %s}`, er.Code, er.Msg)
}

func (er *Err) String() string {
	b, _ := json.Marshal(er)
	return strings.ReplaceAll(string(b), "\"", "'")
}

func NewErr(code int, msg string, v interface{}) error {
	return &Err{code, msg, v}
}
