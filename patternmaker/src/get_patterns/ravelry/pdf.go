package ravelry

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

func GetPDF(c *Client, urlString string) ([]string, error) {

	req, err := http.NewRequest(http.MethodGet, urlString, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		// This is totally sketch but I am saving the redirect URL
		// as an error in the redirect policy because that's the only
		// concurrencys-safe way I know to do it.
		redirectURL := err.(*url.Error).Err.Error()
		_, urlErr := url.Parse(redirectURL)
		if urlErr != nil {
			return nil, err
		}
		return []string{redirectURL}, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get page with status code%d", resp.StatusCode)
	}
	fmt.Println(resp.StatusCode)
	filenames, err := getFilenames(resp.Body)
	if err != nil {
		return nil, err
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
	var bod []byte
	_, err := body.Read(bod)
	if err != nil {
		return nil, err
	}
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
