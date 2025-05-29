package main

import (
	"log"
	"sync"
)

// EventStore хранилище событий
type EventStore struct {
	events  []Event         // Список всех событий
	orderID int             // Счетчик ID заказов
	mu      sync.RWMutex    // Мьютекс для безопасного доступа
}

// NewEventStore создает новое хранилище событий
func NewEventStore() *EventStore {
	return &EventStore{
		events:  make([]Event, 0),
		orderID: 0,
	}
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
	s.mu.Lock()
	defer s.mu.Unlock()

	// Добавляем событие в список
	s.events = append(s.events, event)

	log.Printf("Событие сохранено: %s для заказа #%d", event.GetType(), event.GetOrderID())
	return nil
}

// GetEventsForOrder возвращает все события для указанного заказа
func (s *EventStore) GetEventsForOrder(orderID int) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Event, 0)

	// Находим все события для заказа
	for _, event := range s.events {
		if event.GetOrderID() == orderID {
			result = append(result, event)
		}
	}

	return result
}

// GetAllEvents возвращает все события
func (s *EventStore) GetAllEvents() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Копируем список событий
	result := make([]Event, len(s.events))
	copy(result, s.events)

	return result
}
