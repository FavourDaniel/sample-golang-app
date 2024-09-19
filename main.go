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
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func createFlatSpansIterative(ctx context.Context, tracer trace.Tracer, count, offset int) {
	for i := 0; i < count; i++ {
		_, span := tracer.Start(ctx, fmt.Sprintf("flat-span-%d", offset+i))
		span.SetAttributes(attribute.Int("numbered", offset+i))
		span.End()
		if (offset+i+1)%1000 == 0 {
			fmt.Printf("Created span %d\n", offset+i+1)
		}
	}
}

// Rename the function to avoid redeclaration
func httpHandlerWithSpans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("github.com/SigNoz/sample-golang-app/controllers")
	// Start with the root span
	ctx, rootSpan := tracer.Start(ctx, "FindBooks-root")
	defer rootSpan.End()

	totalSpans := 20000
	batchSize := 1000
	for i := 0; i < totalSpans; i += batchSize {
		count := batchSize
		if i+batchSize > totalSpans {
			count = totalSpans - i
		}
		createFlatSpansIterative(ctx, tracer, count, i)

		// Force flush after each batch
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))
	}

	fmt.Fprintf(w, "Hello, World")
}

func main() {
	// use otelconfig to setup OpenTelemetry SDK
	otelShutdown, err := otelconfig.ConfigureOpenTelemetry()
	if err != nil {
		log.Fatalf("error setting up OTel SDK - %e", err)
	}
	defer otelShutdown()

	// Initialize HTTP handler instrumentation
	handler := http.HandlerFunc(httpHandlerWithSpans)
	wrappedHandler := otelhttp.NewHandler(handler, "hello")
	http.Handle("/hello", wrappedHandler)

	log.Fatal(http.ListenAndServe(":3030", nil))
}
