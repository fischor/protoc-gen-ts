package main

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

type Printer struct {
	G      *protogen.GeneratedFile
	indent int
}

func newPrinter(g *protogen.GeneratedFile) *Printer {
	return &Printer{
		G:      g,
		indent: 0,
	}
}

func (p *Printer) P(v ...interface{}) {
	// Do not indent on empty lines, i.e. if v is nil.
	if v != nil {
		ind := []interface{}{strings.Repeat(" ", p.indent)}
		v = append(ind, v...)
	}
	p.G.P(v...)
}

func (p *Printer) F(format string, a ...interface{}) {
	p.G.P(fmt.Sprintf(strings.Repeat(" ", p.indent)+format, a...))
}

func (p *Printer) C(comm protogen.Comments) {
	lines := strings.Split(string(comm), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		p.P(" *", line)
	}
}

func (p *Printer) Indent() {
	p.indent += 2
}

func (p *Printer) Outdent() {
	p.indent -= 2
	if p.indent < 0 {
		panic("outdented")
	}
}

func (p *Printer) Indented(f func()) {
	p.Indent()
	f()
	p.Outdent()
}
