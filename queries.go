package main

const (
	QueryGetVulnerabilities = `
		SELECT
		t1.System, t1.Name, t1.Version,
		t2.Source, t2.SourceID, t2.SourceURL, t2.Title, t2.Description,
		t2.CVSS3Score, t2.Severity, t2.GitHubSeverity, t2.Disclosed,
		t2.ReferenceURLs

		FROM
		(SELECT DISTINCT pv.System, pv.Name, pv.Version, pv_adv.SourceID
		FROM ` +
		"`bigquery-public-data.deps_dev_v1.PackageVersions`" + `AS pv
		INNER JOIN 
		UNNEST(pv.Advisories) AS pv_adv
		WHERE 
		pv.System = "%s" -- Input 1: dependency system
		AND
		pv.Name = "%s"  -- Input 2: dependency name
		AND
		pv.Version = "%s"  -- Input 3: dependency version
		) AS t1

		INNER JOIN` +
		"`bigquery-public-data.deps_dev_v1.Advisories`" + `
		AS t2
		ON
		t1.SourceID = t2.SourceID

		ORDER BY t2.CVSS3Score
		DESC
	`

	QueryGetDependencies = `
		SELECT
		dep.Dependency
		FROM ` +
		"`bigquery-public-data.deps_dev_v1.Dependencies`" + ` AS dep
		WHERE
		dep.System = "%s" -- Input 1: dependency system
		AND
		dep.Name = "%s" -- Input 2: dependency name
		AND
		dep.Version = "%s" -- Input 3: dependency version
	`
)
