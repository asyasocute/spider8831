package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"github.com/gocolly/colly"
)

var mutex sync.Mutex
var i int = 0

func main() {
	c := colly.NewCollector(
		colly.MaxDepth(4),
		colly.Async(true),
	)

	os.RemoveAll("./tmp")
	os.MkdirAll("./tmp", 0755)
	c.Limit(&colly.LimitRule{DomainGlob: "*.neocities.org", Parallelism: 16})
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 32})

	c.WithTransport(&http.Transport{
		TLSHandshakeTimeout: 5 * time.Second,
	})
	c.SetRequestTimeout(5 * time.Second)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)
		var skipSites = []string{"youtube.com", "fedora.org", "github.com", ".gov", "deviantart.net"}
		for _, site := range skipSites {
			if strings.Contains(absoluteURL, site) {
				return
			}
		}
		// if !strings.Contains(absoluteURL, "neocities.org") {
		// 	return
		// }
		e.Request.Visit(absoluteURL)
	})

	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		url := e.Request.AbsoluteURL(e.Attr("src"))
		lower := strings.ToLower(url)
		if strings.HasSuffix(lower, ".svg") || strings.HasSuffix(lower, ".ico") {
			return
		}
		ctx := colly.NewContext()
		ctx.Put("page", e.Request.URL.String())
		// e.Request.Visit(url)
		c.Request("GET", url, nil, ctx, nil)
	})

	c.OnResponse(func(r *colly.Response) {
		if !strings.Contains(r.Headers.Get("Content-Type"), "image") {
			return
		}
		sourcePage := r.Ctx.Get("page")
		data, err := io.ReadAll(bytes.NewReader(r.Body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read body %s: %v\n", r.Request.URL, err)
			return
		}
		m, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get image %s: %v\n", r.Request.URL, err)
			return
		}
		bounds := m.Bounds()
		if bounds.Dx() == 88 && bounds.Dy() == 31 {
			fmt.Println("found badge", r.Request.URL)
			mutex.Lock()
			i += 1
			mutex.Unlock()
			file, err := os.Create("./tmp/" + strconv.Itoa(i) + "-" + r.FileName())
			if err != nil {
				panic("failed to create file " + r.FileName())
			}
			defer file.Close()
			_, err = file.Write(data)
			if err != nil {
				log.Fatalln("failed write to file")
			}
			fmt.Println(sourcePage)
			c.Visit(sourcePage)
		}
	})

	c.OnError(func(r *colly.Response, e error) {
		// log.Println("error:", e, r.Request.URL, string(r.Body))
		log.Println("error:", e, r.Request.URL)
	})

	// c.Visit("http://localhost:4321")
	c.Visit("https://ranfren.neocities.org/")
	// c.Visit("https://replace_this_by_real_url")

	c.Wait()
}
