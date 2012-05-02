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

	// DeclScope returns the declaring scope of this identifier
	DeclScope() *Scope
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
	scope *Scope
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

func (id *idBase) DeclScope() *Scope {
	return id.scope
}

// Function represents a function identifier
//  e.g. std::fabs
type Function struct {
	idBase   `cxxtypes:"function"`
	qual TypeQualifier
	tspec    TypeSpecifier // or'ed value of virtual/inline/...
	variadic bool          // whether this function is variadic
	params   []Parameter   // the parameters to this function
	ret      Type          // return type of this function
}

// NewFunction returns a new function identifier
func NewFunction(name, scoped_name string, qual TypeQualifier, specifiers TypeSpecifier, variadic bool, params []Parameter, ret Type, scope *Scope) *Function {
	id := &Function{
		idBase: idBase{
			name:        name,
			scoped_name: scoped_name,
			kind:        IK_Fct,
			scope: scope,
		},
		qual: qual,
		tspec:    specifiers,
		variadic: variadic,
		params:   make([]Parameter, 0, len(params)),
		ret:      ret,
	}
	add_id(id)
	return id
}

func (t *Function) to_commonType() *commonType {
	return nil //fixme
}

func (t *Function) Name() string {
	return t.scoped_name
}

func (t *Function) Size() uintptr {
	return 0
}

func (t *Function) Kind() TypeKind {
	return TK_FunctionProto //fixme ?
}

func (t *Function) Qualifiers() TypeQualifier {
	return t.qual
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

// OverloadFunctionSet is a set of functions which are part of the same overload
type OverloadFunctionSet struct {
	idBase `cxxtypes:"overloadfctset"`
	fcts   []*Function
}

// NumFunction returns the number of overloads in that set
func (id *OverloadFunctionSet) NumFunction() int {
	return len(id.fcts)
}

// Function returns the i-th overloaded function in the set
// It panics if i is not in the range [0, NumFunction())
func (id *OverloadFunctionSet) Function(i int) *Function {
	if i < 0 || i >= id.NumFunction() {
		panic("cxxtypes: Function index out of range")
	}
	return id.fcts[i]
}

// ----------------------------------------------------------------------------
// id-related utils

// add_id adds a given identifier to the repository of identifiers.
// 
func add_id(id Id) Id {
	n := id.IdScopedName()
	switch id := id.(type) {
	case *Function:
		if _, exists := g_ids[n]; !exists {
			g_ids[n] = &OverloadFunctionSet{
				idBase: idBase{
					name:        id.IdName(),
					scoped_name: id.IdScopedName(),
					kind:        id.IdKind(),
				},
				fcts: make([]*Function, 1),
			}
		}
		o := g_ids[n].(*OverloadFunctionSet)
		o.fcts = append(o.fcts, id)
		println(":: added ["+n+"] to overload-fct-set...")
	default:
		if _, exists := g_ids[n]; exists {
			panic("cxxtypes: identifier [" + n + "] already in id-registry")
		}
		g_ids[n] = id
	}
	return g_ids[n]
}

// ----------------------------------------------------------------------------
// make sure the interfaces are implemented

var _ Id = (*Function)(nil)
var _ Id = (*OverloadFunctionSet)(nil)

func init() {
	g_ids = make(map[string]Id)
}

// EOF
