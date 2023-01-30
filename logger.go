package golog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
	"time"
)

type Type string

const (
	TypeError   Type = "error"
	TypeInfo    Type = "info"
	TypeWarning Type = "warning"
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

	outputPaths := []string{"stdout"}
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
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "status",
			NameKey:        zapcore.OmitKey,
			CallerKey:      zapcore.OmitKey,
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, _ := cfg.Build(zap.WithCaller(false))
	defer logger.Sync()
	instance = logger

}

func LogRequest(ctx Context, body []byte) {
	instance.Info("Received request", createFields(ctx, body, nil)...)
}

func LogResponse(ctx Context, body []byte, status int) {
	instance.Info("Response", createFields(ctx, body, &status)...)
}

func LogError(ctx Context, err error) {
	instance.Error(err.Error(), createFields(ctx, nil, nil)...)
}

func LogInfo(ctx Context, message string, data []byte) {
	instance.Info(message, createFields(ctx, data, nil)...)
}

func LogWarning(ctx Context, message string) {
	fields := append(createFields(ctx, nil, nil), zap.String("message", message))
	instance.Warn("warning", fields...)
}

func LogReturn(ctx Context, t Type, err error) error {
	switch t {
	case TypeError:
		LogError(ctx, err)
	case TypeWarning:
		LogWarning(ctx, err.Error())
	case TypeInfo:
		LogInfo(ctx, err.Error(), nil)
	}
	return err
}

func createFields(ctx Context, data []byte, httpStatus *int) []zap.Field {
	fields := []zap.Field{
		zap.String("correlation_id", ctx.CorrelationID()),
		zap.String("source", ctx.Source()),
		zap.Duration("duration", time.Now().Sub(ctx.StartTime())),
	}

	if ctx.Method() != "" {
		fields = append(fields, zap.String("http.method", ctx.Method()))
	}

	if ctx.URL() != "" {
		fields = append(fields, zap.String("http.url", ctx.URL()))
	}

	if data != nil {
		fields = append(fields, zap.ByteString("body", data))
	}

	if httpStatus != nil {
		fields = append(fields, zap.Int("http.status_code", *httpStatus))
	}

	return fields
}
