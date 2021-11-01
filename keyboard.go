package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var keyboards = map[string]tgbotapi.InlineKeyboardMarkup{
	"sugaroid:yesno": tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Yes", "yes"),
			tgbotapi.NewInlineKeyboardButtonData("No", "no"),
			tgbotapi.NewInlineKeyboardButtonData("🤷", "idk"),
		},
	),
}
