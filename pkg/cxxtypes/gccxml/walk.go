package gccxml

import (
	"fmt"
)

// An inspector can visit xmlTrees
type inspector func(node i_id) bool

func (f inspector) visit(node i_id) i_visitor {
	if f(node) {
		return f
	}
	return nil
}

// An i_visitor's visit method is invoked for each node encountered by walk
type i_visitor interface {
	visit(node i_id) (w i_visitor)
}

func walk(v i_visitor, node i_id) {
	if v = v.visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *xmlTree:
		for _,vv := range n.FundamentalTypes {
			walk(v, vv)
		}
		for _,vv := range n.Arrays {
			walk(v, vv)
		}
		for _,vv := range n.Classes {
			walk(v, vv)
		}
		for _,vv := range n.Constructors {
			walk(v, vv)
		}
		for _,vv := range n.Converters {
			walk(v, vv)
		}
		for _,vv := range n.CvQualifiedTypes {
			walk(v, vv)
		}
		for _,vv := range n.Destructors {
			walk(v, vv)
		}
		for _,vv := range n.Enumerations {
			walk(v, vv)
		}
		for _,vv := range n.Fields {
			walk(v, vv)
		}
		for _,vv := range n.Files {
			walk(v, vv)
		}
		for _,vv := range n.Functions {
			walk(v, vv)
		}
		for _,vv := range n.FunctionTypes {
			walk(v, vv)
		}
		for _,vv := range n.FundamentalTypes {
			walk(v, vv)
		}
		for _,vv := range n.Methods {
			walk(v, vv)
		}
		for _,vv := range n.MethodTypes {
			walk(v, vv)
		}
		for _,vv := range n.Namespaces {
			walk(v, vv)
		}
		for _,vv := range n.NamespaceAliases {
			walk(v, vv)
		}
		for _,vv := range n.OperatorFunctions {
			walk(v, vv)
		}
		for _,vv := range n.OperatorMethods {
			walk(v, vv)
		}
		for _,vv := range n.OffsetTypes {
			walk(v, vv)
		}
		for _,vv := range n.PointerTypes {
			walk(v, vv)
		}
		for _,vv := range n.ReferenceTypes {
			walk(v, vv)
		}
		for _,vv := range n.Structs {
			walk(v, vv)
		}
		for _,vv := range n.Typedefs {
			walk(v, vv)
		}
		for _,vv := range n.Unimplementeds {
			walk(v, vv)
		}
		for _,vv := range n.Unions {
			walk(v, vv)
		}
		for _,vv := range n.Variables {
			walk(v, vv)
		}

	case *xmlArgument:
	case *xmlArray:
	//case *xmlBase: //FIXME
	//case *xmlEllipsis: //FIXME
	case *xmlClass:
	case *xmlConstructor:
	case *xmlConverter:
	case *xmlCvQualifiedType:
	case *xmlDestructor:
	//case *xmlEnumValue: //FIXME
	case *xmlEnumeration:
	case *xmlField:
	case *xmlFile:
	case *xmlFunction:
	case *xmlFunctionType:
	case *xmlFundamentalType:
	case *xmlMethod:
	case *xmlMethodType:
	case *xmlNamespace:
	case *xmlNamespaceAlias:
	case *xmlOffsetType:
	case *xmlOperatorFunction:
	case *xmlOperatorMethod:
	case *xmlPointerType:
	case *xmlReferenceType:
	case *xmlStruct:
	case *xmlTypedef:
	case *xmlUnimplemented:
	case *xmlUnion:
	case *xmlVariable:
	default:
		panic(fmt.Sprintf("unknown type [%T]", n))
	}
}

