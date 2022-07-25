package pkg

import (
	"encoding/json"
	"fmt"
	"io"

	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
)

//nolint
type jsonCheckResult struct {
	Name       string
	Details    []string
	Confidence int
	Pass       bool
}

type jsonScorecardResult struct {
	Repo     string
	Date     string
	Checks   []jsonCheckResult
	Metadata []string
}

type jsonCheckDocumentationV2 struct {
	URL   string `json:"url"`
	Short string `json:"short"`
	// Can be extended if needed.
}

//nolint
type jsonCheckResultV2 struct {
	Details []string                 `json:"details"`
	Score   int                      `json:"score"`
	Reason  string                   `json:"reason"`
	Name    string                   `json:"name"`
	Doc     jsonCheckDocumentationV2 `json:"documentation"`
}

type jsonRepoV2 struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
}

type jsonScorecardV2 struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type jsonFloatScore float64

func (s jsonFloatScore) MarshalJSON() ([]byte, error) {
	// Note: for integers, this will show as X.0.
	return []byte(fmt.Sprintf("%.1f", s)), nil
}

//nolint:govet
// JSONScorecardResultV2 exports results as JSON for new detail format.
type JSONScorecardResultV2 struct {
	Date           string              `json:"date"`
	Repo           jsonRepoV2          `json:"repo"`
	Scorecard      jsonScorecardV2     `json:"scorecard"`
	AggregateScore jsonFloatScore      `json:"score"`
	Checks         []jsonCheckResultV2 `json:"checks"`
	Metadata       []string            `json:"metadata"`
}

// JSONDependencydiffResult exports dependency-diff check results as JSON for new detail format.
type JSONDependencydiffResult struct {
	ChangeType          *ChangeType            `json:"changeType"`
	PackageURL          *string                `json:"packageUrl"`
	SourceRepository    *string                `json:"sourceRepository"`
	ManifestPath        *string                `json:"manifestPath"`
	Ecosystem           *string                `json:"ecosystem"`
	Version             *string                `json:"packageVersion"`
	JSONScorecardResult *JSONScorecardResultV2 `json:"scorecardResult"`
	Name                string                 `json:"packageName"`
}

// DependencydiffResultsAsJSON exports dependencydiff results as JSON. This cannot be defined as the OOP-like
// ScorecardResult.AsJSON since we return a slice of DependencyCheckResult.
func DependencydiffResultsAsJSON(depdiffResults []DependencyCheckResult,
	logLevel log.Level, doc docs.Doc, writer io.Writer,
) error {
	out := []JSONDependencydiffResult{}
	for _, dr := range depdiffResults {
		// Copy every DependencydiffResult struct to a JSONDependencydiffResult for exporting as JSON.
		jsonDepdiff := JSONDependencydiffResult{
			ChangeType:       dr.ChangeType,
			PackageURL:       dr.PackageURL,
			SourceRepository: dr.SourceRepository,
			ManifestPath:     dr.ManifestPath,
			Ecosystem:        dr.Ecosystem,
			Version:          dr.Version,
			Name:             dr.Name,
		}
		scResult := dr.ScorecardResultWithError.ScorecardResult
		if scResult != nil {
			score, err := scResult.GetAggregateScore(doc)
			if err != nil {
				return err
			}
			jsonResult := JSONScorecardResultV2{
				Repo: jsonRepoV2{
					Name:   scResult.Repo.Name,
					Commit: scResult.Repo.CommitSHA,
				},
				Scorecard: jsonScorecardV2{
					Version: scResult.Scorecard.Version,
					Commit:  scResult.Scorecard.CommitSHA,
				},
				Date:           scResult.Date.Format("2006-01-02"),
				Metadata:       scResult.Metadata,
				AggregateScore: jsonFloatScore(score),
			}
			for _, c := range scResult.Checks {
				doc, e := doc.GetCheck(c.Name)
				if e != nil {
					return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %s: %v", c.Name, e))
				}
				tmpResult := jsonCheckResultV2{
					Name: c.Name,
					Doc: jsonCheckDocumentationV2{
						URL:   doc.GetDocumentationURL(scResult.Scorecard.CommitSHA),
						Short: doc.GetShort(),
					},
					Reason: c.Reason,
					Score:  c.Score,
				}
				for i := range c.Details {
					d := c.Details[i]
					m := DetailToString(&d, logLevel)
					if m == "" {
						continue
					}
					tmpResult.Details = append(tmpResult.Details, m)
				}
				jsonResult.Checks = append(jsonResult.Checks, tmpResult)
				jsonDepdiff.JSONScorecardResult = &jsonResult
			}
		}
		out = append(out, jsonDepdiff)
	}
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}
