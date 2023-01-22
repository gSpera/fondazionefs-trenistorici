package main

import (
	"bytes"
	_ "embed"
	"net/http"
	"strings"
	"text/template"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

//go:embed calendar.tmpl
var calendarTemplateSource string

var calendarTemplate = template.Must(template.New("calendar").Parse(calendarTemplateSource))

func startAndListenHttpServer(addr string) {
	http.HandleFunc("/ics/", httpHandleTrainCreateICal)
	log.Println("Listening on: " + addr)
	http.ListenAndServe(addr, nil)
}

func httpAddressForTrain(t Train, baseUrl string) (ok bool, url string) {
	ok = false
	url = ""
	ok, _, _ = t.DepartureArriveTime()
	if !ok {
		return
	}

	url = baseUrl + "/ics/" + t.UniqueID()
	return
}

func httpHandleTrainCreateICal(w http.ResponseWriter, r *http.Request) {
	hostname := r.Host // Not the best way, but it shouldn't be a problem

	trainID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/ics/"), ".ics")
	trains, err := LoadTrains()
	if err != nil {
		log.Errorln("Cannot load trains:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var train Train
	for _, t := range trains {
		if trainID == t.UniqueID() {
			train = t
			break
		}
	}

	if train.Link == "" {
		// Cannot find the train
		log.Errorln("Cannot find train:", trainID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ok, outboundDeparture, outboundArrive := train.DepartureArriveTime()
	if !ok {
		log.Errorln("Cannot retrieve train outbound time:", train, r.Form.Get("train"))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hasReturn, returnDeparture, returnArrive := train.ReturnDepartureArriveTime()

	var description bytes.Buffer
	err = calendarTemplate.Execute(&description, train)
	if err != nil {
		log.Errorln("Cannot execute template:", r.Form.Get("train"), ":", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)

	cal.SetName(train.Title)
	cal.SetTzid("Europe/Rome")
	ev := cal.AddEvent(train.Hash() + "@trenistorici" + hostname)

	ev.SetSummary(train.String())
	ev.SetURL(BaseURL + strings.TrimPrefix(train.Link, "/"))
	ev.SetLocation("Stazione di " + train.DepartureStation)
	ev.SetDescription(description.String())
	ev.SetStartAt(outboundDeparture)
	ev.SetEndAt(outboundArrive)
	ev.SetClass(ics.ClassificationPublic)

	if hasReturn {
		ret := cal.AddEvent(train.Hash() + "-return" + "@trenistorici" + hostname)
		ret.SetSummary(train.String())
		ret.SetURL(BaseURL + strings.TrimPrefix(train.Link, "/"))
		ret.SetLocation("Stazione di " + train.ArriveStation)
		ret.SetDescription(description.String())
		ret.SetStartAt(returnDeparture)
		ret.SetEndAt(returnArrive)
		ret.SetClass(ics.ClassificationPublic)
	}

	w.Header().Add("Content-Type", "text/calendar")
	err = cal.SerializeTo(w)
	if err != nil {
		log.Errorln("Cannot encode calendar:", err)
	}
}
