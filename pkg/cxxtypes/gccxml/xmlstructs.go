package gccxml

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
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
	fmt.Printf("loaded [%d] array-types\n", len(x.Arrays))

	// if g_dbg {
	// 	for i, _ := range x.FundamentalTypes {
	// 		fmt.Printf(" => %v\n", x.FundamentalTypes[i])
	// 	}
	// 	for _, a := range x.Arrays {
	// 		nn := genTypeName(a.id(), gtnCfg{})
	// 		fmt.Printf(" => '%s' -> '%s' %v\n", a.name(), nn, a)
	// 	}
	// }
}

// fixup fixes a few of the "features" of GCC-XML data
func (x *xmlTree) fixup() {

	for _, v := range x.Arrays {
		v.Max = strings.TrimRight(v.Max, "u")
	}

	// classes and structs
	for _, v := range x.Classes {
		v.fixup_name() // side-effects into fixing empty names...
		patchTemplateName(v)
		v.Members = strings.TrimSpace(v.Members)
	}
	for _, v := range x.Structs {
		v.fixup_name() // side-effects into fixing empty names...
		patchTemplateName(v)
		v.Members = strings.TrimSpace(v.Members)
	}

	// functions
	for _, v := range x.Functions {
		v.set_name(addTemplateToName(v.name(), v.demangled()))
		patchTemplateName(v)
	}

	// operators
	for _, v := range x.OperatorFunctions {
		name := v.name()
		if strings.HasPrefix(name, "operator") {
			if name[8] == ' ' {
				if isNum(name[9]) {
					demangled := v.Demangled
					pos := strings.Index(demangled, "::operator")
					if pos == -1 {
						pos = strings.Index(demangled, "operator")
					} else {
						pos += 2 // for the "::"
					}
					end := strings.Index(demangled, "(")
					if end == -1 {
						end = len(demangled) - 1
					}
					name = demangled[pos:end]
					//fmt.Printf(";-; [%s] -> [%s] (%s) [%d:%d]\n", demangled, name, v.name(), pos, end)
					v.set_name(name)
				} else if !isAlpha(name[9]) && name[9] != '_' {
					v.set_name("operator" + name[9:])
				}
			}
		} else {
			if isAlpha(name[0]) {
				v.set_name("operator " + name)
			} else {
				v.set_name("operator" + name)
			}
		}
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

	// dtors
	for _, v := range x.Destructors {
		if len(v.name()) > 1 && string(v.name()[0]) != "~" {
			v.set_name("~" + v.name())
		}
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
		name := v.name()
		if strings.HasPrefix(name, "operator") {
			if name[8] == ' ' {
				if isNum(name[9]) {
					demangled := v.Demangled
					pos := strings.Index(demangled, "::operator")
					if pos == -1 {
						pos = strings.Index(demangled, "operator")
					} else {
						pos += 2 // for the "::"
					}
					end := strings.Index(demangled, "(")
					if end == -1 {
						end = len(demangled) - 1
					}
					name = demangled[pos:end]
					//fmt.Printf(";;; [%s] -> [%s] (%s) [%d:%d]\n", demangled, name, v.name(), pos, end)
					v.set_name(name)
				} else if !isAlpha(name[9]) && name[9] != '_' {
					v.set_name("operator" + name[9:])
				}
			}
		} else {
			if isAlpha(name[0]) {
				v.set_name("operator " + name)
			} else {
				v.set_name("operator" + name)
			}
		}
		if !strings.HasPrefix(v.name(), "_ZT") {
			v.set_name(addTemplateToName(v.name(), v.demangled()))
			patchTemplateName(v)
		}
	}

	// op-methods
	for _, v := range x.Converters {
		name := v.name()
		if strings.HasPrefix(name, "operator") {
			if name[8] == ' ' {
				if isNum(name[9]) {
					demangled := v.Demangled
					pos := strings.Index(demangled, "::operator")
					if pos == -1 {
						pos = strings.Index(demangled, "operator")
					} else {
						pos += 2 // for the "::"
					}
					end := strings.Index(demangled, "(")
					if end == -1 {
						end = len(demangled) - 1
					}
					name = demangled[pos:end]
					//fmt.Printf(";;; [%s] -> [%s] (%s) [%d:%d]\n", demangled, name, v.name(), pos, end)
					v.set_name(name)
				} else if !isAlpha(name[9]) && name[9] != '_' {
					v.set_name("operator" + name[9:])
				}
			}
		} else {
			if isAlpha(name[0]) {
				v.set_name("operator " + name)
			} else {
				v.set_name("operator" + name)
			}
		}
		if !strings.HasPrefix(v.name(), "_ZT") {
			v.set_name(addTemplateToName(v.name(), v.demangled()))
			patchTemplateName(v)
		}
	}

	// // variables
	// for _, v := range x.Variables {
	// 	name := v.name()
	// 	if len(name) >= 4 && name[:4] == "_ZTV" {

	// 	}
	// }

	// fields
	for _, v := range x.Fields {
		v.fixup_name()
	}

	// builtins
	for _, v := range x.FundamentalTypes {
		alltmpl := false
		v.set_name(normalizeName(v.name(), alltmpl))
	}

}

func (x *xmlTree) gencxxtypes() error {

	// first, generate builtins.
	for _, v := range x.FundamentalTypes {
		if v.name() == "" {
			panic("empty builtin type name !")
		}
		tk, ok := g_n2tk[v.name()]
		if !ok {
			panic("no such builtin type [" + v.name() + "]")
		}
		ct := cxxtypes.NewFundamentalType(
			v.name(),
			str_to_uintptr(v.Size),
			tk,
			"::", //scope,
		)
		g_processed_ids[v.id()] = ct.TypeName()
	}

	for _, v := range x.Namespaces {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Arrays {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Classes {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Constructors {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Converters {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.CvQualifiedTypes {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Destructors {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Enumerations {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Functions {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.FunctionTypes {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.MethodTypes {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Methods {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.OperatorFunctions {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.OperatorMethods {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.PointerTypes {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.ReferenceTypes {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Structs {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Typedefs {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	for _, v := range x.Unions {
		//fmt.Printf("\n%s... (%s) [%v]\n", v.name(), v.id(), genTypeName(v.id(), gtnCfg{}))
		gen_id_from_gccxml(v)
		//fmt.Printf("%v\n", t)
	}

	return nil
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
	return fmt.Sprintf("%s[%s]", tt.name(), x.Max)
}

func (x *xmlArray) set_name(n string) {
	panic("fixme: can set name of an array yet")
}

func (x *xmlArray) context() string {
	t := g_ids[x.Type]
	return t.(i_context).context()
}

func (x *xmlArray) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_ConstantArray
}

func (x *xmlArray) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlArray) size() uintptr {
	return str_to_uintptr(x.Size)
}

func (x *xmlArray) typename() string {
	return genTypeName(x.Id, gtnCfg{})
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

func (x *xml_record) fixup_name() {
	if x.Name == "" && x.Demangled != "" {
		x.Name = x.Demangled
	}
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

func (x *xml_record) context() string {
	return x.Context
}

func (x *xml_record) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_Record
}

func (x *xml_record) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xml_record) typename() string {
	return genTypeName(x.id(), gtnCfg{})
}

func (x *xml_record) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xml_record) offset() uintptr {
	return 0
}

func (x *xml_record) size() uintptr {
	return str_to_uintptr(x.Size)
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

func (x *xmlConstructor) context() string {
	return x.Context
}

func (x *xmlConstructor) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlConstructor) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Fct
}

func (x *xmlConstructor) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlConstructor) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlConstructor) offset() uintptr {
	return 0
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

func (x *xmlConverter) context() string {
	return x.Context
}

func (x *xmlConverter) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlConverter) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Fct
}

func (x *xmlConverter) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlConverter) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlConverter) offset() uintptr {
	return 0
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

func (x *xmlCvQualifiedType) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlCvQualifiedType) set_name(n string) {
	panic("fixme: can set the name of a cv-qualified type - yet")
}

func (x *xmlCvQualifiedType) context() string {
	t := g_ids[x.Type]
	return t.(i_context).context()
}

func (x *xmlCvQualifiedType) kind() cxxtypes.TypeKind {
	t := g_ids[x.Type]
	return t.(i_kind).kind()
}

func (x *xmlCvQualifiedType) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlCvQualifiedType) size() uintptr {
	return str_to_uintptr(x.Size)
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

func (x *xmlDestructor) context() string {
	return x.Context
}

func (x *xmlDestructor) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlDestructor) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Fct
}

func (x *xmlDestructor) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlDestructor) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlDestructor) offset() uintptr {
	return 0
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

func (x *xmlEnumeration) context() string {
	return x.Context
}

func (x *xmlEnumeration) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_Enum
}

func (x *xmlEnumeration) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlEnumeration) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlEnumeration) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlEnumeration) offset() uintptr {
	return 0
}

func (x *xmlEnumeration) size() uintptr {
	return str_to_uintptr(x.Size)
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

func (x *xmlField) fixup_name() {
	//TODO: figure out why we get such weirdos...
	// ex: struct __vmi_class_type_info_pseudo1 
	//     has a number of unnamed fields
	if x.Name == "" && x.Demangled == "" {
		x.Name = "__fake__name__" + x.Id + "__"
		x.Demangled = x.Name
	}
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

func (x *xmlField) context() string {
	return x.Context
}

func (x *xmlField) kind() cxxtypes.TypeKind {
	t := g_ids[x.Type]
	return t.(i_kind).kind()
}

func (x *xmlField) idkind() cxxtypes.IdKind {
	t := g_ids[x.Type]
	return t.(i_idkind).idkind()
}

func (x *xmlField) typename() string {
	return genTypeName(x.Type, gtnCfg{})
}

func (x *xmlField) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlField) offset() uintptr {
	return str_to_uintptr(x.Offset)
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

func (x *xmlFunction) typename() string {
	return genTypeName(x.Id, gtnCfg{})
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

func (x *xmlFunction) context() string {
	return x.Context
}

func (x *xmlFunction) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlFunction) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Fct
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

func (x *xmlFunctionType) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlFunctionType) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlFunctionType) name() string {
	//FIXME - we should perhaps return "id-name (argtype...)"
	return x.Id
}

func (x *xmlFunctionType) set_name(n string) {
	panic("cannot set name of a functiontype")
}

func (x *xmlFunctionType) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlFunctionType) context() string {
	return g_globalns_id
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

func (x *xmlFundamentalType) context() string {
	return g_globalns_id
}

func (x *xmlFundamentalType) kind() cxxtypes.TypeKind {
	return g_n2tk[x.name()]
}

func (x *xmlFundamentalType) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlFundamentalType) size() uintptr {
	return str_to_uintptr(x.Size)
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

func (x *xmlMethod) context() string {
	return x.Context
}

func (x *xmlMethod) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlMethod) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Fct
}

func (x *xmlMethod) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlMethod) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlMethod) offset() uintptr {
	return 0
}

