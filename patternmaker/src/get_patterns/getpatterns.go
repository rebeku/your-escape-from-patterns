package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/rebeku/patternmaker/src/get_patterns/model"
)

func main() {
	username := os.Getenv("RAVELRY_CONSUMER")
	password := os.Getenv("RAVELRY_SECRET")
	endpoint := "https://api.ravelry.com//patterns/search.json?craft=knitting&availability=ravelry%2Bfree"

	c := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
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

	var psr model.PatternSearchResult
	err = json.Unmarshal(body, &psr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(psr.Patterns)
	/*
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(body))
	*/
}
