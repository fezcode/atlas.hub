package model

type Tool struct {
	Name        string `piml:"name"`
	Description string `piml:"description"`
	Repo        string `piml:"repo"`
	Bin         string `piml:"bin"`
	Category    string `piml:"category"`
	IsHub       bool   `piml:"is_hub"`
	Selected    bool   `piml:"-"`
	Status      string `piml:"-"` // "pending", "installing", "done", "error"
	Error       error  `piml:"-"`
}
