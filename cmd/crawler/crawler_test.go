package crawler_test

import (
	"github.com/jwambugu/crawler/cmd/crawler"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	m.Run()
}

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

	defer func() {
		err := os.RemoveAll("storage/tests")
		require.NoError(t, err)
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go cl.Crawl("http://localhost.com", &wg)
	wg.Wait()

}

func BenchmarkCrawler_Crawl(b *testing.B) {
	for i := 0; i < b.N; i++ {
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
	}
}

func BenchmarkCrawler_CrawlWithoutConcurrency(b *testing.B) {
	for i := 0; i < b.N; i++ {
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

		cl.CrawlWithoutConcurrency("http://localhost.com")
	}
}
