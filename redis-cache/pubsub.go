package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v8"
)

// Данная функция запускает пример Redis Pub/Sub
// Запуск: go run pubsub.go publisher - для издателя
// Запуск: go run pubsub.go subscriber - для подписчика
func main() {
	// Проверяем, передан ли аргумент командной строки
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run pubsub.go [publisher|subscriber]")
		os.Exit(1)
	}

	// Инициализируем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем подключение к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Адрес Redis сервера
		Password: "",               // Пароль (если требуется)
		DB:       0,                // Используемая база данных Redis
	})

	// Проверяем соединение
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	log.Println("Успешное подключение к Redis")

	// Настраиваем обработку сигнала прерывания для корректного завершения
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Определяем режим работы: издатель или подписчик
	mode := os.Args[1]

	switch mode {
	case "publisher":
		// Запускаем издателя в отдельной горутине
		go func() {
			counter := 0
			for {
				select {
				case <-ctx.Done():
					// Завершаем работу при отмене контекста
					return
				default:
					counter++
					// Формируем сообщение с порядковым номером и временем
					message := fmt.Sprintf("Сообщение #%d: %s", counter, "Приветствую!!!!")

					// Публикуем сообщение в канал notifications
					err := rdb.Publish(ctx, "notifications", message).Err()
					if err != nil {
						log.Printf("Ошибка при публикации: %v", err)
					} else {
						log.Printf("Опубликовано: %s", message)
					}

					// Ждем 2 секунды перед отправкой следующего сообщения
					time.Sleep(2 * time.Second)
				}
			}
		}()

		log.Println("Издатель запущен. Отправляем сообщения в канал 'notifications'. Для выхода нажмите Ctrl+C.")

	case "subscriber":
		// Подписываемся на канал notifications
		pubsub := rdb.Subscribe(ctx, "notifications")
		defer pubsub.Close()

		// Получаем канал сообщений
		channel := pubsub.Channel()

		// Запускаем обработку сообщений в отдельной горутине
		go func() {
			for {
				select {
				case msg := <-channel:
					// Выводим полученное сообщение
					log.Printf("Получено сообщение из канала %s: %s", msg.Channel, msg.Payload)
				case <-ctx.Done():
					// Завершаем работу при отмене контекста
					return
				}
			}
		}()

		log.Println("Подписчик запущен. Слушаем канал 'notifications'. Для выхода нажмите Ctrl+C.")
	default:
		fmt.Println("Использование: go run pubsub.go [publisher|subscriber]")
		os.Exit(1)
	}

	// Ожидаем сигнала прерывания
	<-c
	log.Println("Завершение работы...")
}
