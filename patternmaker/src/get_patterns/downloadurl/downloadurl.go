package downloadurl

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rebeku/patternmaker/src/get_patterns/patterndetail"
	"github.com/rebeku/patternmaker/src/get_patterns/ravelry"
	"golang.org/x/net/html"
)

type DownloadLoc struct {
	ID       int
	StartURL string
	URLs     []string
}

const nWorkers = 1

func DownloadURLSource(c *ravelry.Client, pdc chan *patterndetail.Pattern) chan DownloadLoc {
	out := make(chan DownloadLoc)
	done := make(chan struct{})

	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		go func() {
			for {
				workDone := worker(c, &wg, pdc, out)
				if workDone {
					done <- struct{}{}
					return
				}
			}
		}()
	}

	go func() {
		<-done
		wg.Wait()
		fmt.Println("Closing downloadURL channel")
		close(out)
	}()
	return out
}

var ErrNoURLs = errors.New("No URLs for pattern")

func worker(c *ravelry.Client, wg *sync.WaitGroup, in chan *patterndetail.Pattern, out chan DownloadLoc) bool {
	pat, more := <-in
	if !more {
		return true
	}
	wg.Add(1)
	defer wg.Done()

	urlString := pat.DownloadLocation.URL
	urls, err := getDownloadURL(c, urlString)
	if err != nil {
		fmt.Println("Error getting download URL: ", err)
		return false
	}
	if len(urls) == 0 {
		fmt.Printf("No URLs for job %d\n", pat.ID)
		return false
	}
	fmt.Printf("Got URL for pattern %d\n", pat.ID)
	out <- DownloadLoc{
		ID:       pat.ID,
		StartURL: urlString,
		URLs:     urls,
	}
	return false
}

// GetDownloadURL scrapes URL of actual pattern downloads from page content
// unfortunately, this is not avialable through the API
func getDownloadURL(c *ravelry.Client, urlString string) ([]string, error) {
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
			//fmt.Printf("Error parsing redirect URL %s: %v\n", redirectURL, urlErr)
			return nil, err
		}
		time.Sleep(10 * time.Second)
		return []string{redirectURL}, nil
	} else if resp.StatusCode == http.StatusServiceUnavailable {
		time.Sleep(time.Duration(int64(rand.Float64() * 1e9)))
		return getDownloadURL(c, urlString)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get page with status code%d", resp.StatusCode)
	}
	defer resp.Body.Close()
	return getFilenames(resp.Body)
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
