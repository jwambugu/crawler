package crawler

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var source, downloadsDir string

func init() {
	rootCmd.Flags().StringVarP(&source, "source", "s", "", "URL to scrap data from.")
	rootCmd.Flags().StringVarP(&downloadsDir, "directory", "d", "downloads", "Directory to store downloaded src contents")

	_ = rootCmd.MarkFlagRequired("source")
}

var rootCmd = &cobra.Command{
	Use:   "crawler",
	Short: "crawler is a minimalistic web crawler",
	Long:  `A simple web crawl that crawls all relative links for the provided URL. It aims to be fast and flexible web crawler.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}

			client = &http.Client{
				Transport: transport,
			}
		)

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			_, _ = fmt.Fprintf(os.Stderr, "Signal interrupt - exiting\n")
			os.Exit(1)
		}()

		cl := NewCrawler(client, downloadsDir)
		source = strings.Trim(source, "/")

		var wg sync.WaitGroup
		wg.Add(1)

		go cl.Crawl(source, &wg)
		wg.Wait()
	},
}

// Execute runs the command using the provided args.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "There was an error while executing your command '%s'", err)
		os.Exit(1)
	}
}
