package main

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

	err = ldapRun() // инициалзируем ldap
	if err != nil {
		return err
	}

	err = telegramRun() // инициалзируем бета телеграм
	if err != nil {
		return err
	}

	err = radiusRun() // инициалзируем радиус сервер
	if err != nil {
		return err
	}

	return nil
}
