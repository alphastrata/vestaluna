package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector()

	var xmls []string

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		if strings.Contains(link, ".xml") {
			xmls = append(xmls, link)
		}
		// Visit link found on page
		// Only those links are visited which are matched by  any of the URLFilter regexps
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on http://httpbin.org
	c.Visit("./entries.html")
}
