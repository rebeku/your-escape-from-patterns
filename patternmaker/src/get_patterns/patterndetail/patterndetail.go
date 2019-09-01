package patterndetail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/rebeku/patternmaker/src/get_patterns/patternsearch"
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
	Patterns map[string]Pattern
}

const patternDetailsEndpoint = "patterns.json?ids="

// GetResults looks up patterns by ID and extracts details.
// Most importantly, this includes the download location
func GetResults(c *ravelry.Client, psr *patternsearch.Result) map[string]Pattern {
	ids := strings.Join(psr.GetPatternIDs(), "+")
	detailsEndpoint := ravelry.Endpoint + patternDetailsEndpoint + ids

	req, err := http.NewRequest(http.MethodGet, detailsEndpoint, nil)
	if err != nil {
		log.Println(err)
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var pdsr PatternDetailSearchResult
	err = json.Unmarshal(body, &pdsr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got details for %d patterns\n", len(pdsr.Patterns))
	return pdsr.Patterns
}
