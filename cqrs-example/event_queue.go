// event_queue.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// EventQueue представляет очередь событий с возможностью их сохранения и загрузки
type EventQueue struct {
	events     []Event          // Сама очередь событий
	mu         sync.RWMutex     // Мьютекс для безопасного доступа
	logFile    string           // Путь к файлу для хранения событий
	subscribers []func(Event)   // Подписчики на новые события
}

// NewEventQueue создает новую очередь событий
func NewEventQueue(logFilePath string) (*EventQueue, error) {
	queue := &EventQueue{
		events:      make([]Event, 0),
		logFile:     logFilePath,
		subscribers: make([]func(Event), 0),
	}

	// Загружаем события из файла, если он существует
	err := queue.loadEventsFromLog()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("ошибка при загрузке событий из лога: %w", err)
	}

	return queue, nil
}

// Subscribe подписывает обработчик на новые события
func (q *EventQueue) Subscribe(handler func(Event)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.subscribers = append(q.subscribers, handler)
}

// Enqueue добавляет событие в очередь и записывает его в лог
func (q *EventQueue) Enqueue(event Event) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Добавляем событие в очередь
	q.events = append(q.events, event)

	// Сохраняем событие в лог
	err := q.appendEventToLog(event)
	if err != nil {
		return fmt.Errorf("ошибка при записи события в лог: %w", err)
	}

	// Уведомляем подписчиков о новом событии
	for _, handler := range q.subscribers {
		go handler(event)
	}

	return nil
}

// GetAll возвращает все события из очереди
func (q *EventQueue) GetAll() []Event {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Создаем копию слайса событий
	result := make([]Event, len(q.events))
	copy(result, q.events)

	return result
}

// GetByOrderID возвращает все события для указанного заказа
func (q *EventQueue) GetByOrderID(orderID int) []Event {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]Event, 0)
	for _, event := range q.events {
		if event.GetOrderID() == orderID {
			result = append(result, event)
		}
	}

	return result
}

// Сериализация событий для хранения

// EventDTO структура для сериализации событий
type EventDTO struct {
	Type      string          `json:"type"`
	OrderID   int             `json:"order_id"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// appendEventToLog сохраняет событие в лог-файл
func (q *EventQueue) appendEventToLog(event Event) error {
	// Открываем файл для добавления
	file, err := os.OpenFile(q.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Сериализуем событие в зависимости от его типа
	var data []byte
	var eventType string

	switch e := event.(type) {
	case OrderCreatedEvent:
		eventType = "OrderCreated"
		data, err = json.Marshal(e)
	case OrderPaidEvent:
		eventType = "OrderPaid"
		data, err = json.Marshal(e)
	case OrderCancelledEvent:
		eventType = "OrderCancelled"
		data, err = json.Marshal(e)
	default:
		return fmt.Errorf("неизвестный тип события: %T", e)
	}

	if err != nil {
		return err
	}

	// Создаем DTO для сохранения
	dto := EventDTO{
		Type:      eventType,
		OrderID:   event.GetOrderID(),
		Timestamp: event.GetTimestamp(),
		Data:      data,
	}

	// Сериализуем DTO в JSON
	jsonData, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	// Записываем событие в файл
	_, err = file.Write(append(jsonData, '\n'))
	return err
}

// loadEventsFromLog загружает события из лог-файла
func (q *EventQueue) loadEventsFromLog() error {
	// Проверяем существование файла
	_, err := os.Stat(q.logFile)
	if os.IsNotExist(err) {
		return err // Файл не существует, это не ошибка
	}

	// Открываем файл для чтения
	file, err := os.Open(q.logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Читаем файл построчно
	decoder := json.NewDecoder(file)

	q.events = make([]Event, 0)

	for decoder.More() {
		var dto EventDTO
		if err := decoder.Decode(&dto); err != nil {
			return err
		}

		// Десериализуем событие в зависимости от его типа
		var event Event

		switch dto.Type {
		case "OrderCreated":
			var e OrderCreatedEvent
			if err := json.Unmarshal(dto.Data, &e); err != nil {
				return err
			}
			// Восстанавливаем время из строки
			t, err := time.Parse(time.RFC3339, dto.Timestamp)
			if err != nil {
				return err
			}
			e.Timestamp = t
			event = e
		case "OrderPaid":
			var e OrderPaidEvent
			if err := json.Unmarshal(dto.Data, &e); err != nil {
				return err
			}
			t, err := time.Parse(time.RFC3339, dto.Timestamp)
			if err != nil {
				return err
			}
			e.Timestamp = t
			event = e
		case "OrderCancelled":
			var e OrderCancelledEvent
			if err := json.Unmarshal(dto.Data, &e); err != nil {
				return err
			}
			t, err := time.Parse(time.RFC3339, dto.Timestamp)
			if err != nil {
				return err
			}
			e.Timestamp = t
			event = e
		default:
			return fmt.Errorf("неизвестный тип события: %s", dto.Type)
		}

		q.events = append(q.events, event)
	}

	return nil
}
