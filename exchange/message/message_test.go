package message

import "testing"

func TestNilMessag(t *testing.T) {
	var m Message
	if m.Nil() == false {
		t.Errorf("Uninitialized message should return true on Nil()")
	}
}
