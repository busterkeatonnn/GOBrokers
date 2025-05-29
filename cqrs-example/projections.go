// projections.go
package main

import (
	"sync"
)

// OrderProjection проекция для заказов
type OrderProjection struct {
	store    *EventStore             // Хранилище событий
	orders   map[int]*OrderState     // Кэш состояний заказов
	mu       sync.RWMutex            // Мьютекс для безопасного доступа
}

// NewOrderProjection создает новую проекцию заказов
func NewOrderProjection(store *EventStore) *OrderProjection {
	projection := &OrderProjection{
		store:  store,
		orders: make(map[int]*OrderState),
	}

	// Обрабатываем все существующие события
	projection.rebuildProjection()

	// Подписываемся на новые события
	store.queue.Subscribe(func(event Event) {
		projection.UpdateProjection(event)
	})

	return projection
}

// rebuildProjection пересоздает проекцию на основе всех событий
func (p *OrderProjection) rebuildProjection() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Очищаем текущее состояние
	p.orders = make(map[int]*OrderState)

	// Получаем все события
	events := p.store.GetAllEvents()

	// Группируем события по заказам
	orderEvents := make(map[int][]Event)
	for _, event := range events {
		orderID := event.GetOrderID()
		orderEvents[orderID] = append(orderEvents[orderID], event)
	}

	// Строим состояние для каждого заказа
	for orderID, events := range orderEvents {
		p.orders[orderID] = buildOrderState(events)
	}
}

// GetOrder возвращает состояние заказа по ID
func (p *OrderProjection) GetOrder(orderID int) *OrderState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.orders[orderID]
}

// GetAllOrders возвращает все заказы
func (p *OrderProjection) GetAllOrders() []*OrderState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*OrderState, 0, len(p.orders))
	for _, order := range p.orders {
		result = append(result, order)
	}

	return result
}

// UpdateProjection обновляет проекцию на основе новых событий
func (p *OrderProjection) UpdateProjection(event Event) {
	p.mu.Lock()
	defer p.mu.Unlock()

	orderID := event.GetOrderID()

	// Получаем текущее состояние заказа или создаем новое
	var state *OrderState
	existingState, found := p.orders[orderID]
	if found {
		state = existingState
	} else {
		state = &OrderState{
			ID:     orderID,
			Status: "unknown",
		}
	}

	// Применяем событие к состоянию
	applyEvent(state, event)

	// Сохраняем обновленное состояние
	p.orders[orderID] = state
}
