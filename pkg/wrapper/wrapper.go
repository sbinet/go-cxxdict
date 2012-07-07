package wrapper

import (
	"fmt"
	"io"
)

// the list of registered plugins
var g_plugins []Plugin

// RegisterPlugin installs a plugin for later use.
// returns an error if the same plugin is registered twice.
func RegisterPlugin(p Plugin) error {
	n := p.Name()
	for i, _ := range g_plugins {
		if g_plugins[i].Name() == n {
			return fmt.Errorf("wrapper: plugin [%s] already registered", n)
		}
	}
	g_plugins = append(g_plugins, p)
	return nil
}

// FileDescriptor describes a wrapper-output
type FileDescriptor struct {
	// name of this file descriptor, usually the library 
	// name (w/o prefix and suffix. ie: Foo, not libFoo.so)
	Name string

	// name of the package this file descriptor wraps
	Package string

	// name of the header containing the declarations for this library
	Header string

	// list of dependencies (pkgs, other file descriptors)
	Dependency []string

	// the products of the wrapping
	Files map[string]io.WriteCloser
}

// Generator is the type whose methods generate the output, 
// stored in the associated response structure.
type Generator struct {
	plugins []Plugin
	Fd      FileDescriptor
	Args    map[string]interface{}
}

// P prints the arguments to the generated output
func (g *Generator) P(str ...interface{}) {
}

// GenerateAllFiles generates the output for all the files we're outputting.
func (g *Generator) GenerateAllFiles() error {
	for _, p := range g_plugins {
		err := p.Init(g)
		if err != nil {
			return err
		}
		g.plugins = append(g.plugins, p)
	}

	for _, p := range g.plugins {
		err := p.Generate(&g.Fd)
		if err != nil {
			return err
		}
	}
	return nil
}

// Plugins returns the names of the active plugins
func (g *Generator) Plugins() []string {
	names := make([]string, 0, len(g.plugins))
	for _, p := range g.plugins {
		names = append(names, p.Name())
	}
	return names
}

func NewGenerator() *Generator {
	gen := &Generator{}
	gen.plugins = make([]Plugin, 0)
	gen.Fd.Files = make(map[string]io.WriteCloser)
	gen.Args = make(map[string]interface{})
	return gen
}

func init() {
	g_plugins = make([]Plugin, 0, 1)
}

// EOF
