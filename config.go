package main

import (
	"github.com/spf13/viper"
	"log"
)

// Глобальные переменные для хранения конфигурации
var (
	configGlobalS configS       // Глобальная структура для хранения конфигурации
	configViper   = viper.New() // Инициализация нового экземпляра viper для работы с конфигурацией
)

// Функция loadConfig загружает конфигурацию из файла
func loadConfig() error {
	Config := viper.New()            // Создание нового экземпляра viper для загрузки конфигурации
	Config.AddConfigPath("./config") // Установка пути к файлу конфигурации
	Config.SetConfigName("config")   // Установка имени файла конфигурации
	Config.SetConfigType("yaml")     // Установка типа файла конфигурации

	err := Config.ReadInConfig() // Загрузка конфигурации из файла
	if err != nil {
		return err // Возвращаем ошибку, если не удалось прочитать файл конфигурации
		//panic(fmt.Errorf("fatal error config file: %s", err.Error()))
	}
	var configSv configS
	err = Config.Unmarshal(&configSv) // Распаковка конфигурации в структуру
	if err != nil {
		return err // Возвращаем ошибку, если не удалось распаковать конфигурацию
	}
	configGlobalS = configSv // Установка глобальной конфигурации
	configViper = Config     // Установка экземпляра viper в глобальную переменную
	if err = checkConfigLdap(); err != nil {
		return err // Проверка конфигурации LDAP
	}
	if err = checkConfigTelegram(); err != nil {
		return err // Проверка конфигурации Telegram
	}
	return nil
}

// Функция checkConfigLdap проверяет конфигурацию LDAP
func checkConfigLdap() error {
	// Проверка обязательных полей конфигурации LDAP
	if configGlobalS.Ldap.Dn == "" {
		return errorGetFromId(400)
	}
	if configGlobalS.Ldap.User == "" {
		return errorGetFromId(401)
	}
	if configGlobalS.Ldap.Password == "" {
		return errorGetFromId(402)
	}
	return nil
}

// Функция checkConfigTelegram проверяет конфигурацию Telegram
func checkConfigTelegram() error {
	// Проверка обязательных полей конфигурации Telegram
	if configGlobalS.Telegram.Token == "" {
		return errorGetFromId(303)
	}
	if configGlobalS.Telegram.HookPort == 0 {
		return errorGetFromId(304)
	}
	if configGlobalS.Telegram.HookPort > 65535 && configGlobalS.Telegram.HookPort < 1 {
		return errorGetFromId(305)
	}
	if configGlobalS.Telegram.HookDomain == "" {
		return errorGetFromId(306)
	}
	// Проверка дополнительных параметров конфигурации Telegram
	if configGlobalS.Telegram.PoolAddress == "" {
		log.Println("telegram.PoolAddress config is empty, set default 0.0.0.0")
		configGlobalS.Telegram.PoolAddress = "0.0.0.0"
	}
	if configGlobalS.Telegram.HookCertPub == "" || configGlobalS.Telegram.HookCertKey == "" {
		log.Println("telegram.HookCertPub or telegram.HookCertKey config is empty, used self-sign")
		// Генерация самоподписанного сертификата, если не указаны ключи
		err := generateCertificate()
		if err != nil {
			return errorGetFromIdAddSuffix(500, err.Error())
		}
	}
	return nil
}
