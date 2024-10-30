package exchange

import (
	"context"
	"testing"

	"github.com/JonHarder/glue/exchange/message"
)

type constantReceiver struct {
	always interface{}
}

func (c *constantReceiver) Receive() (*message.Message, error) {
	return message.New(c.always), nil
}

func TestExchangeWithoutFromOrTwoPanics(t *testing.T) {
	ex := New("test")
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
	ex := New("test")
	var msg message.Message
	ex.From(n).To(FuncSender(func(m *message.Message) error {
		msg = *m
		return nil
	}))

	ctx := context.Background()
	ctx = context.WithValue(ctx, "RUN_ONCE", true)
	ex.Run(ctx)

	if msg.Body.(int) != expect {
		t.Errorf("Message body expected: %v, got: %v", expect, msg.Body.(int))
	}
}
