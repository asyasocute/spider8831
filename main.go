package main

import (
	"fmt"
	"image"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"github.com/gocolly/colly"
)

func saveFile(img *http.Response, path string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = io.Copy(file, img.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("saved to", path)
}

func main() {
	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.Async(true),
	)
	i := 0
	c.Limit(&colly.LimitRule{DomainGlob: "*.neocities.org", Parallelism: 4})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)
		fmt.Println("link ", absoluteURL)
		// c.Visit(absoluteURL)
		e.Request.Visit(absoluteURL)
	})

	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		absoluteURL := e.Request.AbsoluteURL(link)
		response, err := http.Get(absoluteURL)
		if err != nil {
			fmt.Println("error??", err)
			return
		}
		defer response.Body.Close()
		m, _, err := image.Decode(response.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get image %s: %v\n", absoluteURL, err)
			return
		}
		bounds := m.Bounds()
		if bounds.Dx() == 88 && bounds.Dy() == 31 {
			fmt.Println("found badge", "./tmp/"+strconv.Itoa(i))
			saveFile(response, "./tmp/"+strconv.Itoa(i))
			i += 1
		}
		// fmt.Printf("%s %d %d\n", link, m.Width, m.Height)
		fmt.Println("img ", absoluteURL)
		// e.Request.Visit(link)
	})
	c.OnError(func(r *colly.Response, e error) {
		log.Println("error:", e, r.Request.URL, string(r.Body))
	})

	// Start scraping on https://en.wikipedia.org
	// c.Visit("https://ranfren.neocities.org/")
	// c.Visit("https://asyasocute.online/")
	c.Visit("http://localhost:4321/")
	// Wait until threads are finished
	c.Wait()
}
