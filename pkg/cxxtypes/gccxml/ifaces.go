package gccxml

type i_id interface {
	id() string
}

type i_align interface {
	align() uintptr
}

type i_name interface {
	name() string
	set_name(n string)
}

type i_mangler interface {
	mangled() string
	demangled() string
	set_mangled(n string)
	set_demangled(n string)
}

type i_repr interface {
	repr() string 
}

// EOF
