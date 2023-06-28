package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnv() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func NewInnerEnv(outer *Environment) (e *Environment) {
	e = NewEnv()
	e.outer = outer
	return e
}

func (e *Environment) Get(name string) (obj Object, ok bool) {
	obj, ok = e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return
}

func (e *Environment) Set(name string, obj Object) Object {
	e.store[name] = obj
	return obj
}
