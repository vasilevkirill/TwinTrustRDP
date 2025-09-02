package main

import "sync"

// Глобальная очередь
var qu = queue{
	Map: make(map[int64]queueMsg),
}

// queue — структура очереди
type queue struct {
	rw  sync.RWMutex       // RWMutex для безопасного доступа
	Map map[int64]queueMsg // Карта для хранения сообщений
}

// queueMsg — элемент очереди
type queueMsg struct {
	Chan  chan int // Канал для передачи сигналов
	MsgId int64    // Идентификатор сообщения
}

// AddKey — добавляет новый ключ в очередь
func (q *queue) AddKey(key int64) {
	q.rw.Lock()
	defer q.rw.Unlock()

	msg := queueMsg{
		Chan: make(chan int, 1), // Буферизированный канал
	}
	q.Map[key] = msg
}

// IssetKey — проверяет наличие ключа в очереди
func (q *queue) IssetKey(key int64) bool {
	q.rw.RLock()
	defer q.rw.RUnlock()
	_, ok := q.Map[key]
	return ok
}

// RemoveKey — удаляет ключ из очереди и закрывает канал
func (q *queue) RemoveKey(key int64) {
	q.rw.Lock()
	defer q.rw.Unlock()

	if msg, ok := q.Map[key]; ok {
		close(msg.Chan) // Закрываем канал, чтобы горутины не зависли
		delete(q.Map, key)
	}
}

// GetMsg — возвращает сообщение по ключу
func (q *queue) GetMsg(key int64) (queueMsg, bool) {
	q.rw.RLock()
	defer q.rw.RUnlock()
	msg, ok := q.Map[key]
	return msg, ok
}

// SetMsgId — устанавливает идентификатор сообщения
func (q *queue) SetMsgId(key, msgid int64) {
	q.rw.Lock()
	defer q.rw.Unlock()

	if msg, ok := q.Map[key]; ok {
		msg.MsgId = msgid
		q.Map[key] = msg
	}
}
