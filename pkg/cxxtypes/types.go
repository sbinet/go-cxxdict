// Package cxxtypes describes C++ types (classes, structs, functions, ...) which
// have been somehow loaded into memory (from gccxml, clang, ...)
package cxxtypes

import (
	"fmt"
	"strings"
	"unsafe"
)

// TypeKind represents the specific kind of type that a Type represents.
// The zero TypeKind is not a valid kind.
type TypeKind uint

const (
	TK_Invalid TypeKind = iota
	TK_Unexposed

	// builtin types

	TK_Void
	TK_Bool
	TK_Char_U
	TK_UChar
	TK_Char16
	TK_Char32
	TK_UShort
	TK_UInt
	TK_ULong
	TK_ULongLong
	TK_UInt128
	TK_Char_S
	TK_SChar
	TK_WChar
	TK_Short
	TK_Int
	TK_Long
	TK_LongLong
	TK_Int128
	TK_Float
	TK_Double
	TK_LongDouble
	TK_NullPtr
	TK_Overload
	TK_Dependent
	TK_ObjCId
	TK_ObjCClass
	TK_ObjCSel

	TK_Complex
	TK_Ptr
	TK_BlockPtr
	TK_LValueRef
	TK_RValueRef
	TK_Record
	TK_Enum
	TK_Typedef
	TK_ObjCInterface
	TK_ObjCObjectPointer
	TK_FunctionNoProto
	TK_FunctionProto
	TK_ConstantArray

	TK_FirstBuiltin = TK_Void
	TK_LastBuiltin  = TK_ObjCSel
)

func (tk TypeKind) String() string {
	switch tk {
	case TK_Invalid:
		return "Invalid"
	case TK_Void:
		return "Void"
	case TK_Bool:
		return "Bool"
	case TK_Char_U:
		return "Char_U"
	case TK_UChar:
		return "UChar"
	case TK_Char16:
		return "Char16"
	case TK_Char32:
		return "Char32"
	case TK_UShort:
		return "UShort"
	case TK_UInt:
		return "UInt"
	case TK_ULong:
		return "ULong"
	case TK_ULongLong:
		return "ULongLong"
	case TK_UInt128:
		return "UInt128"
	case TK_Char_S:
		return "Char_S"
	case TK_SChar:
		return "SChar"
	case TK_WChar:
		return "WChar"
	case TK_Short:
		return "Short"
	case TK_Int:
		return "Int"
	case TK_Long:
		return "Long"
	case TK_LongLong:
		return "LongLong"
	case TK_Int128:
		return "Int128"
	case TK_Float:
		return "Float"
	case TK_Double:
		return "Double"
	case TK_LongDouble:
		return "LongDouble"
	case TK_NullPtr:
		return "NullPtr"
	case TK_Overload:
		return "Overload"
	case TK_Dependent:
		return "Dependent"
	case TK_ObjCId:
		return "ObjCId"
	case TK_ObjCClass:
		return "ObjCClass"
	case TK_ObjCSel:
		return "ObjCSel"
	// TK_FirstBuiltin = TK_Void
	// TK_LastBuiltin  = TK_ObjCSel

	case TK_Complex:
		return "Complex"
	case TK_Ptr:
		return "Ptr"
	case TK_BlockPtr:
		return "BlockPtr"
	case TK_LValueRef:
		return "LValueRef"
	case TK_RValueRef:
		return "RValueRef"
	case TK_Record:
		return "Record"
	case TK_Enum:
		return "Enum"
	case TK_Typedef:
		return "Typedef"
	case TK_ObjCInterface:
		return "ObjCInterface"
	case TK_ObjCObjectPointer:
		return "ObjCObjectPointer"
	case TK_FunctionNoProto:
		return "FunctionNoProto"
	case TK_FunctionProto:
		return "FunctionProto"
	case TK_ConstantArray:
		return "ConstantArray"
	default:
		panic(fmt.Sprintf("unknown TypeKind: %d", tk))
	}
	panic("unreachable")
}

// TypeQualifier represents the set of qualifiers (const,volatile,restrict) which decorate a type
// The zero TypeQualifier denotes no qualifier being applied.
type TypeQualifier uintptr

const (
	TQ_None  TypeQualifier = 0
	TQ_Const TypeQualifier = 1 << iota
	TQ_Restrict
	TQ_Volatile
)

