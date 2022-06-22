package main

import (
	"fmt"
	"os"
)

func main() {
	// Args[0] is the program path, so use args from Args[1].
	// Args should include:
	// (0) owner name, (1) repo name,
	// (2) GitHub Access Token, (3) base commit SHA, (4) head commit SHA.
	args := os.Args[1:]
	if len(args) != 5 {
		fmt.Println("len of args not equals to 5")
		return
	}

	// Get direct dependenies with change types (e.g. added/removed) using the GitHub Dependency Review REST API
	// on the two specified code commits.
	directDeps, err := GetDiffFromCommits(args[2], args[0], args[1], args[3], args[4])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get dependencies of direct dependencies, i.e. indirect dependencies.
	fmt.Println("> Supplementing vuln info for direct dependencies...")
	for i := range directDeps {
		if directDeps[i].Ecosystem == "pip" {
			directDeps[i].Ecosystem = "pypi"
		}
		if directDeps[i].ChangeType == "removed" {
			continue
		}
		for j, v := range directDeps[i].Vulnerabilities {
			vuln, err := GetVulnerabilityByAdvID(v.ID)
			if err != nil {
				fmt.Println(err)
				return
			}
			if vuln.ID != "" {
				// Use the vuln obtained from BQ to replace the one obtained from GH to supplement information.
				directDeps[i].Vulnerabilities[j] = vuln
			}
		}
		fmt.Println("> Retrieving dependnecies of the current dependency...")
		indirectDeps, err := GetDependenciesOfDependencyBySystemNameVersion(
			directDeps[i].Ecosystem,
			directDeps[i].Name,
			directDeps[i].Version,
		)
		if err != nil {
			fmt.Println(err)
			return
		}
		// for _, d := range indirectDeps {
		// 	fmt.Println(d.Name)
		// }
		directDeps[i].Dependencies = indirectDeps
	}

	// Retrieve vulnerabilities by traversing all dependencies.
	// Since we only have (1) direct dependencies and (2) one layer of indirect dependencies,
	// two iterations here are enough to traverse over all nodes.
	// This might be a graph traversal in the future if more indirect dependency layers are added.
	for i, d := range directDeps {
		// Skip removed dependencies, only focus on added dependencies.
		if d.ChangeType == "removed" {
			continue
		}
		fmt.Println("> Retrieving vulns from BQ")
		vuln, err := GetVulnerabilitiesBySystemNameVersion(d.Ecosystem, d.Name, d.Version)
		if err != nil {
			fmt.Println(err)
			return
		}

		alreadyHave := map[string]bool{}
		for _, v := range d.Vulnerabilities {
			alreadyHave[v.ID] = true
		}
		for _, v := range vuln {
			if alreadyHave[v.ID] {
				// fmt.Printf("found duplicates: %s\n", v.ID)
				continue
			}
			directDeps[i].Vulnerabilities = append(directDeps[i].Vulnerabilities, v)
			alreadyHave[v.ID] = true
		}

		// Handle vulnerabilities in indirect dependencies.
		for j, dd := range directDeps[i].Dependencies {
			// fmt.Printf("Getting vulns from BQ for indirect dependency %s\n", dd.Name)
			v, err := GetVulnerabilitiesBySystemNameVersion(dd.Ecosystem, dd.Name, dd.Version)
			if err != nil {
				fmt.Println(err)
				return
			}
			// fmt.Println(v)
			directDeps[i].Dependencies[j].Vulnerabilities = append(
				directDeps[i].Dependencies[j].Vulnerabilities,
				v...,
			)
			if len(directDeps[i].Dependencies[j].Vulnerabilities) != 0 {
				fmt.Println("Vulnerbaility found in indirect dependencies")
				PrintDependencyToStdOut(directDeps[i].Dependencies[j])
			}
		}

		if len(directDeps[i].Vulnerabilities) != 0 {
			fmt.Println("Vulnerbaility found in direct dependencies")
			PrintDependencyToStdOut(directDeps[i])
		}
	}
}
