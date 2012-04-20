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
	XMLName xml.Name `xml:"GCC_XML"`

	Arrays            []xmlArray            `xml:"ArrayType"`
	Classes           []xmlClass            `xml:"Class"`
	Constructors      []xmlConstructor      `xml:"Constructor"`
	Converters        []xmlConverter        `xml:"Converter"`
	CvQualifiedTypes  []xmlCvQualifiedType  `xml:"CvQualifiedType"`
	Destructors       []xmlDestructor       `xml:"Destructor"`
	Enumerations      []xmlEnumeration      `xml:"Enumeration"`
	Fields            []xmlField            `xml:"Field"`
	Files             []xmlFile             `xml:"File"`
	Functions         []xmlFunction         `xml:"Function"`
	FunctionTypes     []xmlFunctionType     `xml:"FunctionType"`
	FundamentalTypes  []xmlFundamentalType  `xml:"FundamentalType"`
	Methods           []xmlMethod           `xml:"Method"`
	MethodTypes       []xmlMethodType       `xml:"MethodType"`
	Namespaces        []xmlNamespace        `xml:"Namespace"`
	NamespaceAliases  []xmlNamespaceAlias   `xml:"NamespaceAlias"`
	OperatorFunctions []xmlOperatorFunction `xml:"OperatorFunction"`
	OperatorMethods   []xmlOperatorMethod   `xml:"OperatorMethod"`
	OffsetTypes       []xmlOffsetType       `xml:"OffsetType"`
	PointerTypes      []xmlPointerType      `xml:"PointerType"`
	ReferenceTypes    []xmlReferenceType    `xml:"ReferenceType"`
	Structs           []xmlStruct           `xml:"Struct"`
	Typedefs          []xmlTypedef          `xml:"Typedef"`
	Unimplementeds    []xmlUnimplemented    `xml:"Unimplemented"`
	Unions            []xmlUnion            `xml:"Union"`
	Variables         []xmlVariable         `xml:"Variable"`
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

type xmlArgument struct {
	xmlLocation

	Name    string `xml:"name,attr"`
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr"`
}

type xmlArray struct {
	Id         string `xml:"id,attr"`
	Min        string `xml:"min,attr"`
	Max        string `xml:"max,attr"`
	Type       string `xml:"type,attr"`
	Size       string `xml:"size,attr"`
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"`
}

type xmlBase struct {
	Type    string `xml:"type,attr"`
	Access  string `xml:"access,attr"`
	Virtual string `xml:"virtual,attr"`
	Offset  string `xml:"offset,attr"`
}

type xmlEllipsis struct {
	XMLName xml.Name `xml:"Ellipsis"`
}

type xml_record struct {
	Abstract   string `xml:"abstract,attr"`
	Access     string `xml:"access,attr"` // default "public"
	Align      string `xml:"align,attr"`
	Artificial string `xml:"artificial,attr"`
	Attributes string `xml:"attributes,attr"`
	XBases     string `xml:"bases,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Incomplete string `xml:"incomplete,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Members    string `xml:"members,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`

	Bases []xmlBase `xml:"Base"`
}

type xmlConstructor struct {
	Access     string `xml:"access,attr"`     // default "public"
	Artifical  string `xml:"artifical,attr"`  // implied
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Throw      string `xml:"throw,attr"`
}

type xmlConverter struct {
	Access     string `xml:"access,attr"`     // default "public"
	Attributes string `xml:"attributes,attr"` // implied
	Const      string `xml:"const,attr"`      // default "0"
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Returns    string `xml:"returns,attr"`
	Throw      string `xml:"throw,attr"`
	Virtual    string `xml:"virtual,attr"` // default "0"
}

type xmlCvQualifiedType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Const      string `xml:"const,attr"`
	Id         string `xml:"id,attr"`
	Restrict   string `xml:"restrict,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
	Volatile   string `xml:"volatile,attr"`
}

type xmlDestructor struct {
	Access     string `xml:"access,attr"`     // default "public"
	Artifical  string `xml:"artifical,attr"`  // implied
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Throw      string `xml:"throw,attr"`
	Virtual    string `xml:"virtual,attr"` // default "0"
}

type xmlEnumValue struct {
	Init string `xml:"init,attr"`
	Name string `xml:"name,attr"`
}

type xmlEnumeration struct {
	Access     string `xml:"access,attr"` // default "public"
	Align      string `xml:"align,attr"`
	Artifical  string `xml:"artifical,attr"`  // implied
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`
}

type xmlField struct {
	Access     string `xml:"access,attr"`     // default "public"
	Attributes string `xml:"attributes,attr"` // implied
	Bits       string `xml:"bits,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Mutable    string `xml:"mutable,attr"`
	Name       string `xml:"name,attr"`
	Offset     string `xml:"offset,attr"`
	Type       string `xml:"type,attr"`
}

