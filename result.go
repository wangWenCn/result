package result

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

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
		causeErr := errors.Cause(err)
		if e, ok := causeErr.(*xerr.CodeError); ok { //自定义错误类型
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
				} else {
					errCode = uint32(s.Code())
					errMsg = s.Message()
				}
				logx.WithContext(r.Context()).Errorf("【RPC-ERR】 : %+v ", err)
				httpx.WriteJson(w, http.StatusOK, Error(errCode, errMsg))
			} else {
				logx.WithContext(r.Context()).Errorf("【RPC-ERR】 : %+v ", err)
				httpx.WriteJson(w, http.StatusOK, Error(errCode, errMsg))
			}
		}
	}
}

func ParamErrorResult(r *http.Request, w http.ResponseWriter, err error) {
	errMsg := fmt.Sprintf("%s ,%s", xerr.MapErrMsg(xerr.RequestParamError), err.Error())
	httpx.WriteJson(w, http.StatusOK, Error(xerr.RequestParamError, errMsg))
}

func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		causeErr := errors.Cause(err)                // err类型
		if e, ok := causeErr.(*xerr.CodeError); ok { //自定义错误类型
			logx.WithContext(ctx).Errorf("【RPC-SRV-ERR】 %+v", err)
			//转成grpc err
			err = status.Error(codes.Code(e.Code), e.Message)
		} else {
			logx.WithContext(ctx).Errorf("【RPC-SRV-ERR】 %+v", err)
		}
	}
	return resp, err
}
