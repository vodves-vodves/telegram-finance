package main

import (
	"log"
	"os"

	"github.com/NicoNex/echotron/v3"
	"github.com/joho/godotenv"
)

// var (
//
//	mainKeyboard = tgbotapi.NewReplyKeyboard(
//		tgbotapi.NewKeyboardButtonRow(
//			tgbotapi.NewKeyboardButton("Отчеты"),
//			tgbotapi.NewKeyboardButton("Долги"),
//		))
//
// )
var db *Db

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err)
		return
	}
	token := os.Getenv("TOKEN")
	db, err = NewStorage("data/db.db")
	if err != nil {
		log.Println(err)
		return
	}

	defer db.CloseDB()
	if err := db.Init(); err != nil {
		log.Println(err)
		return
	}
	log.Println("started")
	dsp := echotron.NewDispatcher(token, Newbot)
	log.Println(dsp.Poll())
}
