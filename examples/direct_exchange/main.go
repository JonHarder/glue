package main

import (
	"context"
	"log"
	"time"

	"github.com/JonHarder/glue/exchange"
	"github.com/JonHarder/glue/exchange/message"
)

func main() {
	ex1 := exchange.New("test")
	ex1.From(exchange.FuncReceiver(func() (*message.Message, error) {
		m := message.New(4)
		time.Sleep(time.Second)
		return m, nil
	})).
		To(exchange.Direct("ex1"))

	ex2 := exchange.New("test2")

	ex2.From(exchange.Direct("ex1")).
		To(exchange.FuncSender(func(m *message.Message) error {
			log.Printf("Got message: %v", m.Body)
			return nil
		}))

	ctx := context.Background()
	go ex1.Run(ctx)
	ex2.Run(ctx)
}
