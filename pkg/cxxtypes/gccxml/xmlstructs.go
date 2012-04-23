package gccxml

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// helper function to "convert" a 0|1 string into a boolean
func to_bool(v string) bool {
	if v == "" || v == "0" {
		return false
	}
	return true
}

type xmlTree struct {
	XMLName xml.Name `xml:"GCC_XML"`

	Arrays            []*xmlArray            `xml:"ArrayType"`
	Classes           []*xmlClass            `xml:"Class"`
	Constructors      []*xmlConstructor      `xml:"Constructor"`
	Converters        []*xmlConverter        `xml:"Converter"`
	CvQualifiedTypes  []*xmlCvQualifiedType  `xml:"CvQualifiedType"`
	Destructors       []*xmlDestructor       `xml:"Destructor"`
	Enumerations      []*xmlEnumeration      `xml:"Enumeration"`
	Fields            []*xmlField            `xml:"Field"`
	Files             []*xmlFile             `xml:"File"`
	Functions         []*xmlFunction         `xml:"Function"`
	FunctionTypes     []*xmlFunctionType     `xml:"FunctionType"`
	FundamentalTypes  []*xmlFundamentalType  `xml:"FundamentalType"`
	Methods           []*xmlMethod           `xml:"Method"`
	MethodTypes       []*xmlMethodType       `xml:"MethodType"`
	Namespaces        []*xmlNamespace        `xml:"Namespace"`
	NamespaceAliases  []*xmlNamespaceAlias   `xml:"NamespaceAlias"`
	OperatorFunctions []*xmlOperatorFunction `xml:"OperatorFunction"`
	OperatorMethods   []*xmlOperatorMethod   `xml:"OperatorMethod"`
	OffsetTypes       []*xmlOffsetType       `xml:"OffsetType"`
	PointerTypes      []*xmlPointerType      `xml:"PointerType"`
	ReferenceTypes    []*xmlReferenceType    `xml:"ReferenceType"`
	Structs           []*xmlStruct           `xml:"Struct"`
	Typedefs          []*xmlTypedef          `xml:"Typedef"`
	Unimplementeds    []*xmlUnimplemented    `xml:"Unimplemented"`
	Unions            []*xmlUnion            `xml:"Union"`
	Variables         []*xmlVariable         `xml:"Variable"`
}

func (x *xmlTree) printStats() {

	fmt.Printf("loaded [%d] namespaces\n", len(x.Namespaces))
	fmt.Printf("loaded [%d] classes\n", len(x.Classes))
	fmt.Printf("loaded [%d] functions\n", len(x.Functions))
	fmt.Printf("loaded [%d] typedefs\n", len(x.Typedefs))
	fmt.Printf("loaded [%d] files\n", len(x.Files))
	fmt.Printf("loaded [%d] structs\n", len(x.Structs))
	fmt.Printf("loaded [%d] unions\n", len(x.Unions))
	fmt.Printf("loaded [%d] pointer-types\n", len(x.PointerTypes))
	fmt.Printf("loaded [%d] ref-types\n", len(x.ReferenceTypes))
	fmt.Printf("loaded [%d] cv-types\n", len(x.CvQualifiedTypes))
	fmt.Printf("loaded [%d] fundamentals\n", len(x.FundamentalTypes))

	if g_dbg {
		for i, _ := range x.FundamentalTypes {
			fmt.Printf(" => %v\n", x.FundamentalTypes[i])
		}
	}

	fmt.Printf("loaded [%d] array-types\n", len(x.Arrays))
	for _, a := range x.Arrays {
		fmt.Printf(" => '%s' %v\n", a.name(), a)
	}
}

