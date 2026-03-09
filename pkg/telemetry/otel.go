package telemetry

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/version"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Otel struct {
	shutdown []func(context.Context) error
}

func (o *Otel) Shutdown(ctx context.Context) error {
	var err error
	for _, fn := range o.shutdown {
		err = errors.Join(err, fn(ctx))
	}
	return err
}

func New(ctx context.Context) (o *Otel, err error) {
	resourceAttrs := []attribute.KeyValue{
		attribute.String("service.version", version.Get().String()),
	}
	if strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")) == "" {
		resourceAttrs = append(resourceAttrs, attribute.String("service.name", "nanobot"))
	}

	resource, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(resourceAttrs...),
	)
	if err != nil {
		return nil, err
	}

	o = new(Otel)
	defer func() {
		if err != nil {
			err = errors.Join(err, o.Shutdown(context.Background()))
		}
	}()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	traceProvider, err := newTracerProvider(ctx, resource, exportEnabled())
	if err != nil {
		return nil, err
	}
	o.shutdown = append(o.shutdown, traceProvider.Shutdown)
	otel.SetTracerProvider(traceProvider)

	return o, nil
}

func exportEnabled() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_TRACES_EXPORTER")), "none") {
		return false
	}

	for _, k := range []string{
		"OTEL_TRACES_EXPORTER",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
	} {
		v := strings.TrimSpace(os.Getenv(k))
		if v == "" {
			continue
		}
		return true
	}
	return false
}

func newTracerProvider(ctx context.Context, resource *resource.Resource, enabled bool) (*sdktrace.TracerProvider, error) {
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(resource)}
	if !enabled {
		opts = append(opts, sdktrace.WithSampler(sdktrace.NeverSample()))
		return sdktrace.NewTracerProvider(opts...), nil
	}

	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, err
	}
	if autoexport.IsNoneSpanExporter(exporter) {
		opts = append(opts, sdktrace.WithSampler(sdktrace.NeverSample()))
		return sdktrace.NewTracerProvider(opts...), nil
	}

	opts = append(opts, sdktrace.WithBatcher(exporter))
	return sdktrace.NewTracerProvider(opts...), nil
}
