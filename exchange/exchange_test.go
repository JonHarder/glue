package exchange

import (
	"github.com/JonHarder/glue/exchange/message"
	"testing"
)

type nilReceiver struct{}

func (n *nilReceiver) Receive() (*message.Message, error) {
	return message.New(nil), nil
}

func TestExchangeWithoutFromOrTwoPanics(t *testing.T) {
	ex := New("test")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("An exchange without a from and to didn't panic when ran.")
		}
	}()

	// The following is the code under test
	ex.Run()
}

func TestMessageIntegrityThroughSimpleExchange(t *testing.T) {
	t.Error("Not sure how to just run exchange processing once.")
}
