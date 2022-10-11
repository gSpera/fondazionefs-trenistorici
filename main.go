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

	h, err := LoadHashSetFromFile[Train]("trains.hash")
	if err != nil {
		log.Fatalln("Cannot load hashset:", err)
	}
	log.Infof("HashSet loaded, %d hashes", len(h.hash))

	bot, err := NewTelegramBot(cfg)
	if err != nil {
		log.Fatalln("Cannot create telegram bot:", err)
	}
	log.Infoln("Telegram bot loaded")

	ticker := time.NewTicker(time.Hour)
	for {
		run(bot, h)
		<-ticker.C
	}
}

func run(bot TelegramBot, h *HashSet[Train]) {
	log.Infoln("Running")
	trains, err := LoadTrains()
	if err != nil {
		log.Errorln("Cannot load trains:", err)
	}

	hashDirty := false
	log.Println("Hash", len(h.hash))
	for _, train := range trains {
		if h.IsSaved(train) {
			log.Infoln("Skipping train, already sent:", train)
			continue
		}

		if train.When().After(time.Now().AddDate(bot.TrainsUntilYearsInFuture, bot.TrainsUntilMonthsInFuture, bot.TrainsUntilDaysInFuture)) {
			log.Infof("Skipping train %q, too far in the future: %q", train, train.Date)
			continue
		}

		err := bot.SendTrain(train)
		if err != nil {
			log.Errorln("Cannot send train:", err)
			continue
		}
		h.Add(train)
		hashDirty = true
	}

	if hashDirty {
		log.Infoln("Saving hashes")
		h.SaveAsFile("trains.hash")
	}

	log.Infoln("Done running")
}
