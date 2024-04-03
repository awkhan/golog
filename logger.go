package golog

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
	"strings"
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
	UserID() *string
	URL() *url.URL
	HTTPMethod() *string
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

func LogRequest(ctx Context, body []byte, method string) {
	instance.Info(parseData(body), createFields(ctx, nil)...)
}

func LogResponse(ctx Context, body []byte, status int) {
	instance.Info(parseData(body), createFields(ctx, &status)...)
}

func LogError(ctx Context, err error) {
	instance.Error(err.Error(), createFields(ctx, nil)...)
}

func LogInfo(ctx Context, message string) {
	instance.Info(message, createFields(ctx, nil)...)
}

func LogWarning(ctx Context, message string) {
	fields := append(createFields(ctx, nil), zap.String("message", message))
	instance.Warn("warning", fields...)
}

func LogReturn(ctx Context, t Type, err error) error {
	switch t {
	case TypeError:
		LogError(ctx, err)
	case TypeWarning:
		LogWarning(ctx, err.Error())
	case TypeInfo:
		LogInfo(ctx, err.Error())
	}
	return err
}

func parseData(d []byte) string {
	var m map[string]interface{}
	err := json.Unmarshal(d, &m)
	if err != nil {
		// it's probably an array instead of a map
		var ma []map[string]interface{}
		json.Unmarshal(d, &ma)

		if ma == nil || len(ma) == 0 {
			return ""
		}

		s := ""
		for _, v := range ma {
			s = fmt.Sprintf("%s,%s", s, mapToString(v))
		}
		return fmt.Sprintf("[%s]", strings.TrimLeft(s, ", "))
	} else {
		return strings.TrimLeft(mapToString(m), "")
	}
}

func mapToString(m map[string]interface{}) string {
	s := ""
	for k, v := range m {
		s = fmt.Sprintf("%s %s=%v", s, k, v)
	}
	return s
}

func createFields(ctx Context, httpStatus *int) []zap.Field {
	fields := []zap.Field{
		zap.String("correlation_id", ctx.CorrelationID()),
		zap.Duration("duration", time.Now().Sub(ctx.StartTime())),
	}

	userID := ctx.UserID()
	if userID != nil {
		fields = append(fields, zap.String("user_id", *userID))
	}

	method := ctx.HTTPMethod()
	if method != nil {
		fields = append(fields, zap.String("http.method", *method))
	}

	u := ctx.URL()
	if u != nil {
		fields = append(fields, zap.String("http.url_details.host", u.Host))
		fields = append(fields, zap.String("http.url_details.path", u.Path))
		fields = append(fields, zap.String("http.url_details.queryString", u.RawQuery))
	}

	if httpStatus != nil {
		fields = append(fields, zap.Int("http.status_code", *httpStatus))
	}

	return fields
}