func (tq TypeQualifier) String() string {
	if tq == TQ_None {
		return "<none>"
	}
	s := []string{}
	if (tq & TQ_Const) != 0 {
		s = append(s, "const")
	}
	if (tq & TQ_Restrict) != 0 {
		s = append(s, "restrict")
	}
	if (tq & TQ_Volatile) != 0 {
		s = append(s, "volatile")
	}
	return strings.Join(s, "|")
}

// Type is the representation of a C/C++ type
//
// Not all methods apply to all kinds of types.  Restrictions,
// if any, are noted in the documentation for each method.
// Use the Kind method to find out the kind of type before
// calling kind-specific methods.  Calling a method
// inappropriate to the kind of type causes a run-time panic.
type Type interface {

	// Name returns the name of the type
	// mod qualifiers can be or'ed
	// FIXME: implement qualifiers ?
	TypeName( /*mod FINAL|QUALIFIED|SCOPED*/) string

	// Size returns the number of bytes needed to store
	// a value of the given type
	TypeSize() uintptr

	// Kind returns the specific kind of this type.
	TypeKind() TypeKind

	// Qualifiers returns the or'ed values of qualifiers applied to this type.
	Qualifiers() TypeQualifier

	// Specifiers returns the or'ed values of specifiers applied to this type.
	Specifiers() TypeSpecifier

	// DeclScope returns the declaring scope of this type
	DeclScope() Id

	// CanonicalType returns the underlying type with all
	// the "sugar" removed.
	//CanonicalType() Type
}

// size of a pointer for this platform
var g_ptrsz uintptr

// IsConstQualified returns whether this type is const-qualified.
// This doesn't look through typedefs that may have added 'const' at 
// different level.
func IsConstQualified(t Type) bool {
	return (t.Qualifiers() & TQ_Const) != 0
}

// IsRestrictQualified returns whether this type is restrict-qualified.
// This doesn't look through typedefs that may have added 'restrict' at 
// different level.
func IsRestrictQualified(t Type) bool {
	return (t.Qualifiers() & TQ_Restrict) != 0
}

// IsVolatileQualified returns whether this type is volatile-qualified.
// This doesn't look through typedefs that may have added 'volatile' at 
// different level.
func IsVolatileQualified(t Type) bool {
	return (t.Qualifiers() & TQ_Volatile) != 0
}

// IsAbstractType returns whether this type is abstract.
func IsAbstractType(t Type) bool {
	return (t.Specifiers() & TS_Abstract) != 0
}

// BaseType is the common implementation of most types.
// It is embedded in other, public struct types, but always
// with a unique tag like `cxxtypes:"array"` or `cxxtypes:"ptr"`
// so that code cannot convert from, say, *ArrayType to *PtrType.
type BaseType struct {
	Size  uintptr       // size in bytes
	Kind  TypeKind      // the specific kind of this type
	Qual  TypeQualifier // the qualifiers applied to this type
	Spec  TypeSpecifier // the specifiers applied to this type
	Scope string        // declaring scope of this type
	Name  string        // the fully qualified name of the type
	//canon Type          // the canonical type of this type
}

func (t *BaseType) IdName() string {
	return t.Name
}

//fixme
func (t *BaseType) IdScopedName() string {
	return t.Name
}

func (t *BaseType) IdKind() IdKind {
	return IK_Typ
}

func (t *BaseType) TypeName() string {
	return t.Name
}

func (t *BaseType) TypeSize() uintptr {
	return t.Size
}

func (t *BaseType) TypeKind() TypeKind {
	return t.Kind
}

func (t *BaseType) Qualifiers() TypeQualifier {
	return t.Qual
}

func (t *BaseType) Specifiers() TypeSpecifier {
	return t.Spec
}

func (t *BaseType) DeclScope() Id {
	return get_scope_from_name(t.Scope)
}

func (t *BaseType) String() string {
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		t.TypeName(), t.TypeSize(), t.TypeKind(), t.Qualifiers())
}

func get_scope_from_name(name string) Id {
	return IdByName(name)
}

type placeHolderType struct {
	Name   string
	Synced bool
}

func NewPlaceHolder(name string) Type {
	return &placeHolderType{Name: name, Synced: false}
}

func (t *placeHolderType) sync() bool {
	if !t.Synced {
		_, ok := IdByName(t.Name).(Type)
		if !ok {
			panic("could not sync [" + t.Name + "]")
		}
		t.Synced = true
	}
	return t.Synced
}

func (t *placeHolderType) IdName() string {
	return t.Name
}

//fixme
func (t *placeHolderType) IdScopedName() string {
	return t.Name
}

func (t *placeHolderType) IdKind() IdKind {
	return IK_Typ
}

