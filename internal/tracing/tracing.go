/**
 * OpenTelemetry Tracing
 * 实现 redo.md 3.1：分布式追踪
 * 使用 OTLP HTTP exporter（替代已弃用的 Jaeger exporter）
 */
package tracing

import (
	"context"
	"fmt"
	icfg "short-link/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
)

// InitTracing 初始化 OpenTelemetry Tracing
func InitTracing(cfg *icfg.Config) (func(), error) {
	// 如果未配置 Jaeger endpoint，则不启用追踪
	if cfg.JaegerEndpoint == "" {
		return func() {}, nil
	}

	// 创建 OTLP HTTP exporter（替代已弃用的 Jaeger exporter）
	// Jaeger 现在支持 OTLP，所以我们可以使用 OTLP exporter
	ctx := context.Background()
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.JaegerEndpoint),
		otlptracehttp.WithInsecure(), // 如果使用 HTTPS，请移除此选项并配置 TLS
	)
	if err != nil {
		return nil, fmt.Errorf("创建 OTLP exporter 失败: %w", err)
	}

	// 创建 resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("nsl-go"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 resource 失败: %w", err)
	}

	// 创建 TracerProvider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 创建 tracer
	tracer = otel.Tracer("nsl-go")

	// 返回清理函数
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Printf("关闭 TracerProvider 失败: %v\n", err)
		}
	}, nil
}

// GetTracer 获取 tracer
func GetTracer() trace.Tracer {
	if tracer == nil {
		return otel.Tracer("nsl-go")
	}
	return tracer
}

