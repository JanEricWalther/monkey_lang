package eval

import (
	"monkey/object"
	"testing"
)

func TestQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`quote(5)`, `5`},
		{`quote("foobar")`, `foobar`},
		{`quote(foobar + barfoo)`, `(foobar + barfoo)`},
	}

	for _, test := range tests {
		evaluated := testEval(test.input)
		quote, ok := evaluated.(*object.Quote)
		if !ok {
			t.Errorf("object is not Quote. got %T (%+v)", evaluated, evaluated)
			continue
		}
		if quote.Node == nil {
			t.Errorf("Node is nil")
			continue
		}
		if quote.Node.String() != test.expected {
			t.Errorf("not equal. got %q, expected %q", quote.Node.String(), test.expected)
		}
	}
}
