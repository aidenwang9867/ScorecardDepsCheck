package main

import "fmt"

func PrintDepDiffToStdOut(depDiff []Dependency) {
	for _, d := range depDiff {
		result := ""
		if d.ChangeType.IsValid() {
			result += fmt.Sprintf("name: %v \nstatus: %v \nversion: %v \necosys: %v \n",
				d.Name, d.ChangeType, d.Version, d.Ecosystem)
		}
		if d.PackageURL != nil && *d.PackageURL != "" {
			result += fmt.Sprintf("pkg_url: %v\n", *d.PackageURL)
		}
		if d.SrcRepoURL != nil && *d.SrcRepoURL != "" {
			result += fmt.Sprintf("src_url: %v\n", *d.SrcRepoURL)
		}
		if d.Vulnerabilities != nil {
			result += fmt.Sprintln("vulnerabilities: ")
			for _, v := range d.Vulnerabilities {
				result += fmt.Sprintf("\tseverity: %v\n", v.Severity)
				result += fmt.Sprintf("\tadvisory_ghsa_id: %v\n", v.AdvisoryGHSAId)
				result += fmt.Sprintf("\tadvisory_summary: %v\n", *v.AdvisorySummary)
				result += fmt.Sprintf("\tadvisory_url: %v\n\n", *v.AdvisoryURL)
			}
		}
		fmt.Println(result)
	}
}
