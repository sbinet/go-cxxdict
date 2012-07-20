package cxxtypes

import (
	"encoding/gob"
	"fmt"
	"io"
	"sort"
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

// LoadIds loads identifiers using the specified identifier distiller
func LoadIds(distillerName string, r io.Reader) error {
	distiller, ok := g_distillers[distillerName]
	if !ok {
		return fmt.Errorf("cxxtypes: unknown distiller %q (forgotten import?)", distillerName)
	}
	return distiller.LoadIdentifiers(r)
}

// Dict stores dictionary informations about a library and the types/identifiers
// this library defines and exports.
// type Dict struct {
// 	Library string
// 	Keys    []string
// 	Content []Id
// }

// SaveIds dumps all cxxtypes.Id into the specified io.Writer
func SaveIds(dst io.Writer, metadata map[string]interface{}) error {

	enc := gob.NewEncoder(dst)
	if enc == nil {
		return fmt.Errorf("cxxtypes: could not create gob-encoder")
	}
	d := make(map[string]interface{})
	keys := make([]string, 0, len(g_ids))
	vals := make([]Id, 0, len(g_ids))
	for k, _ := range g_ids {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, g_ids[k])
	}
	d["Keys"] = keys
	d["Content"] = vals
	d["MetaData"] = metadata
	return enc.Encode(d)
}

type gobDistiller struct {
}

func (g *gobDistiller) LoadIdentifiers(src io.Reader) error {

	dec := gob.NewDecoder(src)
	if dec == nil {
		return fmt.Errorf("cxxtypes: could not create gob-decoder")
	}

	d := make(map[string]interface{})
	err := dec.Decode(&d)
	if err != nil {
		return err
	}

	keys := d["Keys"].([]string)
	ids := d["Content"].([]Id)

	//fmt.Printf("n-keys: %v\n", len(keys))
	//fmt.Printf("n-vals: %v\n", len(ids))

	for i, k := range keys {
		// FIXME: handle duplicates, if any
		g_ids[k] = ids[i]
	}
	return err
}

func init() {
	// register types with gob
	gob.Register(&ArrayType{})
	gob.Register(&ClassType{})
	gob.Register(&CvrQualType{})
	gob.Register(&EnumType{})
	gob.Register(&FunctionType{})
	gob.Register(&FundamentalType{})
	gob.Register(&PtrType{})
	gob.Register(&RefType{})
	gob.Register(&StructType{})
	gob.Register(&TypedefType{})
	gob.Register(&UnionType{})
	gob.Register(&placeHolderType{})

	// register identifiers with gob
	gob.Register([]Id{})
	gob.Register(&Namespace{})
	gob.Register(&Function{})
	gob.Register(&OverloadFunctionSet{})
	gob.Register(&Member{})

	// register the metadata type with gob.
	gob.Register(map[string]interface{}{})

	// register the "default" distiller
	RegisterDistiller("gob", &gobDistiller{})
	//RegisterDistiller("default", &gobDistiller{})

}

// EOF
