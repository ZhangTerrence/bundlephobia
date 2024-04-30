package main

import (
	types "bundlephobia/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	UrlFormat = "https://registry.npmjs.org/%s"
)

func findDependencies(package_ string, cache *map[string]*types.DependencyTree) *types.DependencyTree {
	dependencyTree := &types.DependencyTree{}

	response, err := http.Get(fmt.Sprintf(UrlFormat, package_))
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var parsedBody types.Package
	err = json.Unmarshal(data, &parsedBody)
	if err != nil {
		panic(err)
	}

	packageExists := (*cache)[parsedBody.Name]
	if packageExists != nil {
		return (*cache)[parsedBody.Name]
	}

	dependencyTree.Name = parsedBody.Name
	dependencyTree.Dependencies = []*types.DependencyTree{}
	for name := range parsedBody.Versions[parsedBody.Tags.Latest].Dependencies {
		dependencyTree.Dependencies = append(dependencyTree.Dependencies, findDependencies(name, cache))
	}

	(*cache)[parsedBody.Name] = dependencyTree
	return dependencyTree
}

func traverseTree(dependencyTree *types.DependencyTree) {
	if dependencyTree == nil {
		return
	}

	for _, dependency := range dependencyTree.Dependencies {
		traverseTree(dependency)
	}
}

func main() {
	package_ := os.Args[1]
	cache := &map[string]*types.DependencyTree{}
	dependencyTree := findDependencies(package_, cache)
	traverseTree(dependencyTree)
}