func (t *placeHolderType) get_type() Type {
	if t.sync() {
		return IdByName(t.Name).(Type)
	}
	return nil
}

func (t *placeHolderType) TypeName() string {
	if t.sync() {
		return t.get_type().TypeName()
	}
	return t.Name
}

func (t *placeHolderType) TypeSize() uintptr {
	if t.sync() {
		return t.get_type().TypeSize()
	}
	return uintptr(0)
}

func (t *placeHolderType) TypeKind() TypeKind {
	if t.sync() {
		return t.get_type().TypeKind()
	}
	return TK_Invalid
}

func (t *placeHolderType) Qualifiers() TypeQualifier {
	if t.sync() {
		return t.get_type().Qualifiers()
	}
	return TQ_None
}

func (t *placeHolderType) Specifiers() TypeSpecifier {
	if t.sync() {
		return t.get_type().Specifiers()
	}
	return TS_None
}

func (t *placeHolderType) DeclScope() Id {
	if t.sync() {
		return t.get_type().DeclScope()
	}
	return nil
}

func (t *placeHolderType) String() string {
	t.sync()
	tt := t.get_type()
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		tt.TypeName(), tt.TypeSize(), tt.TypeKind(), tt.Qualifiers())
}

// func (t *BaseType) CanonicalType() Type {
// 	return t.canon
// }

// NewFundamentalType creates a C/C++ builtin type.
func NewFundamentalType(name string, size uintptr, kind TypeKind, scope string) Type {
	tt := &FundamentalType{
		BaseType: BaseType{
			Size:  size,
			Kind:  kind,
			Qual:  TQ_None,
			Spec:  TS_None,
			Scope: scope,
			Name:  name,
		},
	}
	add_type(tt)
	return tt
}

// FundamentalType represents a builtin type
type FundamentalType struct {
	BaseType //`cxxtypes:"builtin"`
}

// NewQualType creates a new const-restrict-volatile qualified type.
// The new qualifiers are added to the old ones of the base type.
func NewQualType(n string, tn string, scope string, qual TypeQualifier) (q Type) {
	q = &CvrQualType{
		Name:  n,
		Qual:  qual,
		Type:  tn,
		Scope: scope,
	}
	add_type(q)
	return
}

type CvrQualType struct {
	Name  string
	Qual  TypeQualifier
	Type  string // the decorated type (actually, its name)
	Scope string
}

func (t *CvrQualType) IdName() string {
	return t.Name
}

//fixme
func (t *CvrQualType) IdScopedName() string {
	return t.Name
}

func (t *CvrQualType) IdKind() IdKind {
	return IK_Typ
}

func (t *CvrQualType) TypeName() string {
	return t.Name
}

func (t *CvrQualType) get_type() Type {
	return IdByName(t.Type).(Type)
}

func (t *CvrQualType) TypeSize() uintptr {
	return t.get_type().TypeSize()
}

func (t *CvrQualType) TypeKind() TypeKind {
	return t.get_type().TypeKind()
}

func (t *CvrQualType) Qualifiers() TypeQualifier {
	return t.Qual | t.get_type().Qualifiers()
}

func (t *CvrQualType) Specifiers() TypeSpecifier {
	return t.get_type().Specifiers()
}

func (t *CvrQualType) DeclScope() Id {
	return get_scope_from_name(t.Scope)
}

func (t *CvrQualType) String() string {
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		t.TypeName(), t.TypeSize(), t.TypeKind(), t.Qualifiers())
}

// NewPtrType creates a new pointer type from an already existing type t.
func NewPtrType(name string, tn string, scope string) *PtrType {
	p := &PtrType{
		Name:  name,
		Scope: scope,
		Type:  tn,
	}
	add_type(p)
	return p
}

// PtrType represents a typed ptr
type PtrType struct {
	Name  string // the fully qualified name of the type
	Scope string // declaring scope of this type
	Type  string // the pointee type, possibly cvr-qualified
}

func (t *PtrType) get_type() Type {
	return IdByName(t.Type).(Type)
}

func (t *PtrType) IdName() string {
	return t.TypeName()
}

//fixme
func (t *PtrType) IdScopedName() string {
	return t.TypeName()
}

func (t *PtrType) IdKind() IdKind {
	return IK_Typ
}

func (t *PtrType) TypeName() string {
	return t.Name
}

func (t *PtrType) TypeSize() uintptr {
	return g_ptrsz
}

func (t *PtrType) TypeKind() TypeKind {
	return TK_Ptr
}

