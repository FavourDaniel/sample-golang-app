package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/honeycombio/otel-config-go/otelconfig"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func createNestedSpans(ctx context.Context, tracer trace.Tracer, remaining int, current int) {
	if remaining == 0 {
		return
	}

	_, span := tracer.Start(ctx, fmt.Sprintf("nested-span-%d", current))
	defer span.End()

	span.SetAttributes(attribute.Int("depth", current))

	// Recursively create the next nested span using the current span's context
	createNestedSpans(trace.ContextWithSpan(ctx, span), tracer, remaining-1, current+1)
}

// Implement an HTTP Handler func to be instrumented
func httpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("github.com/SigNoz/sample-golang-app/controllers")

	// Start with the root span
	ctx, rootSpan := tracer.Start(ctx, "FindBooks-root")
	defer rootSpan.End()

	// Create 1000 nested spans
	createNestedSpans(ctx, tracer, 1000, 0)
	fmt.Fprintf(w, "Hello, World")
}

// Wrap the HTTP handler func with OTel HTTP instrumentation
func wrapHandler() {
	handler := http.HandlerFunc(httpHandler)
	wrappedHandler := otelhttp.NewHandler(handler, "hello")
	http.Handle("/hello", wrappedHandler)
}

func main() {
	// use otelconfig to setup OpenTelemetry SDK
	otelShutdown, err := otelconfig.ConfigureOpenTelemetry()
	if err != nil {
		log.Fatalf("error setting up OTel SDK - %e", err)
	}
	defer otelShutdown()

	// Initialize HTTP handler instrumentation
	wrapHandler()
	log.Fatal(http.ListenAndServe(":3030", nil))
}
