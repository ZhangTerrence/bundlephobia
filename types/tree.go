package types

type DependencyTree struct {
	Name         string
	Dependencies []*DependencyTree
}
