package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/streadway/amqp"
)

// EmailTask представляет задачу на отправку email
type EmailTask struct {
	To      string    `json:"to"`      // Адрес получателя
	Subject string    `json:"subject"` // Тема письма
	Body    string    `json:"body"`    // Текст письма
	Created time.Time `json:"created"` // Время создания задачи
}

// runPublisher запускает процесс публикации задач в очередь RabbitMQ
func runPublisher() {
	// Подключаемся к серверу RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Ошибка подключения к RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Открываем канал для работы с RabbitMQ
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Ошибка открытия канала: %v", err)
	}
	defer ch.Close()

	// Объявляем очередь для задач на отправку email
	q, err := ch.QueueDeclare(
		"email_tasks", // имя очереди
		true,          // durable - сохранение очереди после перезапуска сервера
		false,         // auto-delete - удаление очереди, когда нет потребителей
		false,         // exclusive - доступна только для текущего соединения
		false,         // no-wait - не ждать ответа от сервера
		nil,           // аргументы
	)
	if err != nil {
		log.Fatalf("Ошибка объявления очереди: %v", err)
	}

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настраиваем обработку сигнала прерывания
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Запускаем отправку задач в очередь
	go func() {
		counter := 0
		for {
			select {
			case <-ctx.Done():
				// Завершаем работу при отмене контекста
				return
			default:
				counter++

				// Создаем новую задачу на отправку email
				task := EmailTask{
					To:      fmt.Sprintf("user%d@example.com", counter),
					Subject: fmt.Sprintf("Важное сообщение #%d", counter),
					Body:    fmt.Sprintf("Это тестовое сообщение #%d, отправленное %s", counter, time.Now().Format(time.RFC3339)),
					Created: time.Now(),
				}

				// Сериализуем задачу в JSON
				body, err := json.Marshal(task)
				if err != nil {
					log.Printf("Ошибка сериализации: %v", err)
					continue
				}

				// Публикуем сообщение в очередь
				err = ch.Publish(
					"",     // exchange - используем exchange по умолчанию
					q.Name, // routing key - имя очереди
					false,  // mandatory - сообщение не должно быть возвращено, если не может быть доставлено
					false,  // immediate - сообщение не должно быть возвращено, если нет активных потребителей
					amqp.Publishing{
						DeliveryMode: amqp.Persistent, // сохраняем сообщения на диск
						ContentType:  "application/json",
						Body:         body,
					})
				if err != nil {
					log.Printf("Ошибка публикации: %v", err)
					continue
				}

				log.Printf("Задача на отправку email добавлена в очередь: %s", task.Subject)

				// Ждем 1 секунду перед отправкой следующей задачи
				time.Sleep(1 * time.Second)
			}
		}
	}()

	log.Println("Издатель запущен. Отправляем задачи в очередь 'email_tasks'. Для выхода нажмите Ctrl+C.")
	<-c
	log.Println("Завершение работы издателя...")
}
