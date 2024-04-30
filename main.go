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
	URL = "https://registry.npmjs.org/%s"
)

func findDependencies(package_ string, cache *map[string]*types.Tree) *types.Tree {
	dependencyTree := &types.Tree{}

	response, err := http.Get(fmt.Sprintf(URL, package_))
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var parsedBody types.NodePackage
	err = json.Unmarshal(data, &parsedBody)
	if err != nil {
		panic(err)
	}

	packageExists := (*cache)[parsedBody.Name]
	if packageExists != nil {
		return (*cache)[parsedBody.Name]
	}

	dependencyTree.Data = parsedBody.Name
	dependencyTree.Children = []*types.Tree{}
	for name := range parsedBody.Versions[parsedBody.Tags.Latest].Dependencies {
		dependencyTree.Children = append(dependencyTree.Children, findDependencies(name, cache))
	}

	(*cache)[parsedBody.Name] = dependencyTree
	return dependencyTree
}

func traverseTree(dependencyTree *types.Tree) {
	if dependencyTree == nil {
		return
	}

	for _, child := range dependencyTree.Children {
		traverseTree(child)
	}
}

func main() {
	package_ := os.Args[1]
	cache := &map[string]*types.Tree{}
	dependencyTree := findDependencies(package_, cache)
	traverseTree(dependencyTree)
}
