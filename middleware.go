package result

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wangWenCn/traceLog"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

type responseWriter struct {
	http.ResponseWriter
	Body   *bytes.Buffer
	Status int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.Body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.Status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
func LogTraceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceLog.SetGoroutineContext(r.Context())
		w.Header().Set("trace-id", trace.TraceIDFromContext(r.Context()))
		startTime := time.Now()
		fields := make([]logx.LogField, 0)
		if token := r.Header.Get("x-token"); len(token) != 0 {
			fields = append(fields,
				logx.Field("token", token),
			)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logx.WithContext(r.Context()).Errorw("读取请求参数失败", logx.Field("err", err))
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		if strings.Contains(r.Header.Get("Content-Type"), "json") && len(body) > 0 {
			var reqBody map[string]any
			err := json.Unmarshal(body, &reqBody)
			if err != nil {
				logx.WithContext(r.Context()).Errorw("解析请求参数失败", logx.Field("err", err))
			}
			fields = append(fields, logx.Field("body", reqBody))
		} else {
			fields = append(fields, logx.Field("body", string(body)))
		}
		
		var buffer bytes.Buffer
		rw := &responseWriter{
			ResponseWriter: w,
			Body:           &buffer,
		}
		next(rw, r)
		header := w.Header().Get("Content-Type")
		if strings.Contains(header, "json") && len(rw.Body.Bytes()) > 0 {
			resBody := make(map[string]any)
			err := json.Unmarshal(rw.Body.Bytes(), &resBody)
			if err != nil {
				logx.WithContext(r.Context()).Errorw("解析返回数据失败", logx.Field("err", err))
			}
			fields = append(fields, logx.Field("response", resBody))
		} else {
			fields = append(fields, logx.Field("response", rw.Body.String()))
		}
		fields = append(fields,
			logx.Field("host", r.Host),
			logx.Field("method", r.Method),
			logx.Field("path", r.URL.Path),
			logx.Field("req-host", r.URL.Host),
			logx.Field("duration", strconv.Itoa(int(time.Since(startTime).Milliseconds()))+"ms"),
		)
		logx.WithContext(r.Context()).Infow(
			"请求: "+r.URL.String(),
			fields...,
		)
		traceLog.DelGoroutineContext()
	}
}
