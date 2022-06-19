package main

type Package struct {
	System             string  `bigquery:"System"`
	Name               string  `bigquery:"Name"`
	AffectedVersions   string  `bigquery:"AffectedVersions"`
	UnaffectedVersions *string `bigquery:"UnaffectedVersions"`
}