func (t *PtrType) Qualifiers() TypeQualifier {
	return t.get_type().Qualifiers()
}

func (t *PtrType) Specifiers() TypeSpecifier {
	return t.get_type().Specifiers()
}

func (t *PtrType) DeclScope() Id {
	return get_scope_from_name(t.Scope)
}

func (t *PtrType) String() string {
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		t.TypeName(), t.TypeSize(), t.TypeKind(), t.Qualifiers())
}

// UnderlyingType returns the type of the pointee
func (t *PtrType) UnderlyingType() Type {
	return t.get_type()
}

// NewRefType creates a new reference type from an already existing type t.
func NewRefType(name string, tn string, scope string) *RefType {
	r := &RefType{
		Name:  name,
		Scope: scope,
		Type:  tn,
	}
	add_type(r)
	return r
}

// RefType represents a typed reference
type RefType struct {
	Name  string // the fully qualified name of the type
	Scope string // declaring scope of this type
	Type  string // the referenced type, possibly cvr-qualified
}

func (t *RefType) get_type() Type {
	return IdByName(t.Type).(Type)
}

func (t *RefType) IdName() string {
	return t.TypeName()
}

//fixme
func (t *RefType) IdScopedName() string {
	return t.TypeName()
}

func (t *RefType) IdKind() IdKind {
	return IK_Typ
}

func (t *RefType) TypeName() string {
	return t.Name
}

func (t *RefType) TypeSize() uintptr {
	return g_ptrsz
}

func (t *RefType) TypeKind() TypeKind {
	return TK_LValueRef
}

func (t *RefType) Qualifiers() TypeQualifier {
	return t.get_type().Qualifiers()
}

func (t *RefType) Specifiers() TypeSpecifier {
	return t.get_type().Specifiers()
}

func (t *RefType) DeclScope() Id {
	return get_scope_from_name(t.Scope)
}

func (t *RefType) String() string {
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		t.TypeName(), t.TypeSize(), t.TypeKind(), t.Qualifiers())
}

// UnderlyingType returns the referenced type
func (t *RefType) UnderlyingType() Type {
	return t.get_type()
}

// NewTypedefType creates a new typedef from an already existing type t.
func NewTypedefType(n string, tn string, tsz uintptr, scope string) *TypedefType {
	tt := &TypedefType{
		BaseType: BaseType{
			Size:  tsz,
			Kind:  TK_Typedef,
			Qual:  TQ_None,
			Spec:  TS_None,
			Scope: scope,
			Name:  n,
		},
		Type: tn,
	}
	add_type(tt)
	return tt
}

// TypedefType represents a typedef
type TypedefType struct {
	BaseType `cxxtypes:"typedef"`
	Type     string // the typedef'd type, possible cvr-qualified
}

func (t *TypedefType) get_type() Type {
	return IdByName(t.Type).(Type)
}

// UnderlyingType returns the type of the typedef'd type
func (t *TypedefType) UnderlyingType() Type {
	return t.get_type()
}

// NewArrayType creates a new array of type T[n].
func NewArrayType(sz uintptr, tn string, tsz uintptr, scope string) *ArrayType {
	tt := &ArrayType{
		BaseType: BaseType{
			Size:  tsz * sz,
			Kind:  TK_ConstantArray,
			Qual:  TQ_None,
			Spec:  TS_None,
			Scope: scope,
			Name:  tn + fmt.Sprintf("[%d]", sz),
		},
		ArrElem: tn,
		ArrLen:  sz,
	}
	add_type(tt)
	return tt
}

// ArrayType represents a fixed array type
type ArrayType struct {
	BaseType `cxxtypes:"array"`
	ArrElem  string  // array element type
	ArrLen   uintptr // array length
}

// Elem returns the type of the array's elements
func (t *ArrayType) Elem() Type {
	return IdByName(t.ArrElem).(Type)
}

// Len returns the size of the array
func (t *ArrayType) Len() uintptr {
	return t.ArrLen
}

// NewStructType creates a new struct type.
func NewStructType(n string, sz uintptr, scope string) *StructType {
	t := &StructType{
		BaseType: BaseType{
			Size:  sz,
			Kind:  TK_Record,
			Qual:  TQ_None,
			Spec:  TS_None,
			Scope: scope,
			Name:  n,
		},
		Bases:   make([]Base, 0),
		Members: make([]Member, 0),
	}
	// t.members = append(t.members, members...)
	// set_scope(t.members, t, scope)
	add_type(t)
	return t
}

// StructType represents a C-struct type
type StructType struct {
	BaseType `cxxtypes:"struct"`
	Bases    []Base
	Members  []Member
}

