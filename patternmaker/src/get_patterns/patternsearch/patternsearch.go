package patternsearch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rebeku/patternmaker/src/get_patterns/ravelry"
)

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

type Result struct {
	PatternIDs []PatternID `json:"patterns"`
	Paginator  Paginator   `json:"paginator"`
}

func (psr Result) GetPatternIDs() []string {
	ids := make([]string, len(psr.PatternIDs))
	for i, p := range psr.PatternIDs {
		ids[i] = fmt.Sprintf("%d", p.ID)
	}
	return ids
}

const patternSearchEndpoint = "patterns/search.json?craft=knitting&availability=ravelry%2Bfree"

// GetResults returns an iterator through all free ravelry knitting patterns.
func GetResults(c *ravelry.Client) *Result {
	req, err := http.NewRequest(http.MethodGet, ravelry.Endpoint+patternSearchEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var psr *Result
	err = json.Unmarshal(body, &psr)
	if err != nil {
		fmt.Println(string(body))
		log.Fatal(err)
	}
	return psr
}
