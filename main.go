package main

import (
	"encoding/json"
	"math"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	TelegramBotToken          string
	ChannelId                 int64
	AdminId                   int64
	TrainsUntilYearsInFuture  int
	TrainsUntilMonthsInFuture int
	TrainsUntilDaysInFuture   int
}

func main() {
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

	if cfg.AdminId != 0 {
		log.AddHook(Logger{cfg, bot.bot})
	}

	ticker := time.NewTicker(time.Hour)
	for {
		run(bot, h)
		<-ticker.C
	}
}

func run(bot TelegramBot, h *HashSet[Train]) {
	log.Debugln("Running")
	trains, err := LoadTrains()
	if err != nil {
		log.Errorln("Cannot load trains:", err)
	}

	hashDirty := false
	var (
		trainsSent    int
		trainsSkipped int
	)
	for _, train := range trains {
		if h.IsSaved(train) {
			log.Debugln("Skipping train, already sended", train)
			continue
		}

		if train.When().After(time.Now().AddDate(bot.TrainsUntilYearsInFuture, bot.TrainsUntilMonthsInFuture, bot.TrainsUntilDaysInFuture)) {
			log.Debugln("Skipping train %q, too far in the future: %q", train, train.Date)
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
		log.Debugln("Saving hashes")
		h.SaveAsFile("trains.hash")
	}

	log.Debugln("Done running")

	log.Infof("Executed run, %d new trains sent, %d trains skipped, %d hashes, %t dirty", trainsSent, trainsSkipped, len(h.hash), hashDirty)
	hashDirty = false
}
