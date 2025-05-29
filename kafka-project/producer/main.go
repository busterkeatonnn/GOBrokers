package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Order представляет структуру заказа
type Order struct {
	ID        string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Настройка подключения к Kafka
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders",
		// Настройка балансировки сообщений (желательно для продакшена)
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	log.Println("Kafka Producer запущен. Ctrl+C для выхода.")

	// Бесконечный цикл для генерации и отправки заказов
	for i := 1; ; i++ {
		// Создаем новый заказ
		order := Order{
			ID:        fmt.Sprintf("order-%d", i),
			CustomerID: fmt.Sprintf("customer-%d", i%10+1),
			Amount:    float64(i*10) + 0.99,
			Status:    "created",
			CreatedAt: time.Now(),
		}

		// Преобразуем заказ в JSON
		orderJSON, err := json.Marshal(order)
		if err != nil {
			log.Printf("Ошибка сериализации заказа: %v", err)
			continue
		}

		// Создаем сообщение с ключом (ID заказа) и значением (JSON заказа)
		message := kafka.Message{
			Key:   []byte(order.ID),
			Value: orderJSON,
			Time:  time.Now(),
			Headers: []kafka.Header{
				{
					Key:   "event_type",
					Value: []byte("order.created"),
				},
			},
		}

		// Отправляем сообщение в Kafka
		err = writer.WriteMessages(context.Background(), message)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
			continue
		}

		log.Printf("Отправлен заказ: %s, сумма: %.2f", order.ID, order.Amount)

		// Пауза между отправками сообщений
		time.Sleep(2 * time.Second)
	}
}
