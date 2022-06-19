package main

import (
	"github.com/hashicorp/go-version"
)

type ChangeType string

const (
	Added   ChangeType = "added"
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
	// ChangeType indicates whether the dependency is added or removed.
	ChangeType ChangeType `json:"change_type"`
	// ManifestFileName is the name of the manifest file of the dependency, such as go.mod for Go.
	ManifestFileName string `json:"manifest"`
	// Ecosystem is the name of the package manager, such as NPM, GO, PYPI.
	Ecosystem string `json:"ecosystem"`
	// Name is the name of the dependency.
	Name string `json:"name"`
	// Version is the package version of the dependency.
	Version version.Version `json:"version"`
	// Package URL is a short link for a package.
	PackageURL *string `json:"package_url"`
	// License is ...
	License *string `json:"license"`
	// SrcRepoURL is the source repository URL of the dependency.
	SrcRepoURL *string `json:"source_repository_url"`
	// Vulnerabilities is a list of Vulnerability.
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}
