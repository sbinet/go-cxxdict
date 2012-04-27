// Package gccxml reads an XML file produced by GCC_XML and fills in the cxxtypes' registry.
package gccxml

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
)

// globals

var g_dbg bool = true

// map of builtin-name to cxxtypes.TypeKind
var g_n2tk map[string]cxxtypes.TypeKind

// id of the global namespace
var g_globalns_id string = ""

// all ids
var g_ids idDB = make(idDB, 128)

// a cache of already processed ids (and their fully qualified name)
var g_processed_ids map[string]string

// a cache of ids being processed
var g_processing_ids map[string]bool = make(map[string]bool)

type gidname struct {
	id string
	cfg gtnCfg
}

// a cache of id->name filled by genTypeName
var g_ids_name map[gidname]string = make(map[gidname]string)

// LoadTypes reads an XML file produced by GCC_XML and fills the cxxtypes' registry accordingly.
func LoadTypes(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	root := &xmlTree{}
	err = xml.Unmarshal(data, root)
	if err != nil {
		return err
	}

	// walk over the xmlTree to fill in the db of ids.
	v := func(node i_id) bool {
		//println("-->",node.id())
		g_ids[node.id()] = node
		return true
	}
	walk(inspector(v), root)
	fmt.Printf("ids: %d\n", len(g_ids))
	g_processed_ids = make(map[string]string, len(g_ids))

	root.fixup()

	fmt.Printf("== gccxml data ==\n")
	root.printStats()

	// generate the scopes in cxxtypes' "ast"
	err = root.gencxxscopes()
	if err != nil {
		return err
	}

	// generate cxxtypes
	err = root.gencxxtypes()
	if err != nil {
		return err
	}

	{
		{
			names := []string{
				"const char*",
				"char const*",
				"char*const",
				"int",
				"Foo",
				"IFoo",
				"Math::do_hello",
				"TT::foo_t",
				"TT::baz_t",
				"MyEnum",
				"Enum0",
				"LongStr_t",
				"Ssiz_t",
				"Func_t",
				"std::vector<Foo>::push_back",
			}
			fmt.Printf("++++++++++++++++++++++++++\n")
			for _, n := range names {
				t := cxxtypes.TypeByName(n)
				if t == nil {
					fmt.Printf("::could not find type [%s]\n", n)
				} else {
					fmt.Printf("[%s]: %v\n", n, t)
				}
			}
			fmt.Printf("++++++++++++++++++++++++++\n")
		}
		{
			names := []string{
				"Foo",
				"Alg",
				"WithPrivateBase",
				"LongStr_t",
				//"std::vector<Foo>",
			}
			for _,n := range names {
				t := cxxtypes.TypeByName(n)
				if t == nil {
					fmt.Printf("could not inspect type [%s]\n", n)
					continue
				}
				fmt.Printf(":: inspecting [%s]...\n", n)
				switch tt := t.(type) {
				case *cxxtypes.ClassType:
					fmt.Printf(" #bases: %d\n", tt.NumBase())
					for i := 0; i < tt.NumBase(); i++ {
						b :=  tt.Base(i)
						fmt.Printf(" %d: %v\n", i, b)
					}
					fmt.Printf(" #mbrs: %d\n", tt.NumMember())
					for i := 0; i < tt.NumMember(); i++ {
						m :=  tt.Member(i)
						fmt.Printf(" %d: %v\n", i, m)
					}
				case *cxxtypes.StructType:
					fmt.Printf(" #bases: %d\n", tt.NumBase())
					for i := 0; i < tt.NumBase(); i++ {
						b :=  tt.Base(i)
						fmt.Printf(" %d: %v\n", i, b)
					}
					fmt.Printf(" #mbrs: %d\n", tt.NumMember())
					for i := 0; i < tt.NumMember(); i++ {
						m :=  tt.Member(i)
						fmt.Printf(" %d: %v\n", i, m)
					}
				}
			}
		}
		names := cxxtypes.TypeNames()
		fmt.Printf("== distilled [%d] types.\n", len(names))
		// for _,n := range names {
		// 	t := cxxtypes.TypeByName(n)
		// 	fmt.Printf("[%s]: %v\n", n, t)
		// }

	}
	return err
}

func init() {
	g_n2tk = map[string]cxxtypes.TypeKind{
		"void":           cxxtypes.TK_Void,
		"bool":           cxxtypes.TK_Bool,
		"char":           gccxml_get_char_type(),
		"signed char":    cxxtypes.TK_SChar,
		"unsigned char":  cxxtypes.TK_UChar,
		"wchar_t":        cxxtypes.TK_WChar,
		"short":          cxxtypes.TK_Short,
		"unsigned short": cxxtypes.TK_UShort,
		"int":            cxxtypes.TK_Int,
		"unsigned int":   cxxtypes.TK_UInt,

		"long":               cxxtypes.TK_Long,
		"unsigned long":      cxxtypes.TK_ULong,
		"long long":          cxxtypes.TK_LongLong,
		"unsigned long long": cxxtypes.TK_ULongLong,

		"float":       cxxtypes.TK_Float,
		"double":      cxxtypes.TK_Double,
		"long double": cxxtypes.TK_LongDouble,

		"float complex":       cxxtypes.TK_Complex,
		"double complex":      cxxtypes.TK_Complex,
		"long double complex": cxxtypes.TK_Complex,
	}
}

// EOF
