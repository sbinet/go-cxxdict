// Package cxxtypes describes C++ types (classes, structs, functions, ...) which
// have been somehow loaded into memory (from gccxml, clang, ...)
package cxxtypes

import (
	"errors"
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
	Name( /*mod FINAL|QUALIFIED|SCOPED*/) string

	// Size returns the number of bytes needed to store
	// a value of the given type
	Size() uintptr

	// Kind returns the specific kind of this type.
	Kind() TypeKind

	// Qualifiers returns the or'ed values of qualifiers applied to this type.
	Qualifiers() TypeQualifier

	// DeclScope returns the declaring scope of this type
	DeclScope() *Scope

	// CanonicalType returns the underlying type with all
	// the "sugar" removed.
	//CanonicalType() Type

	to_commonType() *commonType
}

// the db of all types
var g_types map[string]Type

// size of a pointer for this platform
var g_ptrsz uintptr

// TypeByName retrieves a type by its fully qualified name.
// Returns nil if no such type exists.
func TypeByName(n string) Type {
	t, ok := g_types[n]
	if ok {
		return t
	}
	return nil
}

// TypeNames returns the list of type names currently defined.
func TypeNames() []string {
	names := make([]string, 0, len(g_types))
	for k, _ := range g_types {
		names = append(names, k)
	}
	return names
}

// NumType returns the number of currently defined types
func NumType() int {
	return len(g_types)
}

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

// commonType is the common implementation of most types.
// It is embedded in other, public struct types, but always
// with a unique tag like `cxxtypes:"array"` or `cxxtypes:"ptr"`
// so that code cannot convert from, say, *ArrayType to *PtrType.
type commonType struct {
	size  uintptr       // size in bytes
	kind  TypeKind      // the specific kind of this type
	qual  TypeQualifier // the qualifiers applied to this type
	scope *Scope        // declaring scope of this type
	name  string        // the fully qualified name of the type
	//canon Type          // the canonical type of this type
}

func (t *commonType) Name() string {
	return t.name
}

func (t *commonType) Size() uintptr {
	return t.size
}

func (t *commonType) Kind() TypeKind {
	return t.kind
}

func (t *commonType) Qualifiers() TypeQualifier {
	return t.qual
}

func (t *commonType) DeclScope() *Scope {
	return t.scope
}

func (t *commonType) String() string {
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		t.Name(), t.Size(), t.Kind(), t.Qualifiers())
}

func (t *commonType) to_commonType() *commonType {
	return t
}

type placeHolderType struct {
	name string
	typ  Type
}

func NewPlaceHolder(name string) Type {
	return &placeHolderType{name: name, typ: nil}
}

func (t *placeHolderType) sync() bool {
	if t.typ == nil {
		t.typ = TypeByName(t.name)
	}
	return t.typ != nil
}
func (t *placeHolderType) Name() string {
	if t.sync() {
		return t.typ.Name()
	}
	return t.name
}

func (t *placeHolderType) Size() uintptr {
	if t.sync() {
		return t.typ.Size()
	}
	return uintptr(0)
}

func (t *placeHolderType) Kind() TypeKind {
	if t.sync() {
		return t.typ.Kind()
	}
	return TK_Invalid
}

func (t *placeHolderType) Qualifiers() TypeQualifier {
	if t.sync() {
		return t.typ.Qualifiers()
	}
	return TQ_None
}

func (t *placeHolderType) DeclScope() *Scope {
	if t.sync() {
		return t.typ.DeclScope()
	}
	return nil
}

func (t *placeHolderType) String() string {
	t.sync()
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		t.Name(), t.Size(), t.Kind(), t.Qualifiers())
}

func (t *placeHolderType) to_commonType() *commonType {
	t.sync()
	return t.typ.to_commonType()
}

// func (t *commonType) CanonicalType() Type {
// 	return t.canon
// }

// NewFundamentalType creates a C/C++ builtin type.
func NewFundamentalType(name string, size uintptr, kind TypeKind, scope *Scope) Type {
	tt := &FundamentalType{
		commonType: commonType{
			size:  size,
			kind:  kind,
			qual:  TQ_None,
			scope: scope,
			name:  name,
		},
	}
	add_type(tt)
	return tt
}

// FundamentalType represents a builtin type
type FundamentalType struct {
	commonType `cxxtypes:"builtin"`
}

// NewQualType creates a new const-restrict-volatile qualified type.
// The new qualifiers are added to the old ones of the base type.
func NewQualType(n string, t Type, scope *Scope, qual TypeQualifier) (q Type) {
	q = &CvrQualType{
		name:  n,
		qual:  qual,
		typ:   t,
		scope: scope,
	}
	add_type(q)
	return
}

