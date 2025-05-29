// consumer.go
package main

import (
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "time"
    "errors"

    "github.com/streadway/amqp"
)

// simulateEmailSending имитирует отправку email и может возвращать ошибку
// для тестирования механизма подтверждений
func simulateEmailSending(task EmailTask) error {
    // Для демонстрационных целей - считаем, что адреса с "error" в них вызывают ошибку
    if contains(task.To, "error") {
        return errors.New("ошибка при отправке email")
    }

    // Имитация работы
    log.Printf("Отправка email to: %s, subject: %s", task.To, task.Subject)
    time.Sleep(2 * time.Second)
    return nil
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
    return len(s) >= len(substr) && s[0:len(substr)] == substr
}

// runConsumer запускает обработчик задач из очереди RabbitMQ
func runConsumer() {
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

    // Объявляем очередь (должна совпадать с очередью в издателе)
    q, err := ch.QueueDeclare(
        "email_tasks", // имя очереди
        true,         // durable - сохранение очереди после перезапуска сервера
        false,        // auto-delete - удаление очереди, когда нет потребителей
        false,        // exclusive - доступна только для текущего соединения
        false,        // no-wait - не ждать ответа от сервера
        nil,          // аргументы
    )
    if err != nil {
        log.Fatalf("Ошибка объявления очереди: %v", err)
    }

    // Настраиваем Qos (качество обслуживания)
    err = ch.Qos(
        1,      // prefetch count - количество сообщений, которые будут получены за раз
        0,      // prefetch size - размер сообщений (0 = без ограничения)
        false,  // global - применять ко всем потребителям на этом соединении
    )
    if err != nil {
        log.Fatalf("Ошибка установки Qos: %v", err)
    }

    // Регистрируем потребителя сообщений
    msgs, err := ch.Consume(
        q.Name, // имя очереди
        "",     // consumer - пустая строка означает генерацию уникального имени
        false,  // auto-ack - автоматическое подтверждение обработки (отключено для manual ack)
        false,  // exclusive - доступна только для текущего соединения
        false,  // no-local - не получать сообщения, опубликованные этим соединением
        false,  // no-wait - не ждать ответа от сервера
        nil,    // аргументы
    )
    if err != nil {
        log.Fatalf("Ошибка регистрации потребителя: %v", err)
    }

    // Настраиваем обработку сигнала прерывания
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    // Запускаем обработку сообщений
    go func() {
        for d := range msgs {
            // Десериализуем задачу из JSON
            var task EmailTask
            err := json.Unmarshal(d.Body, &task)
            if err != nil {
                log.Printf("Ошибка десериализации: %v", err)
                // Отклоняем сообщение с ошибкой десериализации,
                // оно будет возвращено в очередь для повторной обработки
                d.Nack(false, true) // false - не групповое подтверждение, true - вернуть в очередь
                continue
            }

            log.Printf("Получена задача на отправку email: %s", task.Subject)

            // Имитация отправки email с возможной ошибкой
            err = simulateEmailSending(task)
            if err != nil {
                log.Printf("Ошибка отправки email: %v", err)
                // В случае ошибки не подтверждаем сообщение,
                // оно будет возвращено в очередь для повторной обработки
                d.Nack(false, true) // false - не групповое подтверждение, true - вернуть в очередь
                continue
            }

            // Логируем успешную отправку
            log.Printf("Email успешно отправлен: %s", task.To)

            // Подтверждаем успешную обработку сообщения
            err = d.Ack(false) // false - не групповое подтверждение
            if err != nil {
                log.Printf("Ошибка подтверждения сообщения: %v", err)
            }
        }
    }()

    log.Println("Потребитель запущен. Ожидание задач из очереди 'email_tasks'. Для выхода нажмите Ctrl+C.")
    <-c
    log.Println("Завершение работы потребителя...")
}
