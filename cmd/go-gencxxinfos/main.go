// go-gencxxinfos creates the C++ types registry from some source (gccxml, 
// clang) and saves it under some form.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sbinet/go-cxxdict/pkg/cxxtypes"
	_ "github.com/sbinet/go-cxxdict/pkg/cxxtypes/gccxml"
)

const dbg = 1

var fname *string = flag.String("fname", "", "gccxml file or clang file from which to distill identifiers")
var oname *string = flag.String("o", "ids.db", "output file in which to store cxxinfos")

func main() {
	fmt.Printf("== go-gencxxinfos ==\n")
	flag.Parse()

	f, err := os.Open(*fname)
	if err != nil {
		fmt.Printf("**err** %v\n", err)
		os.Exit(1)
	}

	err = cxxtypes.LoadIds("gccxml", f)
	if err != nil {
		fmt.Printf("**err** %v\n", err)
		os.Exit(1)
	}

	if dbg == 1 {
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
				id := cxxtypes.IdByName(n)
				if id == nil {
					fmt.Printf("::could not find identifier [%s]\n", n)
				} else {
					fmt.Printf("[%s]: %v\n", n, id)
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
				"",
				"std",
				"std::abs",
				"Math::do_hello",
				"Math2::do_hello",
				"Foo::setDouble",
				"Foo::getme",
				"std::vector<Foo>::push_back",
				"std::vector<Foo>::size",
				//"std::locale",
			}
			for _, n := range names {
				t := cxxtypes.IdByName(n)
				if t == nil {
					fmt.Printf("could not inspect identifier [%s]\n", n)
					continue
				}
				fmt.Printf(":: inspecting [%s]...\n", n)
				switch tt := t.(type) {
				case *cxxtypes.Namespace:
					fmt.Printf(" -> %s (#mbrs: %d)\n", tt.IdScopedName(), tt.NumMember())
				case *cxxtypes.ClassType:
					fmt.Printf(" #bases: %d\n", tt.NumBase())
					for i := 0; i < tt.NumBase(); i++ {
						b := tt.Base(i)
						fmt.Printf(" %d: %v\n", i, b)
					}
					fmt.Printf(" #mbrs: %d\n", tt.NumMember())
					for i := 0; i < tt.NumMember(); i++ {
						m := tt.Member(i)
						fmt.Printf(" %d: %v\n", i, m)
					}
				case *cxxtypes.StructType:
					fmt.Printf(" #bases: %d\n", tt.NumBase())
					for i := 0; i < tt.NumBase(); i++ {
						b := tt.Base(i)
						fmt.Printf(" %d: %v\n", i, b)
					}
					fmt.Printf(" #mbrs: %d\n", tt.NumMember())
					for i := 0; i < tt.NumMember(); i++ {
						m := tt.Member(i)
						fmt.Printf(" %d: %v\n", i, m)
					}
				case *cxxtypes.OverloadFunctionSet:
					for i := 0; i < tt.NumFunction(); i++ {
						fmt.Printf("%s: %s\n",
							tt.IdName(), tt.Function(i).Signature())
					}
				default:
					fmt.Printf("%v\n", tt)
				}
			}
		}
		ids := cxxtypes.IdNames()
		fmt.Printf("== distilled [%d] identifiers.\n", len(ids))
	}

	dst, err := os.Create(*oname)
	if err != nil {
		fmt.Printf("**err** %v\n", err)
		os.Exit(1)
	}

	err = cxxtypes.SaveIds(dst)
	if err != nil {
		fmt.Printf("**err** %v\n", err)
		os.Exit(1)
	}

	err = dst.Sync()
	if err != nil {
		fmt.Printf("**err** %v\n", err)
		os.Exit(1)
	}

	err = dst.Close()
	if err != nil {
		fmt.Printf("**err** %v\n", err)
		os.Exit(1)
	}

}

// EOF
