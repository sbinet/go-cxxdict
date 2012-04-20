// Package gccxml reads an XML file produced by GCC_XML and fills in the cxxtypes' registry.
package gccxml

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	_ "bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
)

var g_dbg bool = true

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

	ids := collectIds(root)
	fmt.Printf("ids: %d\n", len(ids))

	return err
}

// EOF
