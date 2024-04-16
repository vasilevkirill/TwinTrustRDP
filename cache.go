package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Объявление переменных и структур

var Mpw = MapCache{}                  // Глобальная переменная для хранения кэша
var CacheFile = "./config/cache.json" // Путь к файлу кэша

// MapCache Структура MapCache для хранения кэша
type MapCache struct {
	Map map[int64]time.Time // Карта для хранения времени создания записей в кэше
	Rw  sync.RWMutex        // Mutex для безопасного доступа к карте
}

// Функция initCache инициализирует кэш
func initCache() error {
	// Проверка существования файла кэша
	if fileExist(CacheFile) {
		bt, err := os.ReadFile(CacheFile)
		if err != nil {
			// Обработка ошибки чтения файла кэша
			log.Printf("Произошла ошибка при чтении файла %s: %s", CacheFile, err.Error())

			// Получение текущего времени
			currentTime := time.Now()

			// Форматирование текущего времени в строку "2006-01-02_15-04"
			timestamp := currentTime.Format("2006-01-02_15-04")

			// Получение расширения файла
			extension := filepath.Ext(CacheFile)

			// Формирование нового имени файла с добавлением времени перед расширением
			newFileName := fmt.Sprintf("%s_%s%s",
				CacheFile[:len(CacheFile)-len(extension)],
				timestamp,
				extension)

			// Переименование файла
			err := os.Rename(CacheFile, newFileName)
			if err != nil {
				log.Printf("Ошибка при переименовании файла из %s в %s: %s", CacheFile, newFileName, err.Error())
				return err
			}
		}

		// Декодирование содержимого файла в карту кэша
		err = json.Unmarshal(bt, &Mpw.Map)
		if err != nil {
			// Обработка ошибки декодирования файла кэша
			log.Printf("Ошибка при декодирования в json файла %s", CacheFile)

			// Удаление поврежденного файла кэша
			err := os.Remove(CacheFile)
			if err != nil {
				log.Printf("Ошибка при удалении файла %s", CacheFile)
				return err
			}
		}
	} else {
		// Создание новой карты, если файл кэша не существует
		mp := make(map[int64]time.Time)
		Mpw.Map = mp
	}

	return nil
}

func (w *MapCache) remove(TelegramId int64) {
	w.Rw.Lock()
	defer w.Rw.Unlock()
	delete(w.Map, TelegramId)
}

// Функция check проверяет наличие записи в кэше по TelegramId
func (w *MapCache) check(TelegramId int64) bool {
	// Блокировка для чтения
	w.Rw.Lock()
	defer w.Rw.Unlock()

	// Поиск записи в кэше
	t, ok := w.Map[TelegramId]
	if ok == false {
		return false
	}
	// Получение текущего времени
	currentTime := time.Now()
	// Вычисление разницы времени
	diff := t.Sub(currentTime)
	seconds := int(diff.Seconds())
	// Проверка времени жизни записи в кэше
	if seconds < configGlobalS.Cache.Timeout {
		return true
	}
	return false
}

// Функция add добавляет запись в кэш
func (w *MapCache) add(TelegramId int64) {
	// Получение текущего времени
	currentTime := time.Now()
	// Добавление заданного количества секунд к текущему времени
	second := configGlobalS.Cache.Timeout
	newTime := currentTime.Add(time.Duration(second) * time.Second)
	// Блокировка для записи
	w.Rw.Lock()
	w.Map[TelegramId] = newTime
	w.Rw.Unlock()
	// Сохранение кэша в файл
	w.save()
	return
}

// Функция save сохраняет содержимое кэша в файл
func (w *MapCache) save() {
	// Блокировка для записи
	w.Rw.Lock()
	defer w.Rw.Unlock()

	// Кодирование карты кэша в формат JSON
	bt, err := json.Marshal(w.Map)
	if err != nil {
		// Обработка ошибки кодирования данных в JSON
		log.Printf("Произошла ошибка при конвертации данных в json cach: %s", err.Error())
		return
	}
	// Запись данных в файл кэша
	err = os.WriteFile(CacheFile, bt, os.ModePerm)
	if err != nil {
		// Обработка ошибки сохранения данных в файл
		log.Printf("Произошла ошибка при сохраниении данных в файл %s: %s", CacheFile, err.Error())
		return
	}
	return
}
