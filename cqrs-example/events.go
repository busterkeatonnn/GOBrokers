package main

import (
	"time"
)

// Event интерфейс для всех событий
type Event interface {
	GetOrderID() int      // Получение ID заказа
	GetType() string      // Получение типа события
	GetTimestamp() string // Получение времени события
}

// BaseEvent базовая структура для всех событий
type BaseEvent struct {
	OrderID   int       // ID заказа
	Timestamp time.Time // Время события
}

// GetOrderID возвращает ID заказа
func (e BaseEvent) GetOrderID() int {
	return e.OrderID
}

// GetTimestamp возвращает время события в формате RFC3339
func (e BaseEvent) GetTimestamp() string {
	return e.Timestamp.Format(time.RFC3339)
}

// OrderCreatedEvent событие создания заказа
type OrderCreatedEvent struct {
	BaseEvent
	CustomerID string   // ID клиента
	Items      []string // Список товаров в заказе
}

// GetType возвращает тип события
func (e OrderCreatedEvent) GetType() string {
	return "OrderCreated"
}

// OrderPaidEvent событие оплаты заказа
type OrderPaidEvent struct {
	BaseEvent
}

// GetType возвращает тип события
func (e OrderPaidEvent) GetType() string {
	return "OrderPaid"
}

// OrderCancelledEvent событие отмены заказа
type OrderCancelledEvent struct {
	BaseEvent
	Reason string // Причина отмены
}

// GetType возвращает тип события
func (e OrderCancelledEvent) GetType() string {
	return "OrderCancelled"
}

// OrderState представляет текущее состояние заказа
type OrderState struct {
	ID         int       // ID заказа
	CustomerID string    // ID клиента
	Items      []string  // Товары
	Status     string    // Статус (created, paid, cancelled)
	CreateTime time.Time // Время создания
	UpdateTime time.Time // Время последнего обновления
}

// buildOrderState восстанавливает состояние заказа из списка событий
func buildOrderState(events []Event) *OrderState {
	if len(events) == 0 {
		return nil
	}

	// Начальное состояние
	state := &OrderState{
		ID:     events[0].GetOrderID(),
		Status: "unknown",
	}

	// Применяем все события последовательно
	for _, event := range events {
		applyEvent(state, event)
	}

	return state
}

// applyEvent применяет событие к состоянию заказа
func applyEvent(state *OrderState, event Event) {
	switch e := event.(type) {
	case OrderCreatedEvent:
		// Применяем событие создания
		state.CustomerID = e.CustomerID
		state.Items = e.Items
		state.Status = "created"
		state.CreateTime = e.Timestamp
		state.UpdateTime = e.Timestamp
	case OrderPaidEvent:
		// Применяем событие оплаты
		state.Status = "paid"
		state.UpdateTime = e.Timestamp
	case OrderCancelledEvent:
		// Применяем событие отмены
		state.Status = "cancelled"
		state.UpdateTime = e.Timestamp
	}
}
