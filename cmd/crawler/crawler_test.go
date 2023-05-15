package crawler_test

import (
	"github.com/jwambugu/crawler/cmd/crawler"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
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
	cl.Crawl("http://localhost.com")

	filenames := cl.GetFilenames()

	for _, filename := range filenames {
		_, err := os.Stat(filename)
		require.NoError(t, err)
	}

	cl.Crawl("http://localhost.com")
	require.Len(t, cl.GetFilenames(), len(filenames))
}
