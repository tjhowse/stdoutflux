package main

import (
	"flag"
	"fmt"
	"html"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/microcosm-cc/bluemonday"
	"github.com/muesli/reflow/wordwrap"
	miniflux "miniflux.app/v2/client"
)

type config struct {
	MinifluxURL string `env:"MINIFLUX_URL"`
	APIToken    string `env:"MINIFLUX_API_TOKEN"`
	LineLength  int    `env:"STDOUTFLUX_LINE_LENGTH" envDefault:"32"`
	MarkRead    bool   `env:"STDOUTFLUX_MARK_READ" envDefault:"false"`
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

func lineFilter(s string) bool {
	// This returns false for lines with stuff we don't want to print, like
	// "Read more of this story at Slashdot"
	if strings.Contains(s, "Read more of this story at") {
		return false
	}
	return true
}

func formatForThermalPrinter(s string, lineLength int) string {
	// Filter out lines that are only whitespace
	var cleaned []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !lineFilter(line) {
			continue
		}
		cleaned = append(cleaned, line)
	}

	var outLines []string
	for _, line := range cleaned {
		// Word-wrap without splitting words, respecting rune widths
		wrapped := wordwrap.String(line, lineLength)
		outLines = append(outLines, wrapped)
	}

	if len(outLines) == 0 {
		return ""
	}
	return strings.Join(outLines, "\n")
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
	markRead := flag.Bool("markread", false, "Mark entries as read after fetching (default false)")
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
	if *markRead {
		cfg.MarkRead = *markRead
	}

	// Create Miniflux client.
	client := miniflux.NewClient(cfg.MinifluxURL, cfg.APIToken)

	// Fetch unread entries
	entries, err := client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		fmt.Println(err)
		return
	}

	readEntries := []int64{}
	for _, entry := range entries.Entries {
		fmt.Println("===============")
		fmt.Println(formatForThermalPrinter(entry.Title, cfg.LineLength))
		fmt.Println("---------------")
		entry.Content = bm.Sanitize(entry.Content)
		entry.Content = html.UnescapeString(entry.Content)
		entry.Content = formatForThermalPrinter(entry.Content, cfg.LineLength)
		fmt.Println(entry.Content)
		readEntries = append(readEntries, entry.ID)
	}

	if cfg.MarkRead {
		err = client.UpdateEntries(readEntries, miniflux.EntryStatusRead)
		if err != nil {
			fmt.Println("Failed to mark entry as read:", err)
		}
	}
}
