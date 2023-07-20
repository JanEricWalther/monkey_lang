package object

import (
	"fmt"
	"os"
	"strings"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{"len", &Builtin{Fn: monkeyLen}},
	{"head", &Builtin{Fn: monkeyHead}},
	{"first", &Builtin{Fn: monkeyHead}},
	{"back", &Builtin{Fn: monkeyBack}},
	{"last", &Builtin{Fn: monkeyBack}},
	{"print", &Builtin{Fn: monkeyPrint}},
	{"puts", &Builtin{Fn: monkeyPrint}},
	{"tail", &Builtin{Fn: monkeyTail}},
	{"rest", &Builtin{Fn: monkeyTail}},
	{"push", &Builtin{Fn: monkeyPush}},
	{"append", &Builtin{Fn: monkeyPush}},
	{"exit", &Builtin{Fn: exit}},
	{"quit", &Builtin{Fn: exit}},
}

func GetBuildinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

func monkeyLen(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}

	switch arg := args[0].(type) {
	case *String:
		return &Integer{Value: int64(len(arg.Value))}
	case *Array:
		return &Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

func monkeyHead(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `head` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*Array)
	if len(arr.Elements) == 0 {
		return nil
	}
	return arr.Elements[0]
}

func monkeyBack(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `back` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*Array)
	if len(arr.Elements) == 0 {
		return nil
	}
	return arr.Elements[len(arr.Elements)-1]
}

func monkeyTail(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `tail` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*Array)
	length := len(arr.Elements)
	if length == 0 {
		return nil
	}
	newElements := make([]Object, length-1)
	copy(newElements, arr.Elements[1:])
	return &Array{Elements: newElements}
}

func monkeyPush(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 2)
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*Array)
	length := len(arr.Elements)
	newElements := make([]Object, length+1)
	copy(newElements, arr.Elements)
	newElements[length] = args[1]
	return &Array{Elements: newElements}
}

func monkeyPrint(args ...Object) Object {
	var out []string
	for _, arg := range args {
		out = append(out, arg.Inspect())
	}
	fmt.Println(strings.Join(out, " "))
	return nil
}

func exit(args ...Object) Object {
	fmt.Println("Goodbye!")
	os.Exit(0)
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
