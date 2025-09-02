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
	log.Println("Telegram - Инициализация бота...")
	bt, err := tgbotapi.NewBotAPI(configGlobalS.Telegram.Token)
	if err != nil {
		return errorGetFromIdAddSuffix(600, err.Error())
	}
	bt.Debug = true
	log.Printf("Telegram - Запущен бот в режиме отладки: %v", bt.Debug)

	webHookAddress := fmt.Sprintf("https://%s:%d", configGlobalS.Telegram.HookDomain, configGlobalS.Telegram.HookPort)
	configGlobalS.Telegram.WebHookAddress = webHookAddress
	log.Printf("Telegram - Установка веб хуков по адресу: %s", webHookAddress)

	wh, err := tgbotapi.NewWebhookWithCert(webHookAddress, tgbotapi.FilePath(configGlobalS.Telegram.HookCertPub))
	if err != nil {
		return errorGetFromIdAddSuffix(601, err.Error())
	}

	_, err = bt.Request(wh)
	if err != nil {
		return errorGetFromIdAddSuffix(602, err.Error())
	}
	bot = bt

	log.Println("Telegram - Запуск функции обработки обновлений в отдельной горутине...")
	go updatesWord()
	return nil
}

// updatesWord функция обработки обновлений от бота.
func updatesWord() {
	updates := bot.ListenForWebhook("/")
	log.Println("Telegram - Запуск HTTP сервера для обработки вебхуков...")
	go runHttpServer()

	log.Println("Telegram - Обработка обновлений...")
	for update := range updates {
		// Сначала проверяем нажатие кнопок
		if checkCallbackQuery(update) {
			continue
		}

		if update.Message != nil {
			log.Printf("Telegram - Обновление от пользователя: %d, текст: %s", update.Message.From.ID, update.Message.Text)
			err := checkOldMessage(update.Message)
			if err != nil {
				log.Println(err)
				continue
			}

			if !update.Message.IsCommand() {
				log.Printf("Telegram - Обновление от пользователя %d не является командой: %s", update.Message.From.ID, update.Message.Text)
				continue
			}

			if update.Message.From.IsBot {
				log.Printf("Telegram - Игнорирование сообщения от другого бота: %s", update.Message.Text)
				continue
			}

			// Обрабатываем команды
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
				debug(fmt.Sprintf("Получена команда %s", update.Message.Command()))
			}
		}
	}
}

// checkCallbackQuery проверяет, является ли обновление callback query.
func checkCallbackQuery(update tgbotapi.Update) bool {

	cb := update.CallbackQuery
	if cb == nil {
		return false
	}
	log.Printf("Telegram - Пришёл callback: %+v", cb)

	data := cb.Data
	if data == "" {
		return false
	}

	// Подтверждаем нажатие кнопки, чтобы Telegram "убрал часики"
	callback := tgbotapi.NewCallback(cb.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Println("Telegram - Ошибка при подтверждении callback:", err)
	}

	msg := cb.Message
	log.Printf("Telegram - Пользователь %d нажал кнопку '%s'", msg.Chat.ID, data)

	// Удаляем сообщение с кнопками
	if err := removeMsg(msg); err != nil {
		log.Println("Telegram - Ошибка при удалении сообщения:", err)
		return true
	}

	// Обработка выбора
	m := queueMsg{}
	if m, ok := qu.GetMsg(msg.Chat.ID); ok {
		m.Chan <- 1
	}
	switch data {
	case "yes":
		m.Chan <- 1
		log.Printf("Telegram - Пользователь %d ответил 'да'", msg.Chat.ID)
	case "no":
		m.Chan <- 0
		log.Printf("Telegram - Пользователь %d ответил 'нет'", msg.Chat.ID)
	default:
		log.Printf("Telegram - Неизвестный callback: %s", data)
	}

	return true
}

// sendQuery отправляет запрос пользователю Telegram.
func sendQuery(user ldapUser, timeout int) error {
	log.Printf("Telegram - Отправка запроса пользователю %d с таймаутом %d секунд", user.TelegramId, timeout)
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
		log.Printf("Telegram - Ошибка при отправке сообщения: %s", err.Error())
		return err
	}
	qu.SetMsgId(msgSend.Chat.ID, int64(msgSend.MessageID))
	log.Printf("Telegram - Сообщение успешно отправлено пользователю %d, ID сообщения: %d", user.TelegramId, msgSend.MessageID)
	return nil
}

// removeMsg удаляет сообщение Telegram.
func removeMsg(msg *tgbotapi.Message) error {
	deleteMsgConfig := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
	_, err := bot.Request(deleteMsgConfig)
	if err != nil {
		log.Printf("Telegram - Ошибка при удалении сообщения: %s", err.Error())
		return errorGetFromIdAddSuffix(604, err.Error())
	}
	log.Printf("Telegram - Сообщение удалено: %d", msg.MessageID)
	return nil
}

