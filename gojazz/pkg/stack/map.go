package stack

import (
	"fmt"
)

type MapStack struct {
	stack []Map
}

type Map map[string]bool

var ErrEmptyStack = fmt.Errorf("empty stack")

func NewMapStack() *MapStack {
	return &MapStack{
		stack: []Map{},
	}
}

func (ms *MapStack) Get(ix int) Map {
	return ms.stack[ix]
}

func (ms *MapStack) Push(v Map) {
	ms.stack = append(ms.stack, v)
}

func (ms *MapStack) Peek() Map {
	if ms.Empty() {
		return nil
	}
	return ms.stack[ms.Len()-1]
}

func (ms *MapStack) Pop() (Map, error) {
	l := ms.Len()
	if l == 0 {
		return nil, ErrEmptyStack
	}
	m := ms.stack[l-1]
	ms.stack = ms.stack[:l-1]
	return m, nil
}

func (ms *MapStack) Empty() bool {
	return len(ms.stack) == 0
}

func (ms *MapStack) Len() int {
	return len(ms.stack)
}
