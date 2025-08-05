package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"time"
)

var bot *tgbotapi.BotAPI

// telegramRun функция запуска бота Telegram.
func telegramRun() error {
	// Создаем нового бота с помощью токена из конфига.
	bt, err := tgbotapi.NewBotAPI(configGlobalS.Telegram.Token)
	if err != nil {
		return errorGetFromIdAddSuffix(600, err.Error())
	}

	// Включаем режим отладки, если указано в конфиге.
	bt.Debug = configGlobalS.Telegram.Debug
	// Формируем адрес вебхука для бота.
	webHookAddress := fmt.Sprintf("https://%s:%d", configGlobalS.Telegram.HookDomain, configGlobalS.Telegram.HookPort)
	configGlobalS.Telegram.WebHookAddress = webHookAddress
	// Устанавливаем вебхук для бота с помощью SSL-сертификата.
	wh, err := tgbotapi.NewWebhookWithCert(webHookAddress, tgbotapi.FilePath(configGlobalS.Telegram.HookCertPub))
	if err != nil {
		return errorGetFromIdAddSuffix(601, err.Error())
	}
	// Устанавливаем вебхук для бота.
	_, err = bt.Request(wh)
	if err != nil {
		return errorGetFromIdAddSuffix(602, err.Error())
	}

	bot = bt
	// Запускаем функцию обработки обновлений от бота в отдельной горутине.
	go updatesWord()

	return nil
}

// updatesWord функция обработки обновлений от бота.
func updatesWord() {
	// Получаем обновления от бота.
	updates := bot.ListenForWebhook("/")
	// Запускаем HTTP сервер.
	go runHttpServer()
	// Обрабатываем каждое обновление.
	for update := range updates {
		err := checkOldMessage(update.Message)
		if err != nil {
			log.Println(err)
			continue
		}
		// Проверяем, является ли обновление callback query.
		if checkCallbackQuery(update) {
			continue
		}
		// Игнорируем обновления, не являющиеся сообщениями.
		if update.Message == nil {
			continue
		}
		// Игнорируем не командные сообщения.
		if !update.Message.IsCommand() {
			continue
		}
		// Игнорируем сообщения от других ботов.
		if update.Message.From.IsBot {
			continue
		}
		// Обрабатываем команды.
		switch update.Message.Command() {
		case "start":
			cmdStart(update)
		case "force":
			cmdForce(update)
		case "clear":
			cmdClear(update)
		case "killVasya":
			cmdKillVasya(update)
		case "help":
			cmdHelp(update)
		default:
			debug(fmt.Sprintf("Получена комманда %s", update.Message.Command()))
		}
	}
}

// checkCallbackQuery проверяет, является ли обновление callback query.
func checkCallbackQuery(update tgbotapi.Update) bool {
	CallbackQuery := update.CallbackQuery
	data := ""
	if CallbackQuery != nil {
		data = CallbackQuery.Data
	}

	if data == "" {
		return false
	}
	msg := CallbackQuery.Message

	debug(fmt.Sprintf("Пользователь %d нажал %s", msg.Chat.ID, data))
	err := removeMsg(msg)
	if err != nil {
		log.Println(err)
		return true
	}
	m := qu.GetMsg(msg.Chat.ID)
	if data == "yes" {
		m.Chan <- 1
		return true
	}

	if data == "no" {
		m.Chan <- 0
		return true
	}
	return false
}

// sendQuery отправляет запрос пользователю Telegram.
func sendQuery(user ldapUser, timeout int) error {
	// Формируем сообщение с инлайн клавиатурой.
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", "yes"),
			tgbotapi.NewInlineKeyboardButtonData("Нет", "no"),
		),
	)
	str := fmt.Sprintf("Кто-то пытается авторизоваться под вашей учетной записью\nЭто вы?\n Необходимо ответить в течении %d секунд", timeout)
	msg := tgbotapi.NewMessage(user.TelegramId, str)
	msg.ReplyMarkup = inlineKeyboard
	msgSend, err := bot.Send(msg)
	if err != nil {
		return err
	}
	qu.SetMsgId(msgSend.Chat.ID, int64(msgSend.MessageID))
	return nil
}

// removeMsg удаляет сообщение Telegram.
func removeMsg(msg *tgbotapi.Message) error {
	deleteMsgConfig := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
	_, err := bot.Request(deleteMsgConfig)
	if err != nil {
		return errorGetFromIdAddSuffix(604, err.Error())
	}
	return nil
}

