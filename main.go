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
	// Retrieve dependencies of direct dependencies, i.e. indirect dependencies.
	fmt.Println("> Retrieving dependencies of direct dependencies (indirect dependencies)...")
	for i := range directDeps {
		if directDeps[i].ChangeType == "removed" {
			continue
		}
		indirectDeps, err := GetDependenciesOfDependencyBySystemNameVersion(
			directDeps[i].Ecosystem,
			directDeps[i].Name,
			directDeps[i].Version,
		)
		if err != nil {
			fmt.Println(err)
			return
		}
		directDeps[i].Dependencies = indirectDeps
	}

	// Retrieve vulnerabilities by traversing all dependencies.
	// Since we only have (1) direct dependencies and (2) one layer of indirect dependencies,
	// two iterations here are enough to traverse over all nodes.
	// This might be a graph traversal in the future if more indirect dependency layers are added.
	fmt.Println("> Retrieving vulnerabilities from BQ")
	for i, d := range directDeps {
		// Skip removed dependencies, only focus on added dependencies.
		if d.ChangeType == "removed" {
			continue
		}
		vuln, err := GetVulnerabilitiesBySystemNameVersion(d.Ecosystem, d.Name, d.Version)
		if err != nil {
			fmt.Println(err)
			return
		}
		directDeps[i].Vulnerabilities = append(directDeps[i].Vulnerabilities, vuln...)

		// Handle vulnerabilities in indirect dependencies.
		for j, dd := range directDeps[i].Dependencies {
			v, err := GetVulnerabilitiesBySystemNameVersion(dd.Ecosystem, dd.Name, dd.Version)
			if err != nil {
				fmt.Println(err)
				return
			}
			directDeps[i].Dependencies[j].Vulnerabilities = append(
				directDeps[i].Dependencies[j].Vulnerabilities,
				v...,
			)
			if len(directDeps[i].Dependencies[j].Vulnerabilities) != 0 {
				fmt.Printf("Vulnerbaility found in indirect dependency %s\n", dd.Name)
				PrintDependencyToStdOut(directDeps[i].Dependencies[j])
			} else {
				fmt.Printf("indirect dependency %s is vulnerability-free\n", dd.Name)
			}
		}

		if len(directDeps[i].Vulnerabilities) != 0 {
			fmt.Printf("Vulnerbaility found in direct dependency %s\n", d.Name)
			PrintDependencyToStdOut(directDeps[i])
		} else {
			fmt.Printf("direct dependency %s is vulnerability-free\n", d.Name)
		}
	}
}
