package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NicoNex/echotron/v3"
)

type stateFn func(*echotron.Update) stateFn

type bot struct {
	chatID   int64
	amount   int
	recordId int64
	state    stateFn
	echotron.API
}

func Newbot(chatID int64) echotron.Bot {
	token := os.Getenv("TOKEN")
	bot := &bot{
		chatID: chatID,
		API:    echotron.NewAPI(token),
	}
	bot.state = bot.startBot
	return bot
}

func (b *bot) Update(update *echotron.Update) {
	b.state = b.state(update)
}

func (b *bot) startBot(update *echotron.Update) stateFn {
	switch {
	case update.Message != nil:
		b.handleMessage(update)
	case update.CallbackQuery != nil:
		st := b.handleCallbackQuery(update.CallbackQuery)
		return st
	}
	return b.startBot
}

func (b *bot) handleMessage(update *echotron.Update) stateFn {
	msgText := update.Message.Text
	userName := update.Message.From.Username
	msgDate := update.Message.Date
	log.Printf("[%s] %s", userName, msgText)

	switch {
	case msgText == "/start":
		b.startUser(userName, msgDate)
		//todo добавить
	//case messageText == "/get":
	//	b.sendExpences(userID)
	//case messageText == "/chart":
	//	b.sendChart(userID)
	//case addCommandPattern.MatchString(messageText):
	//	b.addExpence(messageText, userID)
	case strings.HasSuffix(msgText, "Отчеты"):
		b.reports()
	case strings.HasSuffix(msgText, "Расходы за текущий месяц"):
		b.userMonthStats()
	case strings.HasSuffix(msgText, "Потрачено за все время"):
		b.userAllStats()
	case strings.HasSuffix(msgText, "Долги"):
	case strings.HasSuffix(msgText, "Главное меню"):
		b.mainMenu()
	case strings.HasSuffix(msgText, "Настройки"):

	}
	if num, ok := isNumeric(update.Message.Text); ok {
		b.addAmount(num)
	}
	return b.startBot
}

func (b *bot) startUser(userName string, msgDate int) stateFn {
	if err := db.SaveUser(b.chatID, msgDate, userName); err != nil {
		log.Println(userName, err)
	}

	msg := "Отправь число для начала учета твоих финансов"

	opt := echotron.MessageOptions{
		ReplyMarkup: mainButtons(),
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		fmt.Println(message, err)
	}
	return b.startBot
}

func (b *bot) mainMenu() {
	msg := "Отправь число для начала учета твоих финансов"

	opt := echotron.MessageOptions{
		ReplyMarkup: mainButtons(),
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		fmt.Println(message, err)
	}
	//return b.startBot
}

