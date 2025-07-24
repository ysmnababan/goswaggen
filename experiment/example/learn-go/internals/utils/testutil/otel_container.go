package testutil

import (
	"context"
	"fmt"

	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type OtelTestContainer struct {
	ctr      *testcontainers.DockerContainer
	port     nat.Port
	shutdown func(context.Context) error
	tracer   trace.Tracer
}

func StartOtelTestContainer(ctx context.Context) (*OtelTestContainer, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to get caller info")
	}

	baseDir := filepath.Dir(filename)
	configPath := filepath.Join(baseDir, "otel-collector-config.yaml")
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	ctr, err := testcontainers.Run(
		ctx,
		"otel/opentelemetry-collector:latest",
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("4317/tcp").WithStartupTimeout(time.Second*10)),
		testcontainers.WithExposedPorts("4317/tcp"),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            strings.NewReader(string(configBytes)),
			ContainerFilePath: "/etc/otelcol/config.yaml",
			FileMode:          0o644,
		}),
		testcontainers.CustomizeRequest(
			testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: []string{"--config=/etc/otelcol/config.yaml"},
				},
			},
		),
	)
	if err != nil {
		return nil, err
	}

	mappedPort, err := ctr.MappedPort(ctx, "4317/tcp")
	if err != nil {
		return nil, err
	}
	tracer, shutdown, err := initTracer(":" + mappedPort.Port())
	if err != nil {
		return nil, err
	}

	return &OtelTestContainer{
		ctr:      ctr,
		port:     mappedPort,
		shutdown: shutdown,
		tracer:   *tracer,
	}, nil
}

func (o *OtelTestContainer) Terminate() error {
	_ = o.shutdown(context.Background())
	return testcontainers.TerminateContainer(o.ctr)
}

func (o *OtelTestContainer) GetGRPCPort() int {
	return o.port.Int()
}

func (o *OtelTestContainer) GetTracer() trace.Tracer {
	return o.tracer
}

func initTracer(collectorUrl string) (*trace.Tracer, func(context.Context) error, error) {
	name := "tracer_test"
	secureOption := otlptracegrpc.WithInsecure()

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorUrl),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", name),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	// to prevent changing the global otel
	// get local traceProvider instead of 'otel.Tracer'
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	log.Info().Str("service.name", name).Msg("Tracing initialized")
	tracer := tp.Tracer("integration-test")
	return &tracer, exporter.Shutdown, nil
}
