package tracing

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Config Jaeger配置
type Config struct {
	ServiceName string   `json:"required"`
	Endpoint    string   `json:"required"` // Jaeger endpoint (e.g., http://localhost:14268/api/traces)
	Environment string   `json:"default=development"`
	Tags        []string `json:"optional"` // 额外标签 (key:value格式)
}

// TracerProvider Trace提供者
type TracerProvider struct {
	tp *tracesdk.TracerProvider
}

var globalTracerProvider *TracerProvider

// InitJaeger 初始化Jaeger追踪
func InitJaeger(cfg Config) (*TracerProvider, error) {
	// 创建Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Endpoint)))
	if err != nil {
		return nil, fmt.Errorf("创建Jaeger exporter失败: %w", err)
	}

	// 构建资源
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	// 添加额外标签
	if len(cfg.Tags) > 0 {
		attrs := []attribute.KeyValue{}
		for _, tag := range cfg.Tags {
			// 解析key:value格式
			// 简化实现，实际应该更严格解析
			attrs = append(attrs, attribute.String("tag", tag))
		}
		res, _ = resource.Merge(res, resource.NewWithAttributes(semconv.SchemaURL, attrs...))
	}

	// 创建TracerProvider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(1.0)), // 100%采样（生产环境应降低）
	)

	// 设置为全局TracerProvider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	globalTracerProvider = &TracerProvider{tp: tp}
	logx.Infow("Jaeger追踪初始化成功", logx.Field("service", cfg.ServiceName), logx.Field("endpoint", cfg.Endpoint))

	return globalTracerProvider, nil
}

// Shutdown 关闭TracerProvider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.tp != nil {
		return tp.tp.Shutdown(ctx)
	}
	return nil
}

// GetTracer 获取Tracer
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan 开始Span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := GetTracer("ecommerce-system")
	return tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes 开始Span（带属性）
func StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := StartSpan(ctx, name, opts...)
	for k, v := range attrs {
		span.SetAttributes(attribute.String(k, v))
	}
	return ctx, span
}

// EndSpan 结束Span
func EndSpan(span trace.Span) {
	span.End()
}

// SetSpanError 设置Span错误
func SetSpanError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanAttributes 设置Span属性
func SetSpanAttributes(span trace.Span, attrs map[string]string) {
	for k, v := range attrs {
		span.SetAttributes(attribute.String(k, v))
	}
}

// ExtractTraceID 提取TraceID
func ExtractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// ExtractSpanID 提取SpanID
func ExtractSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// SpanFromContext 从Context获取Span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}
