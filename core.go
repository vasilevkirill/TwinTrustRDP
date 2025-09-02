package main

func Run() error {
	err := loadConfig() // поучаем конфиг
	if err != nil {
		return err
	}

	initQ()
	err = initCache() // инициализируем кеш
	if err != nil {
		return err
	}

	err = ldapRun() // инициализируем ldap
	if err != nil {
		return err
	}

	err = telegramRun() // инициализируем бета телеграм
	if err != nil {
		return err
	}

	err = radiusRun() // инициализируем радиус сервер
	if err != nil {
		return err
	}

	return nil
}
