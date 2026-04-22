package model

type Tool struct {
	Name             string `piml:"name"`
	Description      string `piml:"description"`
	Repo             string `piml:"repo"`
	Bin              string `piml:"bin"`
	Category         string `piml:"category"`
	LatestVersion    string `piml:"version"`
	IsHub            bool   `piml:"is_hub"`
	
	// Runtime fields
	InstalledVersion string `piml:"-"`
	Selected         bool   `piml:"-"`
	Status           string `piml:"-"` // "pending", "installing", "done", "error"
	Error            error  `piml:"-"`
	Deleted          bool   `piml:"-"`
}
