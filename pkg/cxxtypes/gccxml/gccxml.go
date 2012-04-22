// Package gccxml reads an XML file produced by GCC_XML and fills in the cxxtypes' registry.
package gccxml

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	_ "bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
)

// globals

var g_dbg bool = true
var g_ids idDB


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

	fmt.Printf("== gccxml data ==\n")
	root.printStats()

	// walk over the xmlTree twice.
	v := func(node i_id) bool {
		//println("-->",node.id())
		g_ids[node.id()] = node
		return true
	}
	println("--walking--...")
	walk(inspector(v), root)

	fmt.Printf("ids: %d\n", len(g_ids))


	// 2) applies fixes to xmlFoobar structs.
	v = func(node i_id) bool {
		nn := g_ids[node.id()]
		if nn != node {
			panic("oops")
		}
		switch nn := node.(type) {
		case i_name:
			name := nn.name()
			if name == "do_hello" || 
			   name =="do_hello_c" {
				println("-->",node.id(), nn.name())
			}
		}
		return true
	}
	walk(inspector(v), root)

	return err
}

func init() {
	g_ids = make(idDB, 128)
}

// EOF
