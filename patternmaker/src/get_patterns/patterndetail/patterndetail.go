package patterndetail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/rebeku/patternmaker/src/get_patterns/ravelry"
)

type Pattern struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	DownloadLocation DownloadLocation `json:"download_location"`
}

type DownloadLocation struct {
	URL string `json:"url"`
}

type PatternDetailSearchResult struct {
	Patterns map[string]*Pattern
}

const patternDetailsEndpoint = "patterns.json?ids="

// GetResults looks up patterns by ID and extracts details.
// Most importantly, this includes the download location
func GetResults(c *ravelry.Client, psc chan []string) (chan *Pattern, chan error) {
	rc := make(chan *Pattern, 1)
	ec := make(chan error, 1)

	var wg sync.WaitGroup

	for patternIDs := range psc {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pats, err := getResults(c, patternIDs)
			if err == nil {
				fmt.Println("Got pattern details")
				for _, p := range pats {
					rc <- p
				}
			} else {
				fmt.Println("Failed to get pattern details.")
				ec <- err
			}
		}()
	}
	go func() {
		wg.Wait()
		fmt.Println("Closing pattern detail channels")
		close(rc)
		close(ec)
	}()
	return rc, ec
}

var badStatusErrorString = "Failed to get pattern detail result for ids %v"

func getResults(c *ravelry.Client, patternIDs []string) (map[string]*Pattern, error) {
	ids := strings.Join(patternIDs, "+")
	// we cannot use url.Encode here because Ravelry API uses unescape '+' sign
	detailsEndpoint := ravelry.Endpoint + patternDetailsEndpoint + ids

	req, err := http.NewRequest(http.MethodGet, detailsEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(badStatusErrorString, ids)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pdsr PatternDetailSearchResult
	err = json.Unmarshal(body, &pdsr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Got details for %d patterns\n", len(pdsr.Patterns))
	return pdsr.Patterns, nil
}
