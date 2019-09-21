package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rebeku/patternmaker/src/get_patterns/downloadurl"
	"github.com/rebeku/patternmaker/src/get_patterns/patterndetail"
	"github.com/rebeku/patternmaker/src/get_patterns/patternsearch"
	"github.com/rebeku/patternmaker/src/get_patterns/ravelry"
)

const idKey = "ids"

func main() {
	username := os.Getenv("RAVELRY_CONSUMER")
	password := os.Getenv("RAVELRY_SECRET")
	c := ravelry.NewClient(username, password)

	psc, psec := patternsearch.GetResults(c)
	pdc, pdec := patterndetail.GetResults(c, psc)

	go func() {
		for err := range psec {
			fmt.Println("Error searching pattern: ", err)
		}
	}()
	go func() {
		for err := range pdec {
			fmt.Println("Error getting pattern detail: ", err)
		}
	}()
	fnameMap := make(map[string]downloadurl.DownloadLoc)

	urlc := downloadurl.DownloadURLSource(c, pdc)
	for dl := range urlc {
		fnameMap[fmt.Sprintf("%d", dl.ID)] = dl
	}

	jPats, err := json.MarshalIndent(fnameMap, "", "    ")

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
