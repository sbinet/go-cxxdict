// This file implements scopes and the objects they contain.
//
// loosely modeled after go/ast/scope.go 

package cxxtypes

import (
	"bytes"
	"fmt"
)

// A Scope maintains the set of named language entities declared
// in the scope and a link to the immediately surrounding (outer)
// scope.
//
type Scope struct {
	Outer   *Scope
	Objects map[string]*Object
}

// NewScope creates a new scope nested in the outer scope.
func NewScope(outer *Scope) *Scope {
	const n = 4 // initial scope capacity
	return &Scope{outer, make(map[string]*Object, n)}
}

// Lookup returns the object with the given name if it is
// found in scope s, otherwise it returns nil. Outer scopes
// are ignored.
//
func (s *Scope) Lookup(name string) *Object {
	return s.Objects[name]
}

// Insert attempts to insert a named object obj into the scope s.
// If the scope already contains an object alt with the same name,
// Insert leaves the scope unchanged and returns alt. Otherwise
// it inserts obj and returns nil."
//
func (s *Scope) Insert(obj *Object) (alt *Object) {
	if alt = s.Objects[obj.Name]; alt == nil {
		s.Objects[obj.Name] = obj
	}
	return
}

// Debugging support
func (s *Scope) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "scope %p {", s)
	if s != nil && len(s.Objects) > 0 {
		fmt.Fprintln(&buf)
		for _, obj := range s.Objects {
			fmt.Fprintf(&buf, "\t%s %s\n", obj.Kind, obj.Name)
		}
	}
	fmt.Fprintf(&buf, "}\n")
	return buf.String()
}

// ----------------------------------------------------------------------------
// Objects

// Object describes a named language entity such as a variable, type, function
// (including methods), or namespace.
type Object struct {
	Kind ObjKind
	Name string      // declared name
	Decl interface{} // declaring scope, or nil
	Data interface{} // corresponding declaration for that object; or nil
	Type interface{} // place holder for type information; may be nil
}

// NewObj creates a new object of a given kind and name.
func NewObj(kind ObjKind, name string) *Object {
	return &Object{Kind: kind, Name: name}
}

// ObjKind describes what an object represents.
type ObjKind int

// The list of possible Object kinds.
const (
	OK_Bad ObjKind = iota // for error handling
	OK_Typ                // type
	OK_Var                // variable
	OK_Fun                // function or method
	OK_Nsp                // namespace
)

var objKindStrings = [...]string{
	OK_Bad: "bad",
	OK_Typ: "type",
	OK_Var: "var",
	OK_Fun: "func",
	OK_Nsp: "namespace",
}

func (kind ObjKind) String() string { return objKindStrings[kind] }


// 
var (
	Universe *Scope
)

func define(kind ObjKind, name string, scope *Scope) *Object {
	obj := NewObj(kind, name)
	if scope.Insert(obj) != nil {
		panic("cxxtypes: internal error - double declaration")
	}
	obj.Decl = scope
	return obj
}

// func defType(name string) Type {
// 	obj := define(OK_Typ, name)
// 	typ := 
// }

func init() {
	Universe = NewScope(nil)

	// add builtins
	define(OK_Typ, "void", Universe)
	define(OK_Typ, "bool", Universe)
	define(OK_Typ, "char", Universe)
	define(OK_Typ, "signed char", Universe)
	define(OK_Typ, "unsigned char", Universe)
	define(OK_Typ, "short", Universe)
	define(OK_Typ, "unsigned short", Universe)
	define(OK_Typ, "int", Universe)
	define(OK_Typ, "unsigned int", Universe)

	define(OK_Typ, "long", Universe)
	define(OK_Typ, "unsigned long", Universe)
	define(OK_Typ, "long long", Universe)
	define(OK_Typ, "unsigned long long", Universe)

	define(OK_Typ, "float", Universe)
	define(OK_Typ, "double", Universe)
	
	define(OK_Typ, "float complex", Universe)
	define(OK_Typ, "double complex", Universe)

	// function builtins
	define(OK_Fun, "sizeof", Universe)

	// stl
	obj_std := define(OK_Nsp, "std", Universe)

	obj_std.Data = NewScope(Universe)

}