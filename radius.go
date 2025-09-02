package main

import (
	"context"
	"errors"
	"fmt"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"log"
	"strings"
	"time"
)

// Функция radiusRun запускает сервер Radius.
func radiusRun() error {
	// Создание сервера Radius.
	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),                                                   // Установка обработчика пакетов.
		SecretSource: radius.StaticSecretSource([]byte(configGlobalS.Radius.Secret)),                // Установка секрета.
		Addr:         fmt.Sprintf("%s:%d", configGlobalS.Radius.Address, configGlobalS.Radius.Port), // Формирование адреса сервера.
	}
	configGlobalS.Radius.ServerAddress = server.Addr
	log.Printf(fmt.Sprintf("Radius - Запуск сервера на %s", server.Addr))

	// Запуск сервера и обработка ошибок.
	if err := server.ListenAndServe(); err != nil {
		return errorGetFromIdAddSuffix(700, err.Error(), server.Addr)
	}
	return nil
}

// Функция handler обрабатывает запросы, поступающие на сервер Radius.
func handler(w radius.ResponseWriter, r *radius.Request) {
	// Получение имени пользователя из запроса.
	username := rfc2865.UserName_GetString(r.Packet)
	username = getUserName(username)
	// Получение данных пользователя из базы данных.
	user, err := getUser(username)
	if err != nil {
		log.Println(err)
		sendAccessReject(w, r)
		return
	}
	// Проверка наличия значения TelegramId у пользователя.
	if user.TelegramId == 0 {
		log.Printf("Radius - Пользователь %s не имеет значения TelegramId", user.SAMAccountName)
		sendAccessReject(w, r)
		return
	}
	// Проверка наличия пользователя в кэше.
	if Mpw.check(user.TelegramId) {
		log.Printf("Radius - пользователь %s уже в кэше пропускаем", user.SAMAccountName)
		Mpw.add(user.TelegramId)
		sendAccessAccept(w, r)
		return
	}
	// Проверка наличия запроса для данного пользователя.
	if qu.IssetKey(user.TelegramId) {
		log.Printf("Radius - Запрос пользователю %s уже отправлен", user.SAMAccountName)
		return
	}
	log.Printf("Radius - Запрос на подключение от пользователя %s", user.SAMAccountName)
	qu.AddKey(user.TelegramId)

	// Отправка запроса пользователю и ожидание ответа.
	err = sendQuery(user, configGlobalS.Radius.AnswerTimeout)
	if err != nil {
		log.Println(err)
		sendAccessReject(w, r)
		return
	}

	// Установка таймаута для ожидания ответа.
	ctx := context.Background()
	ctx, cancelFunctionContext := context.WithTimeout(ctx, time.Duration(configGlobalS.Radius.AnswerTimeout)*time.Second)
	defer func() {
		qu.RemoveKey(user.TelegramId)
		cancelFunctionContext()
		sendAccessReject(w, r)
		return
	}()

	// Получение сообщения из очереди и ожидание ответа.
	msg := qu.GetMsg(user.TelegramId)
	err = waitAnswer(ctx, msg, user)
	if err != nil {
		qu.RemoveKey(user.TelegramId)
		errN := errorGetFromIdAddSuffix(701, err.Error())
		log.Println(errN)
		sendAccessReject(w, r)
		errR := removeMsgByChaiIDMsgIDForce(user.TelegramId, msg.MsgId)
		log.Println(errR)
		return
	}
	qu.RemoveKey(user.TelegramId)
	Mpw.add(user.TelegramId)
	log.Printf("Radius - Пользователь %s aвторизирован", user.SAMAccountName)
	sendAccessAccept(w, r)
}

// Функция waitAnswer ожидает ответа от пользователя.
func waitAnswer(ctx context.Context, msg queueMsg, user ldapUser) error {
	for {
		select {
		case <-ctx.Done():
			// Таймаут
			return errors.New(fmt.Sprintf("Radius - Пользователю %s отказанно: %s", user.SAMAccountName, ctx.Err()))

		case num := <-msg.Chan:
			if num == 0 {
				return errors.New(fmt.Sprintf("Radius - Пользователь %s выбрал No", user.SAMAccountName))
			}
			return nil
		}
	}
}

// Функция sendAccessAccept отправляет ответ Access-Accept.
func sendAccessAccept(w radius.ResponseWriter, r *radius.Request) {
	send(w, r, radius.CodeAccessAccept)
	return
}

// Функция sendAccessReject отправляет ответ Access-Reject.
func sendAccessReject(w radius.ResponseWriter, r *radius.Request) {
	send(w, r, radius.CodeAccessReject)
	return
}

// Функция send отправляет ответ на запрос пользователя.
func send(w radius.ResponseWriter, r *radius.Request, code radius.Code) {
	p := r.Response(code)
	prx := rfc2865.ProxyState_Get(r.Packet)
	p.Add(rfc2865.ProxyState_Type, prx)
	err := w.Write(p)
	if err != nil {
		log.Printf("Radius - send error: %s", err.Error())
	}
}

// Функция getUserName извлекает имя пользователя из формата DOMAIN\UserName.
func getUserName(user string) string {
	userSplit := strings.Split(user, `\`)
	if len(userSplit) == 2 {
		return userSplit[1]
	}
	return user
}

// Функция getUser получает данные пользователя из базы данных.
func getUser(sAMAccountName string) (ldapUser, error) {
	u := ldapUser{}
	u.SAMAccountName = sAMAccountName
	err := u.PullViaSAMAccountName()
	if err != nil {
		errN := errorGetFromIdAddSuffix(702, err.Error())
		return ldapUser{}, errN
	}
	return u, nil
}