// NumMember returns a struct type's member count
func (t *StructType) NumMember() int {
	return len(t.Members)
}

// Member returns a struct type's i'th member
// It panics if i is not in the range [0, NumMember())
func (t *StructType) Member(i int) *Member {
	if i < 0 || i >= len(t.Members) {
		panic("cxxtypes: Member index out of range")
	}
	return &t.Members[i]
}

// SetMembers sets the members of this struct type
func (t *StructType) SetMembers(mbrs []Member) error {
	t.Members = make([]Member, len(mbrs))
	copy(t.Members, mbrs)
	err := set_scope(t.Members, t, t.Name)
	if err != nil {
		return err
	}
	for i, _ := range t.Members {
		mbr := &t.Members[i]
		if mbr.IsDataMember() {
			add_id(mbr)
		}
	}
	return nil
}

// NumBase returns a struct type's base struct count
func (t *StructType) NumBase() int {
	return len(t.Bases)
}

// Base returns a struct type's i'th base
// it panics if i is not in the range [0, NumBase()0)
func (t *StructType) Base(i int) *Base {
	if i < 0 || i >= len(t.Bases) {
		panic("cxxtypes: Base index out of range")
	}
	return &t.Bases[i]
}

// HasBase returns true if a struct has b as one of its base structs
func (t *StructType) HasBase(b Type) bool {
	for i, _ := range t.Bases {
		if t.Bases[i].Type() == b {
			return true
		}
	}
	return false
}

// SetBases sets the bases of this struct type
func (t *StructType) SetBases(bases []Base) error {
	t.Bases = make([]Base, len(bases))
	copy(t.Bases, bases)
	return nil
}

// NewEnumType creates a new enum type.
func NewEnumType(n string, members []Member, scope string) *EnumType {
	var sz uintptr = 0
	if len(members) > 0 {
		// take the size of the first member type,
		// they should all be the same
		sz = members[0].get_type().TypeSize()
	}
	t := &EnumType{
		BaseType: BaseType{
			Size:  sz,
			Kind:  TK_Enum,
			Qual:  TQ_None,
			Spec:  TS_None,
			Scope: scope,
			Name:  n,
		},
		Members: make([]Member, 0, len(members)),
	}
	t.Members = append(t.Members, members...)
	// enum members "leak" into the scope declaring the enum-type
	parent_scope := ""
	if scope != "" && scope != "::" {
		parent_scope = IdByName(scope).DeclScope().IdScopedName()
	}
	set_scope(t.Members, t, parent_scope)

	// add enum-members to the identifier-registry
	for i, _ := range t.Members {
		add_id(&t.Members[i])
	}

	add_type(t)
	return t
}

// EnumType represents an enum type
type EnumType struct {
	BaseType `cxxtypes:"enum"`
	Members  []Member
}

// NumMember returns an enum type's member count
func (t *EnumType) NumMember() int {
	return len(t.Members)
}

// Member returns an enum type's i'th member
// It panics if i is not in the range [0, NumMember())
func (t *EnumType) Member(i int) *Member {
	if i < 0 || i >= len(t.Members) {
		panic("cxxtypes: Member index out of range")
	}
	return &t.Members[i]
}

// NewUnionType creates a new union type.
func NewUnionType(n string, members []Member, scope string) *UnionType {
	t := &UnionType{
		BaseType: BaseType{
			Size:  0,
			Kind:  TK_Record,
			Qual:  TQ_None,
			Spec:  TS_None,
			Scope: scope,
			Name:  n,
		},
		Members: make([]Member, 0, len(members)),
	}
	for i, _ := range members {
		t.Members = append(t.Members, members[i])
		// sz := members[i].get_type().TypeSize()
		// if t.BaseType.Size < sz {
		// 	t.BaseType.Size = sz
		// }
	}
	set_scope(t.Members, t, scope)
	add_type(t)
	return t
}

// UnionType represents a union type
type UnionType struct {
	BaseType `cxxtypes:"union"`
	Members  []Member
}

func (t *UnionType) TypeSize() uintptr {
	if t.BaseType.Size == 0 {
		for i, _ := range t.Members {
			sz := t.Members[i].get_type().TypeSize()
			if t.BaseType.Size < sz {
				t.BaseType.Size = sz
			}
		}
	}
	return t.BaseType.Size
}

