package types

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
