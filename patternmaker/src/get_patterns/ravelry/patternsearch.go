package ravelry

import "fmt"

type PatternID struct {
	ID int `json:"id"`
}

type Paginator struct {
	PageCount int `json:"page_count"`
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	Results   int `json:"results"`
	LastPage  int `json:"last_page"`
}

type PatternSearchResult struct {
	PatternIDs []PatternID `json:"patterns"`
	Paginator  Paginator   `json:"paginator"`
}

func (psr PatternSearchResult) GetPatternIDs() []string {
	ids := make([]string, len(psr.PatternIDs))
	for i, p := range psr.PatternIDs {
		ids[i] = fmt.Sprintf("%d", p.ID)
	}
	return ids
}
