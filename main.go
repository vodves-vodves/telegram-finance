package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"telegram-finance/storage"
)

var keyboard = tgbotapi.InlineKeyboardMarkup{}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err)
		return
	}
	adminId, _ := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	token := os.Getenv("TOKEN")

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

	bot, err := tgbotapi.NewBotAPI(token)
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
					//todo добавить выбор месяца и года
					year := time.Now().Year()
					month := time.Now().Month()

					all, err := db.GetSum(update.Message.Chat.ID, year, month)
					if err != nil {
						log.Println(err)
						return
					}
					if len(all) > 0 {
						msg.Text += fmt.Sprintf("Траты за %s\n", time.Now().Month().String())
						for _, data := range all {
							msg.Text += fmt.Sprintf(" %s			%v 	%s\n", data.Category, data.Data, data.Date.Local().Format("02/01/2006 15:04:05"))
						}
					} else {
						msg.Text += fmt.Sprintf("У вас еще нет записей за %s\n", month)
					}
				case "allsum":
					all, err := db.AllSum(update.Message.Chat.ID)
					if err != nil {
						log.Println(err)
						return
					}
					msg.Text += fmt.Sprintf("Общая сумма трат за все время: %v\n", all)
				default:
					msg.Text += "Такой команды нет"
				}

				//admin
				if update.Message.Chat.ID == adminId {
					msg.Text = ""
					switch update.Message.Command() {
					case "user":
						spl := strings.Split(update.Message.Text, " ")
						userId, _ := strconv.ParseInt(spl[1], 10, 64)
						all, c, err := db.GetUserInfo(userId)
						if err != nil {
							log.Println(err)
							return
						}
						regDate := time.Unix(int64(all.RegDate), 0).Format("02/01/2006 15:04:05")
						msg.Text += fmt.Sprintf("Информация о пользователе:\n\nИмя: @%v\nДата регистрации: %v\nКоличество записей: %v\n", all.UserName, regDate, c)
					case "users":
						all, err := db.GetUsers()
						if err != nil {
							log.Println(err)
							return
						}
						msg.Text += fmt.Sprintf("Пользователи:\n\n")
						for _, data := range all {
							msg.Text += fmt.Sprintf("Имя: @%v\nUserID: %v\nДата регистрации: %v\n\n", data.UserName, data.UserId, time.Unix(int64(data.RegDate), 0).Format("02/01/2006 15:04:05"))
						}
					case "deluser":
						spl := strings.Split(update.Message.Text, " ")
						userId, _ := strconv.ParseInt(spl[1], 10, 64)
						err := db.DeleteUser(userId)
						if err != nil {
							log.Println(err)
							return
						}
						msg.Text += fmt.Sprintf("Пользователь %v удален\n", userId)
					case "deluserdata":
						spl := strings.Split(update.Message.Text, " ")
						userId, _ := strconv.ParseInt(spl[1], 10, 64)
						err := db.DeleteUserData(userId)
						if err != nil {
							log.Println(err)
							return
						}
						msg.Text += fmt.Sprintf("Данные пользователя %v удалены\n", userId)
					}
				}

				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}

			//добавление расходов
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
				if err := db.SaveData(num, spl[0], "", update.CallbackQuery.Message.Chat.ID); err != nil {
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

				mes := fmt.Sprintf("✅️ Внесено %v в %s \n🗓 %s", num, spl[0], time.Unix(int64(update.CallbackQuery.Message.Date), 0).Format("02/01/2006 15:04:05"))
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, mes)
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
