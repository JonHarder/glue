package main

import (
	"log"
	"time"

	"github.com/JonHarder/glue/exchange"
	"github.com/JonHarder/glue/exchange/message"
)

func main() {
	ex := exchange.New("functional")
	i := 0
	ex.From(exchange.FuncReceiver(func() (*message.Message, error) {
		m := message.New(i)
		i += 1
		time.Sleep(time.Second)
		log.Printf("Received message with body: %v", m.Body)
		return m, nil
	})).
		Process(exchange.FuncProcessor(func(m *message.Message) error {
			log.Printf("processing message with body: %v", m.Body)
			return nil
		})).
		To(exchange.FuncSender(func(m *message.Message) error {
			time.Sleep(time.Second)
			log.Printf("sending message with body: %v", m.Body)
			return nil
		}))

	ex.Run()
}
