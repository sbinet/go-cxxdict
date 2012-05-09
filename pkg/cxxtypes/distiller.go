package cxxtypes

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
)

// Distiller is the interface to distill types and identifiers
type Distiller interface {
	LoadIdentifiers(r io.Reader) error
}

var g_distillers = make(map[string]Distiller)

// Register makes a distiller available by the provided name.
func RegisterDistiller(name string, distiller Distiller) {
	if distiller == nil {
		panic("cxxtypes: Register distiller is nil")
	}
	if _, dup := g_distillers[name]; dup {
		panic("cxxtypes: Register called twice for distiller " + name)
	}
	g_distillers[name] = distiller
}

// DistillIdentifiers distills identifiers using the specified distiller
func DistillIdentifiers(distillerName string, r io.Reader) error {
	distiller, ok := g_distillers[distillerName]
	if !ok {
		return fmt.Errorf("cxxtypes: unknown distiller %q (forgotten import?)", distillerName)
	}
	return distiller.LoadIdentifiers(r)
}

// SaveIdentifiers dumps all cxxtypes.Id into the specified io.Writer
func SaveIdentifiers(dst io.Writer) error {
	if true {
		enc := json.NewEncoder(dst)
		if enc == nil {
			return fmt.Errorf("cxxtypes: could not create encoder")
		}
		return enc.Encode(g_ids)
	} else if true {
		enc := gob.NewEncoder(dst)
		if enc == nil {
			return fmt.Errorf("cxxtypes: could not create encoder")
		}
		return enc.Encode(g_ids)
	} else if true {
		enc := xml.NewEncoder(dst)
		if enc == nil {
			return fmt.Errorf("cxxtypes: could not create encoder")
		}
		return enc.Encode(g_ids)
	}
	panic("unreachable")
}

func init() {
	// register types with gob
	gob.Register(ArrayType{})
	gob.Register(ClassType{})
	gob.Register(CvrQualType{})
	gob.Register(EnumType{})
	gob.Register(FunctionType{})
	gob.Register(FundamentalType{})
	gob.Register(PtrType{})
	gob.Register(RefType{})
	gob.Register(StructType{})
	gob.Register(TypedefType{})
	gob.Register(UnionType{})
	gob.Register(placeHolderType{})

	// register identifiers with gob
	gob.Register(Namespace{})
	gob.Register(Scope{})
	gob.Register(Function{})
	gob.Register(OverloadFunctionSet{})
}

// EOF
