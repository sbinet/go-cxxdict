package cxxtypes

import (
	"fmt"
	"strings"
)

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

func (id IdKind) String() string {
	switch id {
	case IK_Invalid:
		return "<invalid>"
	case IK_Var:
		return "IK_Var"
	case IK_Typ:
		return "IK_Typ"
	case IK_Fct:
		return "IK_Fct"
	case IK_Nsp:
		return "IK_Nsp"
	}
	panic("unreachable")
}

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
	name  string
	kind  IdKind
	scope *Scope
}

func (id *idBase) IdName() string {
	n := strings.Split(id.name, "::")
	return n[len(n)-1]
}

func (id *idBase) IdScopedName() string {
	return id.name
}

func (id *idBase) IdKind() IdKind {
	return id.kind
}

func (id *idBase) DeclScope() *Scope {
	return id.scope
}

// Namespace represents a namespace identifier
type Namespace struct {
	idBase `cxxtypes:"namespace"`
}

// NewNamespace creates a new namespace identifier
func NewNamespace(name string, scope *Scope) *Namespace {
	id := &Namespace{
		idBase: idBase{
			name:  name,
			kind:  IK_Nsp,
			scope: scope,
		},
	}
	add_id(id)
	return id
}

// Function represents a function identifier
//  e.g. std::fabs
type Function struct {
	idBase   `cxxtypes:"function"`
	qual     TypeQualifier
	tspec    TypeSpecifier // or'ed value of virtual/inline/...
	variadic bool          // whether this function is variadic
	params   []Parameter   // the parameters to this function
	ret      Type          // return type of this function
}

// NewFunction returns a new function identifier
func NewFunction(name string, qual TypeQualifier, specifiers TypeSpecifier, variadic bool, params []Parameter, ret Type, scope *Scope) *Function {
	id := &Function{
		idBase: idBase{
			name:  name,
			kind:  IK_Fct,
			scope: scope,
		},
		qual:     qual,
		tspec:    specifiers,
		variadic: variadic,
		params:   make([]Parameter, 0, len(params)),
		ret:      ret,
	}
	id.params = append(id.params, params...)
	//copy(id.params, params)
	add_id(id)
	return id
}

func (t *Function) to_commonType() *commonType {
	return nil //fixme
}

func (t *Function) Name() string {
	return t.name
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

func (t *Function) IsConst() bool {
	return (t.qual & TQ_Const) != 0
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

// Signature returns the (C++11) signature of this function
func (t *Function) Signature() string {
	s := []string{}
	if t.IsInline() {
		s = append(s, "inline ")
	}
	if t.IsStatic() {
		s = append(s, "static ")
	}
	s = append(s, t.IdScopedName(), "(")
	if len(t.params) > 0 {
		for i, _ := range t.params {
			s = append(s,
				strings.TrimSpace(t.Param(i).Type().Name()),
				" ",
				strings.TrimSpace(t.Param(i).Name()))
			if i < len(t.params)-1 {
				s = append(s, ", ")
			}
		}
	} else {
		// fixme: we should rather test if C XOR C++...
		if t.IsMethod() {
			//nothing
		} else {
			s = append(s, "void")
		}
	}
	if t.IsVariadic() {
		s = append(s, "...")
	}
	s = append(s, ") ")
	if t.IsConst() {
		s = append(s, "const ")
	}
	//fixme: add fct qualifiers: const|static|inline
	if !t.IsConstructor() && !t.IsDestructor() {
		s = append(s, "-> ", t.ReturnType().Name())
	}
	return strings.TrimSpace(strings.Join(s, ""))
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
					name:  id.IdScopedName(),
					kind:  id.IdKind(),
					scope: id.DeclScope(),
				},
				fcts: make([]*Function, 0, 1),
			}
		}
		o := g_ids[n].(*OverloadFunctionSet)
		o.fcts = append(o.fcts, id)
		//println(":: added [" + id.Signature() + "] to overload-fct-set...")
	default:
		if _, exists := g_ids[n]; exists {
			err := fmt.Errorf("cxxtypes: identifier [%s] already in id-registry (type=%T)", n, g_ids[n])
			panic(err)
		}
		g_ids[n] = id
	}
	//println(":: added [" + n + "]")
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
