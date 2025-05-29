package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// ProfileData представляет данные пользовательского профиля
type ProfileData struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	// Клиент Redis для работы с кешем
	rdb *redis.Client
	// Контекст для Redis операций
	ctx = context.Background()
)

func main() {
	// Инициализация подключения к Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Адрес Redis сервера
		Password: "",               // Пароль (если требуется)
		DB:       0,                // Используемая база данных Redis
	})

	// Проверка соединения с Redis
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	log.Println("Успешное подключение к Redis")

	// Настройка маршрутизатора HTTP
	r := mux.NewRouter()

	// Маршрут для получения профиля пользователя
	r.HandleFunc("/profile/{id}", getProfileHandler).Methods("GET")

	// Запуск HTTP сервера
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// getProfileHandler обрабатывает запросы на получение профиля пользователя
func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	// Формируем ключ для кеша
	cacheKey := fmt.Sprintf("user:%d", id)

	// Пытаемся получить данные из кеша Redis
	val, err := rdb.Get(ctx, cacheKey).Result()

	// Если данные найдены в кеше
	if err == nil {
		log.Printf("Данные для пользователя %d получены из кеша", id)

		// Отправляем клиенту данные из кеша
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		fmt.Fprint(w, val)
		return
	}

	// Если данных нет в кеше или произошла ошибка (не redis.Nil)
	if err != redis.Nil {
		log.Printf("Ошибка при обращении к Redis: %v", err)
	}

	// Получаем данные из "медленного источника"
	log.Printf("Данные для пользователя %d не найдены в кеше, получаем из источника", id)

	// Имитация задержки "медленного источника" данных
	time.Sleep(2 * time.Second)

	// Создаем профиль пользователя (в реальном приложении - из БД)
	profile := ProfileData{
		ID:        id,
		Name:      fmt.Sprintf("Пользователь %d", id),
		Email:     fmt.Sprintf("user%d@example.com", id),
		CreatedAt: time.Now().Add(-time.Duration(id) * 24 * time.Hour),
	}

	// Сериализуем профиль в JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, "Ошибка сериализации данных", http.StatusInternalServerError)
		return
	}

	// Сохраняем данные в кеш на 10 минут
	err = rdb.Set(ctx, cacheKey, profileJSON, 10*time.Minute).Err()
	if err != nil {
		log.Printf("Ошибка при сохранении в кеш: %v", err)
	} else {
		log.Printf("Данные для пользователя %d сохранены в кеш на 10 минут", id)
	}

	// Отправляем ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(profileJSON)
}
