package eval

import (
	"fmt"
	"monkey/object"
	"os"
	"strings"
)

var builtins = map[string]*object.Builtin{
	"len":   {Fn: monkeyLen},
	"head":  {Fn: monkeyHead},
	"tail":  {Fn: monkeyTail},
	"back":  {Fn: monkeyBack},
	"push":  {Fn: monkeyPush},
	"print": {Fn: monkeyPrint},
	"exit":  {Fn: exit},
	"quit":  {Fn: exit},
}

func monkeyLen(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}

	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

func monkeyHead(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `head` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*object.Array)
	if len(arr.Elements) == 0 {
		return NULL
	}
	return arr.Elements[0]
}

func monkeyBack(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `back` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*object.Array)
	if len(arr.Elements) == 0 {
		return NULL
	}
	return arr.Elements[len(arr.Elements)-1]
}

func monkeyTail(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 1)
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `tail` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	if length == 0 {
		return NULL
	}
	newElements := make([]object.Object, length-1)
	copy(newElements, arr.Elements[1:])
	return &object.Array{Elements: newElements}
}

func monkeyPush(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got %d, expected %d", len(args), 2)
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
	}
	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	newElements := make([]object.Object, length+1)
	copy(newElements, arr.Elements)
	newElements[length] = args[1]
	return &object.Array{Elements: newElements}
}

func monkeyPrint(args ...object.Object) object.Object {
	var out []string
	for _, arg := range args {
		out = append(out, arg.Inspect())
	}
	fmt.Println(strings.Join(out, " "))
	return NULL
}

func exit(args ...object.Object) object.Object {
	os.Exit(0)
	return &object.Null{}
}
