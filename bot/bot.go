package bot

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NicoNex/echotron/v3"
	"telegram-finance/sql"
)

const (
	markdownV2 = "MarkdownV2"
)

type stateFn func(*echotron.Update) stateFn

type bot struct {
	chatID    int64
	adminId   int64
	middleId  int64
	amount    int
	recordId  int64
	year      int
	categoryB bool
	db        sql.Db
	month     time.Month
	state     stateFn
	echotron.API
}

func NewBot(chatID int64) echotron.Bot {
	token := os.Getenv("TOKEN")
	adminIdStr := os.Getenv("ADMIN_ID")
	adminId, err := strconv.ParseInt(adminIdStr, 10, 64)
	if err != nil {
		log.Panic(err)
	}
	db, err := sql.NewStorage("data/db.db")
	if err != nil {
		log.Println(err)
	}

	if err := db.Init(); err != nil {
		log.Println(err)
	}
	bot := &bot{
		chatID:  chatID,
		adminId: adminId,
		db:      *db,
		API:     echotron.NewAPI(token),
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
		st := b.handleMessage(update)
		return st
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
	case strings.HasSuffix(msgText, "Расходы по категориям"):
		b.userMonthStatsCategory(b.chatID, time.Now().Year(), time.Now().Month())
	case strings.HasSuffix(msgText, "Расходы по категориям за опр. месяц"):
		b.categoryB = true
		b.getYear()
	case strings.HasSuffix(msgText, "Расходы за опр. месяц"):
		b.getYear()
	case strings.HasSuffix(msgText, "Расходы за текущий месяц"):
		b.userMonthStats(b.chatID, time.Now().Year(), time.Now().Month())
	case strings.HasSuffix(msgText, "Потрачено за все время"):
		b.userAllStats(b.chatID)

	//case strings.HasSuffix(msgText, "Долги"):
	//	b.credits()
	//case strings.HasSuffix(msgText, "Выставить долг"):
	//case strings.HasSuffix(msgText, "Мне должны"):
	//case strings.HasSuffix(msgText, "Я должен"):
	case strings.HasSuffix(msgText, "Главное меню"):
		b.mainMenu()
	case strings.HasSuffix(msgText, "Настройки"):
		b.settings()

	}

	if b.chatID == b.adminId {
		switch {
		case strings.HasSuffix(msgText, "Админ"):
			st := b.addAdminMiddleId()
			return st
		case strings.HasSuffix(msgText, "Написать расход"):
			st := b.addAdminAmount()
			return st
		case strings.HasSuffix(msgText, "Расходы по категориям A"):
			b.userMonthStatsCategory(b.middleId, time.Now().Year(), time.Now().Month())
		case strings.HasSuffix(msgText, "Расходы по категориям за опр. месяц A"):
			b.categoryB = true
			b.getYear()
		case strings.HasSuffix(msgText, "Расходы за опр. месяц A"):
			b.getYear()
		case strings.HasSuffix(msgText, "Расходы за текущий месяц A"):
			b.userMonthStats(b.middleId, time.Now().Year(), time.Now().Month())
		case strings.HasSuffix(msgText, "Потрачено за все время A"):
			b.userAllStats(b.middleId)
		}
	}

	if num, ok := isNumeric(update.Message.Text); ok {
		b.addAmount(num, b.chatID)
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
		if b.categoryB {
			b.userMonthStatsCategory(b.middleId, b.year, b.month)
			b.categoryB = false
			break
		}
		b.userMonthStats(b.middleId, b.year, b.month)
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
	//b.userMonthStats(b.year, b.month)
}

func (b *bot) startUser(userName string, msgDate int) stateFn {
	if err := b.db.SaveUser(b.chatID, msgDate, userName); err != nil {
		log.Println(userName, err)
	}

	msg := "Отправь число для начала учета твоих финансов"
	var opt echotron.MessageOptions
	if b.chatID == b.adminId {
		opt = echotron.MessageOptions{
			ReplyMarkup: adminMainButtons(),
		}
	} else {
		opt = echotron.MessageOptions{
			ReplyMarkup: mainButtons(),
		}
	}

	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		fmt.Println(message, err)
	}
	b.SetMyCommands(nil, echotron.BotCommand{Command: "start", Description: "старт бота"})
	return b.startBot
}

