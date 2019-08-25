package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/rebeku/patternmaker/src/get_patterns/ravelry"
)

const PatternSearchEndpoint = "https://api.ravelry.com/patterns/search.json?craft=knitting&availability=ravelry%2Bfree"
const PatternDetailsEndpoint = "https://api.ravelry.com/patterns.json?ids="
const idKey = "ids"

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

func toMap(sm sync.Map) map[string]interface{} {
	m := make(map[string]interface{})
	saveVals := func(k, v interface{}) bool {
		m[fmt.Sprintf("%v", k)] = v
		return true
	}
	sm.Range(saveVals)
	return m
}

func main() {
	username := os.Getenv("RAVELRY_CONSUMER")
	password := os.Getenv("RAVELRY_SECRET")
	c := ravelry.NewClient(username, password)

	psr := getPatternSearchResults(c, username, password)
	fmt.Printf("Successfully downloaded %d patterns\n", psr.Paginator.Results)
	pats := getPatternDetails(c, username, password, psr)

	fmt.Println(pats)
	var fnameMap sync.Map
	var wg sync.WaitGroup
	wg.Add(len(pats))

	for id, pat := range pats {
		id := id
		pat := pat
		go func() {
			defer wg.Done()
			filenames, err := ravelry.GetPDF(c, pat.DownloadLocation.URL)
			if err != nil {
				fmt.Println("Error scraping download locations: ", err)
				return
			}
			if len(filenames) == 0 {
				fmt.Println("Found no filenames for url: ", pat.DownloadLocation.URL)
			}
			fnameMap.Store(id, filenames)
		}()
	}

	wg.Wait()
	nameMap := toMap(fnameMap)
	fmt.Println(nameMap)

	jPats, err := json.Marshal(nameMap)

	if err != nil {
		fmt.Println("Error marshalling pats to json: ", err)
		return
	}

	f, err := os.Create("data/patterns.json")
	if err != nil {
		fmt.Println("Error creating new file: ", err)
		return
	}
	defer f.Close()

	_, err = f.Write(jPats)
	if err != nil {
		fmt.Println("Error writing patterns to file: ", err)
	}
}
