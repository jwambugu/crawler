package crawler

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const baseDownloadsPath = "storage"

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Crawler struct {
	downloadsDir string
	filenames    map[string]struct{}
	httpClient   HttpClient

	mu           sync.Mutex
	visitedLinks map[string]struct{}
}

func (c *Crawler) GetFilenames() []string {
	files := make([]string, len(c.filenames))
	var counter int
	for s := range c.filenames {
		files[counter] = s
		counter++
	}

	return files
}

func (c *Crawler) PageDownloader(link string) (io.Reader, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, fmt.Errorf("crawl: create request - %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crawl: make request - %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crawl: got status - %d on %s", res.StatusCode, link)
	}

	return res.Body, nil
}

func (c *Crawler) Crawl(link string, wg *sync.WaitGroup) {
	defer wg.Done()

	c.mu.Lock()
	defer c.mu.Unlock()

	uri, err := url.Parse(link)
	if err != nil {
		fmt.Printf("crawl: parse url - %v\n", err)
	}

	filename := uri.Host + strings.ReplaceAll(uri.Path, "/", "_") + `.html`
	filename = c.downloadsDir + `/` + filename

	buffer := new(bytes.Buffer)

	contents, err := os.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("crawl: read file - %v\n", err)
			return
		}

		response, err := c.PageDownloader(uri.String())
		if err != nil {
			fmt.Println(err)
			return
		}

		contents, err = io.ReadAll(response)
		if err != nil {
			fmt.Printf("crawl: read response - %v\n", err)
			return
		}

		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("crawl: create file - %v\n", err)
			return
		}

		if _, err = f.Write(contents); err != nil {
			fmt.Printf("crawl: write to file - %v\n", err)
			return
		}
	}

	buffer.Write(contents)
	allLinks := GetLinks(uri, buffer)

	for _, l := range allLinks {
		if _, exists := c.visitedLinks[l]; !exists {
			c.visitedLinks[l] = struct{}{}
			c.filenames[filename] = struct{}{}

			log.Printf("-- %s\n", l)

			wg.Add(1)
			go c.Crawl(l, wg)
		}
	}
}

func NewCrawler(cl HttpClient, dir string) *Crawler {
	dir = baseDownloadsPath + `/` + dir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0750); err != nil {
			log.Panicf("crawler: create downloads dir - %v", err)
		}
	}

	return &Crawler{
		downloadsDir: dir,
		filenames:    make(map[string]struct{}),
		httpClient:   cl,
		visitedLinks: make(map[string]struct{}),
	}
}

func GetLinks(uri *url.URL, r io.Reader) []string {
	links := make(map[string]struct{})

	defer func() {
		// Remove the source url if present. We don't want to crawl it
		delete(links, uri.String())
	}()

	tokenizer := html.NewTokenizer(r)
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			var foundLinks []string
			for s := range links {
				foundLinks = append(foundLinks, s)
			}
			return foundLinks
		}

		token := tokenizer.Token()
		if tokenType == html.StartTagToken && token.DataAtom == atom.A {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					// Remove email links
					if strings.Contains(attr.Val, "mailto") {
						continue
					}

					// Remove links with a # since most of them are references to current page
					if strings.Contains(attr.Val, "#") {
						continue
					}

					linkURL, err := url.Parse(attr.Val)
					if err != nil {
						return nil
					}

					if linkURL.Host == "" {
						link := uri.Scheme + `://` + uri.Host + attr.Val
						link = strings.Trim(link, "/")
						links[link] = struct{}{}
						continue
					}

					if linkURL.Host == uri.Host {
						link := strings.Trim(attr.Val, "/")
						links[link] = struct{}{}
					}
				}
			}
		}
	}
}
