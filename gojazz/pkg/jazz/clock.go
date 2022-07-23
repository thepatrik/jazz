package jazz

import (
	"time"
)

type Clock struct {
}

func (clock *Clock) Arity() int {
	return 0
}

func (clock *Clock) Call(interpreter *Interpreter, args ...interface{}) interface{} {
	return time.Now().UnixMilli()
}

func (clock *Clock) String() string {
	return "<native fn>"
}
