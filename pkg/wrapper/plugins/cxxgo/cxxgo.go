package cxxgo

import (
	"bitbucket.org/binet/go-cxxdict/pkg/wrapper"
)

type plugin struct {
	gen *wrapper.Generator // the generator which is invoking us
}

func (p *plugin) Name() string {
	return "cxxgo.plugin"
}

func (p *plugin) Init(g *wrapper.Generator) error {
	return nil
}

func (p *plugin) Generate(file *wrapper.FileDescriptor) error {
	return nil
}

// test interfaces...

var _ wrapper.Plugin = (*plugin)(nil)

func init() {
	wrapper.RegisterPlugin(&plugin{})
}

// EOF
