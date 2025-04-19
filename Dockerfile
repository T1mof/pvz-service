# Этап сборки
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка необходимых инструментов
RUN apk add --no-cache git make gcc libc-dev

# Копирование и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pvz-service ./cmd/api/main.go

# Финальный образ
FROM alpine:3.19

WORKDIR /app

# Устанавливаем зависимости runtime
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/pvz-service .

# Копируем миграции и конфигурационные файлы
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config/default.yaml ./config/default.yaml

# Копируем скрипты инициализации
COPY --from=builder /app/scripts/wait-for-it.sh /app/scripts/init-db.sh ./scripts/
RUN chmod +x ./scripts/wait-for-it.sh ./scripts/init-db.sh

# Открываем необходимые порты
EXPOSE 8080 3000 9000

# Запускаем приложение
CMD ["/app/pvz-service"]