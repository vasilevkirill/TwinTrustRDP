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
	err := createConn() // Создание соединения с LDAP сервером
	if err != nil {
		return err
	}
	return nil
}

// Функция ping проверяет доступность LDAP сервера
func ping() {
	filter := "(&(objectClass=organizationalPerson)(objectClass=user)(sAMAccountName=test))" // Фильтр для запроса
	attrs := []string{"displayName"}                                                         // Атрибуты для запроса
	_, err := searchFilterAttrs(filter, attrs)                                               // Поиск по фильтру и атрибутам
	if err != nil {
		log.Println("Ошибка ping ldap, пытаемся установить новое соединение с ldap")
		err := createConn() // Попытка установить новое соединение с LDAP сервером
		log.Println(err)
		return
	}
	return
}

// Функция createConn создает соединение с LDAP сервером
func createConn() error {
	for _, ldapServer := range configGlobalS.Ldap.Servers {
		log.Printf("Ldap - Попытка установить соединение с сервером ldap: %s", ldapServer)
		ldapL, err := ldapv3.DialURL(fmt.Sprintf("ldap://%s", ldapServer)) // Подключение к LDAP серверу
		if err != nil {
			log.Printf("Ldap - Ошибка при подключении к серверу ldap: %s", ldapServer)
			log.Println(err)
			continue
		}
		err = ldapL.StartTLS(&tls.Config{InsecureSkipVerify: true}) // Начало TLS сессии
		if err != nil {
			log.Printf("Ldap - Ошибка при установки TLS соединения с сервером ldap: %s", ldapServer)
			log.Println(err)
			continue
		}
		err = ldapL.Bind(configGlobalS.Ldap.User, configGlobalS.Ldap.Password) // Аутентификация на LDAP сервере
		if err != nil {
			log.Printf("Ldap - Ошибка при авторизации (bind) на сервере ldap: %s", ldapServer)
			log.Println(err)
			continue
		}
		log.Printf("Ldap - Установлено соединение с сервером ldap: %s", ldapServer)
		conn = ldapL // Установка глобальной переменной соединения
		schedule()   // Запуск планировщика
		return nil
	}
	// Если не удалось установить соединение с LDAP сервером
	return errorGetFromId(801)
}

// Функция schedule запускает планировщик для периодической проверки доступности LDAP сервера
func schedule() {
	c := cron.New()
	_, _ = c.AddFunc("@every 1m", ping) // Периодический вызов функции ping каждую минуту
	c.Start()
}

// Функция searchFilterAttrs выполняет поиск записей в LDAP по фильтру и атрибутам
func searchFilterAttrs(filter string, attr []string) ([]*ldapv3.Entry, error) {
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
		return []*ldapv3.Entry{}, errN
	}
	return req.Entries, nil
}

// PullViaTelegramId Метод PullViaTelegramId получает данные пользователя LDAP по TelegramId
func (u *ldapUser) PullViaTelegramId() error {
	if u.TelegramId == 0 {
		return errors.New("telegramid не указан")
	}
	filter := fmt.Sprintf("(&(objectClass=organizationalPerson)(objectClass=user)(pager=%d))", u.TelegramId) // Фильтр для запроса
	attrs := []string{"pager", "displayName", "sAMAccountName"}                                              // Атрибуты для запроса
	req, err := searchFilterAttrs(filter, attrs)                                                             // Поиск по фильтру и атрибутам
	if err != nil {
		return err
	}
	if len(req) == 0 {
		return ldapErrUserNotFound
	}
	if len(req) > 1 {
		return ldapErrUserFoundMoreThanOne
	}
	u.DispalyName = req[0].GetAttributeValue("displayName")
	u.SAMAccountName = req[0].GetAttributeValue("sAMAccountName")
	return nil
}

// PullViaSAMAccountName Метод PullViaSAMAccountName получает данные пользователя LDAP по имени учетной записи
func (u *ldapUser) PullViaSAMAccountName() error {
	if u.SAMAccountName == "" {
		return ldapErrUserSAMAccountRequired
	}
	filter := fmt.Sprintf("(&(objectClass=organizationalPerson)(objectClass=user)(sAMAccountName=%s))", u.SAMAccountName) // Фильтр для запроса
	attrs := []string{"pager", "displayName", "sAMAccountName"}                                                           // Атрибуты для запроса
	req, err := searchFilterAttrs(filter, attrs)                                                                          // Поиск по фильтру и атрибутам
	if err != nil {
		return err
	}
	if len(req) == 0 {
		return ldapErrUserNotFound
	}
	if len(req) > 1 {
		return ldapErrUserFoundMoreThanOne
	}
	u.DispalyName = req[0].GetAttributeValue("displayName")
	pager := req[0].GetAttributeValue("pager")
	TelegramId, err := strconv.ParseInt(pager, 10, 64)
	if err != nil {
		errN := errorGetFromIdAddSuffix(803, err.Error())
		return errN
	}
	u.TelegramId = TelegramId
	return nil
}
