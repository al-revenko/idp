package pkgmark

import "fmt"

func New(name string) *Pkg {
	return &Pkg{Name: name}
}

type Pkg struct {
	Name string
}

func (p *Pkg) Op(name string) Op {
	return Op{Pkg: p, Name: fmt.Sprintf("%s.%s", p.Name, name)}
}

type Op struct {
	Pkg  *Pkg
	Name string
}

func (o Op) Err(e error) error {
	return fmt.Errorf("%s -> %w", o.Name, e)
}
