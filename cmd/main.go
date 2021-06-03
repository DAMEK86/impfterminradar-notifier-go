package main

import (
	"fmt"
	"github.com/damek86/go-impfterminradar-notifier"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Cfg struct {
	center         string
	radius         int
	delay          time.Duration
	telegramKey    string
	telegramChatId string
}

func main() {
	cfg := getCfg()

	httpClient := http.Client{
		Timeout: time.Second * 10,
	}
	impfClient := impfterminradar.NewClient(httpClient)
	centers, err := impfClient.GetVacationCenters(cfg.center, cfg.radius)
	if err != nil {
		fmt.Println(err)
	}

	for {
		vaccines, err := impfClient.GetVaccinesIn(centers)
		if err != nil {
			fmt.Println(err)
		}

		found := "nothing"
		for _, vaccine := range vaccines {
			if !vaccine.Available {
				found = fmt.Sprintf("%s in %s", vaccine.FriendlyName, vaccine.Center.Name)
				addressString := fmt.Sprintf("%s\n%s %s\n\nvisit %s",
					vaccine.Center.Address, vaccine.Center.Zip, vaccine.Center.City, vaccine.Center.BaseUrl)
				SendTelegramMessage(cfg,
					fmt.Sprintf("<b>%s available!</b>\n%s", vaccine.FriendlyName, addressString))
			}
		}

		fmt.Printf("found %s, retry after %s...\n", found, cfg.delay)
		time.Sleep(cfg.delay)
	}
}

func getCfg() *Cfg {
	center, ok := os.LookupEnv("center")
	if !ok {
		panic("environment variable `center` not set!")
	}

	radiusStr, ok := os.LookupEnv("radius")
	if !ok {
		panic("environment variable `radius` not set!")
	}
	radius, err := strconv.Atoi(radiusStr)
	if err != nil {
		panic(err)
	}

	delayStr, ok := os.LookupEnv("delay")
	if !ok {
		panic("environment variable `radius` not set!")
	}

	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		panic(err)
	}

	telegramKey, ok := os.LookupEnv("telegram_key")
	if !ok {
		fmt.Println("Warning: environment variable `telegram_key` not set!")
	}

	telegramChatId, ok := os.LookupEnv("telegram_chat_id")
	if !ok {
		fmt.Println("Warning: environment variable `telegram_chat_id` not set!")
	}

	return &Cfg{
		center:         center,
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
