package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otlptrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	logapi "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultOTLPEndpoint     = "http://localhost:4318"
	defaultServiceName      = "potato"
	defaultServiceVersion   = "1.0.0"
	defaultEnvironment      = "development"
	instrumentationName     = "github.com/williamdumont/potato-demo"
	defaultMetricExportFreq = 15 * time.Second
	exporterInitTimeout     = 10 * time.Second
)

type telemetryConfig struct {
	Endpoint        string
	ServiceName     string
	ServiceVersion  string
	Environment     string
	ExtraAttributes []attribute.KeyValue
}

type Observability struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider

	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	errorCounter    metric.Int64Counter

	logger      logapi.Logger
	serviceName string
}

func initOpenTelemetry(ctx context.Context) (*Observability, error) {
	cfg := loadTelemetryConfig()

	res, err := buildResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	traceExp, err := newTraceExporter(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("init trace exporter: %w", err)
	}

	metricExp, err := newMetricExporter(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("init metric exporter: %w", err)
	}

	logExp, err := newLogExporter(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("init log exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	reader := sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(defaultMetricExportFreq))
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	otel.SetMeterProvider(meterProvider)

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	logglobal.SetLoggerProvider(loggerProvider)

	meter := meterProvider.Meter(instrumentationName)

	requestCounter, err := meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Total number of HTTP requests processed by the service"),
	)
	if err != nil {
		return nil, fmt.Errorf("create request counter: %w", err)
	}

	requestDuration, err := meter.Float64Histogram(
		"http.server.request_duration_ms",
		metric.WithDescription("Distribution of HTTP request latency in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("create request duration histogram: %w", err)
	}

	errorCounter, err := meter.Int64Counter(
		"http.server.errors",
		metric.WithDescription("Total number of HTTP requests that returned an error status (>= 400)"),
	)
	if err != nil {
		return nil, fmt.Errorf("create error counter: %w", err)
	}

	telemetry := &Observability{
		tracerProvider:  tracerProvider,
		meterProvider:   meterProvider,
		loggerProvider:  loggerProvider,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
		errorCounter:    errorCounter,
		logger:          loggerProvider.Logger(instrumentationName),
		serviceName:     cfg.ServiceName,
	}

	return telemetry, nil
}

func (o *Observability) Shutdown(ctx context.Context) error {
	var errs []error
	if o == nil {
		return nil
	}

	if o.loggerProvider != nil {
		if err := o.loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if o.meterProvider != nil {
		if err := o.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if o.tracerProvider != nil {
		if err := o.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (o *Observability) WrapHandler(name string, handler http.HandlerFunc) http.Handler {
	if o == nil || handler == nil {
		return handler
	}

	instrumented := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		recorder := newResponseRecorder(w)
		start := time.Now()

		defer func() {
			if rec := recover(); rec != nil {
				recorder.statusCode = http.StatusInternalServerError
				duration := time.Since(start)
				o.recordRequest(ctx, name, r.Method, recorder.statusCode, duration)
				panic(rec)
			}
			duration := time.Since(start)
			o.recordRequest(ctx, name, r.Method, recorder.statusCode, duration)
		}()

		handler(recorder, r)
	})

	return otelhttp.NewHandler(instrumented, name)
}

func (o *Observability) recordRequest(ctx context.Context, route, method string, status int, duration time.Duration) {
	if o == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.route", route),
		attribute.String("http.method", method),
		attribute.Int("http.status_code", status),
		attribute.String("service.name", o.serviceName),
	}

	if o.requestCounter != nil {
		o.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	if o.requestDuration != nil {
		o.requestDuration.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
	}

	if o.errorCounter != nil && status >= http.StatusBadRequest {
		o.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	o.logRequest(ctx, route, method, status, duration)
}

func (o *Observability) logRequest(ctx context.Context, route, method string, status int, duration time.Duration) {
	if o == nil || o.logger == nil {
		return
	}

	record := logapi.Record{}
	record.SetTimestamp(time.Now())
	record.SetBody(logapi.StringValue(fmt.Sprintf("%s %s", method, route)))

	switch {
	case status >= 500:
		record.SetSeverity(logapi.SeverityError)
		record.SetSeverityText("ERROR")
	case status >= 400:
		record.SetSeverity(logapi.SeverityWarn)
		record.SetSeverityText("WARN")
	default:
		record.SetSeverity(logapi.SeverityInfo)
		record.SetSeverityText("INFO")
	}

	record.AddAttributes(
		logapi.String("http.route", route),
		logapi.String("http.method", method),
		logapi.Int("http.status_code", status),
		logapi.Float64("duration_ms", float64(duration.Microseconds())/1000),
		logapi.String("service.name", o.serviceName),
	)

	if span := trace.SpanFromContext(ctx); span != nil {
		if sc := span.SpanContext(); sc.IsValid() {
			record.AddAttributes(
				logapi.String("trace_id", sc.TraceID().String()),
				logapi.String("span_id", sc.SpanID().String()),
			)
		}
	}

	o.logger.Emit(ctx, record)
}

func loadTelemetryConfig() telemetryConfig {
	cfg := telemetryConfig{
		Endpoint:       getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", defaultOTLPEndpoint),
		ServiceName:    getEnv("OTEL_SERVICE_NAME", defaultServiceName),
		ServiceVersion: getEnv("OTEL_SERVICE_VERSION", defaultServiceVersion),
		Environment:    getEnv("DEPLOYMENT_ENVIRONMENT", defaultEnvironment),
	}

	extra := parseResourceAttributes(os.Getenv("OTEL_RESOURCE_ATTRIBUTES"))
	for _, attr := range extra {
		if string(attr.Key) == "deployment.environment" {
			cfg.Environment = attr.Value.AsString()
			break
		}
	}
	cfg.ExtraAttributes = extra

	return cfg
}

func buildResource(ctx context.Context, cfg telemetryConfig) (*resource.Resource, error) {
	base := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
		attribute.String("deployment.environment", cfg.Environment),
	}

	attrs := mergeAttributes(base, cfg.ExtraAttributes)

	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithHost(),
		resource.WithAttributes(attrs...),
	)
}

func mergeAttributes(base, extra []attribute.KeyValue) []attribute.KeyValue {
	merged := make(map[string]attribute.KeyValue, len(base)+len(extra))

	for _, kv := range base {
		merged[string(kv.Key)] = kv
	}

	for _, kv := range extra {
		merged[string(kv.Key)] = kv
	}

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]attribute.KeyValue, 0, len(keys))
	for _, key := range keys {
		out = append(out, merged[key])
	}

	return out
}

func parseResourceAttributes(raw string) []attribute.KeyValue {
	if raw == "" {
		return nil
	}

	pairs := strings.Split(raw, ",")
	attrs := make([]attribute.KeyValue, 0, len(pairs))

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" || value == "" {
			continue
		}

		attrs = append(attrs, attribute.String(key, value))
	}

	return attrs
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func newTraceExporter(ctx context.Context, endpoint string) (*otlptrace.Exporter, error) {
	expCtx, cancel := context.WithTimeout(ctx, exporterInitTimeout)
	defer cancel()
	return otlptracehttp.New(expCtx, otlptracehttp.WithEndpointURL(endpoint))
}

func newMetricExporter(ctx context.Context, endpoint string) (*otlpmetrichttp.Exporter, error) {
	expCtx, cancel := context.WithTimeout(ctx, exporterInitTimeout)
	defer cancel()
	return otlpmetrichttp.New(expCtx, otlpmetrichttp.WithEndpointURL(endpoint))
}

func newLogExporter(ctx context.Context, endpoint string) (*otlploghttp.Exporter, error) {
	expCtx, cancel := context.WithTimeout(ctx, exporterInitTimeout)
	defer cancel()

	// Note: otlploghttp.New behaves differently than trace/metric exporters.
	// It doesn't automatically append /v1/logs when using WithEndpointURL.
	// We need to manually append the path or use WithEndpoint + WithURLPath.
	logEndpoint := endpoint + "/v1/logs"
	return otlploghttp.New(expCtx, otlploghttp.WithEndpointURL(logEndpoint))
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