// fixup fixes a few of the "features" of GCC-XML data
func (x *xmlTree) fixup() {

	// first, fixup classes and structs
	for _, v := range x.Classes {
		patchTemplateName(v)
	}
	for _, v := range x.Structs {
		patchTemplateName(v)
	}

	// functions
	for _, v := range x.Functions {
		v.set_name(addTemplateToName(v.name(), v.demangled()))
		patchTemplateName(v)
	}

	// operators
	for _, v := range x.OperatorFunctions {
		//FIXME

		patchTemplateName(v)
		v.set_name(addTemplateToName(v.name(), v.demangled()))
	}

	// ctors
	for _, v := range x.Constructors {
		if len(v.name()) >= 3 && string(v.name()[0:3]) != "_ZT" {
			v.set_name(addTemplateToName(v.name(), v.demangled()))
			patchTemplateName(v)
		}
	}

	// methods
	for _, v := range x.Methods {
		if len(v.name()) >= 3 && string(v.name()[0:3]) != "_ZT" {
			v.set_name(addTemplateToName(v.name(), v.demangled()))
			patchTemplateName(v)
		}
	}

	// op-methods
	for _, v := range x.OperatorMethods {
		if len(v.name()) >= 3 && string(v.name()[0:3]) != "_ZT" {
			v.set_name(addTemplateToName(v.name(), v.demangled()))
			patchTemplateName(v)
		}
	}

	// builtins
	for _, v := range x.FundamentalTypes {
		alltmpl := false
		v.set_name(normalizeName(v.name(), alltmpl))
	}

}

func (x *xmlTree) id() string {
	return "__0__"
}

// idDB associates the gccxml string id of a type to its parsed xmlFoobar struct
type idDB map[string]i_id

type xmlArgument struct {
	Attributes string `xml:"attributes,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
	Default    string `xml:"default,attr"`
}

func (x *xmlArgument) id() string {
	return x.Id
}

func (x *xmlArgument) name() string {
	return x.Name
}

func (x *xmlArgument) set_name(n string) {
	x.Name = n
}

type xmlArray struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"`
	Id         string `xml:"id,attr"`
	Max        string `xml:"max,attr"`
	Min        string `xml:"min,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlArray) String() string {
	tname := ""
	switch tt := g_ids[x.Type].(type) {
	case i_name:
		tname = " (" + tt.name() + ")"
	default:
	}
	return fmt.Sprintf("Array{Id=%s, Type='%s'%s, Size=%s, Align=%s, Attrs='%s'}",
		x.Id, x.Type, tname, x.Size, x.Align, x.Attributes)
}

func (x *xmlArray) id() string {
	return x.Id
}

func (x *xmlArray) name() string {
	t := g_ids[x.Type]
	tt := t.(i_name)
	return fmt.Sprintf("%s[%s]", tt.name(), x.Size)
}

func (x *xmlArray) set_name(n string) {
	panic("fixme: can set name of an array yet")
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

func (x *xml_record) id() string {
	return x.Id
}

func (x *xml_record) name() string {
	return x.Name
}

func (x *xml_record) set_name(n string) {
	x.Name = n
}

func (x *xml_record) mangled() string {
	return x.Mangled
}

func (x *xml_record) set_mangled(n string) {
	x.Mangled = n
}

func (x *xml_record) demangled() string {
	return x.Demangled
}

func (x *xml_record) set_demangled(n string) {
	x.Demangled = n
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
		x.Name,
		x.Size,
		bases,
		nmbrs)
}

type xmlClass struct {
	xml_record
}

func (x xmlClass) String() string {
	return "Class" + x.xml_record.infos()
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

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlConstructor) id() string {
	return x.Id
}

func (x *xmlConstructor) name() string {
	return x.Name
}

func (x *xmlConstructor) set_name(n string) {
	x.Name = n
}

func (x *xmlConstructor) mangled() string {
	return x.Mangled
}

func (x *xmlConstructor) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlConstructor) demangled() string {
	return x.Demangled
}

func (x *xmlConstructor) set_demangled(n string) {
	x.Demangled = n
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

func (x *xmlConverter) id() string {
	return x.Id
}

func (x *xmlConverter) name() string {
	return x.Name
}

func (x *xmlConverter) set_name(n string) {
	x.Name = n
}

func (x *xmlConverter) mangled() string {
	return x.Mangled
}

func (x *xmlConverter) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlConverter) demangled() string {
	return x.Demangled
}

func (x *xmlConverter) set_demangled(n string) {
	x.Demangled = n
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

func (x *xmlCvQualifiedType) id() string {
	return x.Id
}

func (x *xmlCvQualifiedType) name() string {
	t := g_ids[x.Type]
	return t.(i_name).name()
}

func (x *xmlCvQualifiedType) set_name(n string) {
	panic("fixme: can set the name of a cv-qualified type - yet")
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

func (x *xmlDestructor) id() string {
	return x.Id
}

func (x *xmlDestructor) name() string {
	return x.Name
}

func (x *xmlDestructor) set_name(n string) {
	x.Name = n
}

func (x *xmlDestructor) mangled() string {
	return x.Mangled
}

func (x *xmlDestructor) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlDestructor) demangled() string {
	return x.Demangled
}

func (x *xmlDestructor) set_demangled(n string) {
	x.Demangled = n
}

type xmlEnumValue struct {
	Init string `xml:"init,attr"`
	Name string `xml:"name,attr"`
}

func (x *xmlEnumValue) name() string {
	return x.Name
}

func (x *xmlEnumValue) set_name(n string) {
	x.Name = n
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

	EnumValues []xmlEnumValue `xml:"EnumValue"`
}

func (x *xmlEnumeration) id() string {
	return x.Id
}

func (x *xmlEnumeration) name() string {
	return x.Name
}

func (x *xmlEnumeration) set_name(n string) {
	x.Name = n
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

func (x *xmlField) id() string {
	return x.Id
}

func (x *xmlField) name() string {
	return x.Name
}

func (x *xmlField) set_name(n string) {
	x.Name = n
}

func (x *xmlField) mangled() string {
	return x.Mangled
}

func (x *xmlField) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlField) demangled() string {
	return x.Demangled
}

func (x *xmlField) set_demangled(n string) {
	x.Demangled = n
}

type xmlFile struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

func (x *xmlFile) id() string {
	return x.Id
}

func (x *xmlFile) name() string {
	return x.Name
}

func (x *xmlFile) set_name(n string) {
	x.Name = n
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

func (x *xmlFunction) id() string {
	return x.Id
}

func (x *xmlFunction) name() string {
	return x.Name
}

func (x *xmlFunction) set_name(n string) {
	x.Name = n
}

func (x *xmlFunction) mangled() string {
	return x.Mangled
}

func (x *xmlFunction) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlFunction) demangled() string {
	return x.Demangled
}

func (x *xmlFunction) set_demangled(n string) {
	x.Demangled = n
}

func (x xmlFunction) String() string {
	return fmt.Sprintf(`Function{%s, returns=%s, nargs=%d, ellipsis=%s}`,
		x.Name,
		x.Returns,
		len(x.Arguments),
		x.Ellipsis)
}

type xmlFunctionType struct {
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Returns    string `xml:"returns,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlFunctionType) id() string {
	return x.Id
}

