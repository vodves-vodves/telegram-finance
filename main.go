package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите категорию:")
				keyboard = categoryKeyboard(num)
				msg.ReplyMarkup = keyboard
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}

		} else if update.CallbackQuery != nil {

			//todo при нажатии кнопки добавлять комменатрий
			if update.CallbackQuery.Data == "comment" {

			} else {
				spl := strings.Split(update.CallbackQuery.Data, ":")
				num, _ := strconv.Atoi(spl[1]) //todo поменять на функцию
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

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("✅️ Внесено %v в %s \n🗓 %s", num, spl[0], time.Unix(int64(update.CallbackQuery.Message.Date), 0).Format("02/01/2006 15:04:05")))
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Добавить комментарий", "comment"),
					),
				)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
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

func categoryKeyboard(num int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🍔Еда вне дома", fmt.Sprintf("🍔Еда вне дома:%v", num)),
			tgbotapi.NewInlineKeyboardButtonData("🛒Продукты", fmt.Sprintf("🛒Продукты:%v", num)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏡Дом", fmt.Sprintf("🏡Дом:%v", num)),
			tgbotapi.NewInlineKeyboardButtonData("🚙Машина", fmt.Sprintf("🚙Машина:%v", num)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💊Здоровье", fmt.Sprintf("💊Здоровье:%v", num)),
			tgbotapi.NewInlineKeyboardButtonData("🧙Личное", fmt.Sprintf("🧙Личное:%v", num)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👕Одежда/товары", fmt.Sprintf("👕Одежда/товары:%v", num)),
			tgbotapi.NewInlineKeyboardButtonData("🌐Интернет/связь", fmt.Sprintf("🌐Интернет/связь:%v", num)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎢Развлечения", fmt.Sprintf("🎢Развлечения:%v", num)),
			tgbotapi.NewInlineKeyboardButtonData("🌎Прочие", fmt.Sprintf("🌎Прочие:%v", num)),
		),
	)
}
