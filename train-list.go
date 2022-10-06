package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type TrainImporter interface {
	Trains() []Train
}

func LoadTrains() ([]Train, error) {
	res, err := http.Get("https://www.fondazionefs.it/content/fondazionefs/it/treni-storici.html")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	input := doc.Find("#gridList").First()
	rawJson, ok := input.Attr("value")
	if !ok {
		return nil, errors.New("cannot find trains")
	}

	unmarshal := struct {
		AlreadyLoaded int
		TrainsList    []Train
	}{}
	err = json.Unmarshal([]byte(rawJson), &unmarshal)
	if err != nil {
		return nil, err
	}

	return unmarshal.TrainsList, nil
}
