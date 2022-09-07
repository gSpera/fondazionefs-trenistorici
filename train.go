package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var BaseURL string = "https://www.fondazionefs.it/"
var DateFormat = "Jan 2, 2006 03:04:05 AM"

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
	Title             string `json:"title"`
	Subtitle          string `json:"subtitle"`
	Link              string `json:"link"`
	Region            string `json:"region"`
	Locomotive        string `json:"locomotive"`
	LocomotiveDetails string `json:"locomotiveOtherDetails"`
	Month             string `json:"month"`
	Date              string `json:"dateProp"`
	IsTimeless        bool   `json:"isTimeless"`
	DepartureStation  string `json:"departureStation"`
	ArriveStation     string `json:"arriveStation"`
	ImageURL          string `json:"image"`
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
	bytes, err := json.Marshal(t)
	if err != nil {
		log.Errorln("Cannot marshal:", err)
	}

	hasher.Write(bytes)
	return hex.EncodeToString(hasher.Sum(nil))
}