type xmlFundamentalType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`
}

func (x *xmlFundamentalType) String() string {
	return fmt.Sprintf(
		`builtin{"%s", size=%s, align=%s}`,
		x.Name, x.Size, x.Align,
	)
}

func (x *xmlFundamentalType) id() string {
	return x.Id
}

func (x *xmlFundamentalType) name() string {
	return x.Name
}

func (x *xmlFundamentalType) set_name(n string) {
	x.Name = n
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

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlMethod) id() string {
	return x.Id
}

func (x *xmlMethod) name() string {
	return x.Name
}

func (x *xmlMethod) set_name(n string) {
	x.Name = n
}

func (x *xmlMethod) mangled() string {
	return x.Mangled
}

func (x *xmlMethod) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlMethod) demangled() string {
	return x.Demangled
}

func (x *xmlMethod) set_demangled(n string) {
	x.Demangled = n
}

type xmlMethodType struct {
	Attributes string `xml:"attributes,attr"` // implied
	BaseType   string `xml:"basetype,attr"`
	Id         string `xml:"id,attr"`
	Returns    string `xml:"returns,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlMethodType) id() string {
	return x.Id
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

func (x *xmlNamespace) id() string {
	return x.Id
}

func (x *xmlNamespace) name() string {
	return x.Name
}

func (x *xmlNamespace) set_name(n string) {
	x.Name = n
}

func (x *xmlNamespace) mangled() string {
	return x.Mangled
}

func (x *xmlNamespace) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlNamespace) demangled() string {
	return x.Demangled
}

func (x *xmlNamespace) set_demangled(n string) {
	x.Demangled = n
}

func (x xmlNamespace) String() string {
	return fmt.Sprintf(`Namespace{%s}`, x.Name)
}

type xmlNamespaceAlias struct {
	Context   string `xml:"context,attr"`
	Demangled string `xml:"demangled,attr"`
	Id        string `xml:"id,attr"`
	Mangled   string `xml:"mangled,attr"`
	Name      string `xml:"name,attr"`
	Namespace string `xml:"namespace,attr"`
}

