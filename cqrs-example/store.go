// store.go
package main

import (
	"log"
	"sync"
)

// EventStore хранилище событий
type EventStore struct {
	queue    *EventQueue     // Очередь событий
	orderID  int             // Счетчик ID заказов
	mu       sync.RWMutex    // Мьютекс для безопасного доступа
}

// NewEventStore создает новое хранилище событий
func NewEventStore(logFilePath string) (*EventStore, error) {
	// Создаем очередь событий
	queue, err := NewEventQueue(logFilePath)
	if err != nil {
		return nil, err
	}

	store := &EventStore{
		queue:   queue,
		orderID: 0,
	}

	// Определяем максимальный orderID из загруженных событий
	events := queue.GetAll()
	for _, event := range events {
		if event.GetOrderID() > store.orderID {
			store.orderID = event.GetOrderID()
		}
	}

	return store, nil
}

// NextOrderID генерирует следующий ID заказа
func (s *EventStore) NextOrderID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orderID++
	return s.orderID
}

// SaveEvent сохраняет событие в хранилище
func (s *EventStore) SaveEvent(event Event) error {
	// Добавляем событие в очередь
	err := s.queue.Enqueue(event)
	if err != nil {
		return err
	}

	log.Printf("Событие сохранено: %s для заказа #%d", event.GetType(), event.GetOrderID())
	return nil
}

// GetEventsForOrder возвращает все события для указанного заказа
func (s *EventStore) GetEventsForOrder(orderID int) []Event {
	return s.queue.GetByOrderID(orderID)
}

// GetAllEvents возвращает все события
func (s *EventStore) GetAllEvents() []Event {
	return s.queue.GetAll()
}
