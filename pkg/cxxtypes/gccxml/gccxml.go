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
var g_ids idDB = make(idDB, 128)


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

	root.fixup()

	fmt.Printf("== gccxml data ==\n")
	root.printStats()

	// 2) applies fixes to xmlFoobar structs.
	v = func(node i_id) bool {
		nn := g_ids[node.id()]
		if nn != node {
			panic("oops")
		}
		patchTemplateName(node)
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

// EOF
