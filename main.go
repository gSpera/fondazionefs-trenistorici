package main

import (
	"encoding/json"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	type config struct {
		TelegramBotToken string
		ChannelId        int64
	}
	cfgBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Cannot load config:", err)
	}

	var cfg config
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		log.Fatalln("Cannot unmarshal config:", err)
	}

	h, err := LoadHashSetFromFile[Train]("trains.hash")
	if err != nil {
		log.Fatalln("Cannot load hashset:", err)
	}
	log.Infof("HashSet loaded, %d hashes", len(h.hash))

	bot, err := NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalln("Cannot create telegram bot:", err)
	}
	bot.channelId = cfg.ChannelId
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
			log.Infoln("Skipping train, already sended", train)
			continue
		}

		if train.When().After(time.Now().AddDate(0, 1, 0)) {
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
