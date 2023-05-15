package crawler

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var source, dir string

func init() {
	rootCmd.Flags().StringVarP(&source, "source", "s", "", "URL to scrap data from.")
	rootCmd.Flags().StringVarP(&dir, "directory", "d", "downloads", "Directory to store downloaded src contents")

	_ = rootCmd.MarkFlagRequired("source")
}

var rootCmd = &cobra.Command{
	Use:   "crawler",
	Short: "crawler is a minimalistic web crawler",
	Long:  `A simple web crawl that crawls all relative links for the provided URL. It aims to be fast and flexible web crawler.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(source, dir)
	},
}

// Execute runs the command using the provided args.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "There was an error while executing your command '%s'", err)
		os.Exit(1)
	}
}