// NewClassType creates a new class type.
func NewClassType(n string, sz uintptr, scope string) *ClassType {
	t := &ClassType{
		BaseType: BaseType{
			Size:  sz,
			Kind:  TK_Record,
			Qual:  TQ_None,
			Spec:  TS_None,
			Scope: scope,
			Name:  n,
		},
		Bases:   make([]Base, 0),
		Members: make([]Member, 0),
	}
	add_type(t)
	return t
}

// ClassType represents a C++ class type
type ClassType struct {
	BaseType `cxxtypes:"class"`
	Bases    []Base
	Members  []Member
}

// NumMember returns a class type's member count
func (t *ClassType) NumMember() int {
	return len(t.Members)
}

// Member returns a class type's i'th member
// It panics if i is not in the range [0, NumMember())
func (t *ClassType) Member(i int) *Member {
	if i < 0 || i >= len(t.Members) {
		panic("cxxtypes: Member index out of range")
	}
	return &t.Members[i]
}

// SetMembers sets the members of this class type
func (t *ClassType) SetMembers(mbrs []Member) error {
	t.Members = make([]Member, len(mbrs))
	copy(t.Members, mbrs)
	err := set_scope(t.Members, t, t.Name)
	if err != nil {
		return err
	}
	for i, _ := range t.Members {
		mbr := &t.Members[i]
		if mbr.IsDataMember() {
			add_id(mbr)
		}
	}
	return nil
}

// NumBase returns a class type's base class count
func (t *ClassType) NumBase() int {
	return len(t.Bases)
}

// Base returns a class type's i'th base
// it panics if i is not in the range [0, NumBase()0)
func (t *ClassType) Base(i int) *Base {
	if i < 0 || i >= len(t.Bases) {
		panic("cxxtypes: Base index out of range")
	}
	return &t.Bases[i]
}

// HasBase returns true if a class has b as one of its base classes
func (t *ClassType) HasBase(b Type) bool {
	for i, _ := range t.Bases {
		if t.Bases[i].Type() == b {
			return true
		}
	}
	return false
}

// SetBases sets the bases of this class type
func (t *ClassType) SetBases(bases []Base) error {
	t.Bases = make([]Base, len(bases))
	copy(t.Bases, bases)
	return nil
}

// NewBase creates a new base class/struct
func NewBase(offset uintptr, tn string, access AccessSpecifier, virtual bool) Base {
	return Base{
		OffsetBase: offset,
		TypeBase:   tn,
		Access:     access,
		Virtual:    virtual,
	}
}

// Base represents a base class of a C++ class type
type Base struct {
	OffsetBase uintptr         // the offset to the base class
	TypeBase   string          // Type of this base class
	Access     AccessSpecifier // specifier for the derivation
	Virtual    bool            // whether the derivation is virtual
}

// Offset returns the offset of this base class
func (b *Base) Offset() uintptr {
	return b.OffsetBase
}

// Type returns the Type of this base class
func (b *Base) Type() Type {
	return IdByName(b.TypeBase).(Type)
}

// IsVirtual returns whether the derivation is virtual
func (b *Base) IsVirtual() bool {
	return b.Virtual
}

// IsPublic returns whether the derivation is public
func (b *Base) IsPublic() bool {
	return (b.Access & AS_Public) != 0
}

func (b *Base) IsProtected() bool {
	return (b.Access & AS_Protected) != 0
}

func (b *Base) IsPrivate() bool {
	return (b.Access & AS_Private) != 0
}

func (b *Base) String() string {
	hdr := ""
	if b.IsVirtual() {
		hdr = "virtual "
	}
	return fmt.Sprintf("Base{%s%s %s offset=%d}",
		hdr, b.Access.String(), b.TypeBase, b.OffsetBase)
}

// NewMember creates a new member for a struct, class, enum or union
func NewMember(name string, tn string, idkind IdKind, kind TypeKind, access AccessSpecifier, offset uintptr, scope string) Member {
	mbr := Member{
		BaseId: BaseId{
			Name:  name,
			Kind:  idkind,
			Scope: scope,
		},
		Type:   tn,
		Kind:   kind,
		Access: access,
		Offset: offset,
	}
	return mbr
}

// Member represents a member in a struct, class, enum or union
type Member struct {
	BaseId `cxxtypes:"member"`
	Type   string          // the type of this member
	Kind   TypeKind        // the kind of this member
	Access AccessSpecifier // the access specifier for this member
	Offset uintptr         // the offset in the embedding scope
}

func (t *Member) get_type() Type {
	return IdByName(t.Type).(Type)
}

func (m *Member) IsDataMember() bool {
	return (m.IdKind() == IK_Var)
}

