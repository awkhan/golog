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
	LogRequest(&ctx{}, b, "GET", *u)
	LogResponse(&ctx{}, b, 204)

	nr := []random{r, r, r}
	b, _ = json.Marshal(nr)
	LogRequest(&ctx{}, b, "GET", *u)
	LogResponse(&ctx{}, b, 204)

	nr = []random{}
	b, _ = json.Marshal(nr)
	LogRequest(&ctx{}, b, "GET", *u)
	LogResponse(&ctx{}, b, 204)
}

type random struct {
	String string    `json:"string"`
	Number int       `json:"number"`
	Time   time.Time `json:"time"`
}

type ctx struct{}

func (c *ctx) CorrelationID() string { return "cid" }
func (c *ctx) StartTime() time.Time  { return time.Now() }
func (c *ctx) UserID() *string {
	t := "test"
	return &t
}