func (x *xmlNamespaceAlias) id() string {
	return x.Id
}

func (x *xmlNamespaceAlias) name() string {
	return x.Name
}

func (x *xmlNamespaceAlias) set_name(n string) {
	x.Name = n
}

func (x *xmlNamespaceAlias) mangled() string {
	return x.Mangled
}

func (x *xmlNamespaceAlias) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlNamespaceAlias) demangled() string {
	return x.Demangled
}

func (x *xmlNamespaceAlias) set_demangled(n string) {
	x.Demangled = n
}

type xmlOffsetType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	BaseType   string `xml:"basetype,attr"`
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlOffsetType) id() string {
	return x.Id
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

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlOperatorFunction) id() string {
	return x.Id
}

func (x *xmlOperatorFunction) name() string {
	return x.Name
}

func (x *xmlOperatorFunction) set_name(n string) {
	x.Name = n
}

func (x *xmlOperatorFunction) mangled() string {
	return x.Mangled
}

func (x *xmlOperatorFunction) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlOperatorFunction) demangled() string {
	return x.Demangled
}

func (x *xmlOperatorFunction) set_demangled(n string) {
	x.Demangled = n
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

func (x *xmlPointerType) id() string {
	return x.Id
}

func (x *xmlPointerType) name() string {
	t := g_ids[x.Type]
	switch tt := t.(type) {
	case i_name:
		return tt.name() + "*"
	default:
	}
	return "<unnamed>*"
}

func (x *xmlPointerType) set_name(n string) {
	panic("fixme: can set a name for a pointer-type - yet")
}

type xmlReferenceType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlReferenceType) id() string {
	return x.Id
}

type xmlStruct struct {
	xml_record
}

func (x xmlStruct) String() string {
	return "Struct" + x.xml_record.infos()
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

func (x *xmlTypedef) id() string {
	return x.Id
}

func (x *xmlTypedef) name() string {
	return x.Name
}

func (x *xmlTypedef) set_name(n string) {
	x.Name = n
}

type xmlUnimplemented struct {
	Function     string `xml:"function,attr"`
	Id           string `xml:"id,attr"`
	Node         string `xml:"node,attr"`
	TreeCode     string `xml:"tree_code,attr"`
	TreeCodeName string `xml:"tree_code_name,attr"` // template_type_parm|typename_type|using_decl
}

func (x *xmlUnimplemented) id() string {
	return x.Id
}

type xmlUnion struct {
	Abstract   string `xml:"abstract,attr"`   // default "0"
	Access     string `xml:"access,attr"`     // default "public"
	Align      string `xml:"align,attr"`      // implied
	Artificial string `xml:"artificial,attr"` // "0"
	Attributes string `xml:"attributes,attr"` // implied
	Bases      string `xml:"bases,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Incomplete string `xml:"incomplete,attr"` // default "0"
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`
}

func (x *xmlUnion) id() string {
	return x.Id
}

func (x *xmlUnion) name() string {
	return x.Name
}

func (x *xmlUnion) set_name(n string) {
	x.Name = n
}

func (x *xmlUnion) mangled() string {
	return x.Mangled
}

func (x *xmlUnion) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlUnion) demangled() string {
	return x.Demangled
}

func (x *xmlUnion) set_demangled(n string) {
	x.Demangled = n
}

