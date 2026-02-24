package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.ida.liu.se"),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		c.Visit(e.Request.AbsoluteURL(href))
	})

	c.OnResponse(func(r *colly.Response) {
		filename := fmt.Sprintf("pages/%s", r.FileName())
		err := os.WriteFile(filename, r.Body, 0666)
		if err != nil {
			slog.Error("failed to write page to disk", "filename", filename, "err", err)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	err := c.Visit("https://www.ida.liu.se/")
	if err != nil {
		slog.Error("failed to visit the root page", "err", err)
	}
}
