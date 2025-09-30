package golog

import (
	"encoding/json"
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

	b, _ := json.Marshal(r)

	u, _ := url.Parse("https://google.com/random/path?query=1")
	LogRequest(&ctx{u: u}, b)
	LogResponse(&ctx{}, b, 204)

	nr := []random{r, r, r}
	b, _ = json.Marshal(nr)
	LogRequest(&ctx{}, b)
	LogResponse(&ctx{}, b, 204)

	nr = []random{}
	b, _ = json.Marshal(nr)
	LogRequest(&ctx{}, b)
	LogResponse(&ctx{}, b, 204)
}

type random struct {
	String string    `json:"string"`
	Number int       `json:"number"`
	Time   time.Time `json:"time"`
}

type ctx struct {
	u *url.URL
}

func (c *ctx) URL() *url.URL {
	return c.u
}

func (c *ctx) HTTPMethod() *string {
	return nil
}

func (c *ctx) CorrelationID() string { return "cid" }
func (c *ctx) StartTime() time.Time  { return time.Now() }
func (c *ctx) UserID() *string {
	t := "test"
	return &t
}
func (c *ctx) UserIPAddress() *string {
	t := "127.0.0.1"
	return &t
}
func (c *ctx) OtherFields() map[string]interface{} {
	return map[string]interface{}{
		"test": "test",
		"num":  1,
		"bool": true,
		"obj":  random{},
	}
}
