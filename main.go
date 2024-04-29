package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type NodeTags struct {
	Latest string `json:"latest"`
}

type NodeVersion struct {
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type NodePackage struct {
	Name     string                 `json:"name"`
	Tags     NodeTags               `json:"dist-tags"`
	Versions map[string]NodeVersion `json:"versions"`
}

const (
	URL = "https://registry.npmjs.org/%s"
)

func findDependencies(package_ string) {
	fmt.Println(package_)

	response, err := http.Get(fmt.Sprintf(URL, package_))
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var parsedBody NodePackage
	err = json.Unmarshal(data, &parsedBody)
	if err != nil {
		panic(err)
	}

	for dependency, _ := range parsedBody.Versions[parsedBody.Tags.Latest].Dependencies {
		findDependencies(dependency)
	}
}

func main() {
	package_ := os.Args[1]

	findDependencies(package_)
}
