package gccxml

import (
	"fmt"
)

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

	case *xmlFundamentalType:
	default:
		panic(fmt.Sprintf("unknown type [%T]", n))
	}
}

