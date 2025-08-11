package semantics

import "golite.dev/mvp/internal/object"

type Symbol struct {
	Name  string
	Type  object.ObjectType
	Scope string // "global", "local", etc.
}

type SymbolTable struct {
	store map[string]Symbol
	outer *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.outer = outer
	return s
}

func (st *SymbolTable) Define(name string, ty object.ObjectType) Symbol {
	symbol := Symbol{Name: name, Type: ty}
	if st.outer == nil {
		symbol.Scope = "global"
	} else {
		symbol.Scope = "local"
	}
	st.store[name] = symbol
	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := st.store[name]
	if !ok && st.outer != nil {
		obj, ok = st.outer.Resolve(name)
	}
	return obj, ok
}
