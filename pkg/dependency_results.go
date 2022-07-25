package pkg

import (
	"encoding/json"
	"fmt"
	"io"

	sce "github.com/ossf/scorecard/v4/errors"
	scpkg "github.com/ossf/scorecard/v4/pkg"
)

// ChangeType is the change type (added, updated, removed) of a dependency.
type ChangeType string

const (
	// Added suggests the dependency is a newly added one.
	Added ChangeType = "added"
	// Updated suggests the dependency is updated from an old version.
	Updated ChangeType = "updated"
	// Removed suggests the dependency is removed.
	Removed ChangeType = "removed"
)

// IsValid determines if a ChangeType is valid.
func (ct ChangeType) IsValid() bool {
	switch ct {
	case Added, Updated, Removed:
		return true
	default:
		return false
	}
}

// ScorecardResultWithError is used for the dependency-diff module to record the scorecard result and the error.
type ScorecardResultWithError struct {
	// ScorecardResults is the scorecard result for the dependency repo.
	ScorecardResult *scpkg.ScorecardResult

	// Error is an error returned when running the scorecard checks. A nil Error indicates the run succeeded.
	Error error
}

// DependencyCheckResult is the dependency structure used in the returned results.
type DependencyCheckResult struct {
	// ChangeType indicates whether the dependency is added, updated, or removed.
	ChangeType *ChangeType

	// Package URL is a short link for a package.
	PackageURL *string

	// SourceRepository is the source repository URL of the dependency.
	SourceRepository *string

	// ManifestPath is the path of the manifest file of the dependency, such as go.mod for Go.
	ManifestPath *string

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem *string

	// Version is the package version of the dependency.
	Version *string

	// ScorecardResultWithError contains the scorecard checking result of the dependency
	// and an error field to record potential failing scorecard runs.
	ScorecardResultWithError ScorecardResultWithError

	// Name is the name of the dependency.
	Name string
}

// AsJSON for DependencyCheckResult exports the DependencyCheckResult as a JSON object.
func (dr *DependencyCheckResult) AsJSON(writer io.Writer) error {
	if err := json.NewEncoder(writer).Encode(*dr); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}
