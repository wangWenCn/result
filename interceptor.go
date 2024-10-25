package result

import (
	"context"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/wangWenCn/traceLog"
	"github.com/wangWenCn/xerr"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LogTraceInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	traceLog.SetGoroutineContext(ctx)
	start := time.Now()
	resp, err = handler(ctx, req)
	if err != nil {
		causeErr := errors.Cause(err)                // err类型
		if e, ok := causeErr.(*xerr.CodeError); ok { //自定义错误类型
			//转成grpc err
			err = status.Error(codes.Code(e.Code), e.Message)
			logx.WithContext(ctx).Errorw(info.FullMethod+" 调用失败",
				logx.Field("error", e.Message),
			)
		} else {
			logx.WithContext(ctx).Errorw(info.FullMethod+" 调用失败",
				logx.Field("error", err.Error()),
			)
		}
	}
	logx.WithContext(ctx).Infow("请求: "+info.FullMethod,
		logx.Field("req", pb2map(req)),
		logx.Field("resp", pb2map(resp)),
		logx.Field("duration", strconv.Itoa(int(time.Since(start).Milliseconds()))+"ms"),
	)
	traceLog.DelGoroutineContext()
	return resp, err
}

func pb2map(pb interface{}) map[string]any {
	if pb == nil {
		return nil
	}
	m := make(map[string]any)
	val := reflect.ValueOf(pb)
	if val.IsNil() {
		return nil
	}
	val = val.Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		if valueField.CanInterface() {
			f := valueField.Interface()
			v := reflect.ValueOf(f)
			m[typeField.Name] = v.Interface()
		}
	}
	return m
}
