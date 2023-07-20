package eval

import (
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len":   object.GetBuildinByName("len"),
	"head":  object.GetBuildinByName("head"),
	"tail":  object.GetBuildinByName("tail"),
	"back":  object.GetBuildinByName("back"),
	"push":  object.GetBuildinByName("push"),
	"print": object.GetBuildinByName("print"),
	"exit":  object.GetBuildinByName("exit"),
	"quit":  object.GetBuildinByName("exit"),
}
