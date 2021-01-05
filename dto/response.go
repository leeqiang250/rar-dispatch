package dto

import (
	"dispatch/serializer"
	"dispatch/time"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	TS   int64       `json:"ts"`
	Data interface{} `json:"data"`
}

func (this *Response) SetData(data interface{}) *Response {
	this.Data = data
	return this
}

func (this *Response) Bytes() []byte {
	return serializer.Bytes(this)
}

func Success() *Response {
	return Builder(CodeSuccess)
}

func SuccessBytes() []byte {
	return Success().Bytes()
}

func Fail() *Response {
	return Builder(CodeFail)
}

func FailBytes() []byte {
	return Fail().Bytes()
}

func Builder(code *StatusCode) *Response {
	return &Response{
		Code: code.Code,
		Msg:  code.Msg,
		TS:   time.TimestampNowMs(),
		Data: nil,
	}
}
