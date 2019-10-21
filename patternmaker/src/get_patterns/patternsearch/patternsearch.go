package patternsearch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

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

var maxPages = 2

// GetResults returns an iterator through all free ravelry knitting patterns.
func GetResults(c *ravelry.Client) (chan []string, chan error) {
	rc := make(chan []string, 1)
	ec := make(chan error, 1)

	var wg sync.WaitGroup

	page := 1
	lastPage := sendResult(c, page, rc, ec)

	for !lastPage {
		wg.Add(1)

		page++
		if page >= maxPages {
			fmt.Println("lastPage = true")
			lastPage = true
		}
		page := page
		go func() {
			defer wg.Done()
			lastPage = sendResult(c, page, rc, ec)
		}()
	}
	go func() {
		wg.Wait()
		fmt.Println("Closing pattern search channel")
		close(rc)
		close(ec)
	}()
	return rc, ec
}

func sendResult(c *ravelry.Client, page int, rc chan []string, ec chan error) bool {
	log.Printf("Getting results for page %d", page)
	r, err := getResultPage(c, page)
	if err == nil {
		if r.Paginator.LastPage < maxPages {
			maxPages = r.Paginator.LastPage
		}
		log.Printf("Got %d results for page %d", r.Paginator.PageSize, r.Paginator.Page)
		rc <- r.getPatternIDs()
	} else {
		log.Printf("Failed to get results for page %d", page)
		ec <- err
		return page == 0
	}
	return page == r.Paginator.LastPage
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
