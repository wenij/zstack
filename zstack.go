package zstack // import "github.com/shimmeringbee/zstack"

import (
	"context"
	"github.com/shimmeringbee/unpi/broker"
	"github.com/shimmeringbee/unpi/library"
	"io"
	"time"
)

type RequestResponder interface {
	MessageRequestResponse(ctx context.Context, req interface{}, resp interface{}) error
}

type ZStack struct {
	RequestResponder RequestResponder
}

const DefaultZStackTimeout = 5 * time.Second
const DefaultZStackRetries = 3

func New(uart io.ReadWriter) *ZStack {
	ml := library.NewLibrary()
	registerMessages(ml)

	znp := broker.NewBroker(uart, uart, ml)

	return &ZStack{
		RequestResponder: znp,
	}
}