type xmlVariable struct {
	Access     string `xml:"access,attr"`     // default "public"
	Artificial string `xml:"artificial,attr"` // "0"
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Init       string `xml:"init,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlVariable) id() string {
	return x.Id
}

func (x *xmlVariable) name() string {
	return x.Name
}

func (x *xmlVariable) set_name(n string) {
	x.Name = n
}

func (x *xmlVariable) mangled() string {
	return x.Mangled
}

func (x *xmlVariable) set_mangled(n string) {
	x.Mangled = n
}

func (x *xmlVariable) demangled() string {
	return x.Demangled
}

func (x *xmlVariable) set_demangled(n string) {
	x.Demangled = n
}

// utils ---

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// addTemplateToName
func addTemplateToName(n, demangled string) string {
	name := n
	posargs := strings.LastIndex(demangled, "(")
	println("==",posargs, demangled)
	if posargs > 1 && demangled[posargs-1] == '>' &&
		(isAlphaNum(demangled[posargs-2]) ||
		demangled[posargs-2] == '_') {
		posname := strings.Index(demangled, n+"<")
		println("++", posname, demangled, n+"<")
		if posname >= 0 {
			reui := regexp.MustCompile(`\b(unsigned)(\s+)?([^\w\s])`)
			nn := string(demangled[posname:posargs])
			name = reui.ReplaceAllString(nn, `unsigned int\3`)
			println("~~",nn, posname, posargs, name, n)
		}
	}
	return name
}

// patchTemplateName
func patchTemplateName(node i_id) {
	var n i_name
	// check if this node has a name. early return otherwise
	switch nn := node.(type) {
	case i_name:
		n = nn
	default:
		return
	}

	name := n.name()
	if strings.Contains(name, ">") {
		return
	}

	tmplend := len(name)
	tmplpos := -1

	switch node.(type) {
	case *xmlFunction, *xmlOperatorFunction, *xmlConstructor,
		*xmlMethod, *xmlOperatorMethod:
		tmplend = strings.LastIndex(name, "(")

	default:
		return
	}

	if tmplend > 1 && name[tmplend-1] == '>' {
		tmplpos = strings.Index(name, "<")
	}

	//println("...",tmplpos,tmplend,len(name))
	if tmplpos <= -1 {
		return
	}
	tmplpos += 1
	tmplend -= 1
	tmpl := name[tmplpos:tmplend]
	
	// replace template argument "12u" of "12ul" with "12"
	rep := regexp.MustCompile(
	`\b([\d]+)ul?\b`).ReplaceAllString(tmpl, `\1`)
	// replace -0x00000000000000001 with -1
	rep = regexp.MustCompile(
	`-0x0*([1-9A-Fa-f][0-9A-Fa-f]*)\b`).ReplaceAllString(rep, `-\1`)
	name = string(name[:tmplpos]) + rep + name[tmplend:]
	//FIXME:
	// # replace template argument "12u" or "12ul" by "12":
	// rep = re.sub(r"\b([\d]+)ul?\b", '\\1', name[postmplt:postmpltend])
	// # replace -0x00000000000000001 by -1
	// rep = re.sub(r"-0x0*([1-9A-Fa-f][0-9A-Fa-f]*)\b", '-\\1', rep)
	// name = name[:postmplt] + rep + name[postmpltend:]
	// attrs['name'] = name

	n.set_name(name)
	//println("--", name, tmpl, n.name())

	return
}

func normalizeName(n string, alltmpl bool) string {
	name := strings.TrimSpace(n)
	if !strings.Contains(name, "<") {
		if !strings.Contains(name, "int") {
			return name
		}

		for _, v := range [][]string{
			{"long long unsigned int", "unsigned long long"},
			{"long long int", "long long"},
			{"unsigned short int", "unsigned short"},
			{"short unsigned int", "unsigned short"},
			{"short int", "short"},
			{"long unsigned int", "unsigned long"},
			{"unsigned long int", "unsigned long"},
			{"long int", "long"},
		} {
			name = strings.Replace(name, v[0], v[1], -1)
		}
		return name
	}
	clsname := string(name[:strings.Index(name, "<")])
	var suffix string
	if strings.LastIndex(name, ">") < len(clsname) {
		suffix = ""
	} else {
		suffix = string(name[strings.LastIndex(name, ">")+1:])
	}
	args := getTemplateArgs(name)
	sargs := make([]string, 0, len(args))
	for _, v := range args {
		sargs = append(sargs, normalizeClass(v, alltmpl))
	}

	//fmt.Sprintf("--> '%s' '%s' %v %v\n", suffix, clsname, args, sargs)
	if !alltmpl {
		defargs, ok := g_stldeftable[clsname]
		//println("---", name, defargs, ok, clsname)
		if ok {
			switch defargs_t := defargs.(type) {
			case []string:
				// select only the template parameters different from default ones
				args = sargs
				sargs = make([]string, 0)
				nargs := len(args)
				if len(defargs_t) < nargs {
					nargs = len(defargs_t)
				}
				for i, _ := range args[:nargs] {
					//println("::", i, args[i], defargs_t[i])
					if !strings.Contains(string(args[i]), string(defargs_t[i])) {
						sargs = append(sargs, string(args[i]))
					}
				}
			case map[string][]string:
				//FIXME
				panic("fixme")
				args_tmp := make([]string, 0, len(args))
				for _, v := range args {
					vv := normalizeClass(v, true)
					//println("--", v, "-->", vv)
					args_tmp = append(args_tmp, vv)
				}
				// args = args_tmp
				// defargs_tup := make([]string, 0)
				// for i,_ := range args[1:] {
				// 	defargs_tup = defargs_t[args[:1]]
				// }
			default:
				panic(fmt.Sprintf("unhandled type [%t]", defargs_t))
			}
			sargs_tmp := make([]string, 0, len(sargs))
			for _, v := range sargs {
				vv := normalizeClass(v, alltmpl)
				//println("~~", v, vv, alltmpl)
				sargs_tmp = append(sargs_tmp, vv)
			}
			sargs = sargs_tmp
		}
	}
	name = clsname + "<" + strings.Join(sargs, ",")
	if name[len(name)-1] == '>' {
		name += " >" + suffix
	} else {
		name += ">" + suffix
	}
	return name
}

type stldeftable_t map[string]interface{} //}[]string

var g_stldeftable stldeftable_t = stldeftable_t{
	"deque":        []string{"=", "std::allocator"},
	"list":         []string{"=", "std::allocator"},
	"map":          []string{"=", "=", "std::less", "std::allocator"},
	"multimap":     []string{"=", "=", "std::less", "std::allocator"},
	"queue":        []string{"=", "std::queue"},
	"set":          []string{"=", "std::less", "std::allocator"},
	"multiset":     []string{"=", "std::less", "std::allocator"},
	"stack":        []string{"=", "std::queue"},
	"vector":       []string{"=", "std::allocator"},
	"basic_string": []string{"=", "std::char_traits", "std::allocator"},
	//"basic_ostream": {"=", "std::char_traits"},
	//"basic_istream": {"=", "std::char_traits"},
	//"basic_streambuf": {"=", "std::char_traits"},
	//
	//FIXME: handle win32/gcc
	// if win32:
	// "hash_set":      {"=", "stdext::hash_compare", "std::allocator"},
	// "hash_multiset":      {"=", "stdext::hash_compare", "std::allocator"},
	// "hash_map":      {"=", "=", "stdext::hash_compare", "std::allocator"},
	// "hash_multimap":      {"=", "=", "stdext::hash_compare", "std::allocator"},
	//else:
	"hash_set":      []string{"=", "__gnu_cxx::hash", "std::equal_to", "std::allocator"},
	"hash_multiset": []string{"=", "__gnu_cxx::hash", "std::equal_to", "std::allocator"},
	"hash_map":      []string{"=", "=", "__gnu_cxx::hash", "std::equal_to", "std::allocator"},
	"hash_multimap": []string{"=", "=", "__gnu_cxx::hash", "std::equal_to", "std::allocator"},
}

func normalizeClass(name string, alltempl bool) string {
	names := make([]string, 0)
	count := 0
	// special cases:
	// A< (0>1) >::b
	// A< b::c >
	// a< b::c >::d< e::f >

	for _, s := range strings.Split(name, "::") {
		if count == 0 {
			names = append(names, s)
		} else {
			names[len(names)-1] += "::" + s
		}
		count += strings.Count(s, "<") + strings.Count(s, "(") -
			strings.Count(s, ">") - strings.Count(s, ")")
	}
	lst := make([]string, 0, len(names))
	for _, v := range names {
		vv := normalizeName(v, alltempl)
		//println("~+~", v, vv, alltempl)
		lst = append(lst, vv)
	}
	out := strings.Join(lst, "::")
	//println("**", name, out)
	return out
}

func getTemplateArgs(n string) (args []string) {
	args = []string{}
	beg := strings.Index(n, "<")
	if beg == -1 {
		return
	}
	end := strings.LastIndex(n, ">")
	if end == -1 {
		return
	}
	count := 0
	for _, s := range strings.Split(string(n[beg+1:end]), ",") {
		if count == 0 {
			args = append(args, s)
		} else {
			args[len(args)-1] += "," + s
		}
		count += strings.Count(s, "<") + strings.Count(s, "(") -
			strings.Count(s, ">") -
			strings.Count(s, ")")
	}

	if len(args) > 0 && len(string(args[len(args)-1])) > 0 {
		args[len(args)-1] = strings.TrimSpace(args[len(args)-1])
	}
	return args
}
