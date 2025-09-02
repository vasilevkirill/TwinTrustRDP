package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	ldapv3 "github.com/go-ldap/ldap/v3"
	"github.com/robfig/cron/v3"
	"log"
	"strconv"
)

// Объявление глобальных переменных
var (
	conn *ldapv3.Conn // Подключение к LDAP серверу
)

// Функция ldapRun запускает процесс взаимодействия с LDAP
func ldapRun() error {
	log.Println("LDAP - Запуск процесса взаимодействия с LDAP сервером...")
	err := createConn() // Создание соединения с LDAP сервером
	if err != nil {
		log.Printf("LDAP - Ошибка при создании соединения: %s", err)
		return err
	}
	log.Println("LDAP - Соединение с сервером успешно создано.")
	return nil
}

// Функция ping проверяет доступность LDAP сервера
func ping() {
	log.Println("LDAP - Проверка доступности LDAP сервера...")
	filter := "(&(objectClass=organizationalPerson)(objectClass=user)(sAMAccountName=test))" // Фильтр для запроса
	attrs := []string{"displayName"}                                                         // Атрибуты для запроса
	_, err := searchFilterAttrs(filter, attrs)                                               // Поиск по фильтру и атрибутам
	if err != nil {
		log.Println("Ошибка ping LDAP, пытаемся установить новое соединение с LDAP...")
		err := createConn() // Попытка установить новое соединение с LDAP сервером
		if err != nil {
			log.Printf("LDAP - Не удалось установить новое соединение: %s", err)
		} else {
			log.Println("LDAP - Новое соединение успешно установлено.")
		}
		return
	}
	log.Println("LDAP - Сервер доступен.")
	return
}

// Функция createConn создает соединение с LDAP сервером
func createConn() error {
	log.Println("LDAP - Попытка установить соединение с LDAP сервером...")
	for _, ldapServer := range configGlobalS.Ldap.Servers {
		log.Printf("LDAP - Подключение к серверу LDAP: %s", ldapServer)
		ldapL, err := ldapv3.DialURL(fmt.Sprintf("ldap://%s", ldapServer)) // Подключение к LDAP серверу
		if err != nil {
			log.Printf("LDAP - Ошибка при подключении к серверу LDAP: %s", ldapServer)
			log.Println("Ошибка:", err)
			continue
		}

		log.Printf("LDAP - Установка TLS соединения с сервером LDAP: %s", ldapServer)
		err = ldapL.StartTLS(&tls.Config{InsecureSkipVerify: true}) // Начало TLS сессии
		if err != nil {
			log.Printf("LDAP - Ошибка при установки TLS соединения с сервером LDAP: %s", ldapServer)
			log.Println("Ошибка:", err)
			continue
		}

		log.Printf("LDAP - Попытка аутентификации на сервере LDAP: %s", ldapServer)
		err = ldapL.Bind(configGlobalS.Ldap.User, configGlobalS.Ldap.Password) // Аутентификация на LDAP сервере
		if err != nil {
			log.Printf("LDAP - Ошибка при авторизации (bind) на сервере LDAP: %s", ldapServer)
			log.Println("Ошибка:", err)
			continue
		}

		log.Printf("LDAP - Успешно установлено соединение с сервером LDAP: %s", ldapServer)
		conn = ldapL // Установка глобальной переменной соединения
		schedule()   // Запуск планировщика
		return nil
	}

	log.Println("LDAP - Не удалось установить соединение ни с одним из LDAP серверов.")
	return errorGetFromId(801) // Возвращаем ошибку, если соединение не успешно
}

// Функция schedule запускает планировщик для периодической проверки доступности LDAP сервера
func schedule() {
	log.Println("LDAP - Запуск планировщика для периодической проверки доступности LDAP сервера...")
	c := cron.New()
	_, err := c.AddFunc("@every 1m", ping) // Периодический вызов функции ping каждую минуту
	if err != nil {
		log.Printf("LDAP - Ошибка при добавлении функции ping в планировщик: %s", err)
	}
	c.Start()
	log.Println("LDAP - Планировщик запущен.")
}

