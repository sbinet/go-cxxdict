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
// id of the global namespace
var g_globalns_id string = ""
var g_ids idDB = make(idDB, 128)
// a cache of already processed ids (and their fully qualified name)
var g_processed_ids map[string]string

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
		names := cxxtypes.TypeNames()
		for _,n := range names {
			t := cxxtypes.TypeByName(n)
			fmt.Printf("[%s]: %v\n", n, t)
		}
		
	}
	return err
}

// EOF
