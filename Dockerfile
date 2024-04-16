# Определение базового образа для сборки приложения
FROM golang:1.22.2-alpine3.19 AS builder

# Копирование исходного кода в рабочую директорию контейнера
COPY . $GOPATH/src/app

# Установка рабочей директории
WORKDIR $GOPATH/src/app

# Получение зависимостей
RUN go get -d -v

# Сборка приложения
RUN GO111MODULE=on CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CGO_CFLAGS="-g -O2 -Wno-return-local-addr" go build -o $GOPATH/src/app/app.bin

# Определение конечного образа для работы приложения
FROM alpine

# Установка переменной окружения TZ для установки часового пояса
ENV TZ Europe/Moscow

# Копирование исполняемого файла приложения из образа сборки в конечный образ
COPY --from=builder /go/src/app/app.bin /app.bin

# Открытие портов для взаимодействия с приложением
EXPOSE 8443/tcp
EXPOSE 1812/udp

# Определение точки входа для выполнения приложения при запуске контейнера
CMD ["/app.bin"]
