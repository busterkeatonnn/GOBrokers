package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// Order представляет структуру заказа (аналогично Producer)
type Order struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
}

func main() {
	// Настройка подключения к Kafka
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "orders",
		GroupID:   "order-service", // Группа потребителей
		MinBytes:  10e3,            // Минимальный размер батча (10KB)
		MaxBytes:  10e6,            // Максимальный размер батча (10MB)
		Partition: 0,
	})
	defer reader.Close()

	log.Println("Kafka Consumer запущен. Ожидание сообщений...")

	// Бесконечный цикл для чтения сообщений
	for {
		// Читаем сообщение из Kafka
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v", err)
			continue
		}

		// Обрабатываем заголовки
		var eventType string
		for _, header := range message.Headers {
			if header.Key == "event_type" {
				eventType = string(header.Value)
			}
		}

		// Проверяем, что это событие создания заказа
		if eventType == "order.created" {
			var order Order
			err = json.Unmarshal(message.Value, &order)
			if err != nil {
				log.Printf("Ошибка десериализации заказа: %v", err)
				continue
			}

			// Печатаем информацию о полученном заказе
			log.Printf("Получено событие order.created:")
			log.Printf("  ID заказа: %s", order.ID)
			log.Printf("  ID клиента: %s", order.CustomerID)
			log.Printf("  Сумма: %.2f", order.Amount)
			log.Printf("  Статус: %s", order.Status)
			log.Printf("  Дата создания: %s", order.CreatedAt)
			fmt.Println() // Пустая строка для разделения
		} else {
			log.Printf("Получено неизвестное событие: %s", eventType)
		}
	}
}
