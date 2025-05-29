package main

import (
	"fmt"
	"log"
	"os"
	
)

func main() {
	// Проверяем, что указан режим работы
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run main.go [publisher|consumer|nats]")
		os.Exit(1)
	}

	// Получаем режим работы из аргументов командной строки
	mode := os.Args[1]

	// Выбираем соответствующую функцию в зависимости от режима
	switch mode {
	case "publisher":
		// Запускаем издателя сообщений для RabbitMQ
		runPublisher()
	case "consumer":
		// Запускаем потребителя сообщений для RabbitMQ
		runConsumer()
	case "nats":
		// Запускаем пример работы с NATS
		runNatsExample()
	default:
		// Если указан неизвестный режим, выводим подсказку
		fmt.Println("Использование: go run main.go [publisher|consumer|nats]")
		log.Println("publisher - запускает отправку задач в очередь RabbitMQ")
		log.Println("consumer - запускает обработку задач из очереди RabbitMQ")
		log.Println("nats - запускает пример работы с NATS (публикация и подписка)")
		os.Exit(1)
	}
}
