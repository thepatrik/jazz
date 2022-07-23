package jazz

type Callable interface {
	Arity() int
	Call(interpreter *Interpreter, args ...interface{}) interface{}
	String() string
}
