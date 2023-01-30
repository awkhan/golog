package golog

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Logger_NoInit(t *testing.T) {
	LogError(&ctx{}, errors.New("random error"))
}

func Test_Logger(t *testing.T) {

	called := false
	sf := func(b []byte) {
		called = true
	}

	Initialize(sf)
	LogError(&ctx{}, errors.New("random error"))
	assert.True(t, called)

	type anonBody struct {
		One string `json:"one"`
		Two int    `json:"two"`
	}
	body := anonBody{
		One: "one",
		Two: 1,
	}

	b, _ := json.Marshal(body)

	LogInfo(&ctx{}, "a message for info", b)

}

type ctx struct{}

func (c *ctx) CorrelationID() string { return "cid" }
func (c *ctx) Source() string        { return "source" }
func (c *ctx) StartTime() time.Time  { return time.Now() }
func (c *ctx) Method() string        { return "method" }
func (c *ctx) URL() string           { return "u" }
