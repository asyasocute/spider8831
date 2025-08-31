package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"

	_ "golang.org/x/image/bmp"
	// _ "golang.org/x/image/ico"
	// _ "golang.org/x/image/svg"
	_ "golang.org/x/image/webp"

	"github.com/gocolly/colly"
)

// var mu sync.Mutex
var wg sync.WaitGroup

func main() {
	c := colly.NewCollector(
		colly.MaxDepth(2),
		colly.Async(false),
	)

	// i := 0
	c.Limit(&colly.LimitRule{DomainGlob: "*.neocities.org", Parallelism: 8})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)
		fmt.Println("link ", absoluteURL)
		// c.Visit(absoluteURL)
		e.Request.Visit(absoluteURL)
	})

	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		url := e.Request.AbsoluteURL(link)
		absoluteURL := e.Request.AbsoluteURL(link)
		wg.Add(1)

		// go func(url string) {
		defer wg.Done()
		response, err := http.Get(url)
		if err != nil {
			fmt.Println("error??", err)
			return
		}
		fileName := path.Base(url)
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read body %s: %v\n", url, err)
			return
		}

		m, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get image %s: %v\n", url, err)
			return
		}

		bounds := m.Bounds()
		if bounds.Dx() == 88 && bounds.Dy() == 31 {
			fmt.Println("found badge", fileName)
			// mu.Lock()
			// i += 1
			// mu.Unlock()
			c.Visit(absoluteURL)
			file, err := os.Create("./tmp/" + fileName)
			if err != nil {
				panic("failed to create file " + fileName)
			}
			defer file.Close()

			_, err = file.Write(data)
			if err != nil {

			}

		}

		// fmt.Printf("%s %d %d\n", link, m.Width, m.Height)
		// e.Request.Visit(link)
		// }(absoluteURL)
	})
	c.OnError(func(r *colly.Response, e error) {
		log.Println("error:", e, r.Request.URL, string(r.Body))
	})

	// Start scraping on https://en.wikipedia.org
	// c.Visit("https://ranfren.neocities.org/")
	c.Visit("https://asyasocute.online/")
	// c.Visit("http://localhost:4321/")
	// Wait until threads are finished
	c.Wait()
	wg.Wait()
}
