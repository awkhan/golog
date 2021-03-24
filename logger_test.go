package golog

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Logger_NoInit(t *testing.T) {
	LogError(errors.New("random error"), &ctx{})
}

func Test_Logger(t *testing.T) {

	called := false
	sf := func(b []byte) {
		called = true
	}

	Initialize(sf)
	LogError(errors.New("random error"), &ctx{})
	assert.True(t, called)

	type anonBody struct {
		One string `json:"one"`
		Two int    `json:"two"`
	}
	body := anonBody{
		One: "one",
		Two: 1,
	}
	LogRequest(body, &ctx{})
	LogRequestWithHeaders(body, map[string][]string{"test-header": {"a", "b"}}, &ctx{})
	LogResponse(body, 200, &ctx{})
	LogInfo("a message for info", &ctx{})

}

type ctx struct{}

func (c *ctx) CorrelationID() string { return "cid" }
func (c *ctx) Source() string        { return "source" }
func (c *ctx) StartTime() time.Time  { return time.Now() }
func (c *ctx) Method() string        { return "method" }
func (c *ctx) URL() string           { return "u" }
