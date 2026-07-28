package jazz

// ---- Sentinel signals for break/continue -----------------------------------

type BreakError struct{}
type ContinueError struct{}

func (e *BreakError) Error() string    { return "break" }
func (e *ContinueError) Error() string { return "continue" }

// ---- len() native ----------------------------------------------------------

type LenNative struct{}

func (l *LenNative) Arity() int { return 1 }

func (l *LenNative) Call(_ *Interpreter, args ...interface{}) interface{} {
	switch v := args[0].(type) {
	case *JazzArray:
		return float64(len(v.Elements))
	case string:
		return float64(len(v))
	}
	panic(&InterpreterError{Message: "len() argument must be an array or string"})
}

func (l *LenNative) String() string { return "<native fn>" }

// ---- push() native ---------------------------------------------------------

type PushNative struct{}

func (p *PushNative) Arity() int { return 2 }

func (p *PushNative) Call(_ *Interpreter, args ...interface{}) interface{} {
	arr, ok := args[0].(*JazzArray)
	if !ok {
		panic(&InterpreterError{Message: "push() first argument must be an array"})
	}
	arr.Elements = append(arr.Elements, args[1])
	return float64(len(arr.Elements))
}

func (p *PushNative) String() string { return "<native fn>" }
