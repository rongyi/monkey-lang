package compiler

type SymbolScope string

const (
	LocalScope   SymbolScope = "LOCAL"
	GlobalScope  SymbolScope = "GLOBAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int
	FreeSymbols    []Symbol // for closure
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}

	return &SymbolTable{
		store:       s,
		FreeSymbols: free,
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	ret := Symbol{
		Name:  name,
		Index: s.numDefinitions,
	}
	if s.Outer == nil {
		ret.Scope = GlobalScope
	} else {
		ret.Scope = LocalScope
	}

	s.store[name] = ret
	s.numDefinitions++
	return ret
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: index,
		Scope: BuiltinScope,
	}
	s.store[name] = symbol

	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}
		// not global, not builtin, and not local, cause we are in enclosing scope

		free := s.defineFree(obj)
		return free, ok
	}
	return obj, ok
}

// defineFree 用来转换一个symbol，一倒手，所有symbol都是free类型，并且把之前的symbol缓存在
// 对应的FreeSymbols里。这里有个规律： 如果是在相邻外层找到这个symbol，那么缓存在 freeSymbols
// 里的symbol是local类型，这个很好理解。 如果是在隔了一层或多层的情况下找到的这个symbol呢？
// 这种情况下除了跟这个symbol紧挨着的内层 FreeSymbols缓存的是local类型，其他内层 FreeSymbols
// 缓存的都是 free类型了。因为根据递归，隔了一层之后被这个函数一倒手，给过去的origin都是free
// 类型了。
// e.g. 嵌套关系如下：
// a, b ==> GlobalScope
// c, d ==> firstLocal
// e, f ==> secondLocal
// g ==> thirdLocal

// thirdLocal里的 FreeSymbols 放的c类型是free， 而 secondLocal 里的 FreeSymbols放的c 类型则是local
// 因为 thirdLocal 里 经过递归返回给它的时候被 secondLocal 的 defineFree倒手了，所以在 thirdLocal
// 角度看的时候 original已经是free类型了。这里需要推敲一下
func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	ret := Symbol{
		Name:  original.Name,
		Index: len(s.FreeSymbols) - 1,
		Scope: FreeScope,
	}

	// 如果一symbol递归找着了，满足free variable条件，然后就在这一层缓存了，当成free
	// global不缓存，用到的其他scope的local会被缓存在这里。
	s.store[original.Name] = ret
	return ret
}