// removeMsgByChaiIDMsgIDForce удаляет сообщение Telegram по заданному chatID и msgID.
func removeMsgByChaiIDMsgIDForce(chatId, msgId int64) error {
	deleteMsgConfig := tgbotapi.NewDeleteMessage(chatId, int(msgId))
	_, err := bot.Request(deleteMsgConfig)
	if err != nil {
		return errorGetFromIdAddSuffix(604, err.Error())
	}
	return nil
}

// runHttpServer запускает HTTP сервер для обработки вебхука.
func runHttpServer() {
	strConnect := fmt.Sprintf("%s:%d", configGlobalS.Telegram.PoolAddress, configGlobalS.Telegram.PoolPort)
	err := http.ListenAndServeTLS(strConnect, configGlobalS.Telegram.HookCertPub, configGlobalS.Telegram.HookCertKey, nil)
	if err != nil {
		errN := errorGetFromIdAddSuffix(605, err.Error(), strConnect)
		log.Panic(errN)
	}
}

// debug выводит отладочную информацию.
func debug(str string) {
	if configGlobalS.Telegram.Debug {
		log.Println(str)
	}
}

// cmdKillVasya обрабатывает команду /killVasya.
func cmdKillVasya(update tgbotapi.Update) {
	debug("Система получила команду /killVasya")
	auth, _ := chatAuth(update)
	if !auth {
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.Text = "Ща Усё всё сделаем"
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	debug("Килимся")
	os.Exit(-222)
}

// cmdForce обрабатывает команду /force.
func cmdForce(update tgbotapi.Update) {
	debug("Система получила команду /force")
	auth, user := chatAuth(update)
	if !auth {
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	Mpw.add(user.TelegramId)
	msg.Text = "Принято, можете авторизоваться"
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

// cmdClear обрабатывает команду /clear.
func cmdClear(update tgbotapi.Update) {
	debug("Система получила команду /clear")
	auth, user := chatAuth(update)
	if !auth {
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	Mpw.remove(user.TelegramId)
	qu.RemoveKey(user.TelegramId)
	msg.Text = "Принято, всё почистили"
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

// cmdHelp обрабатывает команду /help.
func cmdHelp(update tgbotapi.Update) {
	debug("Система получила команду /help")
	auth, _ := chatAuth(update)
	if !auth {
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	msg.Text = "/force - прозрачная авторизация"
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
		return
	}
}

// cmdStart обрабатывает команду /start.
func cmdStart(update tgbotapi.Update) {
	debug("Система получила команду /start")
	auth, _ := chatAuth(update)
	if !auth {
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	msg.Text = "Здравствуйте, всё подготовлено, мы уже знакомы."
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
		return
	}
}

// chatAuth аутентификация пользователя через чат Telegram.
func chatAuth(update tgbotapi.Update) (bool, ldapUser) {
	msgWait := tgbotapi.NewMessage(update.Message.Chat.ID, "Ждите...")
	msgW, err := bot.Send(msgWait)

	user := ldapUser{}
	user.TelegramId = update.Message.From.ID
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	err = user.PullViaTelegramId()
	if err != nil {
		switch {
		case errors.Is(err, ldapErrUserNotFound):
			msg.Text = fmt.Sprintf("Привет, мы не знакомы\n отправь с службу поддержки твой ID %d ", user.TelegramId)
		case errors.Is(err, ldapErrUserFoundMoreThanOne):
			msg.Text = fmt.Sprintf("Сообщение об ошибке с кодом 100, пожалуйста, обратитесь в службу технической поддержки.\nid:%d", user.TelegramId)
		default:
			msg.Text = fmt.Sprintf("Сообщение об ошибке с кодом 999, пожалуйста, обратитесь в службу технической поддержки.\nid:%d", user.TelegramId)
		}
		err = removeMsg(&msgW)
		if err != nil {
			log.Println(err)
			return false, ldapUser{}
		}
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
			return false, ldapUser{}
		}
		return false, ldapUser{}
	}
	err = removeMsg(&msgW)
	if err != nil {
		log.Println(err)
	}
	return true, user
}

func checkOldMessage(msg *tgbotapi.Message) error {
	msgTime := time.Unix(int64(msg.Date), 0)
	if time.Since(msgTime) > 3*time.Minute {
		err := removeMsg(msg)
		if err != nil {
			return err
		}
		duration := time.Since(msgTime)
		seconds := int(duration.Seconds())
		return errorGetFromIdAddSuffix(607, fmt.Sprintf("%d seconds", seconds))
	}
	return nil
}
