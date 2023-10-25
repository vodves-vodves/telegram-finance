package bot

import (
	"fmt"
	"time"

	"github.com/NicoNex/echotron/v3"
)

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
				{Text: "💰 Расходы за определенный месяц"},
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
				{Text: "💰 Выставить долг"},
			},
			{
				{Text: "💰 Я должен"},
				{Text: "💰 Мне должны"},
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

func yearsButtons() echotron.InlineKeyboardMarkup {
	nowYear := fmt.Sprintf("%v", time.Now().Year())
	lastYear := fmt.Sprintf("%v", time.Now().Year()-1)
	lastLastYear := fmt.Sprintf("%v", time.Now().Year()-2)
	lastLastYear2 := fmt.Sprintf("%v", time.Now().Year()-3)
	lastLastYear3 := fmt.Sprintf("%v", time.Now().Year()-4)
	return echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{
			{
				{Text: nowYear, CallbackData: "year:" + nowYear},
			},
			{
				{Text: lastYear, CallbackData: "year:" + lastYear},
			},
			{
				{Text: lastLastYear, CallbackData: "year:" + lastLastYear},
			},
			{
				{Text: lastLastYear2, CallbackData: "year:" + lastLastYear2},
			},
			{
				{Text: lastLastYear3, CallbackData: "year:" + lastLastYear3},
			},
		},
	}
}

func monthButtons() echotron.InlineKeyboardMarkup {
	return echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{
			{
				{Text: time.January.String(), CallbackData: fmt.Sprintf("month:%v", 1)},
				{Text: time.February.String(), CallbackData: fmt.Sprintf("month:%v", 2)},
			},
			{
				{Text: time.March.String(), CallbackData: fmt.Sprintf("month:%v", 3)},
				{Text: time.April.String(), CallbackData: fmt.Sprintf("month:%v", 4)},
			},
			{
				{Text: time.May.String(), CallbackData: fmt.Sprintf("month:%v", 5)},
				{Text: time.June.String(), CallbackData: fmt.Sprintf("month:%v", 6)},
			},
			{
				{Text: time.July.String(), CallbackData: fmt.Sprintf("month:%v", 7)},
				{Text: time.August.String(), CallbackData: fmt.Sprintf("month:%v", 8)},
			},
			{
				{Text: time.September.String(), CallbackData: fmt.Sprintf("month:%v", 9)},
				{Text: time.October.String(), CallbackData: fmt.Sprintf("month:%v", 10)},
			},
			{
				{Text: time.November.String(), CallbackData: fmt.Sprintf("month:%v", 11)},
				{Text: time.December.String(), CallbackData: fmt.Sprintf("month:%v", 12)},
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
