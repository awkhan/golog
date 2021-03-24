package golog

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/url"
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

func LogRequestWithHeaders(body interface{}, headers map[string][]string, ctx Context) {
	fields := append(createFields(body, headers, ctx))
	instance.Info("request", fields...)
}

func LogRequest(body interface{}, ctx Context) {
	LogRequestWithHeaders(body, nil, ctx)
}

func LogResponseWitHeaders(body interface{}, status int, headers map[string][]string, ctx Context) {
	fields := append(createFields(body, headers, ctx), zap.Int("status", status))
	instance.Info("response", fields...)
}

func LogResponse(body interface{}, status int, ctx Context) {
	fields := append(createFields(body, nil, ctx), zap.Int("status", status))
	instance.Info("response", fields...)
}

func LogError(err error, ctx Context) {
	fields := append(createFields(nil, nil, ctx), zap.String("error", err.Error()))
	instance.Info("error", fields...)
}

func LogInfo(message string, ctx Context) {
	fields := append(createFields(nil, nil, ctx), zap.String("message", message))
	instance.Info("info", fields...)
}

func createFields(body interface{}, headers map[string][]string, ctx Context) []zap.Field {
	fields := []zap.Field{
		zap.String("correlation_id", ctx.CorrelationID()),
		zap.String("source", ctx.Source()),
		zap.Duration("duration", time.Now().Sub(ctx.StartTime())),
	}

	if headers != nil {
		var s string
		for key, val := range headers {
			s = fmt.Sprintf("%s=%s ", key, val)
		}
		fields = append(fields, zap.String("headers", s))
	}

	if ctx.Method() != "" {
		fields = append(fields, zap.String("method", ctx.Method()))
	}

	if ctx.URL() != "" {
		fields = append(fields, zap.String("url", ctx.URL()))
	}

	if body != nil {
		out, _ := json.Marshal(body)
		fields = append(fields, zap.String("body", string(out)))
	}

	return fields
}
