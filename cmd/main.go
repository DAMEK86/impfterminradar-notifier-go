package main

import (
	"fmt"
	"github.com/damek86/go-impfterminradar-notifier/internal/app"
	"github.com/damek86/go-impfterminradar-notifier/pkg"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Cfg struct {
	zip            string
	radius         int
	delay          time.Duration
	telegramKey    string
	telegramChatId string
}

func main() {
	cfg := getCfg()
	go app.StartHealthEndpoint()
	httpClient := http.Client{
		Timeout: time.Second * 10,
	}
	impfClient := pkg.NewClient(httpClient)
	centers, err := impfClient.GetVacationCenters(cfg.zip, cfg.radius)
	if err != nil {
		fmt.Println(err)
	}

	for {
		err := impfClient.UpdateVaccinesIn(centers)
		if err != nil {
			fmt.Println(err)
		}

		found := "nothing"
		for _, center := range centers {
			for _, vaccine := range center.Vaccines {
				if vaccine.Available {
					found = fmt.Sprintf("%s in %s", vaccine.Id, center.Name)
					addressString := fmt.Sprintf("%s\n%s %s\n\nvisit %s",
						center.Address, center.Zip, center.City, center.BaseUrl)
					SendTelegramMessage(cfg,
						fmt.Sprintf("<b>%s available!</b>\n%s", vaccine.Id, addressString))
				}
			}
		}

		fmt.Printf("found %s, retry after %s...\n", found, cfg.delay)
		time.Sleep(cfg.delay)
	}
}

func getCfg() *Cfg {
	zip, ok := os.LookupEnv("ZIP_CODE")
	if !ok {
		panic("environment variable `ZIP_CODE` not set!")
	}

	radiusStr, ok := os.LookupEnv("RADIUS")
	if !ok {
		panic("environment variable `RADIUS` not set!")
	}
	radius, err := strconv.Atoi(radiusStr)
	if err != nil {
		panic(err)
	}

	delayStr, ok := os.LookupEnv("DELAY")
	if !ok {
		panic("environment variable `DELAY` not set!")
	}

	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		panic(err)
	}

	telegramKey, ok := os.LookupEnv("TELEGRAM_KEY")
	if !ok {
		fmt.Println("Warning: environment variable `TELEGRAM_KEY` not set!")
	}

	telegramChatId, ok := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !ok {
		fmt.Println("Warning: environment variable `TELEGRAM_CHAT_ID` not set!")
	}

	return &Cfg{
		zip:            zip,
		radius:         radius,
		delay:          delay,
		telegramKey:    telegramKey,
		telegramChatId: telegramChatId,
	}
}

func SendTelegramMessage(cfg *Cfg, msg string) {
	if cfg.telegramKey == "" || cfg.telegramChatId == "" {
		fmt.Println("telegram parameters not set - skip telegram send")
		fmt.Println(msg)
		return
	}
	data := url.Values{
		"chat_id":    {cfg.telegramChatId},
		"text":       {msg},
		"parse_mode": {"html"},
	}

	_, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.telegramKey), data)
	if err != nil {
		fmt.Println(err)
	}
}
