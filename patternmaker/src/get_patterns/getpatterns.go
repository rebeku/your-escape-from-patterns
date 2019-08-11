package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rebeku/patternmaker/src/get_patterns/client"
)

const PatternSearchEndpoint = "https://api.ravelry.com/patterns/search.json?craft=knitting&availability=ravelry%2Bfree"
const PatternDetailsEndpoint = "https://api.ravelry.com/patterns.json?"
const idKey = "ids"

func main() {
	username := os.Getenv("RAVELRY_CONSUMER")
	password := os.Getenv("RAVELRY_SECRET")
	c := &http.Client{}

	psr := getPatternSearchResults(c, username, password)
	fmt.Printf("Successfully downloaded %d patterns\n", psr.Paginator.Results)
	pats := getPatternDetails(c, username, password, psr)
	fmt.Println(pats)
}

func getPatternSearchResults(c *http.Client, username, password string) *client.PatternSearchResult {
	req, err := http.NewRequest(http.MethodGet, PatternSearchEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var psr *client.PatternSearchResult
	err = json.Unmarshal(body, &psr)
	if err != nil {
		fmt.Println(string(body))
		log.Fatal(err)
	}
	return psr
}

func getPatternDetails(c *http.Client, username, password string, psr *client.PatternSearchResult) map[string]client.Pattern {
	v := url.Values{}
	ids := strings.Join(psr.GetPatternIDs()[:6], "+")
	v.Set(idKey, ids)

	detailsEndpoint := PatternDetailsEndpoint + v.Encode()
	fmt.Println(detailsEndpoint)

	req, err := http.NewRequest(http.MethodGet, detailsEndpoint, nil)
	if err != nil {
		log.Println(err)
	}
	req.SetBasicAuth(username, password)

	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var pdsr client.PatternDetailSearchResult
	err = json.Unmarshal(body, &pdsr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got details for %d patterns\n", len(pdsr.Patterns))
	return pdsr.Patterns
}
