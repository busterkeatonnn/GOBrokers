// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"path/filepath"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Путь к файлу событий
	eventLogPath := filepath.Join("data", "event_log.json")

	// Создаем директорию для данных, если она не существует
	os.MkdirAll(filepath.Dir(eventLogPath), 0755)

	// Инициализируем хранилище событий
	store, err := NewEventStore(eventLogPath)
	if err != nil {
		log.Fatalf("Ошибка при инициализации хранилища событий: %v", err)
	}

	// Создаем проекцию заказов
	orderProjection := NewOrderProjection(store)

	// Настраиваем HTTP сервер
	r := mux.NewRouter()

	// Маршруты для команд (изменение состояния)
	r.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		// Обработка команды CreateOrder

		// Получаем параметры из формы запроса
		customerID := r.FormValue("customer_id")
		item := r.FormValue("item")

		// Создаем команду
		command := CreateOrderCommand{
			CustomerID: customerID,
			Items:     []string{item},
		}

		// Обрабатываем команду
		orderID, err := HandleCreateOrder(store, command)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Возвращаем ответ
		fmt.Fprintf(w, "Заказ создан, ID: %d", orderID)
	}).Methods("POST")

	r.HandleFunc("/orders/{id}/pay", func(w http.ResponseWriter, r *http.Request) {
		// Обработка команды PayOrder

		// Получаем ID заказа из URL
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Некорректный ID", http.StatusBadRequest)
			return
		}

		// Создаем команду
		command := PayOrderCommand{OrderID: id}

		// Обрабатываем команду
		err = HandlePayOrder(store, command)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Возвращаем ответ
		fmt.Fprint(w, "Заказ оплачен")
	}).Methods("POST")

	r.HandleFunc("/orders/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		// Обработка команды CancelOrder

		// Получаем ID заказа из URL
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Некорректный ID", http.StatusBadRequest)
			return
		}

		// Получаем причину отмены из формы
		reason := r.FormValue("reason")
		if reason == "" {
			reason = "Причина не указана"
		}

		// Создаем команду
		command := CancelOrderCommand{
			OrderID: id,
			Reason:  reason,
		}

		// Обрабатываем команду
		err = HandleCancelOrder(store, command)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Возвращаем ответ
		fmt.Fprint(w, "Заказ отменен")
	}).Methods("POST")

	// Маршруты для запросов (чтение состояния)
	r.HandleFunc("/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Получение данных заказа

		// Получаем ID заказа из URL
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Некорректный ID", http.StatusBadRequest)
			return
		}

		// Получаем заказ из проекции
		order := orderProjection.GetOrder(id)
		if order == nil {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		// Формируем ответ в текстовом формате
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Заказ #%d\n", order.ID)
		fmt.Fprintf(w, "Клиент: %s\n", order.CustomerID)
		fmt.Fprintf(w, "Статус: %s\n", order.Status)
		fmt.Fprintf(w, "Товары: %v\n", order.Items)
		fmt.Fprintf(w, "Создан: %v\n", order.CreateTime)
		fmt.Fprintf(w, "Обновлен: %v\n", order.UpdateTime)
	}).Methods("GET")

	r.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		// Получение списка всех заказов

		// Получаем заказы из проекции
		orders := orderProjection.GetAllOrders()

		// Формируем ответ в текстовом формате
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Список заказов:")

		for _, order := range orders {
			fmt.Fprintf(w, "Заказ #%d - Клиент: %s, Статус: %s, Товары: %v\n",
				order.ID, order.CustomerID, order.Status, order.Items)
		}
	}).Methods("GET")

	// Добавляем маршрут для просмотра лога событий
	r.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Получение всех событий
		events := store.GetAllEvents()

		// Формируем ответ в текстовом формате
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Лог событий:")

		for i, event := range events {
			fmt.Fprintf(w, "[%d] %s - OrderID: %d, Timestamp: %s\n",
				i+1, event.GetType(), event.GetOrderID(), event.GetTimestamp())
		}
	}).Methods("GET")

	// Запуск сервера
	log.Println("CQRS сервер запущен на http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
