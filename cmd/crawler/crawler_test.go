package crawler_test

import (
	"github.com/jwambugu/crawler/cmd/crawler"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"testing"
)

var testsDir = "storage/tests/"

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	m.Run()
}

func cleanup() {
	dirs, err := os.ReadDir(testsDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, dir := range dirs {
		if err = os.RemoveAll(path.Join([]string{testsDir, dir.Name()}...)); err != nil {
			log.Fatal(err)
		}
	}
}

func TestCrawler_CrawlWithoutConcurrency(t *testing.T) {
	t.Parallel()

	defer func() {
		cleanup()
	}()

	httpClient := crawler.NewMockHttpClient()

	httpClient.MockRequest("http://localhost.com", func() (status int, body string) {
		return http.StatusOK, `
	<ul>
		<a href="/">Home</a>
		<a href="/advanced-features">Advance features</a>
		<a href="/advanced-feature-removed">Advance features</a>
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

	httpClient.MockRequest("http://localhost.com/advanced-feature-removed", func() (status int, body string) {
		return http.StatusNotFound, `<p>Not Found</p>`
	})

	cl := crawler.NewCrawler(httpClient, "tests")
	cl.CrawlWithoutConcurrency("http://localhost.com")
}

func TestCrawler_Crawl(t *testing.T) {
	t.Parallel()

	defer func() {
		cleanup()
	}()

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

func BenchmarkCrawler_Crawl(b *testing.B) {
	defer func() {
		cleanup()
	}()

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
	defer func() {
		cleanup()
	}()

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
