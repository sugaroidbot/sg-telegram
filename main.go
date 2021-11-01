package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/sugaroidbot/sg-telegram/sgapi"
	"github.com/withmandala/go-log"
	"strconv"

	"os"
	"strings"
)

var logger = log.New(os.Stdout)

var chanMap = map[int64]*sgapi.WsConn{}

func main() {

	tgBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	prefix := os.Getenv("SG_TG_COMMAND_PREFIX")
	wsEndpoint := os.Getenv("SG_TG_WS_ENDPOINT")
	whitelistedChannels := strings.Split(os.Getenv("SG_TG_CHANNEL_ID_WHITELIST"), ",")

	// initialize all whitelisted channels
	for i := range whitelistedChannels {
		if whitelistedChannels[i] == "" {
			continue
		}
		chanId, err := strconv.Atoi(whitelistedChannels[i])
		if err != nil {
			panic(err)
		}
		chanMap[int64(chanId)] = nil
	}
	if len(chanMap) == 0 {
		logger.Fatal("No channels whitelisted in $SG_TG_CHANNEL_ID_WHITELIST, cowardly refusing to start.")
		return
	}

	// initialize the bot
	bot, err := tgbotapi.NewBotAPI(tgBotToken)
	if err != nil {
		logger.Fatal(err)
	}
	bot.Debug = true

	logger.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message != nil && update.Message.Text == "" {
			continue
		}

		if update.Message != nil && !strings.HasPrefix(update.Message.Text, prefix) {
			continue
		}
		var usrMsg *tgbotapi.Message
		if update.Message != nil {
			usrMsg = update.Message
		} else if update.CallbackQuery != nil {
			usrMsg = update.CallbackQuery.Message
		} else {
			return
		}

		v, ok := chanMap[usrMsg.Chat.ID]
		if !ok {
			continue
		}

		if v == nil {
			uid := uuid.New()
			scheme := "wss"

			// use ws:// for localhost, and similar ones
			if strings.HasPrefix(wsEndpoint, "0.0.0.0") || strings.HasPrefix(wsEndpoint, "127.0.0.1") || strings.HasPrefix(wsEndpoint, "localhost") {
				scheme = "ws"
			}
			wsCon, err := sgapi.New(sgapi.Instance{Endpoint: fmt.Sprintf("%s://%s", scheme, wsEndpoint)}, uid)
			if err != nil {
				logger.Warn(err)
				continue
			}
			v = wsCon
			chanMap[usrMsg.Chat.ID] = wsCon

			go func() {
				err := sgapi.Listen(wsCon, func(resp string) {
					if resp == "" {
						// skip empty responses
						return
					}
					msg := tgbotapi.NewMessage(usrMsg.Chat.ID, resp)

					if strings.Contains(resp, "<sugaroid:yesno>") {
						msg.Text = strings.Replace(resp, "<sugaroid:yesno>", "", -1)
						msg.ReplyMarkup = keyboards["sugaroid:yesno"]
					}
					msg.ParseMode = tgbotapi.ModeHTML

					_, err := bot.Send(msg)
					if err != nil {
						logger.Warn(err)
					}
				})
				if err != nil {
					logger.Warn(err)
					chanMap[usrMsg.Chat.ID] = nil
					return
				}
			}()
		}
		if update.Message != nil {
			logger.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
			err := sgapi.Send(v, strings.TrimSpace(strings.TrimPrefix(update.Message.Text, prefix)))
			if err != nil {
				logger.Warn(err)
				chanMap[usrMsg.Chat.ID] = nil
				continue
			}
			_, err = bot.Send(tgbotapi.NewChatAction(usrMsg.Chat.ID, tgbotapi.ChatTyping))
			if err != nil {
				logger.Warn(err)
			}
		} else if update.CallbackQuery != nil {

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				logger.Warn(err)
				continue
			}
			err := sgapi.Send(v, update.CallbackQuery.Data)
			if err != nil {
				logger.Warn(err)
				chanMap[usrMsg.Chat.ID] = nil
				continue
			}
			editedMsg := fmt.Sprintf("%s <i><b>%s</b> answered <b>%s</b></i>", usrMsg.Text, update.CallbackQuery.From.FirstName, update.CallbackQuery.Data)
			m := tgbotapi.NewEditMessageText(usrMsg.Chat.ID, usrMsg.MessageID, editedMsg)
			m.ParseMode = tgbotapi.ModeHTML
			_, err = bot.Send(m)
			if err != nil {
				logger.Warn(err)
			}

			logger.Infof("[%s] (cb) %s", usrMsg.From.UserName, update.CallbackQuery.Data)
		}

	}
}
