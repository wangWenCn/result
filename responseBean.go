package result

type ResponseBean struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type NullJson struct {
}

type ResponseErrorBean struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
}

func Success(data any) *ResponseBean {
	return &ResponseBean{
		Code: 0,
		Msg:  "OK",
		Data: data,
	}
}

func Error(errCode uint32, errMsg string) *ResponseErrorBean {
	return &ResponseErrorBean{errCode, errMsg}
}
