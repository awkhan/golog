package golog

import (
	"fmt"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"time"
)

type Context interface {
	CorrelationID() string
	Source() string
	StartTime() time.Time
	Method() string
	URL() string
}

type sink struct {
	sf sinkFunc
}

func (s *sink) Close() error {
	return nil
}

func (s *sink) Write(b []byte) (int, error) {
	s.sf(b)
	return 0, nil
}

func (s *sink) Sync() error {
	return nil
}

var instance *zap.Logger

type sinkFunc func(b []byte)

func init() {
	Initialize(nil)
}

func Initialize(sf sinkFunc) {

	outputPaths := []string{"stderr"}
	if sf != nil {
		outputPaths = append(outputPaths, "golog://")
		zap.RegisterSink("golog", func(url *url.URL) (zap.Sink, error) {
			s := sink{sf: sf}
			return &s, nil
		})
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, _ := cfg.Build(zap.WithCaller(false))
	defer logger.Sync()
	instance = logger

}

func LogRequestWithHeaders(ctx Context, body []byte, headers map[string][]string) {
	fields := append(createFields(ctx, body, headers))
	instance.Info("request", fields...)
}

func LogRequest(ctx Context, body []byte) {
	LogRequestWithHeaders(ctx, body, nil)
}

func LogResponseWitHeaders(ctx Context, body []byte, status int, headers map[string][]string) {
	fields := append(createFields(ctx, body, headers), zap.Int("status", status))
	instance.Info("response", fields...)
}

func LogResponse(ctx Context, body []byte, status int) {
	fields := append(createFields(ctx, body, nil), zap.Int("status", status))
	instance.Info("response", fields...)
}

func LogError(ctx Context, err error) {
	fields := append(createFields(ctx, nil, nil), zap.String("error", err.Error()))
	instance.Info("error", fields...)
}

func LogInfo(ctx Context, message string) {
	fields := append(createFields(ctx, nil, nil), zap.String("message", message))
	instance.Info("info", fields...)
}

func createFields(ctx Context, body []byte, headers map[string][]string) []zap.Field {
	fields := []zap.Field{
		zap.String("correlation_id", ctx.CorrelationID()),
		zap.String("source", ctx.Source()),
		zap.Duration("duration", time.Now().Sub(ctx.StartTime())),
	}

	if headers != nil {
		var s string
		for key, val := range headers {
			s = fmt.Sprintf("%s %s=%s", s, key, val)
		}
		fields = append(fields, zap.String("headers", strings.Trim(s, " ")))
	}

	if ctx.Method() != "" {
		fields = append(fields, zap.String("method", ctx.Method()))
	}

	if ctx.URL() != "" {
		fields = append(fields, zap.String("url", ctx.URL()))
	}

	if body != nil {
		fields = append(fields, zap.ByteString("body", body))
	}

	return fields
}
