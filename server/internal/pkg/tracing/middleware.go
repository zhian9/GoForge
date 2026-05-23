package tracing

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPTracingMiddleware HTTP追踪中间件
func HTTPTracingMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从请求头提取Trace上下文
			ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// 开始Span
			ctx, span := StartSpanWithAttributes(ctx, "HTTP "+r.Method+" "+r.URL.Path, map[string]string{
				"http.method":      r.Method,
				"http.path":        r.URL.Path,
				"http.scheme":      r.URL.Scheme,
				"http.host":        r.Host,
				"http.user_agent":  r.UserAgent(),
				"http.remote_addr": r.RemoteAddr,
			})

			// 将Trace上下文注入响应头
			propagation.TraceContext{}.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// 包装ResponseWriter以捕获状态码
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start)

			// 设置Span属性
			span.SetAttributes(
				attribute.Int("http.status_code", rw.statusCode),
				attribute.Int64("http.duration_ms", duration.Milliseconds()),
			)

			// 记录错误
			if rw.statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
			}

			span.End()

			// 记录日志（包含TraceID）
			traceID := ExtractTraceID(ctx)
			logx.Infow("HTTP请求",
				logx.Field("method", r.Method),
				logx.Field("path", r.URL.Path),
				logx.Field("status", rw.statusCode),
				logx.Field("duration", duration.Milliseconds()),
				logx.Field("trace_id", traceID),
			)
		})
	}
}

// responseWriter 包装ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GRPCTracingInterceptor gRPC追踪拦截器
func GRPCTracingInterceptor(serviceName string) func(ctx context.Context, req interface{}, info interface{}, handler func(ctx context.Context, req interface{}) (interface{}, error)) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info interface{}, handler func(ctx context.Context, req interface{}) (interface{}, error)) (interface{}, error) {
		// 获取gRPC方法信息
		// 这里简化实现，实际应该从info中提取方法名
		methodName := "unknown"

		// 开始Span
		ctx, span := StartSpan(ctx, "gRPC "+methodName)
		defer span.End()

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// 设置Span属性
		span.SetAttributes(
			attribute.String("grpc.method", methodName),
			attribute.Int64("grpc.duration_ms", duration.Milliseconds()),
		)

		// 记录错误
		if err != nil {
			SetSpanError(span, err)
		}

		return resp, err
	}
}
