package message

import (
	"github.com/google/uuid"
)

type Message struct {
	id     uuid.UUID
	Header map[string]interface{}
	Body   interface{}
}

func (m *Message) Id() uuid.UUID {
	return m.id
}

func New(body interface{}) *Message {
	return &Message{
		id:     uuid.New(),
		Header: make(map[string]interface{}),
		Body:   body,
	}
}

// Nil checks to see if this [Message] is uninitialized, i.e. the zero value
//
// The assumption is that if its id field is Nil, then the message as a whole
// is initialized, since the only way to construct a message is through the
// [New] function.
func (m *Message) Nil() bool {
	return m.Id() == uuid.Nil
}
