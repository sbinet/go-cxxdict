package cxxtypes

// Id is a C/C++ identifier
type Id interface {
	// IdName returns the name of this identifier
	// e.g. my_fct
	//      MyClass
	//      vector<int>
	//      fabs
	//      g_some_global_variable
	IdName() string

	// IdScopedName returns the scoped name of this identifier
	// e.g. some_namespace::my_fct
	//      SomeOtherNamespace::MyClass
	//      std::vector<int>
	//      std::fabs
	//      g_some_global_variable
	IdScopedName() string

	// IdKind returns the kind of this identifier
	//  IK_Var | IK_Typ | IK_Fct | IK_Nsp
	IdKind() IdKind
}

// IdKind represents the specific kind of identifier an Id represents.
// The zero IdKind is not a valid kind.
type IdKind uint

const (
	IK_Invalid IdKind = iota

	IK_Var // a variable
	IK_Typ // a type (class, fct type, struct, enum, ...)
	IK_Fct // a function, operator function, method, method operator, ...
	IK_Nsp // a namespace
)

// the db of all identifiers
var g_ids map[string]Id

// IdByName retrieves an identifier by its fully qualified name.
// Returns nil if no such identifier exists.
func IdByName(n string) Id {
	id, ok := g_ids[n]
	if ok {
		return id
	}
	return nil
}

// IdNames returns the list of identifier names currently defined.
func IdNames() []string {
	names := make([]string, 0, len(g_ids))
	for k, _ := range g_ids {
		names = append(names, k)
	}
	return names
}

// NumId returns the number of currently defined identifiers
func NumId() int {
	return len(g_ids)
}

// idBase implements the Id interface
type idBase struct {
	name        string
	scoped_name string
	kind        IdKind
}

func (id *idBase) IdName() string {
	return id.name
}

func (id *idBase) IdScopedName() string {
	return id.scoped_name
}

func (id *idBase) IdKind() IdKind {
	return id.kind
}

// Function represents a function identifier
//  e.g. std::fabs
type Function struct {
	idBase `cxxtypes:"function"`
	tspec      TypeSpecifier // or'ed value of virtual/inline/...
	variadic   bool          // whether this function is variadic
	params     []Parameter   // the parameters to this function
	ret        Type          // return type of this function
}

// NewFunction returns a new function identifier
func NewFunction(name, scoped_name string, qual TypeQualifier, specifiers TypeSpecifier, variadic bool, params []Parameter, ret Type) *Function {
	id := &Function{
		idBase: idBase{
			name: name,
			scoped_name: scoped_name,
			kind: IK_Fct,
		},
		tspec:    specifiers,
		variadic: variadic,
		params:   make([]Parameter, 0, len(params)),
		ret:      ret,
	}
	return id
}

// Specifier returns the type specifier for this function
func (t *Function) Specifier() TypeSpecifier {
	return t.tspec
}

func (t *Function) IsVirtual() bool {
	return (t.tspec & TS_Virtual) != 0
}

func (t *Function) IsStatic() bool {
	return (t.tspec & TS_Static) != 0
}

func (t *Function) IsConstructor() bool {
	return (t.tspec & TS_Constructor) != 0
}

func (t *Function) IsDestructor() bool {
	return (t.tspec & TS_Destructor) != 0
}

func (t *Function) IsCopyConstructor() bool {
	return (t.tspec & TS_CopyCtor) != 0
}

func (t *Function) IsOperator() bool {
	return (t.tspec & TS_Operator) != 0
}

func (t *Function) IsMethod() bool {
	return (t.tspec & TS_Method) != 0
}

func (t *Function) IsInline() bool {
	return (t.tspec & TS_Inline) != 0
}

func (t *Function) IsConverter() bool {
	return (t.tspec & TS_Converter) != 0
}

// IsVariadic returns whether this function is variadic
func (t *Function) IsVariadic() bool {
	return t.variadic
}

// NumParam returns a function's input parameter count.
func (t *Function) NumParam() int {
	return len(t.params)
}

// Param returns the i'th parameter of this function.
// It panics if i is not in the range [0, NumParam())
func (t *Function) Param(i int) *Parameter {
	if i < 0 || i >= t.NumParam() {
		panic("cxxtypes: Param index out of range")
	}
	return &t.params[i]
}

// NumDefaultParam returns the number of parameters of a function's input which have a default value.
func (t *Function) NumDefaultParam() int {
	n := 0
	for i, _ := range t.params {
		if t.params[i].defval {
			n += 1
		}
	}
	return n
}

// ReturnType returns the return type of this function.
// FIXME: return nil for 'void' fct ?
// FIXME: return nil for ctor/dtor ?
func (t *Function) ReturnType() Type {
	return t.ret
}


// EOF
