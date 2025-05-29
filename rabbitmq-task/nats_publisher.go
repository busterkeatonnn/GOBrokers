// nats_publisher.go
// ДЛЯ PR13
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "time"

    "github.com/nats-io/nats.go"
)

// runNatsPublisher запускает издателя задач через NATS
func runNatsPublisher() {
    // Подключаемся к серверу NATS
    nc, err := nats.Connect(nats.DefaultURL) // localhost:4222
    if err != nil {
        log.Fatalf("Ошибка подключения к NATS: %v", err)
    }
    defer nc.Close()

    // Настраиваем обработку сигнала прерывания
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    // Запускаем отправку задач в отдельной горутине
    go func() {
        counter := 0
        for {
            counter++
            task := fmt.Sprintf("Task %d", counter)

            // Публикуем задачу в тему jobs.create
            err := nc.Publish("jobs.create", []byte(task))
            if err != nil {
                log.Printf("Ошибка при публикации задачи: %v", err)
            } else {
                log.Printf("Опубликована задача: %s", task)
            }

            // Ждем 1 секунду перед отправкой следующей задачи
            time.Sleep(1 * time.Second)
        }
    }()

    log.Println("NATS Publisher запущен. Отправляем задачи в тему 'jobs.create'. Для выхода нажмите Ctrl+C.")
    <-c
    log.Println("Завершение работы Publisher...")
}
