package golog

import (
	"github.com/google/uuid"
	"time"
)

type RequestContext struct {
	correlationID string
	source        string
	method        string
	url           string
	start         time.Time
	userId        string
}

func New() RequestContext {
	return RequestContext{
		correlationID: uuid.New().String(),
		start:         time.Now(),
	}
}

func NewLinkedContext(source, method, url, correlationId string) RequestContext {
	ctx := RequestContext{}
	ctx.correlationID = correlationId
	ctx.SetSource(source)
	ctx.SetMethod(method)
	ctx.SetURL(url)
	ctx.start = time.Now()
	return ctx
}

func (c *RequestContext) SetSource(source string) {
	c.source = source
}

func (c *RequestContext) SetMethod(method string) {
	c.method = method
}

func (c *RequestContext) SetURL(url string) {
	c.url = url
}

func (c *RequestContext) SetCorrelationID(id string) {
	c.correlationID = id
}

func (c *RequestContext) SetUserID(id string) {
	c.userId = id
}

func (c *RequestContext) CorrelationID() string {
	return c.correlationID
}

func (c *RequestContext) Source() string {
	return c.source
}

func (c *RequestContext) StartTime() time.Time {
	return c.start
}

func (c *RequestContext) Method() string {
	return c.method
}

func (c *RequestContext) URL() string {
	return c.url
}

func (c *RequestContext) GetUserID() string {
	return c.userId
}
