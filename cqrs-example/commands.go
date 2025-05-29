//commands.go
package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// CreateOrderCommand команда для создания заказа
type CreateOrderCommand struct {
	CustomerID string   // ID клиента
	Items      []string // Список товаров
}

// PayOrderCommand команда для оплаты заказа
type PayOrderCommand struct {
	OrderID int // ID заказа
}

// CancelOrderCommand команда для отмены заказа
type CancelOrderCommand struct {
	OrderID int    // ID заказа
	Reason  string // Причина отмены
}

// HandleCreateOrder обрабатывает команду создания заказа
func HandleCreateOrder(store *EventStore, cmd CreateOrderCommand) (int, error) {
	// Валидация данных команды
	if cmd.CustomerID == "" {
		return 0, errors.New("ID клиента не может быть пустым")
	}
	if len(cmd.Items) == 0 {
		return 0, errors.New("заказ должен содержать хотя бы один товар")
	}

	// Генерация нового ID заказа
	orderID := store.NextOrderID()

	// Создание события
	event := OrderCreatedEvent{
		BaseEvent: BaseEvent{
			OrderID:   orderID,
			Timestamp: time.Now(),
		},
		CustomerID: cmd.CustomerID,
		Items:      cmd.Items,
	}

	// Сохранение события
	err := store.SaveEvent(event)
	if err != nil {
		return 0, fmt.Errorf("ошибка при сохранении события: %w", err)
	}

	log.Printf("Заказ #%d создан для клиента %s", orderID, cmd.CustomerID)
	return orderID, nil
}

// HandlePayOrder обрабатывает команду оплаты заказа
func HandlePayOrder(store *EventStore, cmd PayOrderCommand) error {
	// Получение событий для заказа
	events := store.GetEventsForOrder(cmd.OrderID)
	if len(events) == 0 {
		return errors.New("заказ не найден")
	}

	// Восстановление состояния заказа
	orderState := buildOrderState(events)

	// Проверка текущего состояния
	if orderState.Status != "created" {
		return fmt.Errorf("невозможно оплатить заказ в статусе %s", orderState.Status)
	}

	// Создание события оплаты
	event := OrderPaidEvent{
		BaseEvent: BaseEvent{
			OrderID:   cmd.OrderID,
			Timestamp: time.Now(),
		},
	}

	// Сохранение события
	err := store.SaveEvent(event)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении события: %w", err)
	}

	log.Printf("Заказ #%d оплачен", cmd.OrderID)
	return nil
}

// HandleCancelOrder обрабатывает команду отмены заказа
func HandleCancelOrder(store *EventStore, cmd CancelOrderCommand) error {
	// Получение событий для заказа
	events := store.GetEventsForOrder(cmd.OrderID)
	if len(events) == 0 {
		return errors.New("заказ не найден")
	}

	// Восстановление состояния заказа
	orderState := buildOrderState(events)

	// Проверка текущего состояния
	if orderState.Status == "cancelled" {
		return errors.New("заказ уже отменен")
	}
	if orderState.Status == "delivered" {
		return errors.New("невозможно отменить доставленный заказ")
	}

	// Создание события отмены
	event := OrderCancelledEvent{
		BaseEvent: BaseEvent{
			OrderID:   cmd.OrderID,
			Timestamp: time.Now(),
		},
		Reason: cmd.Reason,
	}

	// Сохранение события
	err := store.SaveEvent(event)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении события: %w", err)
	}

	log.Printf("Заказ #%d отменен по причине: %s", cmd.OrderID, cmd.Reason)
	return nil
}
