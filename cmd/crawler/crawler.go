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
	httpClient   HttpClient

	mu           sync.Mutex
	filenames    map[string]struct{}
	visitedLinks map[string]struct{}
}

func (c *Crawler) GetFilenames() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	files := make([]string, len(c.filenames))
	var counter int
	for s := range c.filenames {
		files[counter] = s
		counter++
	}

	return files
}

func (c *Crawler) GetVisitedLinks() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	links := make([]string, len(c.visitedLinks))
	var counter int
	for s := range c.visitedLinks {
		links[counter] = s
		counter++
	}

	return links
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

func (c *Crawler) craw(uri *url.URL) (io.Reader, string, error) {
	filename := uri.Host + strings.ReplaceAll(uri.Path, "/", "_") + `.html`
	filename = c.downloadsDir + `/` + filename

	buffer := new(bytes.Buffer)

	contents, err := os.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, "", fmt.Errorf("crawl: read file - %v", err)
		}

		response, err := c.PageDownloader(uri.String())
		if err != nil {
			return nil, "", err
		}

		contents, err = io.ReadAll(response)
		if err != nil {
			return nil, "", fmt.Errorf("crawl: read response - %v", err)
		}

		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, "", fmt.Errorf("crawl: create file - %v", err)
		}

		if _, err = f.Write(contents); err != nil {
			return nil, "", fmt.Errorf("crawl: write to file - %v\n", err)
		}
	}

	buffer.Write(contents)
	return buffer, filename, nil
}

func (c *Crawler) CrawlWithoutConcurrency(link string) {
	uri, err := url.Parse(link)
	if err != nil {
		log.Printf("crawl: parse url - %v\n", err)
		return
	}

	reader, filename, err := c.craw(uri)
	if err != nil {
		log.Println(err)
		return
	}

	c.filenames[filename] = struct{}{}

	allLinks := GetLinks(uri, reader)
	for _, l := range allLinks {
		if _, exists := c.visitedLinks[l]; !exists {
			c.visitedLinks[l] = struct{}{}

			log.Printf("-- %s\n", l)
			c.CrawlWithoutConcurrency(l)
		}
	}
}

func (c *Crawler) Crawl(link string, wg *sync.WaitGroup) {
	defer wg.Done()

	c.mu.Lock()
	defer c.mu.Unlock()

	uri, err := url.Parse(link)
	if err != nil {
		log.Printf("crawl: parse url - %v\n", err)
		return
	}

	reader, filename, err := c.craw(uri)
	if err != nil {
		log.Println(err)
		return
	}

	c.filenames[filename] = struct{}{}

	allLinks := GetLinks(uri, reader)
	for _, l := range allLinks {
		if _, exists := c.visitedLinks[l]; !exists {
			c.visitedLinks[l] = struct{}{}

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
