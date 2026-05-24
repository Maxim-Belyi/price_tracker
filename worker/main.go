package main

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Task struct {
	Id  int `json:"id"`
	Url string `json:"url"`
}

func main() {
	dsn := "postgres://admin:qwerty@localhost:5432/pricetracker"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к бд: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("БД не отвечает: %v", err)
	}
	log.Println("Успешное подключение к бд")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Не удалось подключиться к RabbitMq: %v", err)
	}
	defer conn.Close()
	log.Println("Успешное подключение к RabbitMq!")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Не удалось открыть канал, %v", err)
	}
	defer ch.Close()
	log.Printf("Успешно подключились к каналу!")

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

	msgs, err := ch.Consume(
		"parsing_tasks",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Ошбибка регистрации консумера: %v", err)
	}

	forever := make(chan struct{})

	go func() {
		for msg := range msgs {
			log.Printf("Получено сообщение: %s", msg.Body)

			t:= Task {
				Id: id,
				Url: msg.Url,
			}

			if err := json.NewDecoder(msg.Body).Decode(t); err != nil {
				log.Printf("Некорректный Json!")
				return
			}


		}
	}





}