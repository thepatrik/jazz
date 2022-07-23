package jazz

import (
	"fmt"
)

type Store map[string]interface{}

type Env struct {
	cfg   *EnvCfg
	store Store
}

type EnvOpt func(*EnvCfg)

type EnvCfg struct {
	enclosing *Env
}

func WithEnclosingEnv(env *Env) EnvOpt {
	return func(cfg *EnvCfg) {
		cfg.enclosing = env
	}
}

func NewEnv(options ...EnvOpt) *Env {
	cfg := &EnvCfg{
		enclosing: nil,
	}
	for _, option := range options {
		option(cfg)
	}
	return &Env{
		cfg:   cfg,
		store: map[string]interface{}{},
	}
}

func (e *Env) AssignAt(depth int, token *Token, val interface{}) error {
	env := e.ancestor(depth)
	if env == nil {
		return fmt.Errorf("could not find env at %v", depth)
	}

	env.store[token.Lexeme] = val

	return nil
}

func (e *Env) Assign(token *Token, val interface{}) error {
	if _, ok := e.store[token.Lexeme]; !ok {
		if e.cfg.enclosing != nil {
			return e.cfg.enclosing.Assign(token, val)
		}
		return fmt.Errorf("undefined variable '%s'", token)
	}

	e.store[token.Lexeme] = val
	return nil
}

func (e *Env) Define(name string, val interface{}) {
	e.store[name] = val
}

func (e *Env) Get(token *Token) (interface{}, error) {
	if val, ok := e.store[token.Lexeme]; ok {
		return val, nil
	}

	if e.cfg.enclosing != nil {
		return e.cfg.enclosing.Get(token)
	}

	return nil, fmt.Errorf("undefined variable '%s'", token.Lexeme)
}

func (e *Env) GetAt(depth int, name string) (interface{}, error) {
	val, ok := e.ancestor(depth).store[name]
	if !ok {
		return nil, fmt.Errorf("could not find variable '%s' at %v", name, depth)
	}

	return val, nil
}

func (e *Env) ancestor(depth int) *Env {
	env := e
	for i := 0; i < depth; i++ {
		env = env.cfg.enclosing
	}

	return env
}
