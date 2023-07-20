package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		"b": Symbol{Name: "b", Scope: GlobalScope, Index: 1},
		"c": Symbol{Name: "c", Scope: LocalScope, Index: 0},
		"d": Symbol{Name: "d", Scope: LocalScope, Index: 1},
		"e": Symbol{Name: "e", Scope: LocalScope, Index: 0},
		"f": Symbol{Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	a := global.Define("a")
	b := global.Define("b")
	c := firstLocal.Define("c")
	d := firstLocal.Define("d")
	e := secondLocal.Define("e")
	f := secondLocal.Define("f")

	if a != expected["a"] {
		t.Errorf("expected a=%+v, got %+v", expected["a"], a)
	}
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got %+v", expected["b"], b)
	}
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got %+v", expected["c"], c)
	}
	if d != expected["d"] {
		t.Errorf("expected d=%+v, got %+v", expected["d"], d)
	}
	if e != expected["e"] {
		t.Errorf("expected e=%+v, got %+v", expected["e"], e)
	}
	if f != expected["f"] {
		t.Errorf("expected f=%+v, got %+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	for _, symbol := range expected {
		result, ok := global.Resolve(symbol.Name)
		if !ok {
			t.Errorf("name %s not resolvable", symbol.Name)
		}
		if result != symbol {
			t.Errorf("expected %s go resolve to %+v, got %+v", symbol.Name, symbol, result)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
		Symbol{Name: "c", Scope: LocalScope, Index: 0},
		Symbol{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got %+v", sym.Name, sym, result)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, test := range tests {
		for _, sym := range test.expectedSymbols {

			result, ok := test.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got %+v", sym.Name, sym, result)
			}
		}
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvabel", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got %+v", sym.Name, sym, result)
			}
		}
	}
}
