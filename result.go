package result

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/wangWenCn/xerr"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HTTPResult(r *http.Request, w http.ResponseWriter, resp any, err error) {
	if err == nil {
		httpx.WriteJson(w, http.StatusOK, Success(resp))
	} else {
		//错误返回
		errCode := xerr.ServerCommonError
		errMsg := "服务器开小差啦，稍后再来试一试"
		var e *xerr.CodeError
		if errors.As(err, &e) { //自定义错误类型
			//自定义CodeError
			errCode = e.Code
			errMsg = e.Message
		} else {
			//rpc 错误
			s, ok := status.FromError(err)
			if ok {
				errMsg = s.Message()
				if s.Code() == codes.Unknown {
					errCode = xerr.MapErrCode(errMsg)
				}
				logx.WithContext(r.Context()).Errorf("【RPC-ERR】 : %+v ", err)
				httpx.WriteJson(w, http.StatusInternalServerError, Error(errCode, errMsg))
			} else {
				logx.WithContext(r.Context()).Errorf("【RPC-ERR】 : %+v ", err)
				httpx.WriteJson(w, http.StatusInternalServerError, Error(errCode, errMsg))
			}
		}
	}
}

func ParamErrorResult(r *http.Request, w http.ResponseWriter, err error) {
	errMsg := fmt.Sprintf("%s ,%s", xerr.MapErrMsg(xerr.RequestParamError), err.Error())
	httpx.WriteJson(w, http.StatusOK, Error(xerr.RequestParamError, errMsg))
}
