package scorecarddepscheck

type ChangeType string

const (
	Added   ChangeType = "added"
	Removed ChangeType = "removed"
)

func (ct *ChangeType) IsValid() bool {
	switch *ct {
	case Added, Removed:
		return true
	default:
		return false
	}
}

type Dependency struct {
	ChangeType ChangeType `json:"change_type"`

	ManifestFileName string `json:"manifest"`
	Ecosystem        string `json:"ecosystem"`
	Name             string `json:"name"`
	Version          string `json:"version"`

	PackageURL *string `json:"package_url"`
	License    *string `json:"license"`
	SrcRepoURL *string `json:"source_repository_url"`

	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}
