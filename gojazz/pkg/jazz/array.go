package jazz

import (
	"fmt"
	"strings"
)

type JazzArray struct {
	Elements []interface{}
}

func NewJazzArray(elements []interface{}) *JazzArray {
	return &JazzArray{Elements: elements}
}

func (a *JazzArray) String() string {
	parts := make([]string, len(a.Elements))
	for i, el := range a.Elements {
		parts[i] = fmt.Sprintf("%v", el)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
