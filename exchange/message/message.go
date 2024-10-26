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
