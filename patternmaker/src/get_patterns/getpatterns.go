package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rebeku/patternmaker/src/get_patterns/ravelry"
)

const PatternSearchEndpoint = "https://api.ravelry.com/patterns/search.json?craft=knitting&availability=ravelry%2Bfree"
const PatternDetailsEndpoint = "https://api.ravelry.com/patterns.json?ids="
const idKey = "ids"

func main() {
	username := os.Getenv("RAVELRY_CONSUMER")
	password := os.Getenv("RAVELRY_SECRET")
	c := ravelry.NewClient(username, password)

	psr := getPatternSearchResults(c, username, password)
	fmt.Printf("Successfully downloaded %d patterns\n", psr.Paginator.Results)
	pats := getPatternDetails(c, username, password, psr)
	fmt.Println(pats)
}

func getPatternSearchResults(c *ravelry.Client, username, password string) *ravelry.PatternSearchResult {
	req, err := http.NewRequest(http.MethodGet, PatternSearchEndpoint, nil)
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

	var psr *ravelry.PatternSearchResult
	err = json.Unmarshal(body, &psr)
	if err != nil {
		fmt.Println(string(body))
		log.Fatal(err)
	}
	return psr
}

func getPatternDetails(c *ravelry.Client, username, password string, psr *ravelry.PatternSearchResult) map[string]ravelry.Pattern {
	ids := strings.Join(psr.GetPatternIDs(), "+")
	detailsEndpoint := PatternDetailsEndpoint + ids

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

	var pdsr ravelry.PatternDetailSearchResult
	err = json.Unmarshal(body, &pdsr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got details for %d patterns\n", len(pdsr.Patterns))
	return pdsr.Patterns
}
