package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Task struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://admin:qwerty@localhost:5432/pricetracker"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к бд: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("БД не отвечает: %v", err)
	}
	log.Println("Успешное подключение к бд")

	rmqUrl := os.Getenv("RMQ_URL")
	if rmqUrl == "" {
		rmqUrl = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(rmqUrl)
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

			var t Task

			if err := json.Unmarshal(msg.Body, &t); err != nil {
				log.Printf("Ошибка декодирования Json: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("Начинаю парсинг для URL: %s", t.Url)
			time.Sleep(2 * time.Second)
			price := float64(rand.Intn(1000)) + 100.00

			query := `
			UPDATE items SET current_price = $1, 
			status = 'processed',
			updated_at = CURRENT_TIMESTAMP WHERE id = $2
			`

			if _, err := db.Exec(query, price, t.Id); err != nil {
				log.Printf("Ошибка обновления БД: %v", err)
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
			log.Printf("Успешно! Товар ID %d получил цену %.2f", t.Id, price)

		}
	}()
	log.Println("Worker запущен! Ожидание сообщений...")
	<-forever
}
