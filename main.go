package main

import (
	"log"
	"os"
	"time"

	"github.com/NicoNex/echotron/v3"
	"github.com/joho/godotenv"
	"telegram-finance/bot"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err)
		return
	}
	token := os.Getenv("TOKEN")

	log.Println("started")
	dsp := echotron.NewDispatcher(token, bot.NewBot)
	for {
		log.Println(dsp.Poll())
		time.Sleep(5 * time.Second)
	}
}