type xmlFile struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type xmlFunction struct {
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Returns    string `xml:"returns,attr"`
	Throw      string `xml:"throw,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

type xmlFunctionType struct {
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Returns    string `xml:"returns,attr"`
}

type xmlFundamentalType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`
}

type xmlMethod struct {
	Access      string `xml:"access,attr"`     // default "public"
	Attributes  string `xml:"attributes,attr"` // implied
	Const       string `xml:"const,attr"`
	Context     string `xml:"context,attr"`
	Demangled   string `xml:"demangled,attr"`
	Endline     string `xml:"endline,attr"`
	Extern      string `xml:"extern,attr"` // default "0"
	File        string `xml:"file,attr"`
	Id          string `xml:"id,attr"`
	Line        string `xml:"line,attr"`
	Location    string `xml:"location,attr"`
	Mangled     string `xml:"mangled,attr"`
	Name        string `xml:"name,attr"`
	PureVirtual string `xml:"pure_virtual,attr"` // default "0"
	Returns     string `xml:"returns,attr"`
	Static      string `xml:"static,attr"` // default "0"
	Throw       string `xml:"throw,attr"`
	Virtual     string `xml:"virtual,attr"` // default "0"
}

type xmlMethodType struct {
	Attributes string `xml:"attributes,attr"` // implied
	BaseType   string `xml:"basetype,attr"`
	Id         string `xml:"id,attr"`
	Returns    string `xml:"returns,attr"`
}

type xmlNamespace struct {
	Attributes string `xml:"attributes,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Id         string `xml:"id,attr"`
	Mangled    string `xml:"mangled,attr"`
	Members    string `xml:"members,attr"`
	Name       string `xml:"name,attr"`
}

type xmlNamespaceAlias struct {
	Context   string `xml:"context,attr"`
	Demangled string `xml:"demangled,attr"`
	Id        string `xml:"id,attr"`
	Mangled   string `xml:"mangled,attr"`
	Name      string `xml:"name,attr"`
	Namespace string `xml:"namespace,attr"`
}

type xmlOffsetType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	BaseType   string `xml:"basetype,attr"`
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

type xmlOperatorFunction struct {
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Returns    string `xml:"returns,attr"`
	Throw      string `xml:"throw,attr"`
}

type xmlOperatorMethod struct {
	xmlMethod
}

type xmlPointerType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

type xmlReferenceType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

type xmlStruct struct {
	xml_record
}

type xmlTypedef struct {
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
}

type xmlUnimplemented struct {
	Function     string `xml:"function,attr"`
	Id           string `xml:"id,attr"`
	Node         string `xml:"node,attr"`
	TreeCode     string `xml:"tree_code,attr"`
	TreeCodeName string `xml:"tree_code_name,attr"` // template_type_parm|typename_type|using_decl
}

type xmlUnion struct {
	Abstract      string `xml:"abstract,attr"`     // default "0"
	Access      string `xml:"access,attr"`     // default "public"
	Align  string `xml:"align,attr"` // implied
	Artificial string `xml:"artificial,attr"` // "0"
	Attributes string `xml:"attributes,attr"` // implied
	Bases     string `xml:"bases,attr"`

	Context     string `xml:"context,attr"`
	Demangled   string `xml:"demangled,attr"`
	Endline     string `xml:"endline,attr"`
	Extern      string `xml:"extern,attr"` // default "0"
	File        string `xml:"file,attr"`
	Id          string `xml:"id,attr"`
	Line        string `xml:"line,attr"`
	Location    string `xml:"location,attr"`
	Mangled     string `xml:"mangled,attr"`
	Name        string `xml:"name,attr"`
	PureVirtual string `xml:"pure_virtual,attr"` // default "0"
	Returns     string `xml:"returns,attr"`
	Static      string `xml:"static,attr"` // default "0"
	Throw       string `xml:"throw,attr"`
	Virtual     string `xml:"virtual,attr"` // default "0"

}

func (x xmlNamespace) String() string {
	return fmt.Sprintf(`Namespace{%s}`, x.xmlCommon.infos())
}

func (x xml_record) infos() string {
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

func (x xmlStruct) String() string {
	return "Struct" + x.xml_record.infos()
}

type xmlClass struct {
	xml_record
}

func (x xmlClass) String() string {
	return "Class" + x.xml_record.infos()
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

	fmt.Printf("loaded [%d] typedefs\n", len(root.Typedefs))

	return err
}

// EOF
