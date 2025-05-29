//nats_example.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
)

// runNatsExample демонстрирует работу с NATS
func runNatsExample() {
	// Подключаемся к серверу NATS
	nc, err := nats.Connect(nats.DefaultURL) // localhost:4222
	if err != nil {
		log.Fatalf("Ошибка подключения к NATS: %v", err)
	}
	defer nc.Close()

	// Настраиваем обработку сигнала прерывания
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Создаем подписку на тему "events"
	sub, err := nc.Subscribe("events", func(msg *nats.Msg) {
		// Обработчик входящих сообщений
		log.Printf("Получено сообщение: %s", string(msg.Data))

		// Если указан канал для ответа, отправляем подтверждение
		if msg.Reply != "" {
			nc.Publish(msg.Reply, []byte("Подтверждение получения"))
		}
	})
	if err != nil {
		log.Fatalf("Ошибка при создании подписки: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Подписка на тему 'events' создана")

	// Запускаем отправку сообщений в отдельной горутине
	go func() {
		counter := 0
		for {
			counter++
			message := fmt.Sprintf("Событие #%d произошло в %s", counter, time.Now().Format(time.RFC3339))

			// Отправляем сообщение и ожидаем ответа (request-reply паттерн)
			reply, err := nc.Request("events", []byte(message), 1*time.Second)
			if err != nil {
				log.Printf("Ошибка при отправке запроса: %v", err)
			} else {
				log.Printf("Опубликовано: %s, Ответ: %s", message, string(reply.Data))
			}

			// Ждем 2 секунды перед отправкой следующего сообщения
			time.Sleep(2 * time.Second)
		}
	}()

	log.Println("NATS пример запущен. Для выхода нажмите Ctrl+C.")
	<-c
	log.Println("Завершение работы...")
}
