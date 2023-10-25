package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NicoNex/echotron/v3"
	"telegram-finance/sql"
)

type stateFn func(*echotron.Update) stateFn

type bot struct {
	chatID   int64
	amount   int
	recordId int64
	year     int
	db       sql.Db
	month    time.Month
	state    stateFn
	echotron.API
}

func NewBot(chatID int64) echotron.Bot {
	token := os.Getenv("TOKEN")
	db, err := sql.NewStorage("data/db.db")
	if err != nil {
		log.Println(err)
	}

	if err := db.Init(); err != nil {
		log.Println(err)
	}
	bot := &bot{
		chatID: chatID,
		db:     *db,
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
	case strings.HasSuffix(msgText, "Отчеты"):
		b.reports()
	case strings.HasSuffix(msgText, "Расходы за определенный месяц"):
		b.getYear()
	case strings.HasSuffix(msgText, "Расходы за текущий месяц"):
		b.userMonthStats(time.Now().Year(), time.Now().Month())
	case strings.HasSuffix(msgText, "Потрачено за все время"):
		b.userAllStats()
	case strings.HasSuffix(msgText, "Долги"):
		b.credits()
	case strings.HasSuffix(msgText, "Выставить долг"):
	case strings.HasSuffix(msgText, "Мне должны"):
	case strings.HasSuffix(msgText, "Я должен"):
	case strings.HasSuffix(msgText, "Главное меню"):
		b.mainMenu(userName)
	case strings.HasSuffix(msgText, "Настройки"):
		b.settings()

	}
	if num, ok := isNumeric(update.Message.Text); ok {
		b.addAmount(num)
	}
	return b.startBot
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
	case strings.HasPrefix(c.Data, "year"):
		st := b.setYear(c)
		return st
	case strings.HasPrefix(c.Data, "month"):
		b.setMonth(c)
	default:
		b.setCategory(c)
	}
	return b.startBot
}

// выбор года для отчета
func (b *bot) getYear() {
	msg := "Выберите год:"

	opt := echotron.MessageOptions{
		ReplyMarkup: yearsButtons(),
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

// выбор месяца для отчета
func (b *bot) setYear(c *echotron.CallbackQuery) stateFn {
	year, _ := isNumeric(strings.Split(c.Data, ":")[1])
	b.year = year

	msg := "Выберите месяц:"

	message := echotron.NewMessageID(b.chatID, c.Message.ID)
	_, err := b.EditMessageText(msg, message, &echotron.MessageTextOptions{
		ReplyMarkup: monthButtons(),
	})
	if err != nil {
		log.Println(err)
	}
	defer b.AnswerCallbackQuery(c.ID, nil)
	return b.startBot
}

func (b *bot) setMonth(c *echotron.CallbackQuery) {
	month, _ := strconv.ParseInt(strings.Split(c.Data, ":")[1], 10, 64)
	b.month = time.Month(month)
	b.DeleteMessage(b.chatID, c.Message.ID)
	defer b.AnswerCallbackQuery(c.ID, nil)
	b.userMonthStats(b.year, b.month)
}

func (b *bot) startUser(userName string, msgDate int) stateFn {
	if err := b.db.SaveUser(b.chatID, msgDate, userName); err != nil {
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
	b.SetMyCommands(nil, echotron.BotCommand{Command: "start", Description: "старт бота"})
	return b.startBot
}

func (b *bot) mainMenu(userName string) {

	msg := "Отправь число для начала учета твоих финансов"

	opt := echotron.MessageOptions{
		ReplyMarkup: mainButtons(),
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		fmt.Println(message, err)
	}
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

func (b *bot) setCategory(c *echotron.CallbackQuery) {
	id, err := b.db.SaveData(b.amount, c.Data, b.chatID)
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
	err := b.db.SetComment(comment, b.recordId)
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
	err := b.db.DeleteUserData(recordId)
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

func (b *bot) userMonthStats(year int, month time.Month) {
	var (
		msg string
		sum int
	)

	all, err := b.db.GetSum(b.chatID, year, month)
	if err != nil {
		log.Println(err)
		return
	}
	if len(all) > 0 {
		msg += fmt.Sprintf("Траты за %s\n\n", time.Now().Month().String())
		for i, data := range all {
			msg += fmt.Sprintf("%v.    %s %v руб - %s - %s\n", i+1, data.Category, data.Data, data.Comment, data.Date.Local().Format("02/01/2006 15:04:05"))
			sum += data.Data
		}
		msg += fmt.Sprintf("\nИтого: %v руб\n", sum)
	} else {
		msg = fmt.Sprintf("У вас нет записей за %s %v\n", month, year)
	}
	message, err := b.SendMessage(msg, b.chatID, nil)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) userAllStats() {
	all, err := b.db.AllSum(b.chatID)
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
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) credits() {
	msg := "Выберите действие:"

	opt := echotron.MessageOptions{
		ReplyMarkup: creditButtons(),
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) settings() {
	msg := "Выберите действие:"

	opt := echotron.MessageOptions{
		ReplyMarkup: settingsButtons(),
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}