type CvrQualType struct {
	name  string
	qual  TypeQualifier
	typ   Type // the decorated type
	scope *Scope
}

func (t *CvrQualType) Name() string {
	return t.name
}

func (t *CvrQualType) Size() uintptr {
	return t.typ.Size()
}

func (t *CvrQualType) Kind() TypeKind {
	return t.typ.Kind()
}

func (t *CvrQualType) Qualifiers() TypeQualifier {
	return t.qual | t.typ.Qualifiers()
}

func (t *CvrQualType) DeclScope() *Scope {
	return t.scope
}

func (t *CvrQualType) String() string {
	return fmt.Sprintf(`{"%s" sz=%d kind=%v qual=%v}`,
		t.Name(), t.Size(), t.Kind(), t.Qualifiers())
}

func (t *CvrQualType) to_commonType() *commonType {
	return t.typ.to_commonType()
}

// NewPtrType creates a new pointer type from an already existing type t.
func NewPtrType(t Type, scope *Scope) *PtrType {
	p := &PtrType{
		commonType: commonType{
			size:  g_ptrsz,
			kind:  TK_Ptr,
			qual:  t.Qualifiers(),
			scope: scope,
			name:  t.Name() + "*",
		},
		typ: t,
	}
	add_type(p)
	return p
}

// PtrType represents a typed ptr
type PtrType struct {
	commonType `cxxtypes:"ptr"`
	typ        Type // the pointee type, possibly cvr-qualified
}

// UnderlyingType returns the type of the pointee
func (t *PtrType) UnderlyingType() Type {
	return t.typ
}

// NewRefType creates a new reference type from an already existing type t.
func NewRefType(t Type, scope *Scope) *RefType {
	r := &RefType{
		commonType: commonType{
			size:  g_ptrsz,
			kind:  TK_LValueRef,
			qual:  t.Qualifiers(),
			scope: scope,
			name:  t.Name() + "&",
		},
		typ: t,
	}
	add_type(r)
	return r
}

// RefType represents a typed reference
type RefType struct {
	commonType `cxxtypes:"ref"`
	typ        Type // the referenced type, possibly cvr-qualified
}

// UnderlyingType returns the referenced type
func (t *RefType) UnderlyingType() Type {
	return t.typ
}

// NewTypedefType creates a new typedef from an already existing type t.
func NewTypedefType(n string, t Type, scope *Scope) *TypedefType {
	tt := &TypedefType{
		commonType: commonType{
			size:  t.Size(),
			kind:  TK_Typedef,
			qual:  TQ_None,
			scope: scope,
			name:  n,
		},
		typ: t,
	}
	add_type(tt)
	return tt
}

// TypedefType represents a typedef
type TypedefType struct {
	commonType `cxxtypes:"typedef"`
	typ        Type // the typedef'd type, possible cvr-qualified
}

// UnderlyingType returns the type of the typedef'd type
func (t *TypedefType) UnderlyingType() Type {
	return t.typ
}

// NewArrayType creates a new array of type T[n].
func NewArrayType(sz uintptr, t Type, scope *Scope) *ArrayType {
	tt := &ArrayType{
		commonType: commonType{
			size:  t.Size() * sz,
			kind:  TK_ConstantArray,
			qual:  TQ_None,
			scope: scope,
			name:  t.Name() + fmt.Sprintf("[%d]", sz),
		},
		elem: t,
		len:  sz,
	}
	add_type(tt)
	return tt
}

// ArrayType represents a fixed array type
type ArrayType struct {
	commonType `cxxtypes:"array"`
	elem       Type    // array element type
	len        uintptr // array length
}

// Elem returns the type of the array's elements
func (t *ArrayType) Elem() Type {
	return t.elem
}

// Len returns the size of the array
func (t *ArrayType) Len() uintptr {
	return t.len
}

// NewStructType creates a new struct type.
func NewStructType(n string, sz uintptr, scope *Scope) *StructType {
	t := &StructType{
		commonType: commonType{
			size:  sz,
			kind:  TK_Record,
			qual:  TQ_None,
			scope: scope,
			name:  n,
		},
		bases:   make([]Base, 0),
		members: make([]Member, 0),
	}
	// t.members = append(t.members, members...)
	// set_scope(t.members, t, scope)
	add_type(t)
	return t
}

// StructType represents a C-struct type
type StructType struct {
	commonType `cxxtypes:"struct"`
	bases      []Base
	members    []Member
}

// NumMember returns a struct type's member count
func (t *StructType) NumMember() int {
	return len(t.members)
}

