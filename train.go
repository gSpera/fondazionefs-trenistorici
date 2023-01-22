package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var BaseURL string = "https://www.fondazionefs.it/"
var DateFormat = "Jan 2, 2006 03:04:05 AM"
var timezone *time.Location

func init() {
	var err error
	timezone, err = time.LoadLocation("Europe/Rome")
	if err != nil {
		panic("Cannot load timezone location Europe/Rome")
	}
}

//	{
//		  "dateProp": "Dec 30, 2022 12:00:00 AM",
//		  "isTimeless": true,
//		  "date": "30/12",
//		  "link": "/content/fondazionefs/it/treni-storici/2022/12/30/ferrovia-dei-parchi--l-alto-sangro.html",
//		  "title": "Ferrovia dei Parchi: l'alto Sangro",
//		  "subtitle": "Treno storico da Sulmona a Castel di Sangro",
//		  "departureStation": "Sulmona",
//		  "departureHour": "",
//		  "arriveStation": "Castel di Sangro",
//		  "arriveHour": "",
//		  "moreInfo": "<p><b>Vai su prenota per scoprire tutti i dettagli dell'evento.</b></p>",
//		  "image": "/content/dam/fondazionefs/fondazione-fs-new/hp-prenota-un-viaggio/card-calendario-treni/abruzzo/Abruzzo_Transiberiana8.jpg",
//		  "priceAdult": "",
//		  "priceChild": "",
//		  "region": "Abruzzo",
//		  "locomotive": "Treno con locomotiva diesel",
//		  "locomotiveOtherDetails": "",
//		  "month": "December",
//		  "timelessBSTConfigPath": "/content/fondazionefs/it/config/binari-senza-tempo/jcr:content/timeless_parsys/tratta_bst_1041103064",
//		  "labelTimelessSearchBST": "",
//		  "returnDepartureHour": "",
//		  "returnArriveHour": "",
//		  "priceAdultReturn": "",
//		  "priceChildReturn": "",
//		  "enableReturn": true,
//		  "singlePrice": false,
//		  "enableReturnPrice": false
//		}

type Train struct {
	Title               string `json:"title"`
	Subtitle            string `json:"subtitle"`
	Link                string `json:"link"`
	Region              string `json:"region"`
	Locomotive          string `json:"locomotive"`
	LocomotiveDetails   string `json:"locomotiveOtherDetails"`
	Month               string `json:"month"`
	Date                string `json:"dateProp"`
	IsTimeless          bool   `json:"isTimeless"`
	DepartureStation    string `json:"departureStation"`
	DepartureTime       string `json:"departureHour"`
	ArriveStation       string `json:"arriveStation"`
	ArriveTime          string `json:"arriveHour"`
	ImageURL            string `json:"image"`
	ReturnDepartureTime string `json:"returnDepartureHour"`
	ReturnArriveTime    string `json:"returnArriveHour"`
	PriceAdult          string `json:"priceAdult,omitempty"`
	PriceChildren       string `json:"priceChild,omitempty"`
	PriceAdultReturn    string `json:"priceAdultReturn,omitempty"`
	PriceChildrenReturn string `json:"priceChildReturn,omitempty"`
}

func (t Train) String() string {
	return t.Title
}

func (t Train) When() time.Time {
	date, err := time.Parse(DateFormat, t.Date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot convert date: %v", err)
		log.Errorln("Cannot convert time:", err)
	}

	return date
}

func (t Train) Hash() string {
	hasher := md5.New()
	hasher.Write([]byte(t.Link))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (t Train) UniqueID() string {
	return strings.TrimSuffix(strings.TrimPrefix(t.Link, "/content/fondazionefs/it/treni-storici/"), ".html")
}

// DepartureArriveTime tries to extract the departure and arrive time of the train.
// Extracting these information is not always possible, if only one of the time can be
// obtained ok will be false
func (t Train) DepartureArriveTime() (ok bool, departure, arrive time.Time) {
	ok = false
	var err error

	date := t.When()

	// Departure
	departure, err = time.Parse("15:04", t.DepartureTime)
	if err != nil {
		log.Errorln("Train Departure time is not empty but cannot parse:", t.DepartureTime, ":", err)
		return
	}

	// Arrive
	arrive, err = time.Parse("15:04", t.ArriveTime)
	if err != nil {
		log.Errorln("Train Departure time is not empty but cannot parse:", t.DepartureTime, ":", err)
		return
	}

	departure = time.Date(date.Year(), date.Month(), date.Day(), departure.Hour(), departure.Minute(), departure.Second(), 0, timezone)
	arrive = time.Date(date.Year(), date.Month(), date.Day(), arrive.Hour(), arrive.Minute(), arrive.Second(), 0, timezone)

	ok = true
	return
}

// ReturnDepartureArriveTime tries to extract the departure and arrive time of the return train.
// Extracting these information is not always possible, if only one of the time can be
// obtained ok will be false
func (t Train) ReturnDepartureArriveTime() (ok bool, departure, arrive time.Time) {
	ok = false
	var err error

	date := t.When()

	// Departure
	departure, err = time.Parse("15:04", t.ReturnDepartureTime)
	if err != nil {
		log.Errorln("Return Train Departure time is not empty but cannot parse:", t.DepartureTime, ":", err)
		return
	}

	// Arrive
	arrive, err = time.Parse("15:04", t.ReturnArriveTime)
	if err != nil {
		log.Errorln("return Train Departure time is not empty but cannot parse:", t.DepartureTime, ":", err)
		return
	}

	departure = time.Date(date.Year(), date.Month(), date.Day(), departure.Hour(), departure.Minute(), departure.Second(), 0, date.Location())
	arrive = time.Date(date.Year(), date.Month(), date.Day(), arrive.Hour(), arrive.Minute(), arrive.Second(), 0, date.Location())

	ok = true
	return
}
