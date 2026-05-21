package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"

)

type TrackRequest struct {
	Url string `json:"url"`
}

func main() {
	dsn := "postgres://admin:qwerty@localhost:5432/pricetracker"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к бд: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("База не отвечает: %v", err)
	}
	log.Println("Успешное подключение к бд")

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
		}
	})

	log.Println("сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
