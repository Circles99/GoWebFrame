package opentelemetry

import (
	"GoWebFrame"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type MiddleWareBuilder struct {
	trace.Tracer
}

func NewMiddleWareBuilder(tracer trace.Tracer) *MiddleWareBuilder {
	return &MiddleWareBuilder{
		Tracer: tracer,
	}
}

func (b MiddleWareBuilder) Builder() GoWebFrame.Middleware {
	return func(next GoWebFrame.HandleFunc) GoWebFrame.HandleFunc {
		return func(ctx *GoWebFrame.Context) {
			reqCtx := ctx.Req.Context()
			// http之间通信， 从入参context中的头部获取
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier{})
			reqCtx, span := b.Tracer.Start(reqCtx, "opentelemetryMiddleware", trace.WithAttributes())
			// 设置值
			span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
			span.SetAttributes(attribute.String("peer.hostname", ctx.Req.Host))
			span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(attribute.String("http.scheme", ctx.Req.URL.Scheme))
			span.SetAttributes(attribute.String("span.kind", "server"))
			span.SetAttributes(attribute.String("component", "web"))
			span.SetAttributes(attribute.String("peer.address", ctx.Req.RemoteAddr))
			span.SetAttributes(attribute.String("http.proto", ctx.Req.Proto))

			// 函数退出时，关闭span
			defer span.End()
			next(ctx)
			span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))
		}

	}
}
