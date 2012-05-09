package cxxtypes

import (
	"fmt"
	"sort"
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
	DeclScope() Id
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

// BaseId implements the Id interface
type BaseId struct {
	Name  string
	Kind  IdKind
	Scope string
}

func (id *BaseId) IdName() string {
	n := strings.Split(id.Name, "::")
	return n[len(n)-1]
}

func (id *BaseId) IdScopedName() string {
	return id.Name
}

func (id *BaseId) IdKind() IdKind {
	return id.Kind
}

func (id *BaseId) DeclScope() Id {
	return get_scope_from_name(id.Scope)
}

// Namespace represents a namespace identifier
type Namespace struct {
	BaseId  `cxxtypes:"namespace"`
	Members []string
}

// NewNamespace creates a new namespace identifier
func NewNamespace(name string, scope string) *Namespace {
	id := &Namespace{
		BaseId: BaseId{
			Name:  name,
			Kind:  IK_Nsp,
			Scope: scope,
		},
		Members: make([]string, 0),
	}
	add_id(id)
	add_id_to_scope(name, scope)
	return id
}

// NumMember returns the number of members of that namespace
func (id *Namespace) NumMember() int {
	return len(id.Members)
}

// Member returns the namespace's i'th member
// It panics if i is not in the range [0, NumMember())
func (t *Namespace) Member(i int) Id {
	if i < 0 || i >= len(t.Members) {
		panic("cxxtypes: Member index out of range")
	}
	return IdByName(t.Members[i])
}

// Function represents a function identifier
//  e.g. std::fabs
type Function struct {
	BaseId   `cxxtypes:"function"`
	Qual     TypeQualifier
	Spec     TypeSpecifier // or'ed value of virtual/inline/...
	Variadic bool          // whether this function is variadic
	Params   []Parameter   // the parameters to this function
	Ret      string        // return type of this function
}

// NewFunction returns a new function identifier
func NewFunction(name string, qual TypeQualifier, specifiers TypeSpecifier, variadic bool, params []Parameter, ret string, scope string) *Function {
	id := &Function{
		BaseId: BaseId{
			Name:  name,
			Kind:  IK_Fct,
			Scope: scope,
		},
		Qual:     qual,
		Spec:     specifiers,
		Variadic: variadic,
		Params:   make([]Parameter, 0, len(params)),
		Ret:      ret,
	}
	id.Params = append(id.Params, params...)
	add_id(id)
	add_id_to_scope(name, scope)
	return id
}

func (t *Function) TypeName() string {
	return t.BaseId.Name
}

func (t *Function) TypeSize() uintptr {
	return 0
}

func (t *Function) TypeKind() TypeKind {
	return TK_FunctionProto //fixme ?
}

func (t *Function) Qualifiers() TypeQualifier {
	return t.Qual
}

// Specifier returns the type specifier for this function
func (t *Function) Specifier() TypeSpecifier {
	return t.Spec
}

func (t *Function) IsConst() bool {
	return (t.Qual & TQ_Const) != 0
}

func (t *Function) IsVirtual() bool {
	return (t.Spec & TS_Virtual) != 0
}

func (t *Function) IsStatic() bool {
	return (t.Spec & TS_Static) != 0
}

func (t *Function) IsConstructor() bool {
	return (t.Spec & TS_Constructor) != 0
}

func (t *Function) IsDestructor() bool {
	return (t.Spec & TS_Destructor) != 0
}

func (t *Function) IsCopyConstructor() bool {
	return (t.Spec & TS_CopyCtor) != 0
}

func (t *Function) IsOperator() bool {
	return (t.Spec & TS_Operator) != 0
}

func (t *Function) IsMethod() bool {
	return (t.Spec & TS_Method) != 0
}

func (t *Function) IsInline() bool {
	return (t.Spec & TS_Inline) != 0
}

func (t *Function) IsConverter() bool {
	return (t.Spec & TS_Converter) != 0
}

// IsVariadic returns whether this function is variadic
func (t *Function) IsVariadic() bool {
	return t.Variadic
}

// NumParam returns a function's input parameter count.
func (t *Function) NumParam() int {
	return len(t.Params)
}

// Param returns the i'th parameter of this function.
// It panics if i is not in the range [0, NumParam())
func (t *Function) Param(i int) *Parameter {
	if i < 0 || i >= t.NumParam() {
		panic("cxxtypes: Param index out of range")
	}
	return &t.Params[i]
}

// NumDefaultParam returns the number of parameters of a function's input which have a default value.
func (t *Function) NumDefaultParam() int {
	n := 0
	for i, _ := range t.Params {
		if t.Params[i].DefVal {
			n += 1
		}
	}
	return n
}

// ReturnType returns the return type of this function.
// FIXME: return nil for 'void' fct ?
// FIXME: return nil for ctor/dtor ?
func (t *Function) ReturnType() Type {
	return IdByName(t.Ret).(Type)
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
	if len(t.Params) > 0 {
		for i, _ := range t.Params {
			s = append(s,
				strings.TrimSpace(t.Param(i).Type),
				" ",
				strings.TrimSpace(t.Param(i).Name))
			if i < len(t.Params)-1 {
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
		s = append(s, "-> ", t.ReturnType().TypeName())
	}
	return strings.TrimSpace(strings.Join(s, ""))
}

// OverloadFunctionSet is a set of functions which are part of the same overload
type OverloadFunctionSet struct {
	BaseId `cxxtypes:"overloadfctset"`
	Fcts   []*Function
}

// NumFunction returns the number of overloads in that set
func (id *OverloadFunctionSet) NumFunction() int {
	return len(id.Fcts)
}

// Function returns the i-th overloaded function in the set
// It panics if i is not in the range [0, NumFunction())
func (id *OverloadFunctionSet) Function(i int) *Function {
	if i < 0 || i >= id.NumFunction() {
		panic("cxxtypes: Function index out of range")
	}
	return id.Fcts[i]
}

func (id *OverloadFunctionSet) TypeName() string {
	return id.Fcts[0].Name
}

func (id *OverloadFunctionSet) TypeSize() uintptr {
	return 0
}

func (id *OverloadFunctionSet) TypeKind() TypeKind {
	return TK_FunctionProto //fixme ?
}

func (id *OverloadFunctionSet) Qualifiers() TypeQualifier {
	//FIXME: is that always true ?
	return id.Fcts[0].Qual
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
				BaseId: BaseId{
					Name:  id.IdScopedName(),
					Kind:  id.IdKind(),
					Scope: id.Scope,
				},
				Fcts: make([]*Function, 0, 1),
			}
		}
		o := g_ids[n].(*OverloadFunctionSet)
		o.Fcts = append(o.Fcts, id)
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

func find_idx(slice []string, x string) int {
	for i, str := range slice {
		if str == x {
			return i
		}
	}
	return -1
}
func add_id_to_scope(id string, scope string) {
	parent := IdByName(scope)
	switch t := parent.(type) {
	case *Namespace:
		sort.Strings(t.Members)
		//FIXME: use sort.SearchStrings when fixed ?
		idx := find_idx(t.Members, id)
		if idx == -1 {
			t.Members = append(t.Members, id)
		}

	default:
		//fmt.Printf("** no handling for [%t] (id=%s, scope=%s)\n", t, id, scope)
	}
}

// ----------------------------------------------------------------------------
// make sure the interfaces are implemented

var _ Id = (*Function)(nil)
var _ Id = (*OverloadFunctionSet)(nil)

func init() {
	g_ids = make(map[string]Id)
}

// EOF
