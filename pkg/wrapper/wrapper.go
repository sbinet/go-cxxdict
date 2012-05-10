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
	Name       string   // name of this file descriptor, for debugging
	Package    string   // name of the package this file descriptor wraps
	Dependency []string // list of dependencies (pkgs, other file descriptors)

	Files map[string]io.WriteCloser // the products of the wrapping
}

// Generator is the type whose methods generate the output, 
// stored in the associated response structure.
type Generator struct {
	plugins []Plugin
	Fd      FileDescriptor
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

	return gen
}

func init() {
	g_plugins = make([]Plugin, 0, 1)
}

// EOF
