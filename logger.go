package golog

import (
	"bytes"
	"encoding/json"
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
		outputPaths = append(outputPaths,  "golog://")
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

func LogRequest(body []byte, ctx Context) {
	fields := append(createFields(body, ctx))
	instance.Info("request", fields...)
}

func LogResponse(body []byte, status int, ctx Context) {
	fields := append(createFields(body, ctx), zap.Int("status", status))
	instance.Info("response", fields...)
}

func LogError(err error, ctx Context) {
	fields := append(createFields(nil, ctx), zap.String("error", err.Error()))
	instance.Info("error", fields...)
}

func LogInfo(message string, ctx Context) {
	fields := append(createFields(nil, ctx), zap.String("message", message))
	instance.Info("info", fields...)
}

func createFields(body []byte, ctx Context) []zap.Field {
	fields := []zap.Field{
		zap.String("correlation_id", ctx.CorrelationID()),
		zap.String("source", ctx.Source()),
		zap.Duration("duration", time.Now().Sub(ctx.StartTime())),
	}

	if ctx.Method() != "" {
		fields = append(fields, zap.String("method", ctx.Method()))
	}

	if ctx.URL() != "" {
		fields = append(fields, zap.String("url", ctx.URL()))
	}

	if body != nil {
		buffer := new(bytes.Buffer)
		err := json.Compact(buffer, body)
		if err != nil {
			// Not json formatted. Lets print it as just a string
			buffer.Write(body)
		}
		fields = append(fields, zap.Any("body", buffer.String()))
	}

	return fields
}

