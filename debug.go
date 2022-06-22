package main

import "fmt"

func PrintDependencyToStdOut(d Dependency) {
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
