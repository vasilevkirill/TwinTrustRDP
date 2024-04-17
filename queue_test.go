package main

import (
	"sync"
	"testing"
)

func TestQueue_AddKey(t *testing.T) {
	// Инициализация тестовой очереди
	qu := queue{
		rw:  sync.RWMutex{},
		Map: make(map[int64]queueMsg),
	}

	// Добавление ключа в очередь
	key := int64(123)
	qu.AddKey(key)

	// Проверка наличия ключа в очереди
	if !qu.IssetKey(key) {
		t.Errorf("AddKey() failed to add key %d to the queue", key)
	}
}

func TestQueue_RemoveKey(t *testing.T) {
	// Инициализация тестовой очереди
	qu := queue{
		rw:  sync.RWMutex{},
		Map: make(map[int64]queueMsg),
	}

	// Добавление ключа в очередь
	key := int64(123)
	qu.AddKey(key)

	// Удаление ключа из очереди
	qu.RemoveKey(key)

	// Проверка отсутствия ключа в очереди
	if qu.IssetKey(key) {
		t.Errorf("RemoveKey() failed to remove key %d from the queue", key)
	}
}

func TestQueue_GetMsg(t *testing.T) {
	// Инициализация тестовой очереди
	qu := queue{
		rw:  sync.RWMutex{},
		Map: make(map[int64]queueMsg),
	}

	// Добавление сообщения с ключом в очередь
	key := int64(123)
	msg := queueMsg{Chan: make(chan int), MsgId: 456}
	qu.Map[key] = msg

	// Получение сообщения по ключу из очереди
	retrievedMsg := qu.GetMsg(key)

	// Проверка корректности полученного сообщения
	if retrievedMsg != msg {
		t.Errorf("GetMsg() returned incorrect message, got %v, want %v", retrievedMsg, msg)
	}
}

func TestQueue_SetMsgId(t *testing.T) {
	// Инициализация тестовой очереди
	qu := queue{
		rw:  sync.RWMutex{},
		Map: make(map[int64]queueMsg),
	}

	// Добавление сообщения с ключом в очередь
	key := int64(123)
	msg := queueMsg{Chan: make(chan int), MsgId: 0}
	qu.Map[key] = msg

	// Установка идентификатора сообщения для ключа
	msgId := int64(456)
	qu.SetMsgId(key, msgId)

	// Получение сообщения по ключу из очереди
	retrievedMsg := qu.Map[key]

	// Проверка корректности установленного идентификатора сообщения
	if retrievedMsg.MsgId != msgId {
		t.Errorf("SetMsgId() failed to set message ID, got %d, want %d", retrievedMsg.MsgId, msgId)
	}
}
