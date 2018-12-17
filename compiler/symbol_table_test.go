package compiler

import (
	"testing"
)

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
	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("expected a= %+v, got=%+v", expected["a"], a)
	}
	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}

	firstLocal := NewEnclosedSymbolTable(global)
	c := firstLocal.Define("c")
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}

	d := firstLocal.Define("d")
	if d != expected["d"] {
		t.Errorf("expected d=%+v, got=%+v", expected["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")
	if e != expected["e"] {
		t.Errorf("expected e=%+v, got=%+v", expected["e"], e)
	}

	f := secondLocal.Define("f")
	if f != expected["f"] {
		t.Errorf("expected f=%+v, got=%+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := map[string]Symbol{
		"a": Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		"b": Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, sym := range expected {
		ret, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if ret != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, ret)
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
		ret, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if ret != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, ret)
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
				Symbol{Name: "c", Scope: FreeScope, Index: 0},
				Symbol{Name: "d", Scope: FreeScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			ret, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if ret != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, ret)
			}
		}
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		Symbol{Name: "a", Scope: BuiltinScope, Index: 0},
		Symbol{Name: "c", Scope: BuiltinScope, Index: 1},
		Symbol{Name: "e", Scope: BuiltinScope, Index: 2},
		Symbol{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, sym := range expected {
			ret, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if ret != sym {
				t.Errorf("expected %s to result to %+v, got=%+v", sym.Name, sym, ret)
			}

		}
	}
}

// TestResolveFree test code as:
// let a = 1;
// let b = 1;
// let firstLocal = fn() {
// 	let c = 3;
// 	let d = 4;
// 	a + b + c + c;

// 	let secondLocal = fn() {
// 		let e = 5;
// 		let f = 6;
// 		a + b + c + d + e + f;
// 	};
// };
func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	secondLocal.Define("e")
	secondLocal.Define("f")

	thirdLocal := NewEnclosedSymbolTable(secondLocal)
	thirdLocal.Define("g")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: FreeScope, Index: 0},
				Symbol{Name: "d", Scope: FreeScope, Index: 1},
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				// specify the origin symbol
				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			thirdLocal,
			// 站在 thirdLocal 的角度看
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: FreeScope, Index: 0},
				Symbol{Name: "d", Scope: FreeScope, Index: 1},
				Symbol{Name: "e", Scope: FreeScope, Index: 2},
				Symbol{Name: "f", Scope: FreeScope, Index: 3},
				Symbol{Name: "g", Scope: LocalScope, Index: 0},
			},
			[]Symbol{
				// 离开 thirdLocal
				Symbol{Name: "c", Scope: FreeScope, Index: 0},
				Symbol{Name: "d", Scope: FreeScope, Index: 1},
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			ret, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if ret != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, ret)
			}
		}
		if len(tt.table.FreeSymbols) != len(tt.expectedFreeSymbols) {
			t.Errorf("wrong number of free symbols. got=%d, want=%d", len(tt.table.FreeSymbols), len(tt.expectedFreeSymbols))
			continue
		}
		for i, sym := range tt.expectedFreeSymbols {
			ret := tt.table.FreeSymbols[i]
			if ret != sym {
				t.Errorf("wrong free symbol. got=%+v, want=%+v", ret, sym)
			}
		}
	}
}

func TestResolveUnresolvableFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "c", Scope: FreeScope, Index: 0},
		Symbol{Name: "e", Scope: LocalScope, Index: 0},
		Symbol{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		ret, ok := secondLocal.Resolve(sym.Name)

		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if ret != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, ret)
		}
	}
	expectedUnresolvable := []string{
		"b",
		"d",
	}
	for _, name := range expectedUnresolvable {
		_, ok := secondLocal.Resolve(name)
		if ok {
			t.Errorf("name %s resolved, but was expected not to", name)
		}
	}
}
