package main

import (
	"bundlephobia/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	URL = "https://registry.npmjs.org/%s"
)

func findDependencies(package_ string) *types.Tree {
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

	dependencyTree.Data = parsedBody.Name
	dependencyTree.Children = []*types.Tree{}
	for name := range parsedBody.Versions[parsedBody.Tags.Latest].Dependencies {
		dependencyTree.Children = append(dependencyTree.Children, findDependencies(name))
	}

	return dependencyTree
}

func traverseTree(dependencyTree *types.Tree, level int) {
	if dependencyTree == nil {
		return
	}

	fmt.Printf("%d %s\n", level, dependencyTree.Data)
	for _, child := range dependencyTree.Children {
		traverseTree(child, level+1)
	}
}

func main() {
	package_ := os.Args[1]
	dependencyTree := findDependencies(package_)
	traverseTree(dependencyTree, 0)
}
