package main

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestCreateTestFolder(t *testing.T) {
	if !fileExist("./testdata/") {
		err := os.Mkdir("./testdata/", os.ModePerm)
		if err != nil {
			t.Errorf("Ошибка: Создания диреткории %s", "./testdata/")
		}
	}
}

func TestAddToCache(t *testing.T) {
	configGlobalS.Cache.Timeout = 3600
	defer func() {
		configGlobalS.Cache.Timeout = 0
	}()
	// Инициализация кэша
	Mpw = MapCache{
		Map: make(map[int64]time.Time),
		Rw:  sync.RWMutex{},
	}

	// Добавление записи в кэш
	telegramID := int64(123456789)
	Mpw.add(telegramID)
	// Проверка, что запись была добавлена
	if !Mpw.check(telegramID) {
		t.Errorf("Ошибка: запись с telegramID %d не была добавлена в кэш", telegramID)
	}
}

func TestRemoveFromCache(t *testing.T) {
	// Инициализация кэша
	Mpw = MapCache{
		Map: make(map[int64]time.Time),
		Rw:  sync.RWMutex{},
	}

	// Добавление записи в кэш
	telegramID := int64(123456789)
	Mpw.add(telegramID)

	// Удаление записи из кэша
	Mpw.remove(telegramID)

	// Проверка, что запись была удалена
	if Mpw.check(telegramID) {
		t.Errorf("Ошибка: запись с telegramID %d не была удалена из кэша", telegramID)
	}
}

func TestCleanOldEntries(t *testing.T) {
	// Инициализация кэша
	Mpw = MapCache{
		Map: make(map[int64]time.Time),
		Rw:  sync.RWMutex{},
	}

	// Добавление записи в кэш с временем, которое уже истекло
	telegramID := int64(123456789)
	Mpw.Map[telegramID] = time.Now().Add(-10 * time.Minute)

	// Очистка старых записей
	cleanOldEntries()

	// Проверка, что запись была удалена из кэша
	if Mpw.check(telegramID) {
		t.Errorf("Ошибка: старая запись с telegramID %d не была удалена из кэша", telegramID)
	}
}

func TestCheckCache(t *testing.T) {
	configGlobalS.Cache.Timeout = 3600
	defer func() {
		configGlobalS.Cache.Timeout = 0
	}()
	// Инициализация кэша
	Mpw = MapCache{
		Map: make(map[int64]time.Time),
		Rw:  sync.RWMutex{},
	}

	// Добавление записи в кэш
	telegramID := int64(123456789)
	Mpw.add(telegramID)

	// Проверка, что запись присутствует в кэше
	if !Mpw.check(telegramID) {
		t.Errorf("Ошибка: запись с telegramID %d не найдена в кэше", telegramID)
	}
}
