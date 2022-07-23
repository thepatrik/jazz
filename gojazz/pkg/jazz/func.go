package jazz

import (
	"fmt"
)

type Func struct {
	Declaration  *FuncStmt
	EnclosingEnv *Env
}

func NewFunc(name *Token, params []*Token, body []Stmt, enclosingEnv *Env) *Func {
	fs := &FuncStmt{Name: name, Params: params, Body: body}

	return &Func{Declaration: fs, EnclosingEnv: enclosingEnv}
}

func (f *Func) Arity() int {
	return len(f.Declaration.Params)
}

func (f *Func) Call(i *Interpreter, args ...interface{}) interface{} {
	enclosingEnv := i.env
	env := NewEnv(WithEnclosingEnv(f.EnclosingEnv))
	for i, param := range f.Declaration.Params {
		env.Define(param.Lexeme, args[i])
	}

	_, err := i.executeBlock(f.Declaration.Body, env)
	if err != nil {
		if rerr, ok := err.(*ReturnError); ok {
			i.env = enclosingEnv
			return rerr.Val
		}
		panic(err)
	}

	return nil
}

func (f *Func) String() string {
	return fmt.Sprintf("<fn %s>", f.Declaration.Name.Lexeme)
}
