# Этап 1: Сборка Go-приложения
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем Go-файлы
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем исполняемый файл
RUN go build -o server main.go

# Этап 2: Финальный образ
FROM debian:bookworm-slim

WORKDIR /app

# 🛠 Установка сертификатов
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Копируем бинарник
COPY --from=builder /app/server .

# Копируем статические файлы (если они нужны)
COPY static/ ./static

EXPOSE 8080

CMD ["./server"]

