package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/draw"

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

	date = monday.Format(tm, "Monday 2 January 2006", monday.LocaleItIT)
	date = strings.ToUpper(string(date[0])) + date[1:] // Not the best method

	return date
}

type TelegramBot struct {
	bot *tgbotapi.BotAPI
	Config

	lastNotification time.Time
}

func NewTelegramBot(cfg Config) (TelegramBot, error) {
	token := cfg.TelegramBotToken
	bot, err := tgbotapi.NewBotAPI(token)

	return TelegramBot{
		bot:              bot,
		Config:           cfg,
		lastNotification: time.Unix(0, 0),
	}, err
}

func (b *TelegramBot) SendTrain(train Train) error {
	text := &bytes.Buffer{}
	err := msgTemplate.Execute(text, train)
	if err != nil {
		return fmt.Errorf("cannot execute template: %w", err)
	}

	link := BaseURL + strings.TrimPrefix(train.Link, "/")
	image := strings.TrimPrefix(train.ImageURL, "/")

	// Check image size
	res, err := http.Head(image)
	var img tgbotapi.RequestFileData = tgbotapi.FileURL(image)
	length, err2 := strconv.Atoi(res.Header.Get("Content-Length"))
	log.Debugf("Train image size: %s %v bytes\n", image, length)
	if err != nil || err2 != nil || length <= 0 {
		log.Warnln("Cannot head the train image:", err, err2)
	} else if length > 10*1024*1024 {
		// Over 10MB need to resize
		log.Infof("Image is over > 10Mb (%vKB), resing\n", length/1024)
		res, err = http.Get(image)
		if err != nil {
			log.Warnln("Cannot get train image:", err)
		}
		resized, err := resizeImage(res.Body)
		if err != nil {
			log.Warnln("Cannot resize image:", err)
		} else {
			img = tgbotapi.FileReader{
				Name:   image,
				Reader: resized,
			}
		}
	} else if length > 4.5*1024*1024 {
		// Image too big for sending with URL, try sending with a reader
		log.Infof("Image is over > 4.5Mb (%vKB), sending with a reader\n", length/1024)
		res, err = http.Get(image)
		if err != nil {
			log.Warnln("Cannot get train image:", err)
		} else {
			defer res.Body.Close()
			img = tgbotapi.FileReader{
				Name:   image,
				Reader: res.Body,
			}
		}
	}

	msg := tgbotapi.NewPhoto(b.ChannelId, img)
	msg.Caption = html.UnescapeString(text.String())
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Maggiori informazioni", link),
	))
	canAddToCalendar, calendarUrl := httpHtmlAddressForTrain(train, b.Config.HttpPublicAddress)
	if canAddToCalendar {
		inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Aggiungi al calendario", calendarUrl)),
		)
	}

	msg.ReplyMarkup = inlineKeyboard
	msg.DisableNotification = b.Config.Silent

	if time.Now().After(b.lastNotification.Add(10 * time.Minute)) {
		b.lastNotification = time.Now()
		log.Info("Sending notification")
	} else {
		msg.DisableNotification = true
	}

	if b.Config.DryRun {
		log.Infof("Skipping train, dry run %q\n", train)
		return nil
	}

	_, err = b.bot.Send(msg)
	if err != nil {
		log.Errorln("Cannot send train, retring without photo:", train, image, err)

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

// resizeImage resizes the given images, it doesn't check
// if the output size is smaller than the requirement
func resizeImage(r io.Reader) (io.Reader, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("cannot decode image: %w", err)
	}
	width := img.Bounds().Max.X / 3
	height := img.Bounds().Max.Y / 3
	resized := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.ApproxBiLinear.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)

	pr, pw := io.Pipe()
	fmt.Println("Starting encoding")
	go func() {
		jpeg.Encode(pw, resized, nil)
		pw.Close()
	}()
	fmt.Println("Done encoding")
	return pr, nil
}
