package golog

import "github.com/nats-io/nats.go"

type NatsSink struct {
	nc *nats.Conn
	subject string
}

func NewNatsSink(context, subject string) NatsSink {
	conn, _ := nats.Connect(context)
	return NatsSink{nc: conn, subject: subject}
}

func (n *NatsSink) Name() string {
	return "nats"
}

func (n *NatsSink) Close() error {
	return nil
}

func (n *NatsSink) Write(b []byte) (int, error) {
	n.nc.Publish(n.subject, b)
	return 0, nil
}

func (n *NatsSink) Sync() error {
	return nil
}