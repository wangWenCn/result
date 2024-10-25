package result

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"runtime"

	"github.com/wangWenCn/xerr"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pkg/errors"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HTTPResult(r *http.Request, w http.ResponseWriter, resp any, err error) {
	if err == nil {
		httpx.WriteJson(w, http.StatusOK, Success(resp))
		return
	}
	errCode := xerr.ServerCommonError
	errMsg := "网络波动，请稍后再试"
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
				errCode = int64(s.Code())
				errMsg = s.Message()
			}
		}
	}
	if errCode == xerr.SystemError || errCode == xerr.ServerCommonError {
		errCode = xerr.SystemError
		errMsg = "网络波动，稍后再试"
	}
	httpx.WriteJson(w, http.StatusOK, Error(errCode, errMsg))
}

func ParamErrorResult(r *http.Request, w http.ResponseWriter, err error) {
	errMsg := fmt.Sprintf("%s ,%s", xerr.MapErrMsg(xerr.RequestParamError), err.Error())
	httpx.WriteJson(w, http.StatusOK, Error(xerr.RequestParamError, errMsg))
}

func DetailInfo(v ...any) []logx.LogField {
	if len(v) == 0 {
		return []logx.LogField{}
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return []logx.LogField{{"param", v}}
	}
	params := getCallParamName(file, line)

	fields := make([]logx.LogField, 0)
	param := make(map[string]any)
	for i := 0; i < len(v) && i < len(params); i++ {
		param[params[i]] = v[i]
	}
	fields = append(fields, logx.LogField{
		Key:   "param",
		Value: param,
	})
	return fields
}
func getCallParamName(file string, line int) []string {
	src, err := os.ReadFile(file)
	if err != nil {
		return []string{}
	}

	fSet := token.NewFileSet()
	fileAST, err := parser.ParseFile(fSet, file, src, 0)
	if err != nil {
		return []string{}
	}

	var callExpr *ast.CallExpr
	ast.Inspect(fileAST, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			pos := fSet.Position(call.Pos())
			if pos.Line == line {
				callExpr = call
				return false
			}
		}
		return true
	})

	if callExpr == nil {
		return []string{}
	}

	var detailInfoCall *ast.CallExpr
	for _, arg := range callExpr.Args {
		if call, ok := arg.(*ast.CallExpr); ok {
			detailInfoCall = call
			break
		}
	}
	if detailInfoCall == nil {
		return []string{}
	}

	paramsNames := make([]string, 0)
	for i := 0; i < len(detailInfoCall.Args); i++ {
		argStart := fSet.Position(detailInfoCall.Args[i].Pos()).Offset
		argEnd := fSet.Position(detailInfoCall.Args[i].End()).Offset
		paramsNames = append(paramsNames, string(src[argStart:argEnd]))
	}
	return paramsNames
}
