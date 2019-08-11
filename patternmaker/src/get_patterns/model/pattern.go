package model

type Pattern struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Paginator struct {
	PageCount int `json:"page_count"`
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	Results   int `json:"results"`
	LastPage  int `json:"last_page"`
}

type PatternSearchResult struct {
	Patterns  []Pattern `json:"patterns"`
	Paginator Paginator `json:"paginator"`
}
