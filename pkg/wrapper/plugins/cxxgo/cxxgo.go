package cxxgo

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sbinet/go-cxxdict/pkg/cxxtypes"
	"github.com/sbinet/go-cxxdict/pkg/wrapper"
)

type bufmap_t map[string]*bytes.Buffer
type idmap_t map[cxxtypes.Id]uint64

type plugin struct {
	gen *wrapper.Generator // the generator which is invoking us
	sel []string           // the identifiers to select
	ids []string           // the selected identifiers
}

func (p *plugin) Name() string {
	return "cxxgo.plugin"
}

func (p *plugin) Init(g *wrapper.Generator) error {
	if g == nil {
		return fmt.Errorf("cxxgo: nil pointer to wrapper.Generator")
	}
	p.gen = g
	p.sel = []string{
		"T*",
		"Math*",
		"NS::*tmpl*",
		"Foo",
		"IFoo",
		"NS*",
		"TT*",
		"Base*",
		"D1",
		"IAlg",
		"App",
		"Alg*",
		"With*Base*",

		"*Enum*",
		"cblas*",
		"*CBLAS_*",

		"add42",
	}
	p.ids = []string{}

	fmt.Printf("cxxgo.Init: args=%v\n", g.Args)
	return nil
}

func (p *plugin) Generate(fd *wrapper.FileDescriptor) error {
	fmt.Printf("cxxgo.Generate...\n")

	// loop over identifiers and filter them out
	for _, n := range cxxtypes.IdNames() {
		selected := false
		for _, sel := range p.sel {
			matched, err := path.Match(sel, n)
			if err != path.ErrBadPattern && matched {
				selected = true
				break
			}
		}
		if selected && is_anon(n) {
			fmt.Printf(":: discarding [%s] (anonymous identifier)\n", n)
			selected = false
		}
		if selected {
			p.ids = append(p.ids, n)
			nn := gen_go_name_from_id(cxxtypes.IdByName(n))
			_cxx2go_typemap[n] = nn
		}
	}
	{
		// select dependent types...
		sel_deps := []string{}
		for _, n := range p.ids {
			id := cxxtypes.IdByName(n)
			sel_deps = append(sel_deps, get_dependent_ids(sel_deps, id)...)
		}
		p.ids = append(p.ids, sel_deps...)
	}
	{
		// make sure we don't wrap a member twice: remove duplicates
		nremoved := 0
		for {
			nremoved = 0
			//fmt.Printf("---\n")
			sel_ids := make([]string, 0, len(p.ids))
			for _, n := range p.ids {
				//fmt.Printf("--> [%s]...\n", n)
				id := cxxtypes.IdByName(n)
				switch iid := id.(type) {
				case *cxxtypes.Member:
					pid := cxxtypes.IdByName(iid.Scope)
					if pid != nil &&
						str_is_in_slice(pid.IdScopedName(), p.ids) {
						// parent is already selected... discard member
						nremoved += 1
						//fmt.Printf("** discard [%s]\n", n)
					} else {
						// keep it
						sel_ids = append(sel_ids, n)
						//fmt.Printf("** keep [%s] (parent=%v)\n", n, pid)
					}
				case *cxxtypes.OverloadFunctionSet:
					pid := cxxtypes.IdByName(iid.Scope)
					if pid != nil &&
						str_is_in_slice(pid.IdScopedName(), p.ids) {
						// parent is already selected... discard member
						nremoved += 1
						//fmt.Printf("** discard [%s]\n", n)
					} else {
						// keep it
						sel_ids = append(sel_ids, n)
						//fmt.Printf("** keep [%s] (parent=%v)\n", n, pid)
					}
				default:
					//fmt.Printf("--- not a member [%s] (%T)\n", n, id)
					// remove any lingering duplicate
					if !str_is_in_slice(n, sel_ids) {
						sel_ids = append(sel_ids, n)
					}
				}
			}
			p.ids = sel_ids
			if nremoved == 0 {
				break
			}
		}
	}
	fmt.Printf("selected ids: ['%v']\n", strings.Join(p.ids, "', '"))
	if len(p.ids) <= 0 {
		fmt.Printf("nothing to wrap\n")
		return nil
	}

	fd_cxx, err := os.Create(fd.Name + "_" + p.Name() + ".cxx")
	if err != nil {
		return err
	}
	defer fd_cxx.Close()

	fd_hdr, err := os.Create(fd.Name + "_" + p.Name() + ".h")
	if err != nil {
		return err
	}
	defer fd_hdr.Close()

	fd_go, err := os.Create(fd.Name + "_" + p.Name() + ".go")
	if err != nil {
		return err
	}
	defer fd_go.Close()

	fd.Files["cxx"] = fd_cxx
	fd.Files["hdr"] = fd_hdr
	fd.Files["go"] = fd_go

	_, err = fd_go.WriteString(fmt.Sprintf(
		_go_hdr,
		fd.Package,
		fd_hdr.Name(),
		fd.Name,
		fd.Name+"_cxxgo.plugin",
	))
	if err != nil {
		return err
	}

	_, err = fd_cxx.WriteString(fmt.Sprintf(
		_cxx_hdr,
		fd_hdr.Name(),
		fd.Header,
	))
	if err != nil {
		return err
	}

	_, err = fd_hdr.WriteString(fmt.Sprintf(
		_hdr_hdr,
		fd.Package,
		fd.Package,
	))
	if err != nil {
		return err
	}

	for _, n := range p.ids {
		id := cxxtypes.IdByName(n)
		cid := get_cxxgo_id(p.gen.Fd.Package, id)
		switch id := id.(type) {
		case *cxxtypes.ClassType:
			err := p.wrapClass(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.StructType:
			err := p.wrapStruct(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.OverloadFunctionSet:
			err := p.wrapFunction(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.Namespace:
			err := p.wrapNamespace(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.Member:
			// will be done by a scope-level thingy...

			// err := p.wrapMember(id)
			// if err != nil {
			// 	return err
			// }

		case *cxxtypes.TypedefType:
			err := p.wrapTypedef(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.RefType:
			err := p.wrapRefType(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.PtrType:
			err := p.wrapPtrType(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.CvrQualType:
			err := p.wrapCvrQualType(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.EnumType:
			err := p.wrapEnum(cid, id)
			if err != nil {
				return err
			}

		case *cxxtypes.FundamentalType:
			// ignore

		default:
			panic(fmt.Errorf("type [%T] unhandled (%s)!", id, id.IdScopedName()))
		}
	}

	_, err = fd_go.WriteString(fmt.Sprintf(
		_go_footer,
		fd.Package,
	))
	if err != nil {
		return err
	}

	_, err = fd_cxx.WriteString(_cxx_footer)
	if err != nil {
		return err
	}

	_, err = fd_hdr.WriteString(fmt.Sprintf(
		_hdr_footer,
		fd.Package,
	))
	if err != nil {
		return err
	}

	fmt.Printf("cxxgo.Generate... [done]\n")
	return nil
}

func (p *plugin) wrapClass(cid *cxxgo_id, id *cxxtypes.ClassType) (err error) {
	fmt.Printf(":: wrapping class [%s]...\n", id.IdScopedName())
	err = nil

	bufs := new_bufmap(
		"cxx_head",
		"cxx_body",
		"cxx_tail",
		"go_impl",
		"go_iface",
		"go_g_iface", //global iface (ie: outside interface context)
	)

	clf := "::" + id.IdScopedName()
	go_cls_impl_name := "Gocxxcptr" + cid.goname

	fmter(bufs["go_iface"],
		`
// %s wraps the C++ class %s
type %s interface {
    /* -- gocxx internals begin -- */
	Gocxxcptr() uintptr
	GocxxIs%s()
    /* -- gocxx internals end -- */

`,
		cid.goname,
		clf,
		cid.goname,
		cid.goname,
	)

	fmter(bufs["go_impl"], "type %s uintptr\n", go_cls_impl_name)
	fmter(bufs["go_impl"],
		"func (p %s) Gocxxcptr() uintptr {\n\treturn uintptr(p)\n}\n",
		go_cls_impl_name,
	)
	fmter(bufs["go_impl"],
		"func (p %s) GocxxIs%s() {\n}\n",
		go_cls_impl_name,
		cid.goname,
	)

	// bases...
	bufs_bases := make([]bufmap_t, 0, len(id.Bases))
	for _, base := range id.Bases {
		bufs_base := bufmap_t{
			"go_impl":  bytes.NewBufferString(""),
			"go_iface": bytes.NewBufferString(""),
			"cxx":      bytes.NewBufferString(""),
		}
		bufs_bases = append(bufs_bases, bufs_base)
		err := p.wrapBaseClass(&base, bufs)
		if err != nil {
			return err
		}
		if base.IsPublic() {
			base_id := base.Type().(cxxtypes.Id)
			bid := get_cxxgo_id(p.gen.Fd.Package, base_id)
			go_base_cls_iface_name := bid.goname
			fmter(bufs["go_iface"],
				"\tGet%s() %s\n",
				go_base_cls_iface_name,
				go_base_cls_iface_name)

			fmter(bufs["go_impl"],
				"func (p %s) Get%s() %s {\n",
				go_cls_impl_name,
				go_base_cls_iface_name,
				go_base_cls_iface_name,
			)
			fmter(bufs["go_impl"],
				"return Gocxxcptr%s(p.Gocxxcptr())\n}\n",
				go_base_cls_iface_name,
			)

			fmter(bufs["go_iface"],
				"\tGocxxIs%s()\n",
				go_base_cls_iface_name,
			)
			fmter(bufs["go_impl"],
				"func (p %s) GocxxIs%s() {\n}\n",
				go_cls_impl_name,
				go_base_cls_iface_name,
			)
		}
	}

	fct_mbr_indices := make([]int, 0, len(id.Members))
	fct_mbr_names := make([]string, 0, len(id.Members))
	// data-members
	for i, mbr := range id.Members {
		if !p.mbr_filter(&mbr) {
			fmt.Printf(":: discarding [%s]...\n", mbr.Name)
			continue
		}
		if mbr.IsFunctionMember() {
			//FIXME: O(n^2)
			if !str_is_in_slice(mbr.Name, fct_mbr_names) {
				fct_mbr_names = append(fct_mbr_names, mbr.Name)
				fct_mbr_indices = append(fct_mbr_indices, i)
			}
			continue
		}
		mid := cxxtypes.IdByName(mbr.Name)
		if mid == nil {
			fmt.Printf("==[%s]==(idx=%d)\n", mbr.Name, i)
			fmt.Printf("==dmbr: %v\n", mbr.IsDataMember())
			fmt.Printf("==fmbr: %v\n", mbr.IsFunctionMember())
			fmt.Printf("==embr: %v\n", mbr.IsEnumMember())
			fmt.Printf("==mkind: %v\n", mbr.Kind)
			fmt.Printf("==mdind: %v\n", mbr.IdKind())
			return fmt.Errorf("cxxgo: could not retrieve identifier [%s]\n%s", mbr.Name, mbr)
		}
		fmt.Printf("--> (%s)[%s]...\n", mbr.IdScopedName(), mbr)
		err := p.wrapMember(&mbr, bufs)
		if err != nil {
			return err
		}
	}

	// fct-members: handle overloads...
	for _, mbr_idx := range fct_mbr_indices {
		mbr := &id.Members[mbr_idx]
		err := p.wrapFctMember(mbr, bufs)
		if err != nil {
			return err
		}
	}

	fmter(bufs["go_iface"], "}\n\n")

	// commit buffers
	_, err = bufs["go_g_iface"].WriteTo(p.gen.Fd.Files["go"])
	if err != nil {
		return err
	}

	_, err = bufs["go_iface"].WriteTo(p.gen.Fd.Files["go"])
	if err != nil {
		return err
	}

	_, err = bufs["go_impl"].WriteTo(p.gen.Fd.Files["go"])
	if err != nil {
		return err
	}

	_, err = bufs["cxx_head"].WriteTo(p.gen.Fd.Files["cxx"])
	if err != nil {
		return err
	}

	_, err = bufs["cxx_body"].WriteTo(p.gen.Fd.Files["cxx"])
	if err != nil {
		return err
	}

	_, err = bufs["cxx_tail"].WriteTo(p.gen.Fd.Files["cxx"])
	if err != nil {
		return err
	}

	fmt.Printf(":: wrapping class [%s]...[ok]\n", id.IdScopedName())
	return err
}

func (p *plugin) wrapStruct(cid *cxxgo_id, id *cxxtypes.StructType) error {
	fmt.Printf(":: wrapping struct [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping struct [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapNamespace(cid *cxxgo_id, id *cxxtypes.Namespace) error {
	fmt.Printf(":: wrapping namespace [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping namespace [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapEnum(cid *cxxgo_id, id *cxxtypes.EnumType) error {
	var err error = nil
	fmt.Printf(":: wrapping enum [%s]...\n", id.IdScopedName())

	n := "::" + id.IdScopedName()
	//tn := g_strtrans.Replace(id.IdScopedName())
	go_enum_iface_name := gen_go_name_from_id(id)

	bufs := new_bufmap(
		"cxx_head",
		"go_iface",
	)

	fmter(bufs["go_iface"],
		"\n// %s wraps the enum %s\n", go_enum_iface_name, n)
	fmter(bufs["go_iface"], "type %s int\n", go_enum_iface_name)

	// commit buffers
	_, err = bufs["go_iface"].WriteTo(p.gen.Fd.Files["go"])
	if err != nil {
		return err
	}

	fmt.Printf(":: wrapping enum [%s]...[ok]\n", id.IdScopedName())
	return err
}

func (p *plugin) wrapMember(id *cxxtypes.Member, bufs bufmap_t) error {
	fmt.Printf(":: wrapping member [%s]...\n", id.IdScopedName())
	if id.IsDataMember() {
		err := p.wrapDataMember(id, bufs)
		if err != nil {
			return err
		}
	} else if id.IsFunctionMember() {
		err := p.wrapFctMember(id, bufs)
		if err != nil {
			return err
		}
	}
	fmt.Printf(":: wrapping member [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapTypedef(cid *cxxgo_id, id *cxxtypes.TypedefType) error {
	fmt.Printf(":: wrapping typedef [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping typedef [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapRefType(cid *cxxgo_id, id *cxxtypes.RefType) error {
	fmt.Printf(":: wrapping ref-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping ref-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapPtrType(cid *cxxgo_id, id *cxxtypes.PtrType) error {
	fmt.Printf(":: wrapping ptr-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping ptr-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapCvrQualType(cid *cxxgo_id, id *cxxtypes.CvrQualType) error {
	fmt.Printf(":: wrapping cvr-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping cvr-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapBaseClass(id *cxxtypes.Base, bufs bufmap_t) error {
	return nil
}

func (p *plugin) wrapDataMember(id *cxxtypes.Member, bufs bufmap_t) (err error) {
	fmt.Printf(":: wrapping data-member [%s]...\n", id.IdScopedName())

	dm_name := id.IdName()
	dm_typename := id.Type

	// declare getters and setters in the go-interface
	fmter(bufs["go_iface"], "\tGet%s() %s\n\tSet%s(%s %s)\n",
		strings.Title(dm_name),
		cxx2go_typename(dm_typename),
		strings.Title(dm_name),
		dm_name,
		cxx2go_typename(dm_typename),
	)

	clsid := cxxtypes.IdByName(id.Scope)
	if clsid == nil {
		return fmt.Errorf("could not find parent-scope [%s] for member [%s]",
			id.Scope, id.IdScopedName())
	}
	//clf := "::" + clsid.IdScopedName()
	//clt := g_strtrans.Replace(clsid.IdScopedName())
	go_cls_iface_name := gen_go_name_from_id(clsid)
	go_cls_impl_name := "Gocxxcptr" + go_cls_iface_name

	// corresponding implementation...
	fmter(bufs["go_impl"],
		`
func (p %s) Get%s() %s {
 var dummy %s
 return dummy
}
`,
		go_cls_impl_name,
		strings.Title(dm_name),
		cxx2go_typename(dm_typename),
		cxx2go_typename(dm_typename),
	)

	fmter(bufs["go_impl"],
		`
func (p %s) Set%s(arg %s) {
}
`,
		go_cls_impl_name,
		strings.Title(dm_name),
		cxx2go_typename(dm_typename),
	)
	fmt.Printf(":: wrapping data-member [%s]...[ok]\n", id.IdScopedName())
	return err
}

func (p *plugin) wrapFctMember(id *cxxtypes.Member, bufs bufmap_t) error {
	var err error = nil
	fmt.Printf(":: wrapping fct-member [%s]...\n", id.IdScopedName())
	ovfct := cxxtypes.IdByName(id.Name).(*cxxtypes.OverloadFunctionSet)
	cid := get_cxxgo_id(p.gen.Fd.Package, ovfct)
	if cid.wrapped {
		fmt.Printf(":: wrapping fct-member [%s]...[already-wrapped]\n",
			id.IdScopedName())
		return err
	}

	cgo_ovfct := p.new_cxxgo_ovfct(ovfct)
	if len(cgo_ovfct.fcts) > 0 {
		fct := cgo_ovfct.fcts[0].f
		if !fct.IsConstructor() &&
			!fct.IsDestructor() &&
			!fct.IsCopyConstructor() {
			fmter(bufs["go_iface"],
				"\t%s\n",
				cgo_ovfct.go_prototype(),
			)
		}
	}

	err = p.wrapFunction(cid, ovfct)
	cid.wrapped = true

	fmt.Printf(":: wrapping fct-member [%s]...[ok]\n", id.IdScopedName())
	return err
}

func (p *plugin) wrapFunction(cid *cxxgo_id, id *cxxtypes.OverloadFunctionSet) error {
	fmt.Printf(":: wrapping fct [%s]...\n", id.IdScopedName())
	ovfct := id
	var err error = nil
	if cid.wrapped {
		fmt.Printf(":: wrapping fct [%s]... [already wrapped]\n",
			id.IdScopedName())
		return err
	}

	pkg := p.gen.Fd.Package

	bufs := new_bufmap(
		"cxx_head",
		"cxx_body",
		"cxx_tail",
		"cgo_head",
		"go_iface",
		"go_impl",
	)

	cgo_ovfct := p.new_cxxgo_ovfct(ovfct)
	needs_dispatch := len(cgo_ovfct.fcts) > 1
	// the table to regroup cgo-functions by number of args
	dispatch_table := map[int][]*cxxgo_function{}
	go_receiver := ""

	for ifct, _ := range cgo_ovfct.fcts {

		cfct := cgo_ovfct.fcts[ifct]
		fct := cfct.f
		nargs := fct.NumParam()

		// // discard private function-member
		// if fct.IsPrivate() {
		// 	continue
		// }

		go_ret_type := ""
		cid_args := make([]*cxxgo_id, 0, nargs)

		if _, ok := dispatch_table[nargs]; !ok {
			dispatch_table[nargs] = []*cxxgo_function{&cfct}
		} else {
			dispatch_table[nargs] = append(
				dispatch_table[nargs],
				&cfct)
		}

		for i, _ := range fct.Params {
			cid_arg := get_cxxgo_id(pkg, cxxtypes.IdByName(fct.Params[i].Type))
			cid_args = append(cid_args, cid_arg)
		}

		if fct.IsMethod() &&
			!fct.IsConstructor() && !fct.IsDestructor() &&
			!fct.IsCopyConstructor() {
			cid_scope := get_cxxgo_id(pkg, cxxtypes.IdByName(fct.BaseId.Scope))
			go_receiver = fmt.Sprintf("(p Gocxxcptr%s)",
				cid_scope.goname,
			)
		}
		fmter(bufs["go_impl"],
			"\n// wraps [%s]\nfunc %s%s {\n",
			cfct.cxx_prototype(),
			go_receiver,
			cfct.go_prototype(),
		)

		// CGo decl.
		cgo_in := []string{}
		cgo_out := []string{}

		fmter(bufs["cgo_head"],
			"\n/* wraps [%s] */\n%s;\n",
			cfct.cxx_prototype(),
			cfct.cgo_prototype(),
		)

		// C++ wrapper
		cxx_in := []string{}
		fmter(
			bufs["cxx_head"], "\n// wraps [%s]\n%s\n{\n",
			cfct.cxx_prototype(),
			cfct.cgo_prototype(),
		)

		if fct.IsMethod() {
			cid_scope := get_cxxgo_id(pkg, cxxtypes.IdByName(fct.BaseId.Scope))
			if fct.IsDestructor() {
				fmter(bufs["go_impl"],
					"\tc_this := unsafe.Pointer(arg.Gocxxcptr())\n",
				)
			} else if fct.IsConstructor() {
				fmter(bufs["go_impl"],
					"\tvar c_ptr Gocxxcptr%s\n",
					cid_scope.goname,
				)
				fmter(bufs["go_impl"],
					"\tc_this := unsafe.Pointer(&c_ptr)\n",
				)
			} else {
				fmter(bufs["go_impl"],
					"\tc_this := unsafe.Pointer(p)\n",
				)
			}
			cgo_in = append(cgo_in, "c_this")
		}
		if fct.Ret != "" && fct.Ret != "void" {
			cid_ret := get_cxxgo_id(pkg, cxxtypes.IdByName(fct.Ret))
			if cid_ret.is_cstring_like() {
				fmter(bufs["go_impl"],
					"\tc_ret := C.CString(\"\")\n")
				fmter(bufs["go_impl"],
					"\tdefer C.free(unsafe.Pointer(c_ret))\n")
				cgo_out = append(cgo_out,
					"\tgo_ret := C.GoString(c_ret)\n",
					"\treturn go_ret\n",
				)

			} else {
				fmter(bufs["go_impl"],
					"\tvar c_ret C.%s\n",
					cid_ret.cgoname,
				)
				cgo_out = append(cgo_out,
					fmt.Sprintf("\tgo_ret := %s(c_ret)\n", cid_ret.goname),
					"\treturn go_ret\n",
				)
			}
			go_ret_type = fct.Ret

		} else {
			if fct.IsConstructor() {
				cgo_out = append(cgo_out,
					"\treturn c_ptr\n",
				)
			}
			fmter(bufs["go_impl"],
				"\tvar c_ret = unsafe.Pointer(nil)\n",
			)
		}
		cgo_in = append(cgo_in, "unsafe.Pointer(&c_ret)")

		for i, _ := range fct.Params {
			cid_arg := cid_args[i]
			if cid_arg.is_cstring_like() {
				fmter(bufs["go_impl"],
					"\tc_arg_%d := C.CString(arg_%d)\n", i, i)
				fmter(bufs["go_impl"],
					"\tdefer C.free(unsafe.Pointer(c_arg_%d))\n", i)
				cgo_in = append(cgo_in,
					fmt.Sprintf("unsafe.Pointer(c_arg_%d)", i))
			} else if cid_arg.is_class_like() {
				fmter(bufs["go_impl"],
					"\tc_arg_%d := unsafe.Pointer(arg_%d.Gocxxcptr())\n",
					i,
					i,
				)
				cgo_in = append(cgo_in,
					fmt.Sprintf("c_arg_%d", i))
			} else {
				fmter(bufs["go_impl"],
					"\tc_arg_%d := C.%s(arg_%d)\n",
					i,
					cid_arg.cgoname,
					i,
				)
				cgo_in = append(cgo_in,
					fmt.Sprintf("unsafe.Pointer(&c_arg_%d)", i))
			}
		}
		fmter(bufs["go_impl"],
			"\tC.%s(%s)\n%s",
			cfct.cgoname,
			strings.Join(cgo_in, ", "),
			strings.Join(cgo_out, ""),
		)
		fmter(bufs["go_impl"], "}\n")

		if go_ret_type != "" {
			cid_ret := get_cxxgo_id(pkg, cxxtypes.IdByName(fct.Ret))
			cxx_type := cid_ret.id.IdScopedName()
			if strings.HasSuffix(cxx_type, "*") {
				// pointer to data member
				if strings.HasSuffix(cxx_type, ":*") {
					fmter(bufs["cxx_head"],
						"  %s* cxx_ret = (%s*)(c_ret);\n",
						cxx_type, cxx_type,
					)
				} else if cid_ret.is_string_like() {
					fmter(bufs["cxx_head"],
						"  %s cxx_ret( ((_gostring_*)c_ret)->p, ((_gostring_*)c_ret)->n);\n",
						cxx_type[:len(cxx_type)-1],
					)
				} else if cid_ret.is_cstring_like() {
					fmter(bufs["cxx_head"],
						"  %s* cxx_ret = (%s*)c_ret;\n",
						cxx_type, cxx_type,
					)
				} else {
					fmter(bufs["cxx_head"],
						"  %s cxx_ret = (%s)(c_ret);\n",
						cxx_type, cxx_type,
					)
				}
			} else if strings.HasSuffix(cxx_type, "&") {
				if cid_ret.is_string_like() {
					fmter(bufs["cxx_head"],
						"  %s cxx_ret( ((_gostring_*)c_ret)->p, ((_gostring_*)c_ret)->n);\n",
						cxx_type[:len(cxx_type)-1],
					)
				} else {
					fmter(bufs["cxx_head"],
						"  %s* cxx_ret = *(%s**)(&c_ret);\n",
						cxx_type[:len(cxx_type)-1],
						cxx_type[:len(cxx_type)-1],
					)
				}
			} else {
				if cid_ret.is_string_like() {
					fmter(bufs["cxx_head"],
						"  %s cxx_ret( ((_gostring_*)c_ret)->p, ((_gostring_*)c_ret)->n);\n",
						cxx_type[:len(cxx_type)-1],
					)
				} else {
					fmter(bufs["cxx_head"],
						"  %s* cxx_ret = (%s*)c_ret;\n",
						cxx_type, cxx_type,
					)
				}
			}
			fmter(bufs["cxx_body"], "  (*cxx_ret) = ")
		} else {
			fmter(bufs["cxx_body"], "  ")
		}
		if fct.IsMethod() {
			cid_scope := get_cxxgo_id(pkg, cxxtypes.IdByName(fct.BaseId.Scope))
			if fct.IsConstructor() {
				fmter(bufs["cxx_body"], "(*((void**)c_this)) = new ")
			} else if fct.IsDestructor() {
				fmter(bufs["cxx_head"],
					"  %s *cxx_this = (%s*)(c_this);\n",
					cid_scope.id.IdScopedName(),
					cid_scope.id.IdScopedName(),
				)
				fmter(bufs["cxx_body"],
					"delete cxx_this; cxx_this = NULL;\n")

			} else {
				fmter(bufs["cxx_head"],
					"  %s *cxx_this = (%s*)(c_this);\n",
					cid_scope.id.IdScopedName(),
					cid_scope.id.IdScopedName(),
				)
				fmter(bufs["cxx_body"], "cxx_this->")
			}
		}

		for i, _ := range fct.Params {
			cid_arg := get_cxxgo_id(pkg, cxxtypes.IdByName(fct.Params[i].Type))
			cxx_type := cid_arg.id.IdScopedName()
			if strings.HasSuffix(cxx_type, "*") ||
				strings.HasSuffix(cxx_type, "* const") {
				// pointer to data member
				if strings.HasSuffix(cxx_type, ":*") ||
					strings.HasSuffix(cxx_type, ":* const") {
					fmter(bufs["cxx_head"],
						"  %s* cxx_arg_%d = (%s*)(c_arg_%d);\n",
						cxx_type, i, cxx_type, i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("*cxx_arg_%d", i))
				} else if cid_arg.is_string_like() {
					fmter(bufs["cxx_head"],
						"  %s cxx_arg_%d( ((_gostring_*)c_arg_%d)->p, ((_gostring_*)c_arg_%d)->n);\n",
						cxx_type[:len(cxx_type)-1], i, i, i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("cxx_arg_%d", i))
				} else if cid_arg.is_cstring_like() {
					fmter(bufs["cxx_head"],
						"  %s cxx_arg_%d = (%s)c_arg_%d;\n",
						cxx_type, i, cxx_type, i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("cxx_arg_%d", i))
				} else {
					fmter(bufs["cxx_head"],
						"  %s cxx_arg_%d = (%s)(c_arg_%d);\n",
						cxx_type, i, cxx_type, i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("cxx_arg_%d", i))
				}
			} else if strings.HasSuffix(cxx_type, "&") {
				if cid_arg.is_string_like() {
					fmter(bufs["cxx_head"],
						"  %s cxx_arg_%d( ((_gostring_*)c_arg_%d)->p, ((_gostring_*)c_arg_%d)->n);\n",
						cxx_type[:len(cxx_type)-1], i, i, i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("*cxx_arg_%d", i))
				} else {
					fmter(bufs["cxx_head"],
						"  %s* cxx_arg_%d = *(%s**)(&c_arg_%d);\n",
						cxx_type[:len(cxx_type)-1], i,
						cxx_type[:len(cxx_type)-1], i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("*cxx_arg_%d", i))
				}
			} else {
				if cid_arg.is_string_like() {
					fmter(bufs["cxx_head"],
						"  %s cxx_arg_%d( ((_gostring_*)c_arg_%d)->p, ((_gostring_*)c_arg_%d)->n);\n",
						cxx_type[:len(cxx_type)-1], i, i, i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("cxx_arg_%d", i))
				} else {
					fmter(bufs["cxx_head"],
						"  %s* cxx_arg_%d = (%s*)c_arg_%d;\n",
						cxx_type, i, cxx_type, i,
					)
					cxx_in = append(cxx_in, fmt.Sprintf("*cxx_arg_%d", i))
				}
			}
		}
		if !fct.IsDestructor() {
			fmter(bufs["cxx_body"],
				"%s(%s);\n",
				cid.id.IdName(),
				strings.Join(cxx_in, ", "),
			)
		} else {
			// noop.
		}
		fmter(bufs["cxx_tail"], "}\n")

		// commit buffers
		_, err = bufs["go_iface"].WriteTo(p.gen.Fd.Files["go"])
		if err != nil {
			return err
		}

		_, err = bufs["go_impl"].WriteTo(p.gen.Fd.Files["go"])
		if err != nil {
			return err
		}

		_, err = bufs["cxx_head"].WriteTo(p.gen.Fd.Files["cxx"])
		if err != nil {
			return err
		}

		_, err = bufs["cxx_body"].WriteTo(p.gen.Fd.Files["cxx"])
		if err != nil {
			return err
		}

		_, err = bufs["cxx_tail"].WriteTo(p.gen.Fd.Files["cxx"])
		if err != nil {
			return err
		}

		_, err = bufs["cgo_head"].WriteTo(p.gen.Fd.Files["hdr"])
		if err != nil {
			return err
		}
	}

	if needs_dispatch {
		go_ret := cgo_ovfct.fcts[0].f.Ret
		if go_ret != "" && go_ret != "void" {
			go_ret = "return"
		} else {
			go_ret = ""
		}
		cxx_protos := make([]string, 0, len(cgo_ovfct.fcts))
		for i, _ := range cgo_ovfct.fcts {
			cxx_protos = append(cxx_protos,
				cgo_ovfct.fcts[i].cxx_prototype())
		}
		fmter(bufs["go_impl"],
			"// dispatch for:\n//  %s\nfunc %s%s {\n",
			strings.Join(cxx_protos, "\n//  "),
			go_receiver,
			cgo_ovfct.go_prototype(),
		)
		fmter(bufs["go_impl"], "\targc := len(args)\n")
		fmter(bufs["go_impl"], "\tswitch argc {\n")
		for nargs, cfcts := range dispatch_table {
			fmter(bufs["go_impl"], "\tcase %d:\n", nargs)
			for _, cfct := range cfcts {
				fmter(bufs["go_impl"], "\t{// %s\n", cfct.cxx_prototype())
				go_casts := make([]string, 0, nargs)
				go_args := make([]string, 0, nargs)
				go_receiver := ""
				if cfct.f.IsMethod() {
					if cfct.f.IsConstructor() {
						go_receiver = ""
					} else {
						go_receiver = "p."
					}
				}
				for iarg, arg_t := range cfct.f.Params {
					arg_cid := get_cxxgo_id(pkg, cxxtypes.IdByName(arg_t.Type))
					fmter(bufs["go_impl"],
						"\targ_%d, ok_%d := args[%d].(%s)\n",
						iarg, iarg, iarg,
						arg_cid.goname,
					)
					go_casts = append(go_casts, fmt.Sprintf("ok_%d", iarg))
					go_args = append(go_args, fmt.Sprintf("arg_%d", iarg))
				}
				if_cond := "true"
				if len(cfct.f.Params) >= 1 {
					if_cond = strings.Join(go_casts, " && ")
				}
				fmter(bufs["go_impl"],
					"\tif %s {\n",
					if_cond,
				)

				if go_ret == "" && go_receiver != "" {
					fmter(bufs["go_impl"],
						"\t\t%s%s(%s)\n\t\treturn\n",
						go_receiver,
						cfct.goname,
						strings.Join(go_args, ", "),
					)
				} else {
					fmter(bufs["go_impl"],
						"\t\treturn %s%s(%s)\n",
						go_receiver,
						cfct.goname,
						strings.Join(go_args, ", "),
					)
				}
				fmter(bufs["go_impl"], "\t} // if-cast-ok\n")
				fmter(bufs["go_impl"], "\t}\n")
			}
		}
		fmter(bufs["go_impl"],
			"} // switch on argc\n\tpanic(\"unreachable\")\n}\n",
		)

		_, err = bufs["go_impl"].WriteTo(p.gen.Fd.Files["go"])
		if err != nil {
			return err
		}
	}

	cid.wrapped = true
	fmt.Printf(":: wrapping fct [%s]...[ok]\n", id.IdScopedName())
	return nil
}

//

// cxxgo_id wraps a cxxtypes.Id and adds a few convenient functions
type cxxgo_id struct {
	id       cxxtypes.Id
	uid      uint64 // unique numeric identifier
	selected bool   // whether this identifier has been selected for wrapping
	wrapped  bool   // whether this identifier has already been wrapped

	goname  string // the GoLang name for this C/C++ identifier
	cgoname string // the CGo name for this C/C++ identifier
}

func (cid *cxxgo_id) is_class_like() bool {
	id := cid.id
	for {
		switch iid := id.(type) {
		case *cxxtypes.ClassType, *cxxtypes.StructType:
			return true
		case *cxxtypes.CvrQualType:
			id = cxxtypes.IdByName(iid.Type).(cxxtypes.Id)
			continue
		case *cxxtypes.PtrType:
			id = iid.UnderlyingType().(cxxtypes.Id)
			continue
		case *cxxtypes.RefType:
			id = iid.UnderlyingType().(cxxtypes.Id)
			continue
		default:
			return false
		}
	}
	panic("unreachable")
}

func (cid *cxxgo_id) is_cstring_like() bool {
	n := cid.id.IdName()
	switch n {
	case "string", "std::string", "TString":
		return true
	case "char*":
		return true
	}

	c := true
	id := cid.id
	for c {
		switch iid := id.(type) {
		case *cxxtypes.CvrQualType:
			id = cxxtypes.IdByName(iid.Type)
		case *cxxtypes.TypedefType:
			id = cxxtypes.IdByName(iid.UnderlyingType().TypeName())
		default:
			c = false
			break
		}
	}
	n = id.IdName()
	if n == "const char*" || n == "char const*" || n == "char*" {
		return true
	}
	return false
}

func (cid *cxxgo_id) is_string_like() bool {
	n := cid.id.IdScopedName()
	switch n {
	case "string", "std::string", "TString", "char*":
		return true
	}
	c := true
	id := cid.id
	for c {
		switch iid := id.(type) {
		case *cxxtypes.CvrQualType:
			id = cxxtypes.IdByName(iid.Type)
		case *cxxtypes.TypedefType:
			id = cxxtypes.IdByName(iid.UnderlyingType().TypeName())
		default:
			c = false
			break
		}
	}
	n = id.IdScopedName()
	switch n {
	case "string", "std::string", "TString", "char*":
		return true
	}
	return false
}

func get_cxxgo_id(pkgname string, id cxxtypes.Id) *cxxgo_id {
	cid, ok := g_cxxgo_idmap[id]
	if ok {
		return cid
	}

	cid = &cxxgo_id{
		id:       id,
		uid:      get_iid(id),
		selected: false,
		wrapped:  false,
		goname:   gen_go_name_from_id(id),
		cgoname:  gen_cgo_name_from_id(pkgname, id),
	}
	g_cxxgo_idmap[id] = cid
	return cid
}

// cxxgo_overload_fct_set_t models all the needed various overloads stemming 
// from a C++ OverloadFunctionSet (handling different signatures and default
// parameters)
type cxxgo_overload_fct_set_t struct {
	cid    *cxxgo_id
	pkg    string
	ovfct  *cxxtypes.OverloadFunctionSet
	fcts   []cxxgo_function
	goname string
}

func (p *plugin) new_cxxgo_ovfct(ovfct *cxxtypes.OverloadFunctionSet) cxxgo_overload_fct_set_t {
	pkg := p.gen.Fd.Package
	o := cxxgo_overload_fct_set_t{
		cid:    get_cxxgo_id(pkg, ovfct),
		pkg:    pkg,
		ovfct:  ovfct,
		fcts:   make([]cxxgo_function, 0, len(ovfct.Fcts)),
		goname: gen_go_name_from_id(ovfct),
	}
	needs_dispatch := fctset_need_dispatch(ovfct)
	for ifct, _ := range ovfct.Fcts {
		fct := ovfct.Function(ifct)
		if fct.IsPrivate() {
			// discard from cxxgo-overload set
			// FIXME: what about protected methods which are meant
			//        to be implemented by, say, derived classes ?
			continue
		}
		if fct.IsMethod() && fct.IsConstructor() {
			// discard if class is abstract...
			scope_id, ok := cxxtypes.IdByName(fct.BaseId.Scope).(cxxtypes.Type)
			if ok && cxxtypes.IsAbstractType(scope_id) {
				continue
			}
		}
		nargs := fct.NumParam()
		ndargs := fct.NumDefaultParam()
		imax := ndargs + 1
		for i := 0; i < imax; i++ {
			idx := len(o.fcts)
			cfct := cxxgo_function{
				f:     *fct, // copy
				pkg:   pkg,
				idx:   idx,
				ovfct: &o,
			}
			cfct.goname = gen_go_name_from_id(ovfct)
			cfct.cgoname = gen_cgo_name_from_id(pkg, ovfct)
			if needs_dispatch {
				cfct.goname = cfct.goname + fmt.Sprintf("__GOCXX_%d", idx)
				cfct.cgoname = cfct.cgoname + fmt.Sprintf("_%d", idx)
			}
			cfct.f.Params = cfct.f.Params[:nargs-ndargs+i]
			o.fcts = append(o.fcts, cfct)
		}
	}
	return o
}

func (f *cxxgo_overload_fct_set_t) cxx_prototype() string {
	return f.fcts[0].cxx_prototype()
}

func (f *cxxgo_overload_fct_set_t) go_prototype() string {

	fct := f.ovfct.Fcts[0]
	s := []string{f.goname, "("}

	if fctset_need_dispatch(f.ovfct) {
		s = append(s, "args ...interface{}")
	} else {
		if fct.IsDestructor() {
			scope_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.BaseId.Scope))
			s = append(s,
				"arg",
				" ", scope_id.goname)
		} else {
			for i, _ := range fct.Params {
				arg_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.Param(i).Type))
				s = append(s,
					fmt.Sprintf("arg_%d ", i),
					arg_id.goname)
				if i < len(fct.Params)-1 {
					s = append(s, ", ")
				}
			}
		}
	}
	s = append(s, ")")
	if fct.Ret != "" && fct.Ret != "void" {
		ret_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.Ret))
		s = append(s, " ", ret_id.goname)
	} else if fct.IsConstructor() || fct.IsCopyConstructor() {
		scope_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.BaseId.Scope))
		s = append(s, " ", scope_id.goname)
	}
	return strings.Join(s, "")
}

type cxxgo_function struct {
	f       cxxtypes.Function
	pkg     string
	idx     int // index in the overload fct set
	ovfct   *cxxgo_overload_fct_set_t
	goname  string
	cgoname string
}

func (f *cxxgo_function) go_prototype() string {

	fct := f.f
	s := []string{f.goname, "("}

	if fct.IsDestructor() {
		scope_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.BaseId.Scope))
		s = append(s,
			"arg",
			" ", scope_id.goname)
	} else {
		for i, _ := range fct.Params {
			arg_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.Param(i).Type))
			s = append(s,
				fmt.Sprintf("arg_%d ", i),
				arg_id.goname)
			if i < len(fct.Params)-1 {
				s = append(s, ", ")
			}
		}
	}
	s = append(s, ")")
	if fct.Ret != "" && fct.Ret != "void" {
		ret_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.Ret))
		s = append(s, " ", ret_id.goname)
	} else if fct.IsConstructor() || fct.IsCopyConstructor() {
		scope_id := get_cxxgo_id(f.pkg, cxxtypes.IdByName(fct.BaseId.Scope))
		s = append(s, " ", scope_id.goname)
	}
	return strings.Join(s, "")
}

func (f *cxxgo_function) cgo_prototype() string {
	fct := f.f

	s := []string{}
	s = append(s,
		"void ",
		f.cgoname,
		"(",
	)

	if fct.IsMethod() {
		s = append(s, "void *c_this, ")
	}

	if fct.Ret != "" && fct.Ret != "void" {
		s = append(s, "void *c_ret")
	} else {
		s = append(s, "void */*c_ret*/")
	}

	for i, _ := range fct.Params {
		s = append(s, fmt.Sprintf(", void *c_arg_%d", i))
	}

	s = append(s, ")")

	//fct := ovfct.Function(0)
	return strings.Join(s, "")
}

func (f *cxxgo_function) cxx_prototype() string {
	fct := f.f

	s := []string{}
	if fct.IsInline() {
		s = append(s, "inline ")
	}
	if fct.IsStatic() {
		s = append(s, "static ")
	}
	//fixme: add fct qualifiers: const|static|inline
	if !fct.IsConstructor() && !fct.IsDestructor() {
		s = append(s, fct.ReturnType().TypeName(), " ")
	}
	s = append(s, fct.IdScopedName(), "(")
	if len(fct.Params) > 0 {
		for i, _ := range fct.Params {
			s = append(s,
				strings.TrimSpace(fct.Param(i).Type),
				" ",
				strings.TrimSpace(fct.Param(i).Name))
			if i < len(fct.Params)-1 {
				s = append(s, ", ")
			}
		}
	} else {
		// fixme: we should rather test if C XOR C++...
		if fct.IsMethod() {
			//nothing
		} else {
			s = append(s, "void")
		}
	}
	if fct.IsVariadic() {
		s = append(s, "...")
	}
	s = append(s, ") ")
	if fct.IsConst() {
		s = append(s, "const ")
	}
	return strings.TrimSpace(strings.Join(s, ""))
}

// utils ------------------------

func fmter(buf *bytes.Buffer, format string, args ...interface{}) (int, error) {
	o := fmt.Sprintf(format, args...)
	return buf.WriteString(o)
}

func new_bufmap(keys ...string) bufmap_t {
	bufs := make(bufmap_t, len(keys))
	for _, k := range keys {
		bufs[k] = bytes.NewBufferString("")
	}
	return bufs
}

// cxx2go_typename converts a C++ type string into its go equivalent
func cxx2go_typename(t string) string {
	if o, ok := _cxx2go_typemap[t]; ok {
		return o
	}
	o := fmt.Sprintf("_go_unknown_%s", t)
	_cxx2go_typemap[t] = o
	return o
}

// cxx2cgo_typemap converts a C++ type string into its cgo equivalent
func cxx2cgo_typename(t string) string {
	if o, ok := _cxx2cgo_typemap[t]; ok {
		return o
	}
	o := fmt.Sprintf("_go_unknown_%s", t)
	_cxx2cgo_typemap[t] = o
	return o
}

func gen_go_name_from_id(id cxxtypes.Id) string {
	n := id.IdScopedName()

	// special cases
	if _, ok := _cxx2go_typemap[n]; ok {
		return cxx2go_typename(n)
	}

	switch id := id.(type) {

	case *cxxtypes.Function:
		cls_id := cxxtypes.IdByName(id.BaseId.Scope)
		cls_name := gen_go_name_from_id(cls_id)
		if id.IsDestructor() {
			n = "Delete" + cls_name //strings.Title(cls_id.IdName())[1:]
		} else if id.IsOperator() {
			switch id.IdName() {
			case "operator+":
				n = "Add_op"
			case "operator+=":
				n = "IAdd_op"
			case "operator-":
				n = "Sub_op"
			case "operator-=":
				n = "ISub_op"
			case "operator*":
				n = "Mul_op"
			case "operator*=":
				n = "IMul_op"
			case "operator/":
				n = "Div_op"
			case "operator/=":
				n = "IDiv_op"
			case "operator=":
				n = "Assign_op"
			case "operator==":
				n = "Eq_op"
			case "operator()":
				n = "Call_op"
			case "operator->":
				n = "Arrow_op"
			case "operator[]":
				n = "At_op"
			case "operator++":
				n = "Inc_op"
			case "operator--":
				n = "Dec_op"
			default:
				panic(fmt.Sprintf("unknown operator [%s] [scoped=%s]",
					id.IdName(),
					id.IdScopedName()))
			}
		} else if id.IsConverter() {

			n = "CnvTo_" + id.IdName()[len("operator "):]
		} else if id.IsConstructor() || id.IsCopyConstructor() {
			n = "New" + cls_name //strings.Title(iid.IdName())
		} else {
			n = strings.Title(id.IdName())
		}

	case *cxxtypes.OverloadFunctionSet:
		iid := id.Function(0)
		n = gen_go_name_from_id(iid)
		/*
			cls_id := cxxtypes.IdByName(iid.BaseId.Scope)
			cls_name := gen_go_name_from_id(cls_id)
			if iid.IsDestructor() {
				n = "Delete" + cls_name // strings.Title(iid.IdName())[1:]
			} else if iid.IsOperator() {
				switch iid.IdName() {
				case "operator+":
					n = "Add_op"
				case "operator+=":
					n = "IAdd_op"
				case "operator-":
					n = "Sub_op"
				case "operator-=":
					n = "ISub_op"
				case "operator*":
					n = "Mul_op"
				case "operator*=":
					n = "IMul_op"
				case "operator/":
					n = "Div_op"
				case "operator/=":
					n = "IDiv_op"
				case "operator=":
					n = "Assign_op"
				case "operator==":
					n = "Eq_op"
				case "operator()":
					n = "Call_op"
				case "operator->":
					n = "Arrow_op"
				case "operator[]":
					n = "At_op"
				case "operator++":
					n = "Inc_op"
				case "operator--":
					n = "Dec_op"
				default:
					panic(fmt.Sprintf("unknown operator [%s] [scoped=%s]",
						iid.IdName(),
						iid.IdScopedName()))
				}
			} else if iid.IsConverter() {

				n = "CnvTo_" + iid.IdName()[len("operator "):]
			} else if iid.IsConstructor() || iid.IsCopyConstructor() {
				n = "New" + cls_name //strings.Title(iid.IdName())
			} else {
				n = strings.Title(iid.IdName())
			}
		*/
	case *cxxtypes.Member:
		iid := cxxtypes.IdByName(id.Type)
		return gen_go_name_from_id(iid)

	case *cxxtypes.ClassType:
		n = strings.Title(n)

	case *cxxtypes.PtrType:
		ptr := "*"
		ptee_id := id.UnderlyingType().(cxxtypes.Id)
		switch ptee_id.(type) {
		case *cxxtypes.ClassType:
			// for a class, the go-type is an interface...
			// having a pointer to an interface isn't really go-ish
			ptr = ""
		}
		return ptr + gen_go_name_from_id(id.UnderlyingType().(cxxtypes.Id))

	case *cxxtypes.RefType:
		return gen_go_name_from_id(id.UnderlyingType().(cxxtypes.Id))

	case *cxxtypes.CvrQualType:
		return gen_go_name_from_id(cxxtypes.IdByName(id.Type))
	}

	// sanitize
	o := g_cxxgo_trans.Replace(n)

	if _, ok := _cxx2go_typemap[o]; ok {
		return cxx2go_typename(o)
	}
	return o
}

func gen_go_name(cxxname string) string {
	o := g_cxxgo_trans.Replace(cxxname)
	if _, ok := _cxx2go_typemap[o]; ok {
		return cxx2go_typename(o)
	}
	return o
}

func gen_cgo_name_from_id(pkgname string, id cxxtypes.Id) string {
	n := id.IdScopedName()

	// special cases
	if _, ok := _cxx2cgo_typemap[n]; ok {
		return cxx2cgo_typename(n) // FIXME
	}

	switch id := id.(type) {

	case *cxxtypes.FundamentalType:
		n = id.IdName()

	case *cxxtypes.Function:
		n = fmt.Sprintf("_gocxx_fct_%s_%s", pkgname, get_iid_str(id))

	case *cxxtypes.OverloadFunctionSet:
		n = fmt.Sprintf("_gocxx_fct_%s_%s", pkgname, get_iid_str(id))

	case *cxxtypes.Member:
		iid := cxxtypes.IdByName(id.Type)
		n = gen_cgo_name_from_id(pkgname, iid)

	case *cxxtypes.ClassType:
		n = "_gocxx_voidptr"

	case *cxxtypes.PtrType:
		n = "_gocxx_voidptr"
		//return "*" + gen_go_name_from_id(id.UnderlyingType().(cxxtypes.Id))

	case *cxxtypes.RefType:
		n = "_gocxx_voidptr"
		//return gen_go_name_from_id(id.UnderlyingType().(cxxtypes.Id))

	case *cxxtypes.CvrQualType:
		n = gen_cgo_name_from_id(pkgname, cxxtypes.IdByName(id.Type))

	default:
		err := fmt.Errorf("unhandled identifier [%v]", id)
		panic(err)
	}
	return n
}

func (p *plugin) mbr_filter(mbr *cxxtypes.Member) bool {
	if mbr == nil {
		return false
	}

	// filter out any non public member
	if mbr.IsPrivate() || mbr.IsProtected() {
		return false
	}

	// filter out any anonymous member
	if n := mbr.Name; is_anon(n) {
		return false
	}

	// filter any copy constructor with a private copy constructor in
	// any base class
	// TODO

	// filter any constructor for pure abstract classes
	// TODO

	// filter methods taking non-public args
	// TODO

	// filter using the exclusion list in the selection file
	// TODO

	return true
}

// get_container_id returns the container-type and stl-class of id
// (if id is actually a container.)
func get_container_id(id cxxtypes.Id) (string, string) {
	n := id.IdScopedName()
	if strings.HasSuffix(n, "iterator") ||
		strings.HasPrefix(n, "__gnu_cxx::__normal_iterator") {
		return "NOCONTAINER", ""
	} else if strings.HasSuffix(n, "iterator<true>") {
		return "NOCONTAINER", ""
	} else if strings.HasSuffix(n, "iterator<false>") {
		return "NOCONTAINER", ""
	} else if strings.HasPrefix(n, "std::deque") {
		return "DEQUE", "list"
	} else if strings.HasPrefix(n, "std::list") {
		return "LIST", "list"
	} else if strings.HasPrefix(n, "std::map") {
		return "MAP", "map"
	} else if strings.HasPrefix(n, "std::multimap") {
		return "MULTIMAP", "map"
	} else if strings.HasPrefix(n, "std::unordered_map") {
		return "HASHMAP", "map"
	} else if strings.HasPrefix(n, "std::unordered_multimap") {
		return "HASHMULTIMAP", "map"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_map") {
		return "HASHMAP", "map"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_multimap") {
		return "HASHMULTIMAP", "map"
	} else if strings.HasPrefix(n, "stdext::hash_map") {
		return "HASHMAP", "map"
	} else if strings.HasPrefix(n, "stdext::hash_multimap") {
		return "HASHMULTIMAP", "map"
	} else if strings.HasPrefix(n, "std::queue") {
		return "QUEUE", "queue"
	} else if strings.HasPrefix(n, "std::set") {
		return "SET", "set"
	} else if strings.HasPrefix(n, "std::multiset") {
		return "MULTISET", "set"
	} else if strings.HasPrefix(n, "std::unordered_set") {
		return "HASHSET", "set"
	} else if strings.HasPrefix(n, "std::unordered_multiset") {
		return "HASHMULTISET", "set"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_set") {
		return "HASHSET", "set"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_multiset") {
		return "HASHMULTISET", "set"
	} else if strings.HasPrefix(n, "stdext::hash_set") {
		return "HASHSET", "set"
	} else if strings.HasPrefix(n, "stdext::hash_multiset") {
		return "HASHMULTISET", "set"
	} else if strings.HasPrefix(n, "std::stack") {
		return "STACK", "stack"
	} else if strings.HasPrefix(n, "std::vector") {
		return "VECTOR", "vector"
	} else if strings.HasPrefix(n, "std::bitset") {
		return "BITSET", "bitset"
	} else {
		return "NOCONTAINER", ""
	}
	panic("unreachable")
}

// is_anon returns true if the given typename looks like an anonymous one
func is_anon(n string) bool {
	// ellipsis would screw us up...
	nn := strings.Replace(n, "...", "", -1)
	if strings.IndexAny(nn, ".$") != -1 {
		return true
	}
	return false
}

func str_is_in_slice(s string, slice []string) bool {
	for _, ss := range slice {
		if s == ss {
			return true
		}
	}
	return false
}

func fctset_need_dispatch(ovfct *cxxtypes.OverloadFunctionSet) bool {
	noverloads := 0
	for i, _ := range ovfct.Fcts {
		f := ovfct.Function(i)
		if f.IsPrivate() {
			continue
		}
		if f.IsMethod() && f.IsConstructor() {
			// discard if class is abstract...
			scope_id, ok := cxxtypes.IdByName(f.BaseId.Scope).(cxxtypes.Type)
			if ok && cxxtypes.IsAbstractType(scope_id) {
				continue
			}
		}
		noverloads += 1 + f.NumDefaultParam()
	}
	return noverloads > 1
}

func get_dependent_ids(in_ids []string, id cxxtypes.Id) (dep_ids []string) {
	return get_dependent_ids_rec(in_ids, id, true)
}

func get_dependent_ids_rec(in_ids []string, id cxxtypes.Id, rec bool) (dep_ids []string) {
	dep_ids = make([]string, 0, len(in_ids))
	for _, dep_id := range in_ids {
		if str_is_in_slice(dep_id, dep_ids) {
			continue
		}
		dep_ids = append(dep_ids, dep_id)
	}
	switch id := id.(type) {
	// case *cxxtypes.ClassType:
	// 	err := p.wrapClass(id)
	// 	if err != nil {
	// 		return err
	// 	}

	// case *cxxtypes.StructType:
	// 	err := p.wrapStruct(id)
	// 	if err != nil {
	// 		return err
	// 	}

	case *cxxtypes.ClassType:
		for _, mbr := range id.Members {
			mbr_id := cxxtypes.IdByName(mbr.Name)
			dep_ids = append(dep_ids,
				get_dependent_ids_rec(dep_ids, mbr_id, false)...)
		}

	case *cxxtypes.OverloadFunctionSet:
		for _, fct := range id.Fcts {
			for i, _ := range fct.Params {
				arg_id := cxxtypes.IdByName(fct.Params[i].Type).IdScopedName()
				dep_ids = append(dep_ids, arg_id)
			}
			if fct.Ret != "" && fct.Ret != "void" {
				ret_id := cxxtypes.IdByName(fct.Ret).IdScopedName()
				dep_ids = append(dep_ids, ret_id)
			}
		}
	}
	// recurse...
	if rec {
		for _, dep_id := range dep_ids {
			if str_is_in_slice(dep_id, in_ids) {
				continue
			}
			iid := cxxtypes.IdByName(dep_id)
			dep_ids = append(dep_ids, get_dependent_ids_rec(dep_ids, iid, false)...)
		}
	}
	return dep_ids
}

var g_id_counter uint64 = 0

func get_iid(id cxxtypes.Id) uint64 {
	c, ok := g_idmap[id]
	if ok {
		return c
	}
	c = g_id_counter
	g_idmap[id] = c
	g_id_counter += 1
	return c
}

func get_iid_str(id cxxtypes.Id) string {
	return fmt.Sprintf("%v", get_iid(id))
}

// globals ----------------------

// g_idmap is a global map of Id to some integer to uniquely identify
// identifiers
var g_idmap idmap_t

type cxxgo_idmap_t map[cxxtypes.Id]*cxxgo_id

// g_cxxgo_idmap is a global map of all cxxgo_ids
var g_cxxgo_idmap cxxgo_idmap_t

var g_strtrans *strings.Replacer = strings.NewReplacer(
	"<", "_",
	">", "_",
	"&", "r",
	"*", "p",
	",", "_",
	":", "_",
	" ", "s",
	"(", "_",
	")", "_",
	".", "_",
	"$", "d",
	"-", "m",
	"[", "_",
	"]", "_",
)

var g_cxxgo_trans *strings.Replacer = strings.NewReplacer(
	"<", "_Sl_",
	">", "_Sg_",
	",", "_Sc_",
	" ", "_",
	"-", "m",
	"::", "_",
	"*", "_Sp_",
)

var _go_hdr string = `
package %s

// #include <stdlib.h>
// #include <string.h>
// #include "%s"
// #cgo LDFLAGS: -l%s -l%s
import "C"
import "unsafe"

// dummy function which uses unsafe
func _gocxx_free_ptr(ptr unsafe.Pointer) {
 C.free(ptr)
}
`

var _cxx_hdr string = `
// C includes
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// C++ includes
#include <string>
#include <vector>

#include "%s"

#include "%s"

#ifdef __cplusplus
extern "C" {
#endif

// helpers for CGo runtime

typedef struct { char *p; int n; } _gostring_;
typedef struct { void* array; unsigned int len; unsigned int cap; } _goslice_;


extern void crosscall2(void (*fn)(void *, int), void *, int);
extern void _cgo_allocate(void *, int);
extern void _cgo_panic(void *, int);

static void *_gocxx_goallocate(size_t len) {
  struct {
    size_t len;
    void *ret;
  } a;
  a.len = len;
  crosscall2(_cgo_allocate, &a, (int) sizeof a);
  return a.ret;
}

static void _gocxx_gopanic(const char *p) {
  struct {
    const char *p;
  } a;
  a.p = p;
  crosscall2(_cgo_panic, &a, (int) sizeof a);
}

static _gostring_ _gocxx_makegostring(const char *p, size_t l) {
  _gostring_ ret;
  ret.p = (char*)_gocxx_goallocate(l + 1);
  memcpy(ret.p, p, l);
  ret.n = l;
  return ret;
}

static char* _gocxx_makecstring(const _gostring_ *s) {
  char* cstr = (char*)malloc((s->n+1) * sizeof(char));
  memcpy(cstr, s->p, s->n);
  cstr[s->n] = '\0';
  return cstr;
}

#define GOCXX_contract_assert(expr, msg) \
  if (!(expr)) { _gocxx_gopanic(msg); } else

#define GOCXX_exception(code, msg) _gocxx_gopanic(msg)

`

var _hdr_hdr string = `
#ifndef _GOCXXDICT_%s_H
#define _GOCXXDICT_%s_H 1

#ifdef __cplusplus
extern "C" {
#endif

typedef void* _gocxx_voidptr;
`

var _go_footer string = `
// EOF %s
`

var _cxx_footer string = `
#ifdef __cplusplus
} /* extern "C" */
#endif
`

var _hdr_footer string = `
#ifdef __cplusplus
} /* extern "C" */
#endif

#endif /* ! %s_H */
`

var _cxx2cgo_typemap = map[string]string{
	"void":     "void",
	"uint64_t": "uint64_t",
	"uint32_t": "uint32_t",
	"uint16_t": "uint16_t",
	"uint8_t":  "uint8_t",
	"uint_t":   "uint_t",
	"int64_t":  "int64_t",
	"int32_t":  "int32_t",
	"int16_t":  "int16_t",
	"int8_t":   "int8_t",
}

var _cxx2go_typemap = map[string]string{
	"void":     "",
	"uint64_t": "uint64",
	"uint32_t": "uint32",
	"uint16_t": "uint16",
	"uint8_t":  "uint8",
	"uint_t":   "uint",
	"int64_t":  "int64",
	"int32_t":  "int32",
	"int16_t":  "int16",
	"int8_t":   "int8",

	"bool":           "bool",
	"char":           "byte",
	"signed char":    "int8",
	"unsigned char":  "byte",
	"short":          "int16",
	"unsigned short": "uint16",
	"int":            "int",
	"unsigned int":   "uint",

	"char*":       "string",
	"const char*": "string",
	"char const*": "string",

	// FIXME: 32/64 platforms... (and cross-compilation)
	//"long":           "int32",
	//"unsigned long":  "uint32",
	"long":          "int64",
	"unsigned long": "uint64",

	"long long":          "int64",
	"unsigned long long": "uint64",

	"float":  "float32",
	"double": "float64",

	"float complex":  "complex64",
	"double complex": "complex128",

	// FIXME: 32/64 platforms
	//"size_t": "int",
	"size_t":      "int64",
	"std::size_t": "int64",

	// stl
	"std::string": "string",

	"std::ptrdiff_t": "int64",   //FIXME !!
	"std::ostream":   "uintptr", //FIXME !!

	// ROOT types
	"Char_t":   "byte",
	"UChar_t":  "byte",
	"Short_t":  "int16",
	"UShort_t": "uint16",
	"Int_t":    "int",
	"UInt_t":   "uint",

	"Seek_t":     "int",
	"Long_t":     "int64",
	"ULong_t":    "uint64",
	"Float_t":    "float32",
	"Float16_t":  "float32", //FIXME
	"Double_t":   "float64",
	"Double32_t": "float64",

	"Bool_t":    "bool",
	"Text_t":    "byte",
	"Byte_t":    "byte",
	"Version_t": "int16",
	"Option_t":  "byte",
	"Ssiz_t":    "int",
	"Real_t":    "float32",
	"Long64_t":  "int64",
	"ULong64_t": "uint64",
	"Axis_t":    "float64",
	"Stat_t":    "float64",
	"Font_t":    "int16",
	"Style_t":   "int16",
	"Marker_t":  "int16",
	"Width_t":   "int16",
	"Color_t":   "int16",
	"SCoord_t":  "int16",
	"Coord_t":   "float64",
	"Angle_t":   "float32",
	"Size_t":    "float32",
}

func init() {
	wrapper.RegisterPlugin(&plugin{})
	g_idmap = make(idmap_t)
	g_cxxgo_idmap = make(cxxgo_idmap_t)
}

// test interfaces...

var _ wrapper.Plugin = (*plugin)(nil)

// EOF
