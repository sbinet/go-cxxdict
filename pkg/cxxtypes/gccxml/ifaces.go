package gccxml

import (
	"github.com/sbinet/go-cxxdict/pkg/cxxtypes"
)

type i_id interface {
	id() string
}

type i_align interface {
	align() uintptr
}

type i_name interface {
	i_id
	name() string
	set_name(n string)
}

type i_mangler interface {
	mangled() string
	demangled() string
	set_mangled(n string)
	set_demangled(n string)
}

type i_context interface {
	i_name
	context() string
}

type i_kind interface {
	i_id
	kind() cxxtypes.TypeKind
}

type i_idkind interface {
	i_id
	idkind() cxxtypes.IdKind
}

type i_field interface {
	i_name
	idkind() cxxtypes.IdKind
	kind() cxxtypes.TypeKind
	typename() string
	access() cxxtypes.AccessSpecifier
	offset() uintptr
}

type i_size interface {
	i_name
	size() uintptr
}

type i_gentypename interface {
	gentypename() string
}

type i_repr interface {
	repr() string
}

// EOF
