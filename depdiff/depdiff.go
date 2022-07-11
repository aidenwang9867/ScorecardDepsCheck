package depdiff

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/ossf/scorecard/v4/pkg"
)

func GetDependencySrcRepo(d Dependency) string {
	return ""
}

func GetDependencyScorecardResults(d Dependency) (*pkg.JSONScorecardResultV2, error) {
	// https://api.securityscorecards.dev/projects/{host}/{owner}/{repository}
	reqURL := path.Join(
		"api.securityscorecards.dev", "projects",
		"github.com", "ossf-tests", "scorecard-action",
	)
	req, err := http.NewRequest("GET", "https://"+reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("generate request error: %w", err)
	}

	// Set a ten-seconds timeout to make sure the client can be created correctly.
	myClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := myClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("not found")
	}
	defer resp.Body.Close()
	results := pkg.JSONScorecardResultV2{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}
	return &results, nil
}
