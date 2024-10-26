package exchange

import (
	"github.com/JonHarder/glue/exchange/message"
)

// A Receiver defines an interface for types which can produce a Message.
//
// Receivers are used by the From method on Exchange objects.
// Receivers are designed to have their Receive method called repeatedly
// to produce Message objects.
type Receiver interface {
	// Gets a message.
	Receive() (*message.Message, error)
}

// FuncReceiver is a type alias allowing functions to operate as a Receiver.
//
//   ex := exchange.New("test")
//   id := 1
//   ex.From(exchange.FuncReceiver(func() (*message.Message, error) {
//       m := message.New(id)
//       id += 1
//       return m, nil
//   })).
//      To(sender)
type FuncReceiver func() (*message.Message, error)

func (f FuncReceiver) Receive() (*message.Message, error) {
	return f()
}
