package internal

type NumberType int

const (
	Unknown NumberType = iota
	Int
	Uint
	Float
	Complex
)

type OutFuncs struct {
	print  func(format string, params ...interface{})
	colour func(str string, colour string) string
}

func (nt NumberType) String() string {
	switch nt {
	case Int:
		return "int"
	case Uint:
		return "uint"
	case Float:
		return "float"
	case Complex:
		return "complex"
	}
	return "unknown"
}

type Outer interface {
	OutShort(with OutFuncs)
	Out(with OutFuncs)
}

type NamedValue struct {
	Name  string
	Value Outer
}

func (nv *NamedValue) OutShort(with OutFuncs) {

}

func (nv *NamedValue) Out(with OutFuncs) {

}

type String struct {
	Value string
}

type ByteArray struct {
	Value []byte
}

type Pointer struct {
	Value Outer
}

type Slice struct {
	ItemValueType string
	Items         []Outer
}

type Chan struct {
	Direction string
	ValueType string
	Value     string
}

type Map struct {
	KeyType   string
	ValueType string
	Items     []Outer
}

type Number struct {
	Type  NumberType
	Bits  int
	Value string
}

type Interface struct {
	Value Outer
}

type Bool struct {
	Value String
}

type Func struct {
	Value string
}

type Struct struct {
	Type  string
	Items []Outer
}

type Invalid struct {
}
