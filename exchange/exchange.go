// Package exchange provides functionality to construct data pipelines.
//
// Data comes into the pipeline from a Receiver then gets transformed through any number of Processor transformations, and finally gets delivered out of the end of the pipeline via a Sender.
// Interanlly, data is stored and trasferred through the pipeline in a Message object.
//
package exchange

import (
	"github.com/JonHarder/glue/exchange/message"
	"log"
)

var directExchangeMap map[string]chan message.Message

// A Processor defines an interface for types which process Message objects.
//
// Processors modify a message in place, and can perform arbitrary side effects
// if desired.
type Processor interface {
	Process(m *message.Message) error
}

// FuncProcessor is a type alias enabling anonymous functions to act as Processors
type FuncProcessor func(m *message.Message) error

func (f FuncProcessor) Process(m *message.Message) error {
	return f(m)
}

// SendReceiver is a composite interface to denote a type implements
// functionality for sending and receiving messages.
type SendReceiver interface {
	Sender
	Receiver
}

type toFromExchange struct {
	buffer chan message.Message
}

func (e *toFromExchange) Receive() (*message.Message, error) {
	m := <-e.buffer
	return &m, nil
}

func (e *toFromExchange) Send(m *message.Message) error {
	e.buffer <- *m
	return nil
}

// Direct creates a [SendReceiver] which sends or receives messages to other
// exchanges by name.
//
//   ex1 := exchange.New("test")
//   ex1.From(receiver).
//      .To(exchange.Direct("other-exchange"))
//
//   ex2 := exchange.New("test2")
//   ex2.From(exchange.Direct("other-exchange")).
//       To(sender)
//
//   go ex1.Run()
//   ex2.Run()
func Direct(name string) SendReceiver {
	buf, ok := directExchangeMap[name]
	if !ok {
		buf = make(chan message.Message)
		directExchangeMap[name] = buf
	}
	return &toFromExchange{buffer: buf}
}

// An Exchange represents a pipeline which connects an input and output stream.
type Exchange struct {
	Name       string
	receiver   Receiver
	sender     Sender
	processors []Processor
	inBuffer   chan message.Message
	outBuffer  chan message.Message
}

// From executes the provided Receiver which acts as the input stream for message.Messages.
//
// Running an exchange using Run without a From causes a panic.
func (ex *Exchange) From(src Receiver) *Exchange {
	ex.receiver = src
	return ex
}

// To executes the provided Sender which acts as the output stream for message.Messages.
//
// Running an exchange using Run without a To causes a panic.
func (ex *Exchange) To(dst Sender) {
	ex.sender = dst
}

// Process adds a Processor to the exchange.
//
// Processors are executed serially in the order in which they were defined on
// the exchange. When multiple Processors are added, you must ensure the
// resulting message transformation is compatible with the next Processor's
// input. Additionally, because the message body is stored as interface{}, it
// is up to each processor to cast the body as necessary to perform the desired
// transformations.
func (ex *Exchange) Process(proc Processor) *Exchange {
	ex.processors = append(ex.processors, proc)
	return ex
}

func (ex *Exchange) runReceive() error {
	for {
		m, err := ex.receiver.Receive()
		if err != nil {
			log.Printf("ERROR: failed to receive message: %v", err)
			continue
		}
		ex.inBuffer <- *m
	}
}

func (ex *Exchange) runSend() error {
	for {
		m := <-ex.outBuffer
		err := ex.sender.Send(&m)
		if err != nil {
			log.Printf("ERROR: could not send message: %v", err)
		}
	}
}

// Run starts the Receive, Process, Send loop for an exchange.
//
// If the exchange does not have both a From and To configured it will panic.
func (ex *Exchange) Run() {
	// TODO: have this take a context to exit.
	// Pass this context to the runReceive and runSend goruitines.
	// If cancel signal is emitted, exit receive loop immediately and close
	// the inBuffer.
	// Exit the send loop only after attempting to process the remainder
	// of the messages in [inBuffer].
	// handle SIGINT and cancel the context.
	if ex.receiver == nil || ex.sender == nil {
		panic("Exchange does not have either a 'from' or a 'to'")
		// log.Panic("Exchange does not have either a 'from' or a 'to'")
	}
	log.Printf("Exchange: '%s' starting up...", ex.Name)
	go ex.runReceive()
	go ex.runSend()

	// read from exchange.inBuffer, run through processors, then send to exchange.outBuffer
	for {
		m := <-ex.inBuffer
		for _, processor := range ex.processors {
			err := processor.Process(&m)
			if err != nil {
				log.Panicf("Failed processing message: %v", err)
			}
		}
		ex.outBuffer <- m
	}
}

// New creates a new [Exchange] with the given name.
//
// Exchanges must have a [From] and [To] in order to be a valid, runnable
// exchange.
func New(name string) *Exchange {
	return &Exchange{
		Name:      name,
		inBuffer:  make(chan message.Message, 25),
		outBuffer: make(chan message.Message, 25),
	}
}

func init() {
	directExchangeMap = make(map[string]chan message.Message)
}
