package depdiff

type ChangeType string

const (
	Added   ChangeType = "added"
	Updated ChangeType = "updated"
	Removed ChangeType = "removed"
)

func (ct *ChangeType) IsValid() bool {
	switch *ct {
	case Added, Removed:
		return true
	default:
		return false
	}
}

// Dependency is a dependency diff in a code commit.
type Dependency struct {
	// IsDirect suggests if the dependency is a direct dependency of a code commit.
	IsDirect bool

	// ChangeType indicates whether the dependency is added or removed.
	ChangeType ChangeType `json:"change_type"`

	// ManifestFileName is the name of the manifest file of the dependency, such as go.mod for Go.
	ManifestFileName string `json:"manifest"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem string `json:"ecosystem" bigquery:"System"`

	// Name is the name of the dependency.
	Name string `json:"name" bigquery:"Name"`

	// Version is the package version of the dependency.
	Version string `json:"version" bigquery:"Version"`

	// Package URL is a short link for a package.
	PackageURL string `json:"package_url"`

	// License is ...
	License string `json:"license"`

	// SrcRepoURL is the source repository URL of the dependency.
	SrcRepoURL string `json:"source_repository_url"`

	// Vulnerabilities is a list of Vulnerability.
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`

	// Dependencies are the dependencies of the dependency, i.e. indirect dependencies.
	Dependencies []Dependency `json:"dependencies"`
}