type xmlMethodType struct {
	Attributes string `xml:"attributes,attr"` // implied
	BaseType   string `xml:"basetype,attr"`
	Const      string `xml:"const,attr"`
	Id         string `xml:"id,attr"`
	Returns    string `xml:"returns,attr"`
	Volatile   string `xml:"volatile,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlMethodType) id() string {
	return x.Id
}

func (x *xmlMethodType) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlMethodType) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlMethodType) name() string {
	return genTypeName(x.id(), gtnCfg{})
}

func (x *xmlMethodType) set_name(n string) {
	panic("gccxml: cannot set name on MethodType")
}

func (x *xmlMethodType) typename() string {
	return genTypeName(x.Id, gtnCfg{})
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

func (x *xmlNamespace) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlNamespace) set_demangled(n string) {
	x.Demangled = n
}

func (x xmlNamespace) String() string {
	return fmt.Sprintf(`Namespace{%s}`, x.Name)
}

func (x *xmlNamespace) context() string {
	return x.Context
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

func (x *xmlNamespaceAlias) context() string {
	return x.Context
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
	Inline     string `xml:"inline,attr"`
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

func (x *xmlOperatorFunction) typename() string {
	return genTypeName(x.Id, gtnCfg{})
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

func (x *xmlOperatorFunction) context() string {
	return x.Context
}

func (x *xmlOperatorFunction) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_FunctionProto
}

func (x *xmlOperatorFunction) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Fct
}

type xmlOperatorMethod struct {
	xmlMethod
}

func (x *xmlOperatorMethod) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Fct
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

func (x *xmlPointerType) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlPointerType) context() string {
	t := g_ids[x.Type]
	return t.(i_context).context()
}

func (x *xmlPointerType) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_Ptr
}

func (x *xmlPointerType) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlPointerType) size() uintptr {
	return str_to_uintptr(x.Size)
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

func (x *xmlReferenceType) name() string {
	t := g_ids[x.Type]
	switch tt := t.(type) {
	case i_name:
		return tt.name() + "&"
	default:
	}
	return "<unnamed>&"
}

func (x *xmlReferenceType) set_name(n string) {
	panic("fixme: can set a name for a ref-type - yet")
}

func (x *xmlReferenceType) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlReferenceType) context() string {
	t := g_ids[x.Type]
	return t.(i_context).context()
}

func (x *xmlReferenceType) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_LValueRef
}

func (x *xmlReferenceType) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlReferenceType) size() uintptr {
	return str_to_uintptr(x.Size)
}

type xmlStruct struct {
	xml_record
}

func (x xmlStruct) String() string {
	return "Struct" + x.xml_record.infos()
}

type xmlTypedef struct {
	Access     string `xml:"access,attr"`     // default "public"
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

func (x *xmlTypedef) context() string {
	return x.Context
}

func (x *xmlTypedef) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_Typedef
}

func (x *xmlTypedef) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlTypedef) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlTypedef) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlTypedef) offset() uintptr {
	return 0
}

func (x *xmlTypedef) size() uintptr {
	t := g_ids[x.Type].(i_size)
	return t.size()
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
	Members    string `xml:"members,attr"`
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

func (x *xmlUnion) context() string {
	return x.Context
}

func (x *xmlUnion) kind() cxxtypes.TypeKind {
	return cxxtypes.TK_Record
}

func (x *xmlUnion) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Typ
}

func (x *xmlUnion) typename() string {
	return genTypeName(x.Id, gtnCfg{})
}

func (x *xmlUnion) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlUnion) offset() uintptr {
	return 0
}

func (x *xmlUnion) size() uintptr {
	return str_to_uintptr(x.Size)
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

func (x *xmlVariable) context() string {
	return x.Context
}

func (x *xmlVariable) kind() cxxtypes.TypeKind {
	t := g_ids[x.Type]
	return t.(i_kind).kind()
}

func (x *xmlVariable) idkind() cxxtypes.IdKind {
	return cxxtypes.IK_Var
}

func (x *xmlVariable) typename() string {
	return genTypeName(x.Type, gtnCfg{})
}

func (x *xmlVariable) access() cxxtypes.AccessSpecifier {
	return str_to_access(x.Access)
}

func (x *xmlVariable) offset() uintptr {
	return 0
}

// utils ---

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore
func isAlphaNum(c uint8) bool {
	return c == '_' ||
		('0' <= c && c <= '9') ||
		('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z')
}

// isNum reports whether the byte is a number
func isNum(c uint8) bool {
	return ('0' <= c && c <= '9')
}

// isAlpha reports whether the byte is an ASCII letter
func isAlpha(c uint8) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// str_to_uintptr returns a uintptr from a string
func str_to_uintptr(s string) uintptr {
	if s == "" {
		return uintptr(0)
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	if i < 0 {
		panic(strconv.ErrRange)
	}
	return uintptr(i)
}

// str_to_bool returns a bool from a string
func str_to_bool(s string) bool {
	if s == "" || s == "0" {
		return false
	}
	return true
}

// str_to_access returns a cxxtypes.AccessSpecifier from a string
func str_to_access(s string) cxxtypes.AccessSpecifier {
	switch s {
	case "", "public":
		return cxxtypes.AS_Public
	case "protected":
		return cxxtypes.AS_Protected
	case "private":
		return cxxtypes.AS_Private
	default:
		panic(fmt.Sprintf("gccxml: unhandled access-string [%s]", s))
	}
	panic("unreachable")
}

// addTemplateToName
func addTemplateToName(n, demangled string) string {
	name := n
	posargs := strings.LastIndex(demangled, "(")
	if posargs > 1 && demangled[posargs-1] == '>' &&
		(isAlphaNum(demangled[posargs-2]) ||
			demangled[posargs-2] == '_') {
		posname := strings.Index(demangled, n+"<")
		if posname >= 0 {
			reui := regexp.MustCompile(`\b(unsigned)(\s+)?([^\w\s])`)
			nn := string(demangled[posname:posargs])
			name = reui.ReplaceAllString(nn, `unsigned int\3`)
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
	n.set_name(name)
	//println("--", name, tmpl, n.name())

	return
}

func normalizeName(n string, alltmpl bool) string {
	name := strings.TrimSpace(n)
	if !strings.Contains(name, "<") {

		if strings.Contains(name, "int") {
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
		}

		if strings.Contains(name, "complex") {
			for _, v := range [][]string{
				{"complex float", "float complex"},
				{"complex double", "double complex"},
				{"complex long double", "long double complex"},
			} {
				name = strings.Replace(name, v[0], v[1], -1)
			}
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

// gtnCfg eases the calling of genTypeName
type gtnCfg struct {
	enum       bool
	const_veto bool
	colon      bool
	alltmpl    bool
}

func genTypeName(id string, cfg gtnCfg) string {
	//fmt.Printf("==gtn==[%s]...\n", id)
	gidx := gidname{id, cfg}
	if gid, ok := g_ids_name[gidx]; ok {
		//fmt.Printf("==gtn==[%s]... (from cache: %s)\n", id, gid)
		return gid
	}

	node := g_ids[id]
	if dnode, ok := node.(i_mangler); ok {
		if isUnnamedType(dnode.demangled()) {
			if cfg.colon {
				g_ids_name[gidx] = "__" + dnode.demangled()
				return g_ids_name[gidx]
			} else {
				g_ids_name[gidx] = dnode.demangled()
				return g_ids_name[gidx]
			}
		}
	}
	if id == "" {
		panic("empty id...")
	}

	nid := ""
	if strings.HasSuffix(id, "r") {
		g_ids_name[gidx] = "restrict " + genTypeName(id[:len(id)-1], cfg)
		return g_ids_name[gidx]
	}
	if strings.HasSuffix(id, "v") ||
		strings.HasSuffix(id, "c") {
		nid = id[:len(id)-1]
		nnid := string(nid[len(nid)-1])
		if strings.ContainsAny(nnid, "cv") {
			nid = nid[:len(nid)-1]
		}

		cvdict := map[uint8]string{
			'c': "const",
			'v': "volatile",
			'r': "restrict",
		}
		tn := ""
		switch node.(type) {
		case *xmlPointerType:
			if cfg.const_veto {
				tn = genTypeName(nid,
					gtnCfg{
						enum:       cfg.enum,
						const_veto: false,
						colon:      cfg.colon,
						alltmpl:    cfg.alltmpl,
					})
			} else {
				tn = genTypeName(nid,
					gtnCfg{
						enum:       cfg.enum,
						const_veto: false,
						colon:      cfg.colon,
						alltmpl:    cfg.alltmpl,
					}) + " " + cvdict[id[len(id)-1]]
			}
		case *xmlReferenceType:
			if cfg.const_veto {
				tn = genTypeName(nid,
					gtnCfg{
						enum:       cfg.enum,
						const_veto: false,
						colon:      cfg.colon,
						alltmpl:    cfg.alltmpl,
					})
			} else {
				tn = cvdict[id[len(id)-1]] + " " + genTypeName(nid,
					gtnCfg{
						enum:       cfg.enum,
						const_veto: false,
						colon:      cfg.colon,
						alltmpl:    cfg.alltmpl,
					})
			}
		case *xmlCvQualifiedType:
			if cfg.const_veto {
				tn = genTypeName(nid,
					gtnCfg{
						enum:       cfg.enum,
						const_veto: false,
						colon:      cfg.colon,
						alltmpl:    cfg.alltmpl,
					})
			} else {
				tn = genTypeName(nid,
					gtnCfg{
						enum:       cfg.enum,
						const_veto: false,
						colon:      cfg.colon,
						alltmpl:    cfg.alltmpl,
					}) + " " + cvdict[id[len(id)-1]]
			}

		default:
			panic("unreachable")
		}
		//fmt.Printf("^^^ %s [%s|%s] (t=%T)'%s' %v\n", id, nid, tn, node, cvdict[id[len(id)-1]], cfg.const_veto)
		g_ids_name[gidx] = tn
		return g_ids_name[gidx]
	}

	//println("***|", id)
	s := genScopeName(id, cfg)
	//println("***||", id)
	switch tt := node.(type) {
	case *xmlNamespace:
		if tt.name() == "" {
			s += "@anonymous@namespace@"
		} else if tt.name() != "::" {
			s += tt.name()
		}

	case *xmlPointerType:
		tn := genTypeName(tt.Type, cfg)
		if strings.HasSuffix(tn, ")") ||
			strings.HasSuffix(tn, ") const") ||
			strings.HasSuffix(tn, ") volatile") {
			tn = strings.Replace(tn, "::*)", "::**)", -1)
			tn = strings.Replace(tn, "::)", "::*)", -1)
			tn = strings.Replace(tn, "(*)", "(**)", -1)
			tn = strings.Replace(tn, "()", "(*)", 1)
			// 's =' is on purpose!
			s = tn
		} else {
			// 's =' is on purpose!
			s = tn + "*"
		}

	case *xmlReferenceType:
		//println("-+-+-&", tt.Type)
		tn := genTypeName(tt.Type, cfg)
		// 's =' is on purpose!
		s = tn + "&"
		//fmt.Printf(";;; tn=[%s] -> ref=[%s]\n", tn, s)

	case *xmlFunctionType:
		// 's =' is on purpose!
		s = genTypeName(tt.Returns, cfg)
		s += "()("
		if len(tt.Arguments) > 0 {
			for i, arg := range tt.Arguments {
				s += genTypeName(arg.Type, cfg)
				if i < len(tt.Arguments)-1 {
					s += ", "
				}
			}
			s += ")"
		} else {
			s += "void)"
		}

	case *xmlMethodType:
		// 's =' is on purpose!
		s = genTypeName(tt.Returns, cfg)
		s += "(" + genTypeName(tt.BaseType, cfg) + "::)("
		if len(tt.Arguments) > 0 {
			for i, arg := range tt.Arguments {
				s += genTypeName(arg.Type, cfg)
				if i < len(tt.Arguments)-1 {
					s += ", "
				}
			}
		} else {
			s += "void)"
		}
		if str_to_bool(tt.Const) {
			s += " const"
		}
		if str_to_bool(tt.Volatile) {
			s += " volatile"
		}

	case *xmlArray:
		max := strings.TrimRight(tt.Max, "u")
		arr := "[]"
		if max != "" {
			dim, err := strconv.Atoi(max)
			if err != nil {
				panic(err)
			}
			arr = fmt.Sprintf("[%d]", dim+1)
		}
		//println("--array--, dim:", arr, "type:", tt.Type)
		tn := genTypeName(tt.Type, cfg)
		// 's =' is on purpose!
		if strings.HasSuffix(tn, "]") {
			pos := strings.Index(tn, "[")
			// 's =' is on purpose!
			s = tn[:pos] + arr + tn[pos:]
		} else {
			// 's =' is on purpose!
			s = tn + arr
		}

	case *xmlUnimplemented:
		s += tt.TreeCodeName

	case *xmlEnumeration:
		if cfg.enum {
			// 's =' is on purpose!
			s = "int" // replace "enum type" with "int"
		} else {
			s += tt.Name // FIXME: not always true
		}

	case *xmlTypedef:
		// on purpose
		s = genScopeName(tt.id(), cfg)
		s += tt.name()

	case *xmlFunction:
		s += tt.name()

	case *xmlOperatorFunction:
		s += tt.name()

	case *xmlOffsetType:
		s += genTypeName(tt.Type, cfg) + " "
		s += genTypeName(tt.BaseType, cfg) + "::"
		//FIXME:
		// OffsetType A::*, different treatment for GCCXML 0.7 and 0.9:
		// 0.7: basetype: A*
		// 0.9: basetype: A - add a "*" here.
		s += "*"

	default:
		if tn, ok := tt.(i_name); ok {
			s += tn.name()
		}
		// normalize STL class namespaces, primitives, etc...
		s = normalizeClass(s, cfg.alltmpl)
	}
	g_ids_name[gidx] = s
	return s
}

func isUnnamedType(name string) bool {
	// ellipsis would screw us up...
	name = strings.Replace(name, "...", "", -1)
	if strings.IndexAny(name, ".$") != -1 {
		return true
	}
	return false
}

func genScopeName(id string, cfg gtnCfg) string {
	s := ""
	if t, ok := g_ids[id].(i_context); ok {
		id = t.context()
	} else {
		panic("gccxml: no context for id=" + id + "!!")
	}

	ns := ""
	if id != "" {
		ns = genTypeName(id, cfg)
	}
	if ns != "" {
		s = ns + "::"
	} else if cfg.colon {
		s = "::"
	}
	return s
}

// getScopeChainIds returns the chain of context-ids from the global-one to the current node.
func getScopeChainIds(node i_context) []string {
	scopeids := make([]string, 0, 1)
	for {
		scopeids = append(scopeids, node.id())
		pid := node.context()
		if pid == "" {
			// we reached the global namespace
			break
		}
		p := g_ids[node.context()].(i_context)
		node = p
	}
	// Reverse scopeids
	for i, j := 0, len(scopeids)-1; i < j; i, j = i+1, j-1 {
		scopeids[i], scopeids[j] = scopeids[j], scopeids[i]
	}
	return scopeids
}

// getScopeChainNames returns the chain of scope names from the global-one to the current node.
func getScopeChainNames(node i_context) []string {
	scopeids := getScopeChainIds(node)
	scopenames := make([]string, 0, len(scopeids))
	for _, id := range scopeids {
		//n := g_ids[id].(i_name).name()
		n := genTypeName(g_ids[id].id(), gtnCfg{})
		scopenames = append(scopenames, n)
	}
	if scopenames[0] == "" {
		scopenames[0] = "::"
	}
	return scopenames
}

// getCxxtypesScope
func getCxxtypesScope(node i_id) string {
	scope := "::"
	if n, ok := node.(i_context); ok {
		ctxt := n.context()
		if ctxt == "" {
			return scope
		}
		return genTypeName(ctxt, gtnCfg{})
	}
	return scope
}

func gen_id_from_gccxml(node i_id) cxxtypes.Id {

	// has that type already been processed ?
	if tname, ok := g_processed_ids[node.id()]; ok {
		return cxxtypes.IdByName(tname)
	}

	// are we processing that id ?
	if proc, ok := g_processing_ids[node.id()]; ok && proc {
		n := genTypeName(node.id(), gtnCfg{})
		//FIXME: panic or not ?
		panic("placeholder:" + n)
		return cxxtypes.NewPlaceHolder(n).(cxxtypes.Id)
	}

	// mark for processing:
	g_processing_ids[node.id()] = true

	var ct cxxtypes.Id = nil

	gen_mbrs := func(mbrs string, scope string) []cxxtypes.Member {
		members := make([]cxxtypes.Member, 0)
		if mbrs == "" {
			return members
		}

		mbrs = strings.TrimSpace(mbrs)
		for _, mbrid := range strings.Split(mbrs, " ") {
			tmbr, ok := g_ids[mbrid]
			if !ok {
				panic(fmt.Sprintf("gccxml: no such id [%s]", mbrid))
			}
			name := genTypeName(tmbr.id(), gtnCfg{})
			mbr := tmbr.(i_field)
			mbr_idkind := mbr.idkind()
			if _, ok := tmbr.(*xmlField); ok {
				mbr_idkind = cxxtypes.IK_Var
			}
			if name == "EnumNs::ESTLtype" {
				fmt.Printf("++ [%s] tn=[%s], idkind=%v, tkind=%v scope=%s\n",
					name, mbr.typename(), mbr_idkind, mbr.kind(), scope)
			}
			members = append(members,
				cxxtypes.NewMember(
					name,
					mbr.typename(),
					mbr_idkind,
					mbr.kind(),
					mbr.access(),
					mbr.offset(),
					scope,
				))
		}
		return members
	}

	gen_enum_mbrs := func(mbrs []xmlEnumValue, scope string) []cxxtypes.Member {
		members := make([]cxxtypes.Member, 0)
		if len(mbrs) == 0 {
			return members
		}

		// fixme: that's not always true... (could be unsigned int...)
		typ := "int"
		for _, mbr := range mbrs {
			n := ""
			if scope == "" || scope == "::" {
				n = mbr.Name
			} else {
				n = strings.Join([]string{scope, mbr.Name}, "::")
			}
			members = append(members,
				cxxtypes.NewMember(
					n,
					typ,
					cxxtypes.IK_Var,
					cxxtypes.TK_Int,
					cxxtypes.AS_Public,
					uintptr(0),
					scope,
				))
		}
		return members
	}

	gen_args := func(args []xmlArgument) []cxxtypes.Parameter {
		params := make([]cxxtypes.Parameter, 0, len(args))
		for _, arg := range args {
			tn := gen_id_from_gccxml(g_ids[arg.Type]).(cxxtypes.Type).TypeName()
			p := cxxtypes.NewParameter(
				arg.Name,
				tn,
				arg.Default != "",
			)
			params = append(params, *p)
		}
		return params
	}

	gen_bases := func(xbases []xmlBase) []cxxtypes.Base {
		bases := make([]cxxtypes.Base, 0)
		for _, b := range xbases {
			access := str_to_access(b.Access)
			typ := gen_id_from_gccxml(g_ids[b.Type]).(cxxtypes.Type).TypeName()
			offset := str_to_uintptr(b.Offset)
			virtual := str_to_bool(b.Virtual)
			bases = append(bases, cxxtypes.NewBase(offset, typ, access, virtual))
		}
		return bases
	}

	switch t := node.(type) {

	case *xmlFundamentalType:
		ct = cxxtypes.IdByName(t.name())

	case *xmlArray:
		sz := str_to_uintptr(t.Size)
		typ := gen_id_from_gccxml(g_ids[t.Type]).(cxxtypes.Type)
		tn := typ.TypeName()
		tsz := typ.TypeSize()
		scope := getCxxtypesScope(t)
		ct = cxxtypes.NewArrayType(sz, tn, tsz, scope)

	case *xmlConstructor:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("--(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		qual := cxxtypes.TQ_None
		spec := cxxtypes.TS_Method | cxxtypes.TS_Constructor
		if str_to_bool(t.Extern) {
			spec |= cxxtypes.TS_Extern
		}
		variadic := strings.Contains(scoped_name, "...")

		scope := getCxxtypesScope(t)
		params := gen_args(t.Arguments)
		ret_type := "void" //FIXME ?
		ct = cxxtypes.NewFunction(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type,
			scope,
		)

	case *xmlConverter:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("--(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		qual := cxxtypes.TQ_None
		spec := cxxtypes.TS_Method | cxxtypes.TS_Converter
		if str_to_bool(t.Extern) {
			spec |= cxxtypes.TS_Extern
		}
		variadic := strings.Contains(scoped_name, "...")

		scope := getCxxtypesScope(t)
		params := []cxxtypes.Parameter{}
		ret_type := gen_id_from_gccxml(g_ids[t.Returns]).(cxxtypes.Type)
		ct = cxxtypes.NewFunction(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type.TypeName(),
			scope,
		)

	case *xmlCvQualifiedType:
		//sz := str_to_uintptr(t.Size)
		typ := gen_id_from_gccxml(g_ids[t.Type]).(cxxtypes.Type)
		scope := getCxxtypesScope(t)
		qual := cxxtypes.TQ_None
		if str_to_bool(t.Const) {
			qual |= cxxtypes.TQ_Const
		}
		if str_to_bool(t.Restrict) {
			qual |= cxxtypes.TQ_Restrict
		}
		if str_to_bool(t.Volatile) {
			qual |= cxxtypes.TQ_Volatile
		}
		n := genTypeName(t.id(), gtnCfg{})
		ct = cxxtypes.NewQualType(n, typ.TypeName(), scope, qual).(cxxtypes.Id)

	case *xmlPointerType:
		typ := gen_id_from_gccxml(g_ids[t.Type]).(cxxtypes.Type).TypeName()
		scope := getCxxtypesScope(t)
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("--(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		ct = cxxtypes.NewPtrType(scoped_name, typ, scope)

	case *xmlReferenceType:
		tn := genTypeName(t.Type, gtnCfg{})
		//typ := gen_id_from_gccxml(g_ids[t.Type]).(cxxtypes.Type).TypeName()
		scope := getCxxtypesScope(t)
		scoped_name := genTypeName(t.id(), gtnCfg{})
		ct = cxxtypes.NewRefType(scoped_name, tn, scope)

	case *xmlDestructor:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("--(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		qual := cxxtypes.TQ_None
		spec := cxxtypes.TS_Method | cxxtypes.TS_Destructor
		if str_to_bool(t.Extern) {
			spec |= cxxtypes.TS_Extern
		}
		if str_to_bool(t.Virtual) {
			spec |= cxxtypes.TS_Virtual
		}
		variadic := strings.Contains(scoped_name, "...")

		params := []cxxtypes.Parameter{}
		ret_type := "void" //FIXME ?
		//ret_type := cxxtypes.IdByName("void").(cxxtypes.Type) //FIXME ?
		scope := getCxxtypesScope(t)
		ct = cxxtypes.NewFunction(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type,
			scope,
		)

	case *xmlEnumeration:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("--enum(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		scope := getCxxtypesScope(t)
		// do note that the enum-values "leak" into the scope holding the
		// declaration of the enum-type.
		mbrs := gen_enum_mbrs(t.EnumValues, scope)
		ct = cxxtypes.NewEnumType(scoped_name, mbrs, scope)

	case *xmlFunctionType:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("--fct-type(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		qual := cxxtypes.TQ_None
		spec := cxxtypes.TS_None
		variadic := strings.Contains(scoped_name, "...")

		scope := getCxxtypesScope(t)
		params := gen_args(t.Arguments)
		ret_type := gen_id_from_gccxml(g_ids[t.Returns]).(cxxtypes.Type)
		ct = cxxtypes.NewFunctionType(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type.TypeName(),
			scope,
		)

	case *xmlFunction:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		qual := cxxtypes.TQ_None
		spec := cxxtypes.TS_None
		variadic := strings.Contains(scoped_name, "...")

		scope := getCxxtypesScope(t)
		params := gen_args(t.Arguments)
		ret_type := gen_id_from_gccxml(g_ids[t.Returns]).(cxxtypes.Type)
		ct = cxxtypes.NewFunction(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type.TypeName(),
			scope,
		)

	case *xmlStruct:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		sz := str_to_uintptr(t.Size)
		scope := getCxxtypesScope(t)
		st := cxxtypes.NewStructType(scoped_name, sz, scope)
		// un-mark from processing:
		delete(g_processing_ids, node.id())
		g_processed_ids[node.id()] = st.TypeName()
		//
		st.SetMembers(gen_mbrs(t.Members, scoped_name))
		st.SetBases(gen_bases(t.Bases))
		ct = st

	case *xmlClass:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		sz := str_to_uintptr(t.Size)
		scope := getCxxtypesScope(t)
		st := cxxtypes.NewClassType(scoped_name, sz, scope)
		// un-mark from processing:
		delete(g_processing_ids, node.id())
		g_processed_ids[node.id()] = st.TypeName()
		//
		st.SetMembers(gen_mbrs(t.Members, scoped_name))
		st.SetBases(gen_bases(t.Bases))
		ct = st

	case *xmlMethod:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("++(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		qual := cxxtypes.TQ_None
		if str_to_bool(t.Const) {
			qual |= cxxtypes.TQ_Const
		}
		spec := cxxtypes.TS_Method
		if str_to_bool(t.Extern) {
			spec |= cxxtypes.TS_Extern
		}
		if str_to_bool(t.Static) {
			spec |= cxxtypes.TS_Static
		}
		if str_to_bool(t.Virtual) {
			spec |= cxxtypes.TS_Virtual
		}
		variadic := strings.Contains(scoped_name, "...")

		scope := getCxxtypesScope(t)
		params := gen_args(t.Arguments)
		//ret_type := gen_id_from_gccxml(g_ids[t.Returns]).(cxxtypes.Type)
		ret_type := genTypeName(t.Returns, gtnCfg{})
		ct = cxxtypes.NewFunction(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type,
			scope,
		)

	case *xmlNamespace:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		scope := getCxxtypesScope(t)
		ct = cxxtypes.NewNamespace(scoped_name, scope)

	case *xmlOperatorMethod:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("+-(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		qual := cxxtypes.TQ_None
		if str_to_bool(t.Const) {
			qual |= cxxtypes.TQ_Const
		}
		spec := cxxtypes.TS_Method | cxxtypes.TS_Operator
		if str_to_bool(t.Extern) {
			spec |= cxxtypes.TS_Extern
		}
		if str_to_bool(t.Static) {
			spec |= cxxtypes.TS_Static
		}
		if str_to_bool(t.Virtual) {
			spec |= cxxtypes.TS_Virtual
		}
		variadic := strings.Contains(scoped_name, "...")

		scope := getCxxtypesScope(t)
		params := gen_args(t.Arguments)
		ret_type := gen_id_from_gccxml(g_ids[t.Returns]).(cxxtypes.Type)
		ct = cxxtypes.NewFunction(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type.TypeName(),
			scope,
		)

	case *xmlOperatorFunction:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//fmt.Printf("+-op-fct(%s)[%s][%s]...\n", t.id(), t.name(), scoped_name)
		qual := cxxtypes.TQ_None
		spec := cxxtypes.TS_Operator
		if str_to_bool(t.Extern) {
			spec |= cxxtypes.TS_Extern
		}
		if str_to_bool(t.Inline) {
			spec |= cxxtypes.TS_Inline
		}
		variadic := strings.Contains(scoped_name, "...")

		scope := getCxxtypesScope(t)
		params := gen_args(t.Arguments)
		ret_type := gen_id_from_gccxml(g_ids[t.Returns]).(cxxtypes.Type)
		ct = cxxtypes.NewFunction(
			scoped_name,
			qual,
			spec,
			variadic,
			params,
			ret_type.TypeName(),
			scope,
		)

	case *xmlTypedef:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//typ := gen_id_from_gccxml(g_ids[t.Type]).(cxxtypes.Type)
		tn := genTypeName(t.Type, gtnCfg{}) //typ.TypeName()
		tsz := g_ids[t.Type].(i_size).size()
		scope := getCxxtypesScope(t)
		ct = cxxtypes.NewTypedefType(scoped_name, tn, tsz, scope)

	case *xmlUnion:
		scoped_name := genTypeName(t.id(), gtnCfg{})
		//sz := str_to_uintptr(t.Size)
		scope := getCxxtypesScope(t)
		//fmt.Printf("**> [%s] [%s] mbrs:[%s]\n", t.name(), scoped_name, t.Members)
		mbrs := gen_mbrs(t.Members, scoped_name)
		ct = cxxtypes.NewUnionType(scoped_name, mbrs, scope)

	default:
		panic(fmt.Sprintf("unhandled type [%T] (%s)", t, t.id()))
	}

	// un-mark from processing:
	delete(g_processing_ids, node.id())
	g_processed_ids[node.id()] = ct.IdName()
	return ct
}
