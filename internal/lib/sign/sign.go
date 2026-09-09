package sign

import "fmt"

func Pkg(name string) Package {
	return Package{Name: name}
}

type Package struct {
	Name string
}

func (p Package) Op(name string) Operation {
	return Operation{Pkg: p, Name: fmt.Sprintf("%s.%s", p.Name, name)}
}

type Operation struct {
	Pkg  Package
	Name string
}

func (o Operation) Err(e error) error {
	return fmt.Errorf("%s: %w", o.Name, e)
}
