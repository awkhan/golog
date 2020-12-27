package golog

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func runNatsServerOnPort(port int) error {
	opts := server.Options{Port: port, Host: "127.0.0.1"}
	s, err := server.NewServer(&opts)
	if err != nil {
		return  err
	}
	go s.Start()
	time.Sleep(1 *time.Second) // Give time for nats server to come up
	return nil
}

func Test_NatsSink(t *testing.T) {

	port := 5688
	context := fmt.Sprintf("http://127.0.0.1:%d", port)
	err := runNatsServerOnPort(port)
	require.Nil(t, err)

	ns := NewNatsSink(context, "logger-subject")

	Initialize([]Sink{&ns})

	nc, err := nats.Connect(context)
	require.Nil(t, err)

	subCh := make(chan *nats.Msg)
	nc.ChanSubscribe("logger-subject", subCh)

	done := make(chan bool)
	go func() {
		<- subCh
		done <- true
	}()

	go func() {
		time.Sleep(1 *time.Second)
		LogError(errors.New("random error"), &ctx{})
	}()
	
	assert.True(t, <-done)

}

type ctx struct {}
func (c *ctx) CorrelationID() string{ return "cid"}
func (c *ctx) Source() string{ return "source"}
func (c *ctx) StartTime() time.Time { return time.Now() }
func (c *ctx) Method() string { return "method"}
func (c *ctx) URL() string { return "u"}