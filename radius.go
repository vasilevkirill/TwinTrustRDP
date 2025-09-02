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
	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),                                                   // Установка обработчика пакетов.
		SecretSource: radius.StaticSecretSource([]byte(configGlobalS.Radius.Secret)),                // Установка секрета.
		Addr:         fmt.Sprintf("%s:%d", configGlobalS.Radius.Address, configGlobalS.Radius.Port), // Формирование адреса сервера.
	}
	configGlobalS.Radius.ServerAddress = server.Addr
	log.Printf("Radius - Запуск сервера на %s", server.Addr)

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
	log.Printf("Radius - Получение имени пользователя из запроса: %s", username)
	username = getUserName(username)

	// Получение данных пользователя из базы данных.
	user, err := getUser(username)
	if err != nil {
		log.Println(err)
		sendAccessReject(w, r)
		return
	}
	log.Printf("Radius - Данные пользователя успешно получены: %+v", user)

	// Проверка наличия значения TelegramId у пользователя.
	if user.TelegramId == 0 {
		log.Printf("Radius - Пользователь %s не имеет значения TelegramId", user.SAMAccountName)
		sendAccessReject(w, r)
		return
	}

	// Проверка наличия пользователя в кэше.
	if Mpw.check(user.TelegramId) {
		log.Printf("Radius - Пользователь %s уже в кэше, отправляем Access-Accept", user.SAMAccountName)
		Mpw.add(user.TelegramId)
		sendAccessAccept(w, r)
		return
	}

	// Проверка наличия запроса для данного пользователя.
	if qu.IssetKey(user.TelegramId) {
		log.Printf("Radius - Запрос пользователю %s уже отправлен, ожидаем ответа", user.SAMAccountName)
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
		log.Printf("Radius - Время ожидания ответа истекло для пользователя %s, отправляем Access-Reject", user.SAMAccountName)
		sendAccessReject(w, r)
		return
	}()

	// Получение сообщения из очереди и ожидание ответа.
	msg := qu.GetMsg(user.TelegramId)
	log.Printf("Radius - Ожидание ответа от пользователя %s на подключение", user.SAMAccountName)
	err = waitAnswer(ctx, msg, user)
	if err != nil {
		log.Println(err)
		qu.RemoveKey(user.TelegramId)
		errN := errorGetFromIdAddSuffix(701, err.Error())
		log.Println(errN)
		sendAccessReject(w, r)
		errR := removeMsgByChaiIDMsgIDForce(user.TelegramId, msg.MsgId)
		if errR != nil {
			log.Printf("Radius - Ошибка при удалении сообщения для пользователя %s: %s", user.SAMAccountName, errR)
		}
		return
	}

	qu.RemoveKey(user.TelegramId)
	Mpw.add(user.TelegramId)
	log.Printf("Radius - Пользователь %s успешно аутентифицирован", user.SAMAccountName)
	sendAccessAccept(w, r)
}

// Функция waitAnswer ожидает ответа от пользователя.
func waitAnswer(ctx context.Context, msg queueMsg, user ldapUser) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("radius - Пользователю %s отказанно: %s", user.SAMAccountName, ctx.Err())
		case num, ok := <-msg.Chan:
			if !ok {
				return errors.New(fmt.Sprintf("Radius - Канал для пользователя %s закрыт", user.SAMAccountName))
			}
			if num == 0 {
				return errors.New(fmt.Sprintf("Radius - Пользователь %s выбрал No", user.SAMAccountName))
			}
			return nil
		}
	}
}

// Функция sendAccessAccept отправляет ответ Access-Accept.
func sendAccessAccept(w radius.ResponseWriter, r *radius.Request) {
	log.Println("Radius - Отправка Access-Accept")
	send(w, r, radius.CodeAccessAccept)
}

// Функция sendAccessReject отправляет ответ Access-Reject.
func sendAccessReject(w radius.ResponseWriter, r *radius.Request) {
	log.Println("Radius - Отправка Access-Reject")
	send(w, r, radius.CodeAccessReject)
}

// Функция send отправляет ответ на запрос пользователя.
func send(w radius.ResponseWriter, r *radius.Request, code radius.Code) {
	p := r.Response(code)
	prx := rfc2865.ProxyState_Get(r.Packet)
	p.Add(rfc2865.ProxyState_Type, prx)
	err := w.Write(p)
	if err != nil {
		log.Printf("Radius - Ошибка при отправке ответа: %s", err.Error())
	}
}

// Функция getUserName извлекает имя пользователя из формата DOMAIN\UserName.
func getUserName(user string) string {
	userSplit := strings.Split(user, `\`)
	if len(userSplit) == 2 {
		log.Printf("Radius - Извлечение имени пользователя: %s", userSplit[1])
		return userSplit[1]
	}
	log.Printf("Radius - Имя пользователя без DOMAIN: %s", user)
	return user
}

// Функция getUser получает данные пользователя из базы данных.
func getUser(sAMAccountName string) (ldapUser, error) {
	log.Printf("Radius - Получение данных пользователя с SAMAccountName: %s", sAMAccountName)
	u := ldapUser{}
	u.SAMAccountName = sAMAccountName
	err := u.PullViaSAMAccountName()
	if err != nil {
		errN := errorGetFromIdAddSuffix(702, err.Error())
		log.Printf("Radius - Ошибка получения данных пользователя: %s", errN)
		return ldapUser{}, errN
	}
	log.Printf("Radius - Данные пользователя успешно загружены: %+v", u)
	return u, nil
}
