package dto

type StatusCode struct {
	Code        int
	Msg         string
	Description string
}

var (
	CodeSuccess = &StatusCode{Code: 0, Msg: "success", Description: "success"}
	CodeFail    = &StatusCode{Code: 500, Msg: "fail", Description: "fail"}
)
