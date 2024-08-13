package result

type ResponseBean struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type NullJson struct {
}

type ResponseErrorBean struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

func Success(data any) *ResponseBean {
	return &ResponseBean{
		Code: 0,
		Msg:  "OK",
		Data: data,
	}
}

func Error(errCode int64, errMsg string) *ResponseErrorBean {
	return &ResponseErrorBean{errCode, errMsg}
}
