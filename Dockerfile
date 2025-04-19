FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/pvz-api ./cmd/api/main.go

# Создаем финальный образ
FROM alpine:3.19

WORKDIR /app

# Копируем исполняемый файл
COPY --from=builder /app/pvz-api .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/configs ./configs

# Запускаем приложение
CMD ["/app/pvz-api"]
