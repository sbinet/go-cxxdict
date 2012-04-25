package gccxml

import (
	"bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
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

type i_field interface {
	i_name
	kind() cxxtypes.TypeKind
	cxxtype() cxxtypes.Type
	access() cxxtypes.AccessSpecifier
	offset() uintptr
}
type i_gentypename interface {
	gentypename() string
}

type i_repr interface {
	repr() string
}

// EOF
