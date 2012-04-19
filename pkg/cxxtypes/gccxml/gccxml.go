// Package gccxml reads an XML file produced by GCC_XML and fills in the cxxtypes' registry.
package gccxml

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"

	_ "bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
)

type xmlTree struct {
	XMLName    xml.Name       `xml:"GCC_XML"`
	Namespaces []xmlNamespace `xml:"Namespace"`
	Classes    []xmlClass     `xml:"Class"`
	Structs    []xmlStruct    `xml:"Struct"`
	Functions  []xmlFunction  `xml:"Function"`
}

type xmlCommon struct {
	Id        string `xml:"id,attr"`
	Name      string `xml:"name,attr"`
	Context   string `xml:"context,attr"`
	Mangled   string `xml:"mangled,attr"`
	Demangled string `xml:"demangled,attr"`
}

func (x xmlCommon) infos() string {
	n := x.Name
	if x.Demangled != "" {
		n = x.Demangled
	}
	return fmt.Sprintf(`id="%s", name="%s"`, x.Id, n)
}

type xmlLocation struct {
	Location string `xml:"location,attr"`
	File     string `xml:"file,attr"`
	Line     string `xml:"line,attr"`
}

type xmlNamespace struct {
	xmlCommon
	Attributes string `xml:"attributes,attr"`
	Members    string `xml:"members,attr"`
}

func (x xmlNamespace) String() string {
	return fmt.Sprintf(`Namespace{%s}`, x.xmlCommon.infos())
}

type xmlArgument struct {
	xmlLocation

	Name    string `xml:"name,attr"`
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr"`
}

type xmlEllipsis struct {
	XMLName xml.Name `xml:"Ellipsis"`
}

type xmlBase struct {
	Type    string `xml:"type,attr"`
	Access  string `xml:"access,attr"`
	Virtual string `xml:"virtual,attr"`
	Offset  string `xml:"offset,attr"`
}

type xmlRecord struct {
	xmlCommon
	xmlLocation

	BaseIds    string    `xml:"bases,attr"`
	Bases      []xmlBase `xml:"Base"`
	Members    string    `xml:"members,attr"`
	Artificial string    `xml:"artificial,attr"`
	Size       string    `xml:"size,attr"`
}

func (x xmlRecord) infos() string {
	bases := ""
	if len(x.Bases) > 0 {
		bases = ", bases=["
		for i, b := range x.Bases {
			if i > 0 {
				bases += " "
			}
			bases += b.Offset
			if i+1 != len(x.Bases) {
				bases += ", "
			}
		}
		bases += "]"
	}
	nmbrs := len(strings.Split(x.Members, " "))
	return fmt.Sprintf(`{%s, size=%s%s, nmbrs=%d}`,
		x.xmlCommon.infos(),
		x.Size,
		bases,
		nmbrs)
}

type xmlStruct struct {
	xmlRecord
}

func (x xmlStruct) String() string {
	return "Struct" + x.xmlRecord.infos()
}

type xmlClass struct {
	xmlRecord
}

func (x xmlClass) String() string {
	return "Class" + x.xmlRecord.infos()
}

type xmlFunction struct {
	xmlCommon
	xmlLocation
	Attributes string        `xml:"attributes,attr"`
	Extern     string        `xml:"extern,attr"`
	Arguments  []xmlArgument `xml:"Argument"`
	Ellipsis   xmlEllipsis   `xml:"Ellipsis"`
	Returns    string        `xml:"returns,attr"`
}

func (x xmlFunction) String() string {
	return fmt.Sprintf(`Function{%s, returns=%s, nargs=%d, ellipsis=%s}`,
		x.xmlCommon.infos(),
		x.Returns,
		len(x.Arguments),
		x.Ellipsis)
}

// LoadTypes reads an XML file produced by GCC_XML and fills the cxxtypes' registry accordingly.
func LoadTypes(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	root := xmlTree{}
	err = xml.Unmarshal(data, &root)
	if err != nil {
		return err
	}

	fmt.Printf("loaded [%d] namespaces\n", len(root.Namespaces))
	for i, v := range root.Namespaces {
		fmt.Printf(" - %d: %v\n", i, v)
	}

	fmt.Printf("loaded [%d] classes\n", len(root.Classes))
	for i, v := range root.Classes {
		fmt.Printf(" - %d: %v\n", i, v)
	}

	fmt.Printf("loaded [%d] functions\n", len(root.Functions))
	for i, v := range root.Functions {
		fmt.Printf(" - %d: %v\n", i, v)
	}
	return err
}

// EOF
