package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aidenwang9867/depdiffvis/pkg"
	"github.com/ossf/scorecard/v4/checks"
)

func main() {
	// Args[0] is the program path, so use args from Args[1].
	// Args should include:
	// (0) owner name, (1) repo name,
	// (2) base commit SHA, (3) head commit SHA.
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Println("len of args not equals to 3")
		return
	}
	repoURI, base, head := args[0], args[1], args[2]
	//Fetch dependency diffs using the GitHub Dependency Review API.
	checksToRun := []string{
		// checks.CheckCodeReview,
		// checks.CheckSAST,
		// checks.CheckBranchProtection,
		checks.CheckLicense,
	}
	changeTypeToCheck := map[pkg.ChangeType]bool{
		pkg.Added:   true,
		pkg.Updated: true,
		// pkg.Removed: true,
	}
	results, err := GetDependencyDiffResults(
		context.Background(), repoURI, base, head, checksToRun, changeTypeToCheck,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(results)
	markdown, err := SprintDependencyChecksToMarkdown(results)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// title := "# [Scorecards' Github action](https://github.com/ossf/scorecard-action) Dependency-diff Report\n\n"
	// title += fmt.Sprintf(
	// 	"Dependency-diffs (changes) between the BASE commit `%s` and the HEAD commit `%s`:\n\n",
	// 	args[2], args[3],
	// )
	// fmt.Print(title)
	if *markdown == "" {
		fmt.Println("No dependency changes found.")
	} else {
		fmt.Println(*markdown)
	}
	// return
}
