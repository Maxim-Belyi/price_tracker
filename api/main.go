package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TrackRequest struct {
	Url string `json:"url"`
}

type TrackResponse struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}

type Task struct {
	Id  int
	Url string
}

func main() {
	dsn := "postgres://admin:qwerty@localhost:5432/pricetracker"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к бд: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("База не отвечает: %v", err)
	}
	log.Println("Успешное подключение к бд")

	conn, err := amqp.Dial("amqp://quest:quest@localhost:5672/")
	if err != nil {
		log.Fatalf("Не удалось подключиться к RabbitMq: %v", err)
	}
	defer conn.Close()
	log.Println("Успешное подключение к RabbitMq!")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Не удалось открыть канал: %v", err)
	}
	defer ch.Close()
	log.Println("Успешно подключились к каналу!")

	q, err := ch.QueueDeclare(
		"parsing_tasks",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Не удалось объявить очередь: %v", err)
	}

	log.Printf("Очередь объявлена! Имя: %s, Сообщений: %d", q.Name, q.Messages)

	http.HandleFunc("/track", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
			return
		}

		var req TrackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}

		var id int
		query := (`INSERT INTO items (url) VALUES ($1) RETURNING id`)
		if err := db.QueryRow(query, req.Url).Scan(&id); err != nil {
			http.Error(w, "Ошибка", http.StatusInternalServerError)
			return
		}

		t := Task{
			Id:  id,
			Url: req.Url,
		}

		bodyBytes, err := json.Marshal(t)
		if err != nil {
			http.Error(w, "Не удалось преобразовать структуру", http.StatusInternalServerError)
			return
		}

		err = ch.PublishWithContext(
			r.Context(),
			"",
			"parsing_tasks",
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        bodyBytes,
			})
		if err != nil {
			http.Error(w, "Ошибка публикации", http.StatusInternalServerError)
			return
		}

		log.Println("Сообщение отправлено!")

		res := TrackResponse{
			Id:     id,
			Status: "Сохранено!",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
		}

	})

	log.Println("сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
