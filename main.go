package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-finance/storage"
)

var keyboard = tgbotapi.InlineKeyboardMarkup{}

func main() {
	db, err := storage.NewStorage("data/db.db")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.CloseDB()
	if err := db.Init(); err != nil {
		log.Println(err)
		return
	}

	bot, err := tgbotapi.NewBotAPI("6345585421:AAEA34EZ4_XD5hLcVKPeEPgLRs6oGApalEo")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.IsCommand() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

				//команды
				//todo добавить команды для получения результата последних 7,14,30 дней
				switch update.Message.Command() {
				case "start":
					msg.Text = "Привет! Отправь число для начала учета твоих финансов"
					if err := db.SaveUser(update.Message.Chat.ID, update.Message.Date, update.Message.From.UserName); err != nil {
						log.Println(err)
						return
					}
				case "sum":
					//todo сделать вывод общих трат
					all, err := db.GetSum(update.Message.Chat.ID)
					if err != nil {
						log.Println(err)
						return
					}

					if len(all) > 0 {
						for _, data := range all {
							msg.Text += fmt.Sprintf("data: %+v\n", data)
						}
					} else {
						msg.Text += fmt.Sprintf("No data\n")
					}
				default:
					msg.Text = "Не знаю такую команду"
				}

				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}

			//добавление расходов
			//todo добавить ответ с кнопками после ввода числа, где идет выбор категорий
			if num, ok := IsNumeric(update.Message.Text); ok {
				//if err := db.SaveData(num, "", "", update.Message.Date, update.Message.Chat.ID); err != nil {
				//	log.Println(err)
				//	return
				//}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите категорию:")
				keyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Еда вне дома", fmt.Sprintf("Еда вне дома:%v", num)),
						tgbotapi.NewInlineKeyboardButtonData("Продукты", fmt.Sprintf("Продукты:%v", num)),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Дом", fmt.Sprintf("Дом:%v", num)),
						tgbotapi.NewInlineKeyboardButtonData("Машина", fmt.Sprintf("Машина:%v", num)),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Здоровье", fmt.Sprintf("Здоровье:%v", num)),
						tgbotapi.NewInlineKeyboardButtonData("Личное", fmt.Sprintf("Личное:%v", num)),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Одежда/товары", fmt.Sprintf("Одежда/товары:%v", num)),
						tgbotapi.NewInlineKeyboardButtonData("Интернет/связь", fmt.Sprintf("Интернет/связь:%v", num)),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Развлечения", fmt.Sprintf("Развлечения:%v", num)),
						tgbotapi.NewInlineKeyboardButtonData("Прочие", fmt.Sprintf("Прочие:%v", num)),
					),
				)
				msg.ReplyMarkup = keyboard
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}

		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}
			spl := strings.Split(update.CallbackQuery.Data, ":")
			num, _ := strconv.Atoi(spl[1])
			if err := db.SaveData(num, spl[0], "", update.CallbackQuery.Message.Date, update.CallbackQuery.Message.Chat.ID); err != nil {
				log.Println(err)
				return
			}
			deleteMessageConfig := tgbotapi.DeleteMessageConfig{
				ChatID:    update.CallbackQuery.Message.Chat.ID,
				MessageID: update.CallbackQuery.Message.MessageID,
			}
			_, err := bot.Request(deleteMessageConfig)
			if err != nil {
				log.Println(err)
				return
			}
			// And finally, send a message containing the data received.
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("Внесено %v в %s", num, spl[0]))

			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}

}

func IsNumeric(s string) (int, bool) {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i, true
	}
	return 0, false
}
