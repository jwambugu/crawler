package crawler_test

import (
	"github.com/jwambugu/crawler/cmd/crawler"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"sync"
	"testing"
)

func TestCrawler_Crawl(t *testing.T) {
	httpClient := crawler.NewMockHttpClient()

	httpClient.MockRequest("http://localhost.com", func() (status int, body string) {
		return http.StatusOK, `
	<ul>
		<a href="/">Home</a>
		<a href="/advanced-features">Advance features</a>
		<a href="https://google.com"> External </a>
	</ul>`
	})

	httpClient.MockRequest("http://localhost.com/advanced-features", func() (status int, body string) {
		return http.StatusOK, `
		<ul>
			<a href="/">Home</a>
			<a href="/advanced-features">Advance features</a>
			<a href="https://google.com"> External </a>
		</ul>

		<section>
			<h2>Progress bar <a href="https://whatsapp.com"></a></h2>
		</section>`
	})

	cl := crawler.NewCrawler(httpClient, "tests")

	var wg sync.WaitGroup
	wg.Add(1)

	go cl.Crawl("http://localhost.com", &wg)
	wg.Wait()

	filenames := cl.GetFilenames()

	defer func() {
		for _, filename := range filenames {
			err := os.Remove(filename)
			require.NoError(t, err)
		}
	}()

	for _, filename := range filenames {
		_, err := os.Stat(filename)
		require.NoError(t, err)
	}

}
