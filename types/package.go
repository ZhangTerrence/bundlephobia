package types

type PackageTags struct {
	Latest string `json:"latest"`
}

type PackageVersion struct {
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type Package struct {
	Name     string                    `json:"name"`
	Tags     PackageTags               `json:"dist-tags"`
	Versions map[string]PackageVersion `json:"versions"`
}