func (b *bot) mainMenu() {

	msg := "Отправь число для начала учета твоих финансов"

	var opt echotron.MessageOptions
	if b.chatID == b.adminId {
		opt = echotron.MessageOptions{
			ReplyMarkup: adminMainButtons(),
		}
	} else {
		opt = echotron.MessageOptions{
			ReplyMarkup: mainButtons(),
		}
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		fmt.Println(message, err)
	}
}

func (b *bot) addAmount(amount int, id int64) {
	b.amount = amount
	b.middleId = id
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
	//id, err := b.db.SaveData(b.amount, c.Data, b.chatID)
	id, err := b.db.SaveData(b.amount, c.Data, b.middleId)
	if err != nil {
		log.Println(c.From.Username, err)
		return
	}

	msg := fmt.Sprintf("✅️ Внесено `%v` руб в %s \n🗓 %s", b.amount, c.Data, time.Unix(int64(c.Message.Date), 0).Format("02/01/2006 15:04:05"))
	message := echotron.NewMessageID(b.chatID, c.Message.ID)
	_, err = b.EditMessageText(msg, message, &echotron.MessageTextOptions{
		ReplyMarkup: recordButtons(id),
		ParseMode:   markdownV2,
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

func (b *bot) userMonthStats(userId int64, year int, month time.Month) {
	var (
		msg string
		sum int
	)

	all, err := b.db.GetSum(userId, year, month)
	if err != nil {
		log.Println(err)
		return
	}
	if len(all) > 0 {
		msg += fmt.Sprintf("``` Траты за %s\n\n", time.Now().Month().String())
		for i, data := range all {
			msg += fmt.Sprintf("%v\\. %s | %s | %v руб | %s\n", i+1, data.Date.Local().Format("02/01/2006"), data.Category, data.Data, data.Comment)
			sum += data.Data
		}
		msg += fmt.Sprintf("\nИтого\\: %v руб```\n", sum)
	} else {
		msg = fmt.Sprintf("У вас нет записей за %s %v\n", month, year)
	}
	opt := echotron.MessageOptions{
		ParseMode: markdownV2,
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) userMonthStatsCategory(userId int64, year int, month time.Month) {
	var (
		msg string
	)

	categories := make(map[string]int)

	all, err := b.db.GetSum(userId, year, month)
	if err != nil {
		log.Println(err)
		return
	}
	if len(all) > 0 {
		msg += fmt.Sprintf("``` Траты за %s по категориям\n\n", time.Now().Month().String())
		for _, data := range all {
			categories[data.Category] += data.Data
		}
		keys := make([]string, 0, len(categories))
		for key := range categories {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return categories[keys[i]] > categories[keys[j]] })
		for _, key := range keys {
			msg += fmt.Sprintf("%s | %v руб \n", key, categories[key])
		}

		msg += fmt.Sprintf("```\n")
	} else {
		msg = fmt.Sprintf("У вас нет записей за %s %v\n", month, year)
	}
	opt := echotron.MessageOptions{
		ParseMode: markdownV2,
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) userAllStats(userId int64) {
	all, err := b.db.AllSum(userId)
	if err != nil {
		log.Println(err)
		return
	}
	msg := fmt.Sprintf("``` Общая сумма трат за все время: %v рублей```\n", all)

	opt := echotron.MessageOptions{
		ParseMode: markdownV2,
	}
	message, err := b.SendMessage(msg, b.chatID, &opt)
	if err != nil {
		log.Println(message, err)
	}
}

func (b *bot) addAdminAmount() stateFn {
	msg := "Напишите расход:"
	message, err := b.SendMessage(msg, b.chatID, nil)
	if err != nil {
		log.Println(message, err)
	}
	return b.getAdminSent
}

func (b *bot) getAdminSent(update *echotron.Update) stateFn {
	num, _ := isNumeric(update.Message.Text)
	b.addAmount(num, b.middleId)

	return b.startBot
}

func (b *bot) addAdminMiddleId() stateFn {
	msg := "Напишите id:"
	message, err := b.SendMessage(msg, b.chatID, nil)
	if err != nil {
		log.Println(message, err)
	}
	return b.getAdminMiddleId
}

func (b *bot) getAdminMiddleId(update *echotron.Update) stateFn {
	userId, _ := strconv.ParseInt(update.Message.Text, 10, 64)
	b.middleId = userId
	b.adminReports()
	return b.startBot
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

func (b *bot) adminReports() {
	msg := "Выберите отчеты A:"

	opt := echotron.MessageOptions{
		ReplyMarkup: adminReportButtons(),
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
