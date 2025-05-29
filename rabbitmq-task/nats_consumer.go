// nats_consumer.go
// ДЛЯ PR13
package main

import (
    "log"
    "os"
    "os/signal"
    "time"

    "github.com/nats-io/nats.go"
)

// runNatsConsumer запускает обработчик задач из NATS
func runNatsConsumer() {
    // Получаем ID обработчика из переменной окружения или используем по умолчанию
    workerID := os.Getenv("WORKER_ID")
    if workerID == "" {
        workerID = "worker-default"
    }

    // Подключаемся к серверу NATS
    nc, err := nats.Connect(nats.DefaultURL) // localhost:4222
    if err != nil {
        log.Fatalf("Ошибка подключения к NATS: %v", err)
    }
    defer nc.Close()

    // Настраиваем обработку сигнала прерывания
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    // Создаем подписку на тему jobs.create с queue группой workers
    // Благодаря queue группе, сообщения будут распределяться между всеми подписчиками в группе
    sub, err := nc.QueueSubscribe("jobs.create", "workers", func(msg *nats.Msg) {
        task := string(msg.Data)
        log.Printf("[%s] Получена задача: %s", workerID, task)

        // Имитируем обработку задачи
        log.Printf("[%s] Обработка задачи: %s", workerID, task)
        time.Sleep(2 * time.Second)

        log.Printf("[%s] Задача выполнена: %s", workerID, task)
    })

    if err != nil {
        log.Fatalf("Ошибка при создании подписки: %v", err)
    }
    defer sub.Unsubscribe()

    log.Printf("[%s] Consumer запущен. Ожидание задач из темы 'jobs.create'. Для выхода нажмите Ctrl+C.", workerID)
    <-c
    log.Println("Завершение работы Consumer...")
}
