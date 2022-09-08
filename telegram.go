package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goodsign/monday"
	log "github.com/sirupsen/logrus"
)

//go:embed telegram.tmpl
var msgTemplateSource string

var msgTemplate = template.Must(template.New("telegram").Funcs(template.FuncMap{
	"escape":      escapeTelegramText,
	"convertDate": convertDate,
}).Parse(msgTemplateSource))

func escapeTelegramText(text string) string {
	return strings.NewReplacer(
		"_", "\\_",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	).Replace(text)
}

func convertDate(date string) string {
	tm, err := time.Parse(DateFormat, date)
	if err != nil {
		log.Errorln("Cannot parse time:", err)
	}

	return monday.Format(tm, "2 January 2006 alle 15:04", monday.LocaleItIT)
}

type TelegramBot struct {
	bot *tgbotapi.BotAPI
	Config
}

func NewTelegramBot(cfg Config) (TelegramBot, error) {
	token := cfg.TelegramBotToken
	bot, err := tgbotapi.NewBotAPI(token)

	return TelegramBot{
		bot:    bot,
		Config: cfg,
	}, err
}

func (b TelegramBot) SendTrain(train Train) error {
	text := &bytes.Buffer{}
	err := msgTemplate.Execute(text, train)
	if err != nil {
		return fmt.Errorf("cannot execute template: %w", err)
	}

	link := BaseURL + strings.TrimPrefix(train.Link, "/")
	image := BaseURL + strings.TrimPrefix(train.Link, "/")
	msg := tgbotapi.NewPhoto(b.ChannelId, tgbotapi.FileURL(image))
	msg.Caption = html.UnescapeString(text.String())
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Maggiori informazioni", link),
	))

	_, err = b.bot.Send(msg)
	if err != nil {
		log.Errorln("cannot send train, retring without photo:", train, image, err)

		safeMsg := tgbotapi.NewMessage(b.ChannelId, msg.Caption)
		safeMsg.ParseMode = tgbotapi.ModeMarkdownV2
		safeMsg.ReplyMarkup = msg.ReplyMarkup
		_, err = b.bot.Send(safeMsg)
		if err != nil {
			return fmt.Errorf("cannot send safe message: %q %w", train, err)
		}
	}

	return nil
}

type Logger struct {
	cfg Config
	bot *tgbotapi.BotAPI
}

var _ log.Hook = Logger{}

func (l Logger) Levels() []log.Level {
	return []log.Level{
		log.InfoLevel,
		log.ErrorLevel,
		log.FatalLevel,
		log.WarnLevel,
	}
}

func (l Logger) Fire(entry *log.Entry) error {
	silent := false
	var level string
	switch entry.Level {
	case log.ErrorLevel:
		level = "*ERRORE*"
	case log.FatalLevel:
		level = "*ERRORE FATALE*"
	case log.WarnLevel:
		level = "*Warning*"
	case log.InfoLevel:
		level = "_Info_"
		silent = true
	default:
		panic("Entry level not recognized")
	}

	msg := tgbotapi.NewMessage(l.cfg.AdminId, fmt.Sprintf("Log: %s\n%s\n`%s`", level, entry.Message, time.Now().Format(time.RubyDate)))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.DisableNotification = silent
	_, err := l.bot.Send(msg)
	return err
}
