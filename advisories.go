package main

import (
	"net/url"

	"github.com/golang/protobuf/ptypes/timestamp"
)

type SeverityLevel string

const (
	Critical SeverityLevel = "CRITICAL"
	High     SeverityLevel = "HIGH"
	Medium   SeverityLevel = "MEDIUM"
	Moderate SeverityLevel = "MODERATE"
	Low      SeverityLevel = "LOW"
	None     SeverityLevel = "NONE"
	Unknown  SeverityLevel = "UNKNOWN"
)

func (sl *SeverityLevel) IsValid() bool {
	switch *sl {
	case Critical, High, Medium, Moderate, Low, None, Unknown:
		return true
	default:
		return false
	}
}

type Advisory struct {
	SnapshotAt     timestamp.Timestamp `bigquery:"SnapshotAt"`
	Source         string              `bigquery:"Source"`
	SourceID       string              `bigquery:"SourceID"`
	SourceURL      string              `bigquery:"SourceURL"`
	Title          *string             `bigquery:"Title"`
	Description    *string             `bigquery:"Description"`
	ReferenceURLs  []url.URL           `bigquery:"ReferenceURLs"`
	CVSS3Score     float32             `bigquery:"CVSS3Score"`
	Severity       *SeverityLevel      `bigquery:"Severity"`
	GitHubSeverity *SeverityLevel      `bigquery:"GitHubSeverity"`
	Disclosed      timestamp.Timestamp `bigquery:"Disclosed"`
}

// Vulnerability is a vulnerability of a dependency.
type Vulnerability struct {
	// Severity is a enum type of the severity level of a vulnerability.
	Severity SeverityLevel `json:"severity"`
	// AdvisoryGHSAId is ...
	AdvisoryGHSAId string `json:"advisory_ghsa_id"`
	// AdvisorySummary is ...
	AdvisorySummary *string `json:"advisory_summary"`
	// AdvisoryURL is ...
	AdvisoryURL *string `json:"advisory_url"`
}
