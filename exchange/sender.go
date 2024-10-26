package exchange

import (
	"github.com/JonHarder/glue/exchange/message"
)

// A Sender defines an interface for types which can send a message somewhere.
//
// Implementing this interface means your type can be used in the [To] method
// of an exchange.
type Sender interface {
	Send(m *message.Message) error
}

// FuncSender is a type alias allowing functions to operate as a Sender.
//
//   ex := exchange.New("test")
//   ex.From(receiver).
//      To(exchange.FuncSender(func(m *message.Message) error {
//          fmt.Printf("message: %v\n", m.Body)
//   }))
type FuncSender func(m *message.Message) error

func (f FuncSender) Send(m *message.Message) error {
	return f(m)
}