func (b *bot) addAmount(amount int) {
	b.amount = amount
	msg := "Выберите категорию:"

	opt := echotron.MessageOptions{
		ReplyMarkup: categoriesButtons(),
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) handleCallbackQuery(c *echotron.CallbackQuery) stateFn {
	switch {
	case strings.HasPrefix(c.Data, "comment"):
		st := b.setComment(c)
		return st
	case strings.HasPrefix(c.Data, "delRecord"):
		recordId, _ := strconv.ParseInt(strings.Split(c.Data, ":")[1], 10, 64)
		b.deleteRecord(recordId, c.Message.ID)
	case c.Data == "cancelCategories":
		b.cancelCategories(c)
	default:
		b.setCategory(c)
	}
	return b.startBot
}

func (b *bot) setCategory(c *echotron.CallbackQuery) {
	id, err := db.SaveData(b.amount, c.Data, b.chatID)
	if err != nil {
		log.Println(c.From.Username, err)
		return
	}

	msg := fmt.Sprintf("✅️ Внесено %v руб в %s \n🗓 %s", b.amount, c.Data, time.Unix(int64(c.Message.Date), 0).Format("02/01/2006 15:04:05"))
	message := echotron.NewMessageID(b.chatID, c.Message.ID)
	_, err = b.EditMessageText(msg, message, &echotron.MessageTextOptions{
		ReplyMarkup: recordButtons(id),
	})
	if err != nil {
		log.Println(err)
	}
}

func (b *bot) setComment(c *echotron.CallbackQuery) stateFn {
	recordId, _ := strconv.ParseInt(strings.Split(c.Data, ":")[1], 10, 64)
	b.recordId = recordId
	msg := "Напишите комментарий:"
	message, err := b.SendMessage(msg, b.chatID, nil)
	if err != nil {
		log.Println(message, err)
	}
	defer b.AnswerCallbackQuery(c.ID, nil)
	return b.getComment
}

func (b *bot) getComment(update *echotron.Update) stateFn {
	comment := update.Message.Text
	err := db.setComment(comment, b.recordId)
	if err != nil {
		log.Println(update.Message.From.Username, err)
		return b.startBot
	}

	msg := "Комментарий добавлен!"
	message, err := b.SendMessage(msg, b.chatID, nil)
	if err != nil {
		log.Println(message, err)
	}
	return b.startBot
}

func (b *bot) deleteRecord(recordId int64, messageId int) {
	err := db.DeleteUserData(recordId)
	if err != nil {
		log.Println(err)
		return
	}
	msg := "Запись удалена!"
	message := echotron.NewMessageID(b.chatID, messageId)
	_, err = b.EditMessageText(msg, message, nil)
	if err != nil {
		log.Println(err)
	}
}

func (b *bot) cancelCategories(c *echotron.CallbackQuery) {
	msg := "Операция отменена"
	message := echotron.NewMessageID(b.chatID, c.Message.ID)
	_, err := b.EditMessageText(msg, message, nil)
	if err != nil {
		log.Println(err)
	}
}

func (b *bot) userMonthStats() {
	//todo сделать выбор года и месяца
	var msg string
	year := time.Now().Year()
	month := time.Now().Month()

	all, err := db.GetSum(b.chatID, year, month)
	if err != nil {
		log.Println(err)
		return
	}
	if len(all) > 0 {
		msg += fmt.Sprintf("Траты за %s\n", time.Now().Month().String())
		for i, data := range all {
			msg += fmt.Sprintf("%v.  %s %v руб - %s - %s\n", i+1, data.Category, data.Data, data.Comment, data.Date.Local().Format("02/01/2006 15:04:05"))
		}
	} else {
		msg = fmt.Sprintf("У вас еще нет записей за %s\n", month)
	}

	message, err := b.SendMessage(msg, b.chatID, nil)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) userAllStats() {
	all, err := db.AllSum(b.chatID)
	if err != nil {
		log.Println(err)
		return
	}
	msg := fmt.Sprintf("Общая сумма трат за все время: %v рублей\n", all)

	message, err := b.SendMessage(msg, b.chatID, nil)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) reports() {
	msg := "Выберите отчеты:"

	opt := echotron.MessageOptions{
		ReplyMarkup: reportButtons(),
	}
	//b.SendMessage("Вам выставлен счет: 5 000 000 руб", 459455590, nil)
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

func mainButtons() echotron.ReplyKeyboardMarkup {
	return echotron.ReplyKeyboardMarkup{
		Keyboard: [][]echotron.KeyboardButton{
			{
				{Text: "📄 Отчеты"},
				{Text: "💰 Долги"}, //credit
			},
			{
				{Text: "⚙️ Настройки"},
			},
		},
		ResizeKeyboard: true,
	}
}

func reportButtons() echotron.ReplyKeyboardMarkup {
	return echotron.ReplyKeyboardMarkup{
		Keyboard: [][]echotron.KeyboardButton{
			{
				{Text: "💰 Расходы за текущий месяц"},
				{Text: "💰 Потрачено за все время"},
			},
			{
				{Text: "💰 Удалить запись"},
				{Text: "💰 Изменить комментарий"},
			},
			{
				{Text: "⬅️ Главное меню"},
			},
		},
		ResizeKeyboard: true,
	}
}

func creditButtons() echotron.ReplyKeyboardMarkup {
	return echotron.ReplyKeyboardMarkup{
		Keyboard: [][]echotron.KeyboardButton{
			{
				{Text: "💰 Выставить счет"},
				{Text: "💰 Мои долги"},
			},
			{
				{Text: "⬅️ Главное меню"},
			},
		},
		ResizeKeyboard: true,
	}
}

func settingsButtons() echotron.InlineKeyboardMarkup {
	return echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{
			{
				{Text: "Напоминание учета расходов (вкл/выкл)", CallbackData: "reminder"},
			},
			{
				{Text: "Напоминание о долгах (вкл/выкл)", CallbackData: "remidnerDolg"},
			},
		},
	}
}

func categoriesButtons() echotron.InlineKeyboardMarkup {
	return echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{
			{
				{Text: "🍔Еда вне дома", CallbackData: "🍔Еда вне дома"},
				{Text: "🛒Продукты", CallbackData: "🛒Продукты"},
			},
			{
				{Text: "🏡Дом", CallbackData: "🏡Дом"},
				{Text: "🚙Машина", CallbackData: "🚙Машина"},
			},
			{
				{Text: "💊Здоровье", CallbackData: "💊Дом"},
				{Text: "🧙Личное", CallbackData: "🧙Личное"},
			},
			{
				{Text: "👕Одежда/товары", CallbackData: "👕Одежда/товары"},
				{Text: "🌐Интернет/связь", CallbackData: "🌐Интернет/связь"},
			},
			{
				{Text: "🎢Развлечения", CallbackData: "🎢Развлечения"},
				{Text: "🌎Прочие", CallbackData: "🌎Прочие"},
			},
			{
				{Text: "❌Отмена", CallbackData: "cancelCategories"},
			},
		},
	}
}

func recordButtons(id int64) echotron.InlineKeyboardMarkup {
	return echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{
			{
				{Text: "⏬ Добавить комментарий", CallbackData: fmt.Sprintf("comment:%v", id)},
			},
			{
				{Text: "❌ Удалить", CallbackData: fmt.Sprintf("delRecord:%v", id)},
			},
		},
	}
}

func isNumeric(s string) (int, bool) {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i, true
	}
	return 0, false
}
