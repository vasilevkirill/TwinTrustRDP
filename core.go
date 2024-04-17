package main

import "time"

func Run() error {
	err := loadConfig() // поулчаем конфиг
	if err != nil {
		return err
	}

	initQ()
	err = initCache() // инициалзируем кеш
	if err != nil {
		return err
	}

	// Запуск горутины для очистки старых записей кэша каждые пять минут
	go func() {
		// Создаем канал, который будет отправляться каждые пять минут
		ticker := time.Tick(5 * time.Minute)
		for {
			select {
			case <-ticker:
				// Вызов функции очистки старых записей
				cleanOldEntries()
			}
		}
	}()

	err = ldapRun() // инициалзируем ldap
	if err != nil {
		return err
	}

	err = telegramRun() // инициалзируем бота телеграм
	if err != nil {
		return err
	}

	err = radiusRun() // инициалзируем радиус сервер
	if err != nil {
		return err
	}

	return nil
}
