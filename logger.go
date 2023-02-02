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
	StartTime() time.Time
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
			StacktraceKey:  zapcore.OmitKey,
			MessageKey:     "message",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, _ := cfg.Build(zap.WithCaller(false))
	defer logger.Sync()
	instance = logger

}

func LogRequest(ctx Context, method string, url url.URL, body []byte) {
	instance.Info(string(body), createFields(ctx, body, nil, &method, &url)...)
}

func LogResponse(ctx Context, body []byte, status int) {
	instance.Info(string(body), createFields(ctx, body, &status, nil, nil)...)
}

func LogError(ctx Context, err error) {
	instance.Error(err.Error(), createFields(ctx, nil, nil, nil, nil)...)
}

func LogInfo(ctx Context, message string, data []byte) {
	instance.Info(message, createFields(ctx, data, nil, nil, nil)...)
}

func LogWarning(ctx Context, message string) {
	fields := append(createFields(ctx, nil, nil, nil, nil), zap.String("message", message))
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

func createFields(ctx Context, data []byte, httpStatus *int, method *string, url *url.URL) []zap.Field {
	fields := []zap.Field{
		zap.String("correlation_id", ctx.CorrelationID()),
		//zap.String("source", ctx.Source()),
		zap.Duration("duration", time.Now().Sub(ctx.StartTime())),
	}

	if method != nil {
		fields = append(fields, zap.String("http.method", *method))
	}

	if url != nil {
		fields = append(fields, zap.String("http.url_details.host", url.Host))
		fields = append(fields, zap.String("http.url_details.path", url.Path))
		fields = append(fields, zap.String("http.url_details.queryString", url.RawQuery))
	}

	if data != nil {
		fields = append(fields, zap.ByteString("body", data))
	}

	if httpStatus != nil {
		fields = append(fields, zap.Int("http.status_code", *httpStatus))
	}

	return fields
}
