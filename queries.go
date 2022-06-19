package main

const (
	QueryGetVulnerabilities = `
	SELECT 
		a.Title, a.GitHubSeverity
	FROM` +
		"`bigquery-public-data.deps_dev_v1.Advisories`" + ` AS a
	WHERE 
		"%s" IN (
			SELECT 
			DISTINCT pkg.Name
			FROM 
			UNNEST(a.Packages) as pkg
		)
	ORDER BY 
	a.Title
	LIMIT 
	10
	`
)
