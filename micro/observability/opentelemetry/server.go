package opentelemetry

import (
	"GoWebFrame/micro/observability"
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const InstrumentationName = "gitee.com/geektime-geekbang/geektime-go/micro/observability/opentelemetry"

type ServerOtelBuilder struct {
	Tracer trace.Tracer
	Port   int
}

func (b *ServerOtelBuilder) Build() grpc.UnaryServerInterceptor {

	if b.Tracer == nil {
		b.Tracer = otel.GetTracerProvider().Tracer(InstrumentationName)
	}
	addr := observability.GetOutBoundIP()

	if b.Port != 0 {
		addr = fmt.Sprintf("%s:%d", addr, b.Port)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = b.extract(ctx)
		spanCtx, span := b.Tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(attribute.String("address", addr))

		defer func() {
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}
			span.End()
		}()

		resp, err = handler(spanCtx, req)

		return
	}
}

func (b *ServerOtelBuilder) extract(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(md))
}
