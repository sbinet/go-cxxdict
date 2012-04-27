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
// EOF
