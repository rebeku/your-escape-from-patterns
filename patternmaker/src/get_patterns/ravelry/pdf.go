package ravelry

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

func GetPDF(c *Client, url string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	filenames, err := getFilenames(resp.Body)
	if err != nil {
		return err
	}
	return filenames, nil
	// TODO: actually select, download, and parse file
	// englishFile := selectEnglishPdf(filenames)
	// return parseWordsFromPDF(englishFile)
}

const class = "class"
const filename = "filename"

func isFilenameClass(z *html.Tokenizer) bool {
	key, val, moreAttr := z.TagAttr()
	if string(key) == class && string(val) == filename {
		return true
	}
	if moreAttr {
		return isFilenameClass(z)
	}
	return false
}

func getFilenames(body io.Reader) ([]string, error) {
	z := html.NewTokenizer(body)
	filenames := make([]string, 0, 10)
	expectFilenameNext := false
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			err := z.Err()
			if err == io.EOF {
				break
			}
			return filenames, z.Err()
		} else if tt != html.StartTagToken {
			continue
		}
		if expectFilenameNext {
			_, hasAttr := z.TagName()
			if !hasAttr {
				fmt.Println("ERROR: Filename has no attributes.")
				continue
			}
			for {
				key, val, moreAttr := z.TagAttr()
				if string(key) == "href" {
					filenames = append(filenames, string(val))
					break
				}
				if !moreAttr {
					break
				}
			}
			expectFilenameNext = false
			continue
		}
		name, hasAttr := z.TagName()
		if string(name) != "div" || !hasAttr {
			continue
		}
		if isFilenameClass(z) {
			expectFilenameNext = true
		}
	}
	// this should never happen.
	return filenames, nil

}
