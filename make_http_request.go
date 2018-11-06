// make_http_request.go
package main

import (
	"fmt"
	"github.com/gocolly/colly/debug"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"

	"github.com/gocolly/colly"
	"scrapy/sources"
)

const (
	delay        = 2 * time.Second
	urlBase      = "https://www.kijiji.ca"
	urlCondoRent = urlBase + "/b-appartement-condo/ville-de-montreal/c37l1700281?ad=offering"
	price        = "price=%s__%s"
	withImg      = "minNumberOfImages=1"
	withPet      = "animaux-acceptes=1"
	furnished    = "meuble=1"
	rentedBy     = "a-louer-par=%s"
)

func main() {

	db, c := setup("postgres", "host=localhost port=5432 user=scrapy dbname=scrapy password=scrapy sslmode=disable", true, delay)

	defer db.Close()

	kijiji.VisitNextPage(c, kijiji.SearchLimitYesterday)
	kijiji.VisitAd(c, db, urlBase)

	kijiji.ParseAd(c, db)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(urlCondoRent)

	c.Wait()
}

func setup(dialect string, connectionString string, async bool, del time.Duration) (db *gorm.DB, c *colly.Collector) {
	db, err := gorm.Open(dialect, connectionString)

	if err != nil {
		panic("failed to connect database.\nWith error:\n" + err.Error())
	}

	kijiji.Setup(db)

	c = colly.NewCollector(
		// Turn on asynchronous requests
		colly.Async(async),
		// Attach a debugger to the collector
		colly.Debugger(&debug.LogDebugger{}),
		//.AllowedDomains("kijiji.ca"),
	)

	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*httpbin.*",
		Parallelism: 1,
		Delay:       del,
	})

	return db, c
}
