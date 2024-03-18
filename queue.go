package main

import "sync"

// Определение структуры очереди
var qu queue

type queue struct {
	rw  sync.RWMutex       // RWMutex для безопасного доступа к данным
	Map map[int64]queueMsg // Карта для хранения сообщений в очереди
}

type queueMsg struct {
	Chan  chan int // Канал для сообщений
	MsgId int64    // Идентификатор сообщения
}

// AddKey Метод добавления ключа в очередь
func (q *queue) AddKey(key int64) {
	q.rw.Lock()               // Блокировка для записи
	defer q.rw.Unlock()       // Обеспечение разблокировки после завершения операции
	msg := queueMsg{}         // Создание нового сообщения
	msg.Chan = make(chan int) // Инициализация канала для сообщений
	q.Map[key] = msg          // Добавление сообщения в карту
}

// IssetKey Метод проверки наличия ключа в очереди
func (q *queue) IssetKey(key int64) bool {
	q.rw.Lock()         // Блокировка для чтения
	defer q.rw.Unlock() // Обеспечение разблокировки после завершения операции
	_, ok := q.Map[key] // Проверка наличия ключа в карте
	return ok           // Возврат результата проверки
}

// RemoveKey Метод удаления ключа из очереди
func (q *queue) RemoveKey(key int64) {
	q.rw.Lock()         // Блокировка для записи
	defer q.rw.Unlock() // Обеспечение разблокировки после завершения операции
	delete(q.Map, key)  // Удаление ключа из карты
	return              // Возврат из метода
}

// GetMsg Метод получения сообщения из очереди по ключу
func (q *queue) GetMsg(key int64) queueMsg {
	q.rw.Lock()         // Блокировка для чтения
	defer q.rw.Unlock() // Обеспечение разблокировки после завершения операции
	return q.Map[key]   // Возврат сообщения из карты по ключу
}

// SetMsgId Метод установки идентификатора сообщения для заданного ключа
func (q *queue) SetMsgId(key, msgid int64) {
	q.rw.Lock()         // Блокировка для записи
	defer q.rw.Unlock() // Обеспечение разблокировки после завершения операции
	_, ok := q.Map[key] // Проверка наличия ключа в карте
	if !ok {            // Если ключ отсутствует, выходим из метода
		return
	}
	msg := q.Map[key] // Получаем сообщение из карты
	msg.MsgId = msgid // Устанавливаем идентификатор сообщения
	q.Map[key] = msg  // Обновляем сообщение в карте
}

// Инициализация очередей
func initQ() {
	qu.Map = make(map[int64]queueMsg) // Инициализация карты для хранения сообщений в очереди
}
