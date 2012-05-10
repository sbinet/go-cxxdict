package cxxgo

import (
	"fmt"
	"path"
	"os"

	"bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
	"bitbucket.org/binet/go-cxxdict/pkg/wrapper"
)

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
		"Alg*",
		"With*Base*",
		
		"*Enum*",
		"cblas*",
		"*CBLAS_*",
	}
	p.ids = []string{}

	fmt.Printf("cxxgo.Init: args=%v\n", g.Args)
	return nil
}

func (p *plugin) Generate(fd *wrapper.FileDescriptor) error {
	fmt.Printf("cxxgo.Generate...\n")

	// loop over identifiers and filter them out
	for _,n := range cxxtypes.IdNames() {
		selected := false
		for _,sel := range p.sel {
			matched, err := path.Match(sel, n)
			if err != path.ErrBadPattern && matched {
				selected = true
				break
			}
		}
		if selected {
			p.ids = append(p.ids, n)
		}
	}
	fmt.Printf("selected ids: %v\n", p.ids)
	if len(p.ids) <= 0 {
		fmt.Printf("nothing to wrap\n")
		return nil
	}

	fd_cxx, err := os.Create(fd.Name+"_"+p.Name()+".cxx")
	if err != nil {
		return err
	}
	defer fd_cxx.Close()

	fd_hdr, err := os.Create(fd.Name+"_"+p.Name()+".h")
	if err != nil {
		return err
	}
	defer fd_hdr.Close()

	fd_go, err := os.Create(fd.Name+"_"+p.Name()+".go")
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

	for _,n := range p.ids {
		id := cxxtypes.IdByName(n)
		switch id := id.(type) {
		case *cxxtypes.ClassType:
			err := p.wrapClass(id)
			if err != nil {
				return err
			}

		case *cxxtypes.StructType:
			err := p.wrapStruct(id)
			if err != nil {
				return err
			}

		case *cxxtypes.OverloadFunctionSet:
			err := p.wrapFunction(id)
			if err != nil {
				return err
			}

		case *cxxtypes.Namespace:
			err := p.wrapNamespace(id)
			if err != nil {
				return err
			}

		case *cxxtypes.Member:
			err := p.wrapMember(id)
			if err != nil {
				return err
			}

		case *cxxtypes.TypedefType:
			err := p.wrapTypedef(id)
			if err != nil {
				return err
			}

		case *cxxtypes.RefType:
			err := p.wrapRefType(id)
			if err != nil {
				return err
			}

		case *cxxtypes.PtrType:
			err := p.wrapPtrType(id)
			if err != nil {
				return err
			}

		case *cxxtypes.CvrQualType:
			err := p.wrapCvrQualType(id)
			if err != nil {
				return err
			}

		case *cxxtypes.EnumType:
			err := p.wrapEnum(id)
			if err != nil {
				return err
			}

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

func (p *plugin) wrapClass(id *cxxtypes.ClassType) error {
	fmt.Printf(":: wrapping class [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping class [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapStruct(id *cxxtypes.StructType) error {
	fmt.Printf(":: wrapping struct [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping struct [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapFunction(id *cxxtypes.OverloadFunctionSet) error {
	fmt.Printf(":: wrapping fct [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping fct [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapNamespace(id *cxxtypes.Namespace) error {
	fmt.Printf(":: wrapping namespace [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping namespace [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapEnum(id *cxxtypes.EnumType) error {
	fmt.Printf(":: wrapping enum [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping enum [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapMember(id *cxxtypes.Member) error {
	fmt.Printf(":: wrapping member [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping member [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapTypedef(id *cxxtypes.TypedefType) error {
	fmt.Printf(":: wrapping typedef [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping typedef [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapRefType(id *cxxtypes.RefType) error {
	fmt.Printf(":: wrapping ref-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping ref-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapPtrType(id *cxxtypes.PtrType) error {
	fmt.Printf(":: wrapping ptr-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping ptr-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapCvrQualType(id *cxxtypes.CvrQualType) error {
	fmt.Printf(":: wrapping cvr-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping cvr-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}


// test interfaces...

var _ wrapper.Plugin = (*plugin)(nil)

// globals ----------------------

var _go_hdr string = `
package %s

// #include <stdlib.h>
// #include <string.h>
// #include "%s"
// #cgo LDFLAGS: -l%s
import "C"
import "unsafe"
 `

var _cxx_hdr string = `
// C includes
#include <stdlib.h>
#include <string.h>

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
`

var _go_footer string = `
// EOF %s
`

var _cxx_footer string = `
#ifdef __cpluspluc
} /* extern "C" */
#endif
`

var _hdr_footer string = `
#ifdef __cplusplus
} /* extern "C" */
#endif

#endif /* ! %s_H */
`

func init() {
	wrapper.RegisterPlugin(&plugin{})
}

// EOF
