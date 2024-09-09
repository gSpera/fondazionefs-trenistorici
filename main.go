package main

import (
	"encoding/json"
	"flag"
	"math"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	TelegramBotToken          string
	ChannelId                 int64
	HttpPublicAddress         string
	HttpListenAddress         string
	TrainsUntilYearsInFuture  int
	TrainsUntilMonthsInFuture int
	TrainsUntilDaysInFuture   int
	DryRun                    bool `json:"-"`
	Silent                    bool `json:"-"`
}

func main() {
	dryRun := flag.Bool("dry", false, "dry run, doesn't send messages on telegram, updates hashes")
	silent := flag.Bool("silent", false, "send silent messages")
	flag.Parse()

	cfgBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Cannot load config:", err)
	}

	cfg := Config{
		TrainsUntilYearsInFuture:  0,
		TrainsUntilMonthsInFuture: 1,
		TrainsUntilDaysInFuture:   0,
	}

	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		log.Fatalln("Cannot unmarshal config:", err)
	}

	cfg.DryRun = *dryRun
	cfg.Silent = *silent

	if cfg.TrainsUntilYearsInFuture < 0 || cfg.TrainsUntilMonthsInFuture < 0 || cfg.TrainsUntilDaysInFuture < 0 {
		cfg.TrainsUntilYearsInFuture = math.MaxInt
	}

	h, err := LoadTrainArchiveFromFile("trains.hash")
	if err != nil {
		log.Fatalln("Cannot load train archive:", err)
	}
	log.Infof("HashSet loaded, %d hashes", len(h.hash))

	bot, err := NewTelegramBot(cfg)
	if err != nil {
		log.Fatalln("Cannot create telegram bot:", err)
	}
	log.Infoln("Telegram bot loaded")

	go startAndListenHttpServer(cfg.HttpListenAddress, cfg.HttpPublicAddress)

	ticker := time.NewTicker(time.Hour)
	for {
		run(&bot, h)
		<-ticker.C
	}
}

func run(bot *TelegramBot, h *TrainArchive) {
	log.Infoln("Running")
	trains, err := LoadTrains()
	if err != nil {
		log.Errorln("Cannot load trains:", err)
	}

	hashDirty := false
	log.Println("Hash", len(h.hash))
	for _, train := range trains {
		when, err := train.When()
		if err != nil {
			log.Errorln("Cannot get train date:", train, err)
			continue
		}
		if when.After(time.Now().AddDate(bot.TrainsUntilYearsInFuture, bot.TrainsUntilMonthsInFuture, bot.TrainsUntilDaysInFuture)) {
			log.Infof("Skipping train %q, too far in the future: %q", train, train.Date)
			continue
		}

		switch h.Compare(train) {
		case TrainSaved:
			log.Infoln("Skipping train, already sent:", train)
			continue
		case TrainChanged:
			if h.GetID(train) == 0 {
				// Train was sent dry, do nothing
				log.Infoln("Skipping updating train sent dry:", train)
				continue
			}
			log.Infoln("Changing train:", train)
			err := bot.EditMessage(train, h.GetID(train))
			if err != nil {
				log.Errorln("Cannot change train:", train, ":", err)
			}
			h.Add(train, h.GetID(train))
			hashDirty = true
		case TrainNotSaved:
			log.Infoln("Sending train:", train)
			msgID, err := bot.SendTrain(train)
			if err != nil {
				log.Errorln("Cannot send train:", err)
				continue
			}
			h.Add(train, msgID)
			hashDirty = true
		}
	}

	if hashDirty {
		log.Infoln("Saving hashes")
		h.SaveAsFile("trains.hash")
	}

	log.Infoln("Done running")
}
