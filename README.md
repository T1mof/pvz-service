# PVZ Service

Сервис для управления пунктами выдачи заказов (ПВЗ), приёмками и товарами.

## Особенности

- RESTful API для управления ПВЗ, приёмками и товарами
- gRPC API для получения списка ПВЗ
- Метрики Prometheus для мониторинга
- Аутентификация с использованием JWT
- Интеграция с PostgreSQL

## Требования

- Docker 20.10+
- Docker Compose 2.0+
- Make (опционально, для удобства использования Makefile)

## Быстрый старт

1. Клонируйте репозиторий
```bash
git clone https://github.com/T1mof/pvz-service.git
cd pvz-service
```

2. Запустите сервисы с помощью Docker Compose
```bash
docker-compose up -d
```
Или, если у вас установлен Make:
```bash
make dev
```

3. Проверьте статус сервисов
```bash
docker-compose ps
```
Или:
```bash
make service-status
```

4. Доступ к сервисам:
- HTTP API: [http://localhost:8080](http://localhost:8080)
- gRPC сервер: localhost:3000
- Метрики Prometheus: [http://localhost:9000/metrics](http://localhost:9000/metrics)
- Prometheus UI: [http://localhost:9090](http://localhost:9090)

## Работа с API

### HTTP API

Базовые эндпоинты:

- `POST /auth/register` - Регистрация нового пользователя
- `POST /auth/login` - Авторизация и получение JWT токена
- `POST /pvz` - Создание нового ПВЗ
- `GET /pvz` - Получение списка ПВЗ
- `GET /pvz/{id}` - Получение информации о конкретном ПВЗ
- `POST /receptions` - Создание новой приёмки
- `PUT /pvz/{pvzId}/close-reception` - Закрытие последней приёмки ПВЗ
- `POST /products` - Добавление нового товара
- `DELETE /products/{pvzId}/last` - Удаление последнего товара из приёмки ПВЗ

Для аутентификации используйте заголовок `Authorization: Bearer <token>`.

### gRPC API

Для доступа к gRPC API можно использовать инструмент grpcurl:
```bash
grpcurl -plaintext localhost:3000 pvz.PVZService/ListPVZ
```

## Метрики

Сервис собирает следующие метрики:

### Технические метрики:
- `http_requests_total` - Общее количество HTTP запросов
- `http_request_duration_seconds` - Время выполнения HTTP запросов

### Бизнес-метрики:
- `pvz_created_total` - Количество созданных ПВЗ
- `receptions_created_total` - Количество созданных приёмок
- `products_added_total` - Количество добавленных товаров

Метрики доступны по эндпоинту `/metrics` на порту 9000 и могут быть визуализированы в инструменте Prometheus.

## Работа с базой данных

### Структура базы данных

База данных автоматически инициализируется при первом запуске, применяя миграции из директории `migrations`. Структура включает:

- `users` - таблица пользователей
- `pvz` - таблица пунктов выдачи заказов
- `receptions` - таблица приёмок
- `products` - таблица товаров

### Подключение напрямую к БД
```bash
make db-shell
```

### Сброс базы данных

Если вам нужно сбросить базу данных и начать с чистого состояния:
```bash
make db-reset
```

## Разработка

### Структура проекта

```bash
pvz-service/
├── cmd/
│ └── api/
│ └── main.go # Точка входа приложения
├── internal/
│ ├── api/ # Обработчики HTTP API
│ ├── config/ # Конфигурация
│ ├── domain/ # Бизнес-модели и интерфейсы
│ ├── grpc/ # gRPC сервер
│ ├── logger/ # Логирование
│ ├── metrics/ # Метрики Prometheus
│ ├── repository/ # Репозитории для работы с БД
│ └── services/ # Сервисный слой
├── migrations/ # SQL миграции
├── scripts/ # Скрипты для контейнеров
├── proto/ # Protobuf определения
├── config/ # Конфигурационные файлы
├── Dockerfile # Инструкция для сборки образа
├── docker-compose.yml # Конфигурация для Docker Compose
├── prometheus.yml # Конфигурация Prometheus
├── Makefile # Команды для упрощения разработки
└── README.md # Документация
```

### Переменные окружения

| Переменная     | Описание                       | Значение по умолчанию |
|----------------|--------------------------------|-----------------------|
| DB_HOST        | Хост базы данных               | db                    |
| DB_PORT        | Порт базы данных               | 5432                  |
| DB_NAME        | Имя базы данных                | pvz                   |
| DB_USER        | Пользователь БД                | postgres              |
| DB_PASSWORD    | Пароль пользователя БД         | postgres              |
| JWT_SECRET     | Секретный ключ для JWT         | your_jwt_secret_key   |
| ENVIRONMENT    | Окружение (dev/prod)           | development           |

## Тестирование

Запуск всех тестов:
```bash
make test
```

## Частые проблемы и их решения

### Проблемы с подключением к БД

Убедитесь, что контейнер БД запущен и доступен:
```bash
docker-compose ps
```

Проверьте переменные окружения, связанные с подключением к БД.

### Проблемы с миграциями

Если миграции не применяются автоматически, вы можете выполнить их вручную:
```bash
make db-migrate
```

Если возникают ошибки с миграциями, проверьте логи:
```bash
make logs
```

### Проблемы с запуском контейнеров

Если контейнеры не запускаются должным образом, проверьте логи:
```bash
make logs
```

Возможно, вам потребуется обновить образы:
```bash
docker-compose pull
make build
```