func (m *Member) IsEnumMember() bool {
	return m.IsDataMember() && (m.Kind == TK_Enum)
}

func (m *Member) IsFunctionMember() bool {
	return (m.Kind == TK_FunctionProto) ||
		(m.Kind == TK_FunctionNoProto)
}

func id_kind_from_tk(tk TypeKind) IdKind {
	ik := IK_Typ
	switch tk {
	case TK_FunctionProto, TK_FunctionNoProto:
		ik = IK_Fct
	}
	return ik
}

// // Specifier returns the or'ed valued of the type specifier.
// func (m *Member) Specifier() TypeSpecifier {
// 	return m.tspec
// }

// func (m *Member) IsConst() bool {
// 	return (m.tspec & TQ_Const) != 0
// }

// func (m *Member) IsMutable() bool {
// 	return (m.tspec & TS_Mutable) != 0
// }

// func (m *Member) IsReference() bool {
// 	return (m.tspec & TS_Reference) != 0
// }

func (m *Member) IsPublic() bool {
	return (m.Access & AS_Public) != 0
}

func (m *Member) IsProtected() bool {
	return (m.Access & AS_Protected) != 0
}

func (m *Member) IsPrivate() bool {
	return (m.Access & AS_Private) != 0
}

func (m *Member) String() string {
	hdr := "DMbr"
	if m.IsEnumMember() {
		hdr = "EMbr"
	}
	if m.IsFunctionMember() {
		hdr = "FMbr"
	}
	return fmt.Sprintf("%s{%s type='%s' kind=%s access=%s offset=%d scope='%s'}",
		hdr,
		m.Name, m.Type, m.Kind.String(), m.Access.String(), m.Offset, m.Scope)
}

// NewFunctionType creates a new function type.
func NewFunctionType(n string, qual TypeQualifier, specifiers TypeSpecifier, variadic bool, params []Parameter, ret string, scope string) *FunctionType {
	t := &FunctionType{
		BaseType: BaseType{
			Size:  0,
			Kind:  TK_FunctionProto,
			Qual:  qual,
			Spec:  specifiers,
			Scope: scope,
			Name:  n,
		},
		Variadic: variadic,
		Params:   make([]Parameter, 0, len(params)),
		Ret:      ret,
	}

	// only add that type to the db if it isn't a method (of a class/struct)
	if !t.IsMethod() {
		add_type(t)
	}
	return t
}

// FunctionType represents a C/C++ function, a member function, c-tor, ...
type FunctionType struct {
	BaseType `cxxtypes:"function"`
	Variadic bool        // whether this function is variadic
	Params   []Parameter // the parameters to this function
	Ret      string      // return type of this function
}

// Specifier returns the type specifier for this function
func (t *FunctionType) Specifier() TypeSpecifier {
	return t.BaseType.Spec
}

func (t *FunctionType) IsVirtual() bool {
	return (t.BaseType.Spec & TS_Virtual) != 0
}

func (t *FunctionType) IsStatic() bool {
	return (t.BaseType.Spec & TS_Static) != 0
}

func (t *FunctionType) IsConstructor() bool {
	return (t.BaseType.Spec & TS_Constructor) != 0
}

func (t *FunctionType) IsDestructor() bool {
	return (t.BaseType.Spec & TS_Destructor) != 0
}

func (t *FunctionType) IsCopyConstructor() bool {
	return (t.BaseType.Spec & TS_CopyCtor) != 0
}

func (t *FunctionType) IsOperator() bool {
	return (t.BaseType.Spec & TS_Operator) != 0
}

func (t *FunctionType) IsMethod() bool {
	return (t.BaseType.Spec & TS_Method) != 0
}

func (t *FunctionType) IsInline() bool {
	return (t.BaseType.Spec & TS_Inline) != 0
}

func (t *FunctionType) IsConverter() bool {
	return (t.BaseType.Spec & TS_Converter) != 0
}

// IsVariadic returns whether this function is variadic
func (t *FunctionType) IsVariadic() bool {
	return t.Variadic
}

// NumParam returns a function type's input parameter count.
func (t *FunctionType) NumParam() int {
	return len(t.Params)
}

// Param returns the i'th parameter of this function's type.
// It panics if i is not in the range [0, NumParam())
func (t *FunctionType) Param(i int) *Parameter {
	if i < 0 || i >= t.NumParam() {
		panic("cxxtypes: Param index out of range")
	}
	return &t.Params[i]
}

