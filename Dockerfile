# Используем официальный образ Golang как базовый
FROM golang:1.23.3 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum перед копированием исходного кода для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем файл .env
COPY .env ./

# Копируем остальные файлы исходного кода
COPY . .

# Устанавливаем переменные окружения для сборки
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Собираем приложение, указывая путь к исполняемому файлу
RUN go build -o user-task-reward-controller ./cmd/main.go

# Изображение для выполнения
FROM alpine:3.17

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем скомпилированное приложение из предыдущего этапа сборки
COPY --from=builder /app/user-task-reward-controller .

# Убираем файл .env из финального образа, если он не нужен на этапе выполнения
COPY --from=builder /app/.env .

# Обеспечиваем наличие прав на выполнение
RUN chmod +x ./user-task-reward-controller

# Проверяем, что файл есть и он исполняемый
RUN ls -l ./user-task-reward-controller

# Указываем команду по умолчанию для запуска приложения
CMD ["./user-task-reward-controller"]