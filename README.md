# <a href="https://github.com/Maxim-Belyi/PriceTracker"> PriceTracker: Асинхронный парсер цен на Go </a>

Микросервисный проект для фонового отслеживания цен на товары. Демонстрирует паттерн асинхронного взаимодействия между сервисами на базе очередей сообщений (RabbitMQ).

## 🚀 О проекте

Приложение демонстрирует концепции языка Go и построения распределенных систем: работу с брокерами сообщений (Publisher/Consumer), ручное управление подтверждениями (Ack/Nack), написание чистых SQL-запросов через `pgx` и микросервисную архитектуру. 

Состоит из двух независимых компонентов:
1.  **API Gateway:** HTTP-сервер, который принимает REST-запросы от клиентов, сохраняет начальное состояние в базу данных и публикует задачу (сообщение) в очередь RabbitMQ. Не заставляет клиента ждать парсинга.
2.  **Parser Worker:** Фоновый демон без HTTP-интерфейса. Непрерывно слушает очередь RabbitMQ, забирает ссылки на товары, "парсит" их цены (с эмуляцией задержек), обновляет информацию в PostgreSQL и отправляет подтверждение брокеру (ACK).

## 🛠️ Стек технологий
<div> 
<img src="https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"/> 
<img src="https://img.shields.io/badge/RabbitMQ-FF6600?style=flat&logo=rabbitmq&logoColor=white" alt="RabbitMQ"/>
<img src="https://img.shields.io/badge/PostgreSQL-336791?style=flat&logo=postgresql&logoColor=white" alt="PostgreSQL"/> 
<img src="https://img.shields.io/badge/pgx-blue?style=flat&logo=go&logoColor=white" alt="pgx"/> 
<img src="https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white" alt="Docker"/> 
</div>

## ⚙️ Как запустить локально

Чтобы развернуть проект у себя, выполните следующие шаги:

### Необходимые компоненты
*   [Go](https://golang.org/dl/) (версия 1.22+)
*   [Docker и Docker Compose](https://www.docker.com/) (для запуска инфраструктуры БД и RabbitMQ)

### Установка и запуск

1.  **Клонируйте репозиторий:**
    ```sh
    git clone https://github.com/Maxim-Belyi/PriceTracker.git
    cd PriceTracker
    ```

2.  **Установите зависимости:**
    ```sh
    go mod tidy
    cd api && go mod tidy
    cd ../worker && go mod tidy
    ```
3.  **Поднимите инфраструктуру (PostgreSQL + RabbitMQ):**
    Для запуска базы данных и брокера сообщений используется Docker Compose.
    ```sh
    docker-compose up -d postgres rabbitmq
    ```
    *Web-панель RabbitMQ будет доступна по адресу `http://localhost:15672` (guest/guest).*

4.  **Создайте таблицу в БД:**
    Подключитесь к `localhost:5432` (пользователь `admin`, пароль `qwerty`, БД `pricetracker`) и выполните SQL скрипт:
    ```sql
    CREATE TABLE IF NOT EXISTS items (
        id SERIAL PRIMARY KEY,
        url TEXT NOT NULL,
        current_price NUMERIC(10, 2) DEFAULT 0.00,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    ```

5.  **Запустите микросервисы (в разных терминалах):**
    
    Сначала запускаем фоновый слушатель:
    ```sh
    go run ./worker/main.go
    ```
    Затем запускаем публичный API (порт 8080):
    ```sh
    go run ./api/main.go
    ```

## 📝 Как управлять парсингом (REST API)

*   **Добавление товара в очередь:** 
    Отправьте POST запрос с URL товара. API сохранит его со статусом `pending`, положит в RabbitMQ и мгновенно вернет ID. Воркер на фоне обработает ссылку.
    ```bash
    curl -X POST http://localhost:8080/track \
    -H "Content-Type: application/json" \
    -d '{"url": "https://mvideo.ru/iphone-15"}'
    ```

## 🌐 Публикация и архитектура

В проекте предусмотрен `docker-compose.yml` для полной оркестрации. Помимо инфраструктуры, можно собрать Docker-образы для самих Go-сервисов (API и Worker), чтобы разворачивать проект одной кнопкой.

---

**Примечание:** Это учебный проект. Логика парсинга HTML-страниц заменена заглушкой (генератором случайных чисел и задержкой), чтобы сфокусироваться на архитектуре очередей и взаимодействии компонентов.