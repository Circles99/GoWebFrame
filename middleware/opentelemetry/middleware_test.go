package opentelemetry

import (
	"GoWebFrame"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"log"
	"os"
	"testing"
	"time"
)

func TestOpenTelemetry(t *testing.T) {
	h := GoWebFrame.NewHttpServer()
	InitZipkin(t)
	tracer := otel.GetTracerProvider().Tracer("")
	h.Get("/user", func(ctx *GoWebFrame.Context) {
		c, span := tracer.Start(ctx.Req.Context(), "first_layer")
		defer span.End()
		c, second := tracer.Start(c, "second_layer")
		time.Sleep(time.Second)
		c, third1 := tracer.Start(c, "third_layer_1")
		time.Sleep(100 * time.Millisecond)
		third1.End()
		c, four := tracer.Start(c, "four_layer_1")
		time.Sleep(300 * time.Millisecond)
		four.End()
		second.End()
		ctx.RespStatusCode = 200
		ctx.RespData = []byte("hello")
	}, NewMiddleWareBuilder(tracer).Builder())

	h.Start(":8082")
}

func InitZipkin(t *testing.T) {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans", zipkin.WithLogger(log.New(os.Stderr, "opentelemetry", log.Ldate|log.Ltime|log.Llongfile)))
	if err != nil {
		t.Fatal(err)
	}
	batcher := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(batcher),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("opentelemetry-demo"),
		)))

	otel.SetTracerProvider(tp)
}
