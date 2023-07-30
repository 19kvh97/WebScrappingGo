package main

import (
	"fmt"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	// c := colly.NewCollector()

	// c.OnRequest(func(r *colly.Request) {
	// 	log.Println("Visiting: ", r.URL)
	// 	r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36 Edg/101.0.1210.47")
	// 	// r.Headers.Set("cookie", "__cf_bm=_wxQLkRiwubRFmgZWpE0V_5w0egNZaEQ_wjeqCJsw.8-1690690604-0-Ad1fCm49xuwaRCHw3dJvohy9SxQP6ozhM0Rn+5vRrbPdRHuMOzAyoiliVUCtEhA/LHAZ5CCkOyFt51IOHFujAw4=; __cflb=02DiuEXPXZVk436fJfSVuuwDqLqkhavJb4fAJVoRrw5MD; _cfuvid=y7F2Y4vqBZhggPoHdCzsW4Jh1GOr4k0WU7tsVUOtNCE-1690690604750-0-604800000; country_code=VN; visitor_gql_token=oauth2v2_e419cb21db1a821452d2511385e7f225; visitor_id=42.118.50.46.1690690605536000")
	// 	r.Headers.Set("Accept", "*/*")
	// 	r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
	// 	r.Headers.Set("Connection", "keep-alive")
	// })

	// c.OnError(func(_ *colly.Response, err error) {
	// 	log.Println("Something went wrong: ", err)
	// })

	// c.OnResponse(func(r *colly.Response) {
	// 	log.Println("Page visited: ", r.Request.URL)
	// })

	// c.OnHTML("li.product", func(e *colly.HTMLElement) {
	// 	// printing all URLs associated with the a links in the page
	// 	log.Printf("%v", e.Attr("href"))
	// })

	// c.OnScraped(func(r *colly.Response) {
	// 	log.Println(r.Request.URL, " scraped!")
	// })

	// c.Visit("https://www.upwork.com")
	// log.Println("Hello World!")

	// Create a new Colly collector
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,               // Set the number of concurrent requests
		RandomDelay: 5 * time.Second, // Add a random delay of up to 5 seconds between requests
	})

	// Set up the callback for when a visited HTML element is found
	c.OnHTML(".job-title-link", func(e *colly.HTMLElement) {
		fmt.Println(e.Text)
	})

	// Set up the callback for when a visited HTML element's attributes are found
	c.OnHTML(".job-tile-title-link", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println("Link:", link)
	})

	// Start the scraping process by visiting the Upwork website
	err := c.Visit("https://www.upwork.com/")
	if err != nil {
		fmt.Println("Error:", err)
	}
}
