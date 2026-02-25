package main

import (
	"log"
	"log/slog"
	"os"

	_ "embed"

	"github.com/gocolly/colly"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

func main() {
	// Set up logging to file
	logFile, err := os.OpenFile("crawler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("failed to open log file:", err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, nil))
	slog.SetDefault(logger)

	db, err := sqlx.Connect("sqlite", "database.sqlite3")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	db.MustExec(schema)

	c := colly.NewCollector(
		colly.AllowedDomains("www.ida.liu.se"),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")

		is_visited, err := IsVisited(db, href)
		if err != nil {
			slog.Error("failed to check if page is visited", "err", err)
		}

		if !is_visited {
			c.Visit(e.Request.AbsoluteURL(href))
		}
	})

	c.OnResponse(func(r *colly.Response) {
		AddVisited(db, r.Request.URL.String())
		content_id, err := AddContent(db, r.Body)
		if err != nil {
			slog.Error("failed to add content", "err", err)
		}

		content_type_raw := r.Headers.Get("Content-Type")
		var content_type *string = nil
		if content_type_raw != "" {
			content_type = &content_type_raw
		}

		_, err = AddPageVisit(db, r.Request.URL.String(), r.StatusCode, content_type, content_id)
		if err != nil {
			slog.Error("failed to add page visit", "err", err)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		slog.Info("Visiting page", "url", r.URL)
	})

	err = c.Visit("https://www.ida.liu.se/")
	if err != nil {
		slog.Error("failed to visit the root page", "err", err)
	}
}
