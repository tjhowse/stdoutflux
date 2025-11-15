package main

import (
	"flag"
	"fmt"
	"html"

	"github.com/caarlos0/env/v11"
	"github.com/microcosm-cc/bluemonday"
	miniflux "miniflux.app/v2/client"
)

type config struct {
	MinifluxURL string `env:"MINIFLUX_URL"`
	APIToken    string `env:"MINIFLUX_API_TOKEN"`
	LineLength  int    `env:"LINE_LENGTH" envDefault:"32"`
}

func trimWhitespace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func splitLines(s string) []string {
	lines := []string{}
	currentLine := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(r)
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}
	return lines
}

func formatForThermalPrinter(s string, lineLength int) string {

	// Strip lines that are just whitespace
	// and trim leading/trailing whitespace
	lines := []rune{}
	currentLine := []rune{}
	for _, r := range s {
		if r == '\n' {
			lineStr := string(currentLine)
			if len(lineStr) > 0 && len(trimWhitespace(lineStr)) > 0 {
				lines = append(lines, currentLine...)
				lines = append(lines, '\n')
			}
			currentLine = []rune{}
		} else {
			currentLine = append(currentLine, r)
		}
	}
	// Add the last line if any
	lineStr := string(currentLine)
	if len(lineStr) > 0 && len(trimWhitespace(lineStr)) > 0 {
		lines = append(lines, currentLine...)
	}
	s = string(lines)

	// Wrap lines to the specified length
	wrappedContent := ""
	for _, line := range splitLines(s) {
		for len(line) > lineLength {
			wrappedContent += line[:lineLength] + "\n"
			line = line[lineLength:]
		}
		wrappedContent += line + "\n"
	}
	return wrappedContent
}

func main() {

	cfg := config{}

	bm := bluemonday.StrictPolicy()

	if err := env.Parse(&cfg); err != nil {
		fmt.Println("Failed to parse env vars:", err)
		return
	}

	minifluxURL := flag.String("url", "", "Miniflux instance URL")
	apiToken := flag.String("token", "", "API token for Miniflux")
	lineLength := flag.Int("linelength", 32, "Maximum line length for content wrapping")
	flag.Parse()

	// Make CLI flags override env vars if provided
	if *minifluxURL != "" {
		cfg.MinifluxURL = *minifluxURL
	}
	if *apiToken != "" {
		cfg.APIToken = *apiToken
	}
	if *lineLength != 32 {
		cfg.LineLength = *lineLength
	}

	// Create Miniflux client.
	client := miniflux.NewClient(cfg.MinifluxURL, cfg.APIToken)

	// Fetch unread entries
	entries, err := client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, entry := range entries.Entries {
		fmt.Println("===============")
		fmt.Println(formatForThermalPrinter(entry.Title, cfg.LineLength))
		fmt.Println("---------------")
		entry.Content = bm.Sanitize(entry.Content)
		entry.Content = html.UnescapeString(entry.Content)
		entry.Content = formatForThermalPrinter(entry.Content, cfg.LineLength)
		fmt.Println(entry.Content)
		// fmt.Println(content2)

	}
}
