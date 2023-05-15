package crawler

import (
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Crawler struct {
	httpClient   HttpClient
	visitedLinks map[string]struct{}
}

func (c *Crawler) PageDownloader(link string) (io.Reader, error) {
	defer func() {
		log.Printf("Downloading %s\n", link)
		c.visitedLinks[link] = struct{}{}
	}()

	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, fmt.Errorf("crawl: create request - %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crawl: make request - %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crawl: got status %d", res.StatusCode)
	}

	return res.Body, nil
}

func (c *Crawler) Crawl(link string) {
	response, err := c.PageDownloader(link)
	if err != nil {
		log.Fatalln(err)
	}

	uri, err := url.Parse(link)
	if err != nil {
		log.Fatalf("crawl: parse url - %v\n", err)
	}

	allLinks := GetLinks(uri, response)
	for _, l := range allLinks {
		if _, exists := c.visitedLinks[l]; !exists {
			c.Crawl(l)
		}
	}
}

func NewCrawler(cl HttpClient) *Crawler {
	return &Crawler{
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
