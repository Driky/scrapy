package kijiji

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
	"time"
)

const (
	SearchLimitHoures    = "heures"
	SearchLimitYesterday = "hier"
)

type House struct {
	Guid        string `gorm:"PRIMARY_KEY"`
	Title       string
	Price       float32
	Address     string
	Attributes  []Attributes
	Description string
	CreatedAt   time.Time
}

type Attributes struct {
	Name  string
	Value string
}

func Setup(db *gorm.DB) {
	// Migrate the schema
	db.AutoMigrate(&House{})
}

func VisitAd(c *colly.Collector, db *gorm.DB, url string) {
	// Find every regular ads and visit link
	c.OnHTML(".regular-ad a.title", func(e *colly.HTMLElement) {
		if !adExist(db, url+strings.TrimSpace(e.Attr("href"))) {
			e.Request.Visit(url + strings.TrimSpace(e.Attr("href")))
		} else {
			fmt.Printf("House already exist with url: %s\n", url+strings.TrimSpace(e.Attr("href")))
		}
	})
}

func VisitNextPage(c *colly.Collector, limit string) {

	c.OnHTML("#mainPageContent", func(e *colly.HTMLElement) {
		stop := false
		e.ForEach(".regular-ad span.date-posted", func(i int, e *colly.HTMLElement) {
			stop = strings.Contains(e.Text, limit)
		})

		if !stop {
			e.Request.Visit(e.ChildAttr("a[title=\"Suivante\"]", "href"))
		} else {
			println("#############################################")
			println("#############################################")
			println("#############################################")
			println("#############################################")
			println("##############FOUND LIMIT MARKER#############")
			println("#############################################")
			println("#############################################")
			println("#############################################")
		}
	})
}

func adExist(db *gorm.DB, ad string) bool {
	var house House
	//db.First(&house, "guid LIKE ?", ad)
	db.Raw("SELECT guid FROM houses WHERE guid LIKE ?", "%"+ad+"%").Scan(&house)
	if house.Guid == "" {
		return false
	}

	return true
}

func getPriceFromAdHTMLElement(e *colly.HTMLElement) float32 {
	price := e.ChildAttr("span[class^=\"currentPrice-\"] span[content]", "content")
	price64, err := strconv.ParseFloat(price, 32)
	if err != nil {
		fmt.Printf("Error while fetching price for %s.\nError: %s", e.Request.URL, err)
		price64 = 0.0
	}
	return float32(price64)
}

func getAttributesFromAdHTMLElement(e *colly.HTMLElement) []Attributes {
	attributes := make([]Attributes, 0)

	e.ForEach("div[id=\"vip-body\"] ul[class^=\"itemAttributeList-\"] li dl", func(i int, e *colly.HTMLElement) {
		attributes = append(attributes, Attributes{e.ChildText("dt"), e.ChildText("dd")})
	})

	return attributes
}

func getDescriptionFromAdHTMLElement(e *colly.HTMLElement) string {
	var description string

	e.ForEach("div[id=\"vip-body\"] div[class^=\"descriptionContainer-\"] p", func(i int, e *colly.HTMLElement) {
		description += e.Text
	})

	return description
}

func saveAd(db *gorm.DB, guid string, title string, price32 float32, address string, attributes *[]Attributes, description string) {
	fmt.Printf("Saving house with guid: %s \n", guid)
	db.Create(&House{Guid: guid, Title: title, Price: price32, Address: address, Attributes: *attributes, Description: description})
}

func ParseAd(c *colly.Collector, db *gorm.DB) {
	c.OnHTML("div[id=\"ViewItemPage\"]", func(e *colly.HTMLElement) {
		guid := e.Request.URL.String()

		title := e.ChildText("h1[class^=\"title-\"]")

		price32 := getPriceFromAdHTMLElement(e)

		address := e.ChildText("span[class^=\"address-\"]")

		attributes := getAttributesFromAdHTMLElement(e)

		description := getDescriptionFromAdHTMLElement(e)

		saveAd(db, guid, title, price32, address, &attributes, description)
	})
}