// NumDefaultParam returns a function type's input with default value parameter count.
func (t *FunctionType) NumDefaultParam() int {
	n := 0
	for i, _ := range t.Params {
		if t.Params[i].DefVal {
			n += 1
		}
	}
	return n
}

// ReturnType returns the return type of this function's type.
// FIXME: return nil for 'void' fct ?
func (t *FunctionType) ReturnType() Type {
	return IdByName(t.Ret).(Type)
}

// NewParameter creates a new parameter.
func NewParameter(n string, tn string, defval bool) *Parameter {
	return &Parameter{
		Name:   n,
		Type:   tn,
		DefVal: defval,
	}
}

// Parameter represents a parameter of a function's signature
type Parameter struct {
	Name   string // name of the parameter
	Type   string // type of this parameter
	DefVal bool   // whether this parameter has a default value
}

// HasDefaultValue returns whether this parameter has a default value
func (p *Parameter) HasDefaultValue() bool {
	return p.DefVal
}

// NewVar creates a new global variable
func NewVar(n string, specifiers TypeSpecifier, tn string, scope string) *Var {
	return &Var{
		Name:  n,
		Spec:  specifiers,
		Type:  tn,
		Scope: scope,
	}
}

// Var represents a variable
type Var struct {
	Name  string        // name of the variable
	Spec  TypeSpecifier // or'ed value of const/extern/static/...
	Type  string        // type of this variable
	Scope string        // the scope holding that variable
}

// TypeSpecifier represents the specifiers which can "decorate" C/C++ types.
// e.g. static,inline,virtual ...
type TypeSpecifier uintptr

const (
	TS_None     TypeSpecifier = 0
	TS_Register TypeSpecifier = 1 << iota
	TS_Virtual
	TS_Static
	TS_Inline
	TS_Extern
	TS_Constructor
	TS_Destructor
	TS_CopyCtor
	TS_Operator
	TS_Converter
	TS_Method
	TS_Explicit
	TS_Auto
	TS_Mutable
	TS_Abstract
	TS_Transient
	TS_Artificial
)

// AccessSpecifier represents the C++ access control level to a base class or a class' member
type AccessSpecifier uintptr

const (
	AS_None AccessSpecifier = 0

	AS_Private AccessSpecifier = 1 << iota
	AS_Protected
	AS_Public
)

func (a AccessSpecifier) String() string {
	switch a {
	case AS_None:
		return "<none>"
	case AS_Private:
		return "private"
	case AS_Protected:
		return "protected"
	case AS_Public:
		return "public"
	}
	panic("unreachable")
}

// ----------------------------------------------------------------------------
// helper functions

func is_anonymous(scope string) bool {
	return strings.HasPrefix(scope, ".")
}

// set_scope decorates a list of members with the correct embedding scope
func set_scope(members []Member, t Type, outer string) error {
	for i, _ := range members {
		members[i].Scope = outer
	}
	return nil
}

// add_type adds a type into the db of types
func add_type(t Type) {
	id := t.(Id)
	add_id(id)
	return

}

// gen_new_name returns a new name for n from the list of qualifiers
func gen_new_name(n string, qual TypeQualifier) string {
	if (qual & TQ_Volatile) != 0 {
		n = "volatile " + n
	}
	if (qual & TQ_Restrict) != 0 {
		n = "restrict " + n
	}
	if (qual & TQ_Const) != 0 {
		n = "const " + n
	}
	return n
}

// ----------------------------------------------------------------------------
// make sure the interfaces are implemented

var _ Type = (*placeHolderType)(nil)
var _ Type = (*CvrQualType)(nil)
var _ Type = (*PtrType)(nil)
var _ Type = (*RefType)(nil)
var _ Type = (*TypedefType)(nil)
var _ Type = (*ArrayType)(nil)
var _ Type = (*StructType)(nil)
var _ Type = (*EnumType)(nil)
var _ Type = (*UnionType)(nil)
var _ Type = (*ClassType)(nil)
var _ Type = (*FunctionType)(nil)

var _ Id = (*placeHolderType)(nil)
var _ Id = (*CvrQualType)(nil)
var _ Id = (*PtrType)(nil)
var _ Id = (*RefType)(nil)
var _ Id = (*TypedefType)(nil)
var _ Id = (*ArrayType)(nil)
var _ Id = (*StructType)(nil)
var _ Id = (*EnumType)(nil)
var _ Id = (*UnionType)(nil)
var _ Id = (*ClassType)(nil)
var _ Id = (*FunctionType)(nil)

func init() {
	g_ptrsz = unsafe.Sizeof(uintptr(0))
}

// EOF
