package main

import (
	"fmt"
	"path"

	"github.com/aidenwang9867/scorecard-bigquery-auth/app/query"
)

func PrintDependencyToStdOut(d query.Dependency) {
	result := ""
	if d.IsDirect {
		result += fmt.Sprintf("name: %v \nstatus: %v \nversion: %v \necosys: %v \n",
			d.Name, d.ChangeType, d.Version, d.Ecosystem)
	} else {
		result += fmt.Sprintf("name: %v \nstatus: %v \nversion: %v \necosys: %v \n",
			d.Name, "indirect", d.Version, d.Ecosystem)
	}
	if d.PackageURL != "" {
		result += fmt.Sprintf("pkg_url: %v\n", d.PackageURL)
	}
	if d.SrcRepoURL != "" {
		result += fmt.Sprintf("src_url: %v\n", d.SrcRepoURL)
	}
	if len(d.Vulnerabilities) != 0 {
		result += fmt.Sprintln("vulnerabilities: ")
		for _, v := range d.Vulnerabilities {
			result += fmt.Sprintf("\tvuln_id: %v\n", v.ID)
			result += fmt.Sprintf("\tvuln_url: %v\n\n", v.SourceURL)
			if v.Title != "" {
				result += fmt.Sprintf("\tvuln_summary: %v\n", v.Title)
			}
			if v.GitHubSeverity != "" {
				result += fmt.Sprintf("\tseverity: %v\n", v.GitHubSeverity)
			}
			if v.Score.Valid {
				result += fmt.Sprintf("\tCVSS3Score: %v\n", v.Score.Float64)
			}
			if !v.DisclosedTime.IsZero() {
				result += fmt.Sprintf("\ttime_disclosed: %v\n", v.DisclosedTime)
			}
		}
	}
	fmt.Println(result)
}

func PrintDependencyChangeInfo(deps []query.Dependency) {
	removed, added := map[string]query.Dependency{}, map[string]query.Dependency{}
	for _, d := range deps {
		switch d.ChangeType {
		case query.Added:
			added[d.Name] = d
		case query.Removed:
			removed[d.Name] = d
		}
	}
	for dName, d := range added {
		if v, ok := removed[dName]; ok {
			fmt.Printf(
				"[Dependency updated] %s@%s (old) --> %s@%s (new)\n",
				v.Name, v.Version,
				d.Name, d.Version,
			)
		} else {
			fmt.Printf(
				"[Dependency added] %s@%s\n",
				d.Name, d.Version,
			)
		}
		if len(d.Vulnerabilities) != 0 {
			srcURL := path.Join("deps.dev", d.Ecosystem, d.Name, d.Version)
			fmt.Printf(
				"Vulnerabilties found in %s@%s, see [%s] for further information\n\n",
				d.Name, d.Version,
				srcURL,
			)
		}
	}
	for dName, d := range removed {
		if _, ok := added[dName]; ok {
			continue
		} else {
			fmt.Printf(
				"[Dependency removed] %s@%s\n",
				d.Name, d.Version,
			)
		}
	}
}
