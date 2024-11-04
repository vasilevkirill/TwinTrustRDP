package main

import (
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	go periodicRestart(5 * time.Hour)
	// Ваша основная логика приложения
	for {
		checkErrorFatal(Run())
	}

}
func checkErrorFatal(err error) {
	if err == nil {
		return
	}
	log.Fatalln(err.Error())
}

// periodicRestart запускает перезапуск процесса через указанный интервал
func periodicRestart(interval time.Duration) {
	for {
		time.Sleep(interval)

		// Получаем путь к текущему исполняемому файлу
		execPath, err := os.Executable()
		if err != nil {
			log.Fatalln("Не удалось получить путь к исполняемому файлу:", err)
		}

		// Перезапускаем процесс с тем же PID
		log.Println("Перезапускаем приложение...")
		if err := syscall.Exec(execPath, os.Args, os.Environ()); err != nil {
			log.Fatalln("Не удалось перезапустить процесс:", err)
		}
	}
}
