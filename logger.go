package golog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
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

type Sink interface {
	Initialize()
	Name() string
	zapcore.WriteSyncer
	io.Closer
}

var instance *zap.Logger

func Initialize(sinks []Sink) {

	outputPaths := []string{"stderr"}
	for _, s := range sinks {
		zap.RegisterSink(s.Name(), func(url *url.URL) (zap.Sink, error) {
			return s, nil
		})
		outputPaths = append(outputPaths, fmt.Sprintf("%s://", s.Name()))
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