// Member returns a struct type's i'th member
// It panics if i is not in the range [0, NumMember())
func (t *StructType) Member(i int) *Member {
	if i < 0 || i >= len(t.members) {
		panic("cxxtypes: Member index out of range")
	}
	return &t.members[i]
}

// SetMembers sets the members of this struct type
func (t *StructType) SetMembers(mbrs []Member) error {
	t.members = make([]Member, len(mbrs))
	copy(t.members, mbrs)
	err := set_scope(t.members, t, t.scope)
	if err != nil {
		return err
	}
	return nil
}

// NumBase returns a struct type's base struct count
func (t *StructType) NumBase() int {
	return len(t.bases)
}

// Base returns a struct type's i'th base
// it panics if i is not in the range [0, NumBase()0)
func (t *StructType) Base(i int) *Base {
	if i < 0 || i >= len(t.bases) {
		panic("cxxtypes: Base index out of range")
	}
	return &t.bases[i]
}

// HasBase returns true if a struct has b as one of its base structs
func (t *StructType) HasBase(b Type) bool {
	for i, _ := range t.bases {
		if t.bases[i].Type() == b {
			return true
		}
	}
	return false
}

// SetBases sets the bases of this struct type
func (t *StructType) SetBases(bases []Base) error {
	t.bases = make([]Base, len(bases))
	copy(t.bases, bases)
	return nil
}

// NewEnumType creates a new enum type.
func NewEnumType(n string, members []Member, scope *Scope) *EnumType {
	var sz uintptr = 0
	if len(members) > 0 {
		// take the size of the first member type,
		// they should all be the same
		sz = members[0].Type().Size()
	}
	t := &EnumType{
		commonType: commonType{
			size:  sz,
			kind:  TK_Enum,
			qual:  TQ_None,
			scope: scope,
			name:  n,
		},
		members: make([]Member, 0, len(members)),
	}
	t.members = append(t.members, members...)
	// FIXME: enum-values *should* leak into scope.Outer()!
	set_scope(t.members, t, scope)

	add_type(t)
	return t
}

// EnumType represents an enum type
type EnumType struct {
	commonType `cxxtypes:"enum"`
	members    []Member
}

// NumMember returns an enum type's member count
func (t *EnumType) NumMember() int {
	return len(t.members)
}

// Member returns an enum type's i'th member
// It panics if i is not in the range [0, NumMember())
func (t *EnumType) Member(i int) *Member {
	if i < 0 || i >= len(t.members) {
		panic("cxxtypes: Member index out of range")
	}
	return &t.members[i]
}

// NewUnionType creates a new union type.
func NewUnionType(n string, members []Member, scope *Scope) *UnionType {
	t := &UnionType{
		commonType: commonType{
			size:  0,
			kind:  TK_Record,
			qual:  TQ_None,
			scope: scope,
			name:  n,
		},
		members: make([]Member, 0, len(members)),
	}
	for i, _ := range members {
		t.members = append(t.members, members[i])
		sz := members[i].Type().Size()
		if t.commonType.size < sz {
			t.commonType.size = sz
		}
	}
	set_scope(t.members, t, scope)
	add_type(t)
	return t
}

// UnionType represents a union type
type UnionType struct {
	commonType `cxxtypes:"union"`
	members    []Member
}

// NewClassType creates a new class type.
func NewClassType(n string, sz uintptr, scope *Scope) *ClassType {
	t := &ClassType{
		commonType: commonType{
			size:  sz,
			kind:  TK_Record,
			qual:  TQ_None,
			scope: scope,
			name:  n,
		},
		bases:   make([]Base, 0),
		members: make([]Member, 0),
	}
	// t.bases = append(t.bases, bases...)
	// t.members = append(t.members, members...)
	// err := set_scope(t.members, t, scope)
	// if err != nil {
	// 	panic(err)
	// }
	add_type(t)
	return t
}

// ClassType represents a C++ class type
type ClassType struct {
	commonType `cxxtypes:"class"`
	bases      []Base
	members    []Member
}

// NumMember returns a class type's member count
func (t *ClassType) NumMember() int {
	return len(t.members)
}

// Member returns a class type's i'th member
// It panics if i is not in the range [0, NumMember())
func (t *ClassType) Member(i int) *Member {
	if i < 0 || i >= len(t.members) {
		panic("cxxtypes: Member index out of range")
	}
	return &t.members[i]
}

// SetMembers sets the members of this class type
func (t *ClassType) SetMembers(mbrs []Member) error {
	t.members = make([]Member, len(mbrs))
	copy(t.members, mbrs)
	err := set_scope(t.members, t, t.scope)
	if err != nil {
		return err
	}
	return nil
}

