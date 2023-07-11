package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Scope: GlobalScope, Index: s.numDefinitions}
	s.store[name] = symbol
	s.numDefinitions += 1
	return symbol
}

func (s *SymbolTable) Resolve(name string) (sym Symbol, ok bool) {
	sym, ok = s.store[name]
	return sym, ok
}