// removeMsgByChaiIDMsgIDForce удаляет сообщение Telegram по заданному chatID и msgID.
func removeMsgByChaiIDMsgIDForce(chatId, msgId int64) error {
	deleteMsgConfig := tgbotapi.NewDeleteMessage(chatId, int(msgId))
	_, err := bot.Request(deleteMsgConfig)
	if err != nil {
		log.Printf("Telegram - Ошибка при удалении сообщения с ID %d у чата %d: %s", msgId, chatId, err.Error())
		return errorGetFromIdAddSuffix(604, err.Error())
	}
	log.Printf("Telegram - Сообщение с ID %d успешно удалено у чата %d", msgId, chatId)
	return nil
}

// runHttpServer запускает HTTP сервер для обработки вебхука.
func runHttpServer() {
	strConnect := fmt.Sprintf("%s:%d", configGlobalS.Telegram.PoolAddress, configGlobalS.Telegram.PoolPort)
	log.Printf("Telegram - Запуск HTTP сервера на %s", strConnect)

	err := http.ListenAndServeTLS(strConnect, configGlobalS.Telegram.HookCertPub, configGlobalS.Telegram.HookCertKey, nil)
	if err != nil {
		errN := errorGetFromIdAddSuffix(605, err.Error(), strConnect)
		log.Panic(errN)
	}
}

// Debug выводит отладочную информацию.
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
		log.Println("Telegram - Не удалось аутентифицировать пользователя для команды /killVasya")
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ща Усё всё сделаем")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Telegram - Ошибка при отправке сообщения пользователю: %s", err.Error())
	}
	log.Println("Telegram - Завершение программы по команде /killVasya")
	os.Exit(-222)
}

// cmdForce обрабатывает команду /force.
func cmdForce(update tgbotapi.Update) {
	debug("Система получила команду /force")
	auth, user := chatAuth(update)
	if !auth {
		log.Println("Telegram - Не удалось аутентифицировать пользователя для команды /force")
		return
	}
	Mpw.add(user.TelegramId)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Принято, можете авторизоваться")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Telegram - Ошибка при отправке сообщения пользователю: %s", err.Error())
	}
}

// cmdClear обрабатывает команду /clear.
func cmdClear(update tgbotapi.Update) {
	debug("Система получила команду /clear")
	auth, user := chatAuth(update)
	if !auth {
		log.Println("Telegram - Не удалось аутентифицировать пользователя для команды /clear")
		return
	}
	Mpw.remove(user.TelegramId)
	qu.RemoveKey(user.TelegramId)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Принято, всё почистили")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Telegram - Ошибка при отправке сообщения пользователю: %s", err.Error())
	}
}

// cmdHelp обрабатывает команду /help.
func cmdHelp(update tgbotapi.Update) {
	debug("Система получила команду /help")
	auth, _ := chatAuth(update)
	if !auth {
		log.Println("Telegram - Не удалось аутентифицировать пользователя для команды /help")
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "/force - прозрачная авторизация")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Telegram - Ошибка при отправке сообщения пользователю: %s", err.Error())
	}
}

// cmdStart обрабатывает команду /start.
func cmdStart(update tgbotapi.Update) {
	debug("Система получила команду /start")
	auth, _ := chatAuth(update)
	if !auth {
		log.Println("Telegram - Не удалось аутентифицировать пользователя для команды /start")
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Здравствуйте, всё подготовлено, мы уже знакомы.")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Telegram - Ошибка при отправке сообщения пользователю: %s", err.Error())
	}
}

// chatAuth аутентификация пользователя через чат Telegram.
func chatAuth(update tgbotapi.Update) (bool, ldapUser) {
	msgWait := tgbotapi.NewMessage(update.Message.Chat.ID, "Ждите...")
	msgW, err := bot.Send(msgWait)
	user := ldapUser{}
	user.TelegramId = update.Message.From.ID

	log.Printf("Telegram - Аутентификация пользователя с ID %d", user.TelegramId)
	err = user.PullViaTelegramId()
	if err != nil {
		log.Printf("Telegram - Ошибка аутентификации пользователя %d: %s", user.TelegramId, err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
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
			log.Println("Telegram - Ошибка при удалении сообщения ожидания:", err)
		}
		_, err = bot.Send(msg)
		if err != nil {
			log.Println("Telegram - Ошибка при отправке сообщения пользователю после ошибочной аутентификации:", err)
		}
		return false, ldapUser{}
	}
	err = removeMsg(&msgW)
	if err != nil {
		log.Println("Telegram - Ошибка при удалении сообщения ожидания:", err)
	}
	log.Printf("Telegram - Пользователь %d успешно аутентифицирован", user.TelegramId)
	return true, user
}

// checkOldMessage проверяет время старых сообщений.
func checkOldMessage(msg *tgbotapi.Message) error {
	if msg == nil {
		return nil
	}
	msgTime := time.Unix(int64(msg.Date), 0)
	if time.Since(msgTime) > 3*time.Minute {
		err := removeMsg(msg)
		if err != nil {
			return err
		}
		duration := time.Since(msgTime)
		seconds := int(duration.Seconds())
		log.Printf("Telegram - Сообщение от пользователя %d старше 3 минут и удалено. Продолжительность: %d секунд", msg.From.ID, seconds)
		return errorGetFromIdAddSuffix(607, fmt.Sprintf("%d seconds", seconds))
	}
	return nil
}