// NumBase returns a class type's base class count
func (t *ClassType) NumBase() int {
	return len(t.bases)
}

// Base returns a class type's i'th base
// it panics if i is not in the range [0, NumBase()0)
func (t *ClassType) Base(i int) *Base {
	if i < 0 || i >= len(t.bases) {
		panic("cxxtypes: Base index out of range")
	}
	return &t.bases[i]
}

// HasBase returns true if a class has b as one of its base classes
func (t *ClassType) HasBase(b Type) bool {
	for i, _ := range t.bases {
		if t.bases[i].Type() == b {
			return true
		}
	}
	return false
}

// SetBases sets the bases of this class type
func (t *ClassType) SetBases(bases []Base) error {
	t.bases = make([]Base, len(bases))
	copy(t.bases, bases)
	return nil
}

// NewBase creates a new base class/struct
func NewBase(offset uintptr, typ Type, access AccessSpecifier, virtual bool) Base {
	return Base{
		offset:  offset,
		typ:     typ,
		access:  access,
		virtual: virtual,
	}
}

// Base represents a base class of a C++ class type
type Base struct {
	offset  uintptr         // the offset to the base class
	typ     Type            // Type of this base class
	access  AccessSpecifier // specifier for the derivation
	virtual bool            // whether the derivation is virtual
}

// Offset returns the offset of this base class
func (b *Base) Offset() uintptr {
	return b.offset
}

// Type returns the Type of this base class
func (b *Base) Type() Type {
	return b.typ
}

// Access returns the access specifier for this base class
// ex:
//   is_public := b.Access() & AS_Public
func (b *Base) Access() AccessSpecifier {
	return b.access
}

// IsVirtual returns whether the derivation is virtual
func (b *Base) IsVirtual() bool {
	return b.virtual
}

func (b *Base) String() string {
	hdr := ""
	if b.IsVirtual() {
		hdr = "virtual "
	}
	return fmt.Sprintf("Base{%s%s %s offset=%d}",
		hdr, b.access.String(), b.typ.Name(), b.offset)
}

// NewMember creates a new member for a struct, class, enum or union
func NewMember(name string, typ Type, kind TypeKind, access AccessSpecifier, offset uintptr, scope *Scope) Member {
	return Member{
		name:   name,
		typ:    typ,
		kind:   kind,
		access: access,
		scope:  scope,
		offset: offset,
	}
}

// Member represents a member in a struct, class, enum or union
type Member struct {
	name   string          // name of this member
	typ    Type            // the type of this member
	kind   TypeKind        // the kind of this member
	access AccessSpecifier // the access specifier for this member
	scope  *Scope          // embedding scope
	offset uintptr         // the offset in the embedding scope
}

// Name returns the name of this member
func (m *Member) Name() string {
	return m.name
}

// Type returns the type of this member
func (m *Member) Type() Type {
	return m.typ
}

// Kind returns the kind of this member.
func (m *Member) Kind() TypeKind {
	return m.kind
}

// Scope returns the embedding scope of this member
func (m *Member) Scope() *Scope {
	return m.scope
}

// Offset return the offset of this member within the embedding scope
func (m *Member) Offset() uintptr {
	return m.offset
}

func (m *Member) IsDataMember() bool {
	return !m.IsFunctionMember() && !m.IsEnumMember()
}

func (m *Member) IsEnumMember() bool {
	return m.kind == TK_Enum
}

func (m *Member) IsFunctionMember() bool {
	return (m.kind == TK_FunctionProto) ||
		(m.kind == TK_FunctionNoProto)
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
	return (m.access & AS_Public) != 0
}

func (m *Member) IsProtected() bool {
	return (m.access & AS_Protected) != 0
}

func (m *Member) IsPrivate() bool {
	return (m.access & AS_Private) != 0
}

func (m *Member) String() string {
	hdr := "DMbr"
	if m.IsEnumMember() {
		hdr = "EMbr"
	}
	if m.IsFunctionMember() {
		hdr = "FMbr"
	}
	return fmt.Sprintf("%s{%s type='%s' kind=%s access=%s offset=%d}",
		hdr,
		m.name, m.typ.Name(), m.kind.String(), m.access.String(), m.offset)
}

// NewFunctionType creates a new function type.
func NewFunctionType(n string, qual TypeQualifier, specifiers TypeSpecifier, variadic bool, params []Parameter, ret Type, scope *Scope) *FunctionType {
	t := &FunctionType{
		commonType: commonType{
			size:  0,
			kind:  TK_FunctionProto,
			qual:  qual,
			scope: scope,
			name:  n,
		},
		tspec:    specifiers,
		variadic: variadic,
		params:   make([]Parameter, 0, len(params)),
		ret:      ret,
	}

	// only add that type to the db if it isn't a method (of a class/struct)
	if !t.IsMethod() {
		add_type(t)
	}
	return t
}

