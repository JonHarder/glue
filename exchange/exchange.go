// Package exchange provides functionality to construct data pipelines.
//
// Data comes into the pipeline from a Receiver then gets transformed through any number of Processor transformations, and finally gets delivered out of the end of the pipeline via a Sender.
// Interanlly, data is stored and trasferred through the pipeline in a Message object.
package exchange

import (
	"context"
	"log"

	"github.com/JonHarder/glue/exchange/message"
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

// toFromExchange represents a [SendReceiver] which will send or receive
// messages directly from (or to) another exchange
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
//		ex1 := exchange.New("test")
//		ex1.From(receiver).
//		   .To(exchange.Direct("other-exchange"))
//
//		ex2 := exchange.New("test2")
//		ex2.From(exchange.Direct("other-exchange")).
//		    To(sender)
//
//	     ctx := context.Background()
//		go ex1.Run(ctx)
//		ex2.Run(ctx)
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

func (ex *Exchange) runReceive(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			close(ex.inBuffer)
			log.Printf("runReceive(): exiting. %v", ctx.Err())
			return nil
		default:
			m, err := ex.receiver.Receive()
			if err != nil {
				log.Printf("runReceive(): failed to receive message: %v", err)
				continue
			}
			ex.inBuffer <- *m
			if ctx.Value("RUN_ONCE") == true {
				close(ex.inBuffer)
				return nil
			}
		}
	}
}

func (ex *Exchange) runSend(ctx context.Context, doneCh chan bool) error {
	for {
		select {
		case <-ctx.Done():
			log.Printf("runSend(): context canceled, attempting exit gracefully")
		case m := <-ex.outBuffer:
			// check if the message is its zero value
			if m.Nil() {
				// if so, the channel has been closed and
				// there are no more messages to process.
				// exit the loop.
				log.Printf("runSend(): Nothing left to send. Exiting.")
				// Indicate to spawning goroutine that runSend has finished
				// processing all the messages in the outBuffer
				doneCh <- true
				return nil
			}
			err := ex.sender.Send(&m)
			if err != nil {
				log.Printf("ERROR: could not send message: %v", err)
			}
		}
	}
}

// Run starts the Receive, Process, Send loop for an exchange.
//
// If the exchange does not have both a From and To configured it will panic.
func (ex *Exchange) Run(ctx context.Context) {
	// handle SIGINT and cancel the context.
	if ex.receiver == nil || ex.sender == nil {
		panic("Exchange does not have either a 'from' or a 'to'")
		// log.Panic("Exchange does not have either a 'from' or a 'to'")
	}
	log.Printf("Exchange: '%s' starting up...", ex.Name)
	// doneCh is used for runSend to indicate back to this method that it has
	// finished processing all messages in the outBuffer.
	// This allows this function to block and wait for runSend to finish before
	// it itself finishes.
	doneCh := make(chan bool)
	go ex.runReceive(ctx)
	go ex.runSend(ctx, doneCh)

	for {
		select {
		case m := <-ex.inBuffer:
			if m.Nil() {
				close(ex.outBuffer)
				log.Printf("Run(): no more messages from inBuffer. exiting loop")
				log.Printf("Run(): There are still %d items in the queue", len(ex.outBuffer))
				// Wait until runSend finishes
				<-doneCh
				return
			}
			for _, processor := range ex.processors {
				err := processor.Process(&m)
				if err != nil {
					log.Panicf("Failed processing message: %v", err)
				}
			}
			ex.outBuffer <- m
		case <-ctx.Done():
			close(ex.outBuffer)
			break
		}
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
