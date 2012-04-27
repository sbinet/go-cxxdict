// go-gencxxinfos creates the C++ types registry from some source (gccxml, 
// clang) and saves it under some form.
package main

import (
	"fmt"
	"os"

	"bitbucket.org/binet/go-cxxdict/pkg/cxxtypes/gccxml"
)

func main() {
	fmt.Printf("== go-gencxxinfos ==\n")
	err := gccxml.LoadTypes("t.xml")
	if err != nil {
		fmt.Printf("**err** %v\n", err)
		os.Exit(1)
	}
}

// EOF
