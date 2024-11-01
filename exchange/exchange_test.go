package exchange

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JonHarder/glue/exchange/message"
)

type constantReceiver struct {
	always interface{}
}

func (c *constantReceiver) Receive() (*message.Message, error) {
	return message.New(c.always), nil
}

func TestExchangeWithoutFromOrTwoPanics(t *testing.T) {
	ex := New("TestExchangeWithoutFromOrTwoPanics")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("An exchange without a from and to didn't panic when ran.")
		}
	}()

	// The following is the code under test
	ex.Run(context.Background())
}

func TestMessageIntegrityThroughSimpleExchange(t *testing.T) {
	expect := 4
	n := &constantReceiver{always: expect}
	ex := New("TestMessageIntegrityThroughSimpleExchange")
	var msg message.Message
	ex.From(n).To(FuncSender(func(m *message.Message) error {
		msg = *m
		return nil
	}))

	ctx := context.Background()
	ex.RunOnce(ctx)

	if msg.Body.(int) != expect {
		t.Errorf("Message body expected: %v, got: %v", expect, msg.Body.(int))
	}
}

func TestContextCancelExitsExchange(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()
	ex := New("TestContextCancelExitsExchange")
	ex.From(FuncReceiver(func() (*message.Message, error) {
		return message.New(1), nil
	})).To(FuncSender(func(m *message.Message) error {
		return nil
	}))
	ex.Run(ctx)
}

func TestErrorReceiverLogsError(t *testing.T) {
	ex := New("TestErrorReceiverLogsError")
	ex.From(FuncReceiver(func() (*message.Message, error) {
		return nil, fmt.Errorf("Failed to receive message")
	})).To(FuncSender(func(m *message.Message) error {
		return nil
	}))
	ctx := context.Background()
	ex.RunOnce(ctx)
}