// FunctionType represents a C/C++ function, a member function, c-tor, ...
type FunctionType struct {
	commonType `cxxtypes:"function"`
	tspec      TypeSpecifier // or'ed value of virtual/inline/...
	variadic   bool          // whether this function is variadic
	params     []Parameter   // the parameters to this function
	ret        Type          // return type of this function
}

// Specifier returns the type specifier for this function
func (t *FunctionType) Specifier() TypeSpecifier {
	return t.tspec
}

func (t *FunctionType) IsVirtual() bool {
	return (t.tspec & TS_Virtual) != 0
}

func (t *FunctionType) IsStatic() bool {
	return (t.tspec & TS_Static) != 0
}

func (t *FunctionType) IsConstructor() bool {
	return (t.tspec & TS_Constructor) != 0
}

func (t *FunctionType) IsDestructor() bool {
	return (t.tspec & TS_Destructor) != 0
}

func (t *FunctionType) IsCopyConstructor() bool {
	return (t.tspec & TS_CopyCtor) != 0
}

func (t *FunctionType) IsOperator() bool {
	return (t.tspec & TS_Operator) != 0
}

func (t *FunctionType) IsMethod() bool {
	return (t.tspec & TS_Method) != 0
}

func (t *FunctionType) IsInline() bool {
	return (t.tspec & TS_Inline) != 0
}

func (t *FunctionType) IsConverter() bool {
	return (t.tspec & TS_Converter) != 0
}

// IsVariadic returns whether this function is variadic
func (t *FunctionType) IsVariadic() bool {
	return t.variadic
}

// NumParam returns a function type's input parameter count.
func (t *FunctionType) NumParam() int {
	return len(t.params)
}

// Param returns the i'th parameter of this function's type.
// It panics if i is not in the range [0, NumParam())
func (t *FunctionType) Param(i int) *Parameter {
	if i < 0 || i >= t.NumParam() {
		panic("cxxtypes: Param index out of range")
	}
	return &t.params[i]
}

// NumDefaultParam returns a function type's input with default value parameter count.
func (t *FunctionType) NumDefaultParam() int {
	n := 0
	for i, _ := range t.params {
		if t.params[i].defval {
			n += 1
		}
	}
	return n
}

// ReturnType returns the return type of this function's type.
// FIXME: return nil for 'void' fct ?
func (t *FunctionType) ReturnType() Type {
	return t.ret
}

// NewParameter creates a new parameter.
func NewParameter(n string, t Type, defval bool) *Parameter {
	return &Parameter{
		name:   n,
		typ:    t,
		defval: defval,
	}
}

// Parameter represents a parameter of a function's signature
type Parameter struct {
	name   string // name of the parameter
	typ    Type   // type of this parameter
	defval bool   // whether this parameter has a default value
}

// Name returns the name of the parameter
func (p *Parameter) Name() string {
	return p.name
}

// Type returns the type of this parameter
func (p *Parameter) Type() Type {
	return p.typ
}

// HasDefaultValue returns whether this parameter has a default value
func (p *Parameter) HasDefaultValue() bool {
	return p.defval
}

// NewVar creates a new global variable
func NewVar(n string, specifiers TypeSpecifier, typ Type, scope *Scope) *Var {
	return &Var{
		name:  n,
		tspec: specifiers,
		typ:   typ,
		scope: scope,
	}
}

// Var represents a variable
type Var struct {
	name  string        // name of the variable
	tspec TypeSpecifier // or'ed value of const/extern/static/...
	typ   Type          // type of this variable
	scope *Scope        // the scope holding that variable
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

// set_scope decorates a list of members with the correct embedding scope
func set_scope(members []Member, t Type, outer *Scope) error {
	// get type as scope
	obj := outer.Lookup(t.Name())
	if obj == nil {
		return errors.New("set_scope: could not find scope [" + t.Name() + "]")
	}

	scope := obj.Data.(*Scope)
	for i, _ := range members {
		members[i].scope = scope
	}
	return nil
}

// add_type adds a type into the db of types
func add_type(t Type) {
	_, exists := g_types[t.Name()]
	if exists {
		panic("cxxtypes: type [" + t.Name() + "] already in registry")
	}
	g_types[t.Name()] = t
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

func init() {
	g_types = make(map[string]Type)
	g_ptrsz = unsafe.Sizeof(uintptr(0))

}

// EOF