// Функция searchFilterAttrs выполняет поиск записей в LDAP по фильтру и атрибутам
func searchFilterAttrs(filter string, attr []string) ([]*ldapv3.Entry, error) {
	log.Printf("LDAP - Выполнение поиска с фильтром: %s", filter)
	// Создание запроса поиска
	searchRequest := ldapv3.NewSearchRequest(
		configGlobalS.Ldap.Dn,
		ldapv3.ScopeWholeSubtree,
		ldapv3.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attr,
		nil,
	)

	// Выполнение запроса поиска
	req, err := conn.Search(searchRequest)
	if err != nil {
		errN := errorGetFromIdAddSuffix(802, err.Error())
		log.Printf("LDAP - Ошибка при выполнении поиска: %s", err)
		return []*ldapv3.Entry{}, errN
	}

	log.Printf("LDAP - Поиск завершен. Найдено записей: %d", len(req.Entries))
	return req.Entries, nil
}

// PullViaTelegramId Метод PullViaTelegramId получает данные пользователя LDAP по TelegramId
func (u *ldapUser) PullViaTelegramId() error {
	log.Printf("LDAP - Получение данных пользователя по TelegramId: %d", u.TelegramId)
	if u.TelegramId == 0 {
		log.Println("LDAP - Ошибка: TelegramId не указан.")
		return errors.New("telegram id не указан")
	}

	filter := fmt.Sprintf("(&(objectClass=organizationalPerson)(objectClass=user)(pager=%d))", u.TelegramId) // Фильтр для запроса
	attrs := []string{"pager", "displayName", "sAMAccountName"}                                              // Атрибуты для запроса
	req, err := searchFilterAttrs(filter, attrs)                                                             // Поиск по фильтру и атрибутам
	if err != nil {
		return err
	}

	if len(req) == 0 {
		log.Printf("LDAP - Пользователь с TelegramId: %d не найден.", u.TelegramId)
		return ldapErrUserNotFound
	}

	if len(req) > 1 {
		log.Printf("LDAP - Найдено более одного пользователя с TelegramId: %d", u.TelegramId)
		return ldapErrUserFoundMoreThanOne
	}

	u.DisplayName = req[0].GetAttributeValue("displayName")
	u.SAMAccountName = req[0].GetAttributeValue("sAMAccountName")
	log.Printf("LDAP - Успешно получены данные пользователя: %+v", u)
	return nil
}

// PullViaSAMAccountName Метод PullViaSAMAccountName получает данные пользователя LDAP по имени учетной записи
func (u *ldapUser) PullViaSAMAccountName() error {
	log.Printf("LDAP - Получение данных пользователя по имени учетной записи: %s", u.SAMAccountName)
	if u.SAMAccountName == "" {
		log.Println("LDAP - Ошибка: имя учетной записи не указано.")
		return ldapErrUserSAMAccountRequired
	}

	filter := fmt.Sprintf("(&(objectClass=organizationalPerson)(objectClass=user)(sAMAccountName=%s))", u.SAMAccountName) // Фильтр для запроса
	attrs := []string{"pager", "displayName", "sAMAccountName"}                                                           // Атрибуты для запроса
	req, err := searchFilterAttrs(filter, attrs)                                                                          // Поиск по фильтру и атрибутам
	if err != nil {
		return err
	}

	if len(req) == 0 {
		log.Printf("LDAP - Пользователь с именем учетной записи: %s не найден.", u.SAMAccountName)
		return ldapErrUserNotFound
	}

	if len(req) > 1 {
		log.Printf("LDAP - Найдено более одного пользователя с именем учетной записи: %s", u.SAMAccountName)
		return ldapErrUserFoundMoreThanOne
	}

	u.DisplayName = req[0].GetAttributeValue("displayName")
	pager := req[0].GetAttributeValue("pager")
	TelegramId, err := strconv.ParseInt(pager, 10, 64)
	if err != nil {
		errN := errorGetFromIdAddSuffix(803, err.Error())
		log.Printf("LDAP - Ошибка при преобразовании pager в TelegramId: %s", err)
		return errN
	}

	u.TelegramId = TelegramId
	log.Printf("LDAP - Успешно получены данные пользователя: %+v", u)
	return nil
}
