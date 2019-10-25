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

func (psr Result) getPatternIDs() []string {
	ids := make([]string, len(psr.PatternIDs))
	for i, p := range psr.PatternIDs {
		ids[i] = fmt.Sprintf("%d", p.ID)
	}
	return ids
}

const patternSearchEndpoint = "patterns/search.json?craft=knitting&availability=ravelry%%2Bfree&page=%d"

const nWorkers = 3

var maxPages = 2

func GetFreeDownloads(c *ravelry.Client) chan []string {
	out := make(chan []string)

	page := 0

	work := func() (bool, func()) {
		page++
		return page > maxPages, func() {
			sendResult(c, page, out)
		}
	}
	close := func() {
		fmt.Println("Closing pattern search channel")
		close(out)
	}

	qr := ravelry.NewQueryRunner(nWorkers, work, close)
	qr.Run()
	return out
}

func sendResult(c *ravelry.Client, page int, rc chan []string) {
	log.Printf("Getting results for page %d", page)
	r, err := getResultPage(c, page)
	if err == nil {
		if r.Paginator.LastPage < maxPages {
			maxPages = r.Paginator.LastPage
		}
		log.Printf("Got %d results for page %d", r.Paginator.PageSize, r.Paginator.Page)
		rc <- r.getPatternIDs()
	} else {
		log.Printf("Failed to get results for page %d: %v", page, err)
	}
}

const badStatusErrorString = "Failed to get pattern search page %d with status %q"

func getResultPage(c *ravelry.Client, page int) (*Result, error) {
	endpoint := fmt.Sprintf(ravelry.Endpoint+patternSearchEndpoint, page)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(badStatusErrorString, page, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var psr *Result
	err = json.Unmarshal(body, &psr)
	return psr, err
}
