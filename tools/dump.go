package main

import (
	"html"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	res, err := http.Get("https://www.fondazionefs.it/content/fondazionefs/it/treni-storici.html")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic(err)
	}
	input := doc.Find("#gridList").First()
	rawJson, ok := input.Attr("value")
	if !ok {
		panic(err)
	}

	os.WriteFile("trains.dump", []byte(html.UnescapeString(rawJson)), 0655)
}
