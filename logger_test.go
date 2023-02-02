package golog

import (
	"errors"
	"net/url"
	"testing"
	"time"
)

func Test_Logger(t *testing.T) {

	LogError(&ctx{}, errors.New("random error"))
	LogInfo(&ctx{}, "a message for info")

	r := random{
		String: "random-string-here",
		Number: 500,
		Time:   time.Now(),
	}

	u, _ := url.Parse("https://google.com/random/path?query=1")
	LogRequest(&ctx{}, r, "GET", *u)

	LogResponse(&ctx{}, r, 204)

}

type random struct {
	String string    `json:"string"`
	Number int       `json:"number"`
	Time   time.Time `json:"time"`
}

type ctx struct{}

func (c *ctx) CorrelationID() string { return "cid" }
func (c *ctx) StartTime() time.Time  { return time.Now() }
