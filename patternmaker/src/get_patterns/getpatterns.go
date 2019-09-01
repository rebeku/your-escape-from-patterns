package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/rebeku/patternmaker/src/get_patterns/downloadurl"
	"github.com/rebeku/patternmaker/src/get_patterns/patterndetail"
	"github.com/rebeku/patternmaker/src/get_patterns/patternsearch"
	"github.com/rebeku/patternmaker/src/get_patterns/ravelry"
)

const idKey = "ids"

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

	psr := patternsearch.GetResults(c)
	fmt.Printf("Successfully downloaded %d patterns\n", psr.Paginator.Results)
	pats := patterndetail.GetResults(c, psr)

	var fnameMap sync.Map
	var wg sync.WaitGroup
	wg.Add(len(pats))

	for id, pat := range pats {
		id := id
		url := pat.DownloadLocation.URL
		go func() {
			defer wg.Done()
			filenames, err, desc := downloadurl.GetDownloadURL(c, url)
			if err != nil {
				fmt.Println("Error scraping download locations: ", err)
				return
			}
			if len(filenames) == 0 {
				fmt.Printf("ID: %s\nURL: %s\ndescription: %s\nfilenames: %v\n\n", id, url, desc, filenames)
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
