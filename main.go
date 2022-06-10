package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	args := os.Args[1:] // Args[0] is the program path.
	// Args should include: (0) owner name, (1) repo name,
	// (2) GH PAT, (3) base commit SHA, (4) head commit SHA, .
	if len(args) != 5 {
		fmt.Println("len of args not equals to 5")
		return
	}

	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/dependency-graph/compare/%s...%s",
		args[0], args[1], args[3], args[4])
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		fmt.Println("get response error")
		return
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+args[2])

	myClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := myClient.Do(req)
	if err != nil {
		fmt.Println("get response error")
		return
	}
	defer resp.Body.Close()
	// body, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(body))
	depDiff := []Dependency{}
	err = json.NewDecoder(resp.Body).Decode(&depDiff)
	if err != nil {
		fmt.Println("parse response error: %w", err)
		return
	}
	for _, d := range depDiff {
		if d.ChangeType.IsValid() {
			fmt.Printf("name: %v \nstatus: %v \nversion: %v \necosys: %v \npkg_url: %v\nsrc_url: %v\n\n",
				d.ChangeType, d.Name, d.Version, d.Ecosystem, *d.PackageURL, *d.SrcRepoURL)
		}
	}
}
