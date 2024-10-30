package main

import (
	"context"
	"log"
	"time"

	"github.com/JonHarder/glue/exchange"
	"github.com/JonHarder/glue/exchange/message"
)

// incrementalReceiver stores an integer for use as a exchange.Receiver.
type incrementalReceiver struct {
	state int
}

// incrementalReceiver.Receive
func (incRec *incrementalReceiver) Receive() (*message.Message, error) {
	time.Sleep(time.Second)
	x := incRec.state
	incRec.state += 1
	m := message.New(x)
	log.Printf("Receiving message with body: %v", m.Body)
	return m, nil
}

type printSender struct{}

func (ps *printSender) Send(m *message.Message) error {
	log.Printf("Got a message with body: %v\n", m.Body)
	time.Sleep(time.Second)
	return nil
}

func main() {
	ex := exchange.New("test")

	ex.From(&incrementalReceiver{state: 0}).
		To(&printSender{})

	ctx := context.Background()
	ex.Run(ctx)
}
