package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var BaseURL string = "https://www.fondazionefs.it/"
var DateFormats = []string{"Jan 2, 2006 03:04:05 PM", "Jan 2, 2006, 03:04:05 PM"}
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
	MonthDay            string `json:"date"`
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

func (t Train) When() (time.Time, error) {
	departureTime := t.DepartureTime
	if len(strings.TrimSpace(departureTime)) == 0 {
		log.Debugln("Train without time, moving to 10:00 AM")
		departureTime = "10:00"
	}

	date, err := time.Parse("2/1 3:4", t.MonthDay+" "+departureTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse train date: (%s) %w", t.MonthDay+" "+departureTime, err)
	}

	date = date.AddDate(time.Now().Year(), 0, 0)
	if time.Now().Month() > 10 && date.Month() < 3 {
		// Rollover
		log.Debugln("Rollover date", date)
		date = date.AddDate(1, 0, 0)
	}

	return date, nil
}

func (t Train) Hash() string {
	hasher := md5.New()
	body, err := json.Marshal(t)
	if err != nil {
		panic(fmt.Sprintf("Cannot hash train, this may not happen: %v", err))
	}

	hasher.Write(body)
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

	date, err := t.When()
	if err != nil {
		log.Errorln("Cannot get train date:", err)
		return
	}

	// Departure
	if t.DepartureTime == "" {
		return
	}

	departure, err = time.Parse("15:04", t.DepartureTime)
	if err != nil {
		log.Errorln("Train Departure time but cannot parse:", t.DepartureTime, ":", err)
		return
	}

	// Arrive
	arrive, err = time.Parse("15:04", t.ArriveTime)
	if err != nil {
		log.Errorln("Train Arrive time but cannot parse:", t.DepartureTime, ":", err)
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

	date, err := t.When()
	if err != nil {
		log.Errorln("Cannot get train date:", err)
		return
	}

	// Departure
	departure, err = time.Parse("15:04", t.ReturnDepartureTime)
	if err != nil {
		log.Errorln("Return Train Departure time is not empty but cannot parse:", t.DepartureTime, ":", err)
		return
	}

	// Arrive
	arrive, err = time.Parse("15:04", t.ReturnArriveTime)
	if err != nil {
		log.Errorln("Return Train Arrive time is not empty but cannot parse:", t.DepartureTime, ":", err)
		return
	}

	departure = time.Date(date.Year(), date.Month(), date.Day(), departure.Hour(), departure.Minute(), departure.Second(), 0, date.Location())
	arrive = time.Date(date.Year(), date.Month(), date.Day(), arrive.Hour(), arrive.Minute(), arrive.Second(), 0, date.Location())

	ok = true
	return
}
