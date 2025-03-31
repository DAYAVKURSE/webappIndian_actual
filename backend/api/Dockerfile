FROM golang:1.20-alpine AS builder

WORKDIR /build

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY ["go.sum", "go.mod", "./"]

# Загрузка зависимостей
RUN --mount=type=cache,target=/gomod-cache go mod download

# Копируем исходники
COPY . .

# Приводим go.mod/go.sum в порядок
RUN go mod tidy

# Сборка
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    go build -o api cmd/app/main.go

# Финальный образ
FROM alpine

RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app
COPY --from=builder /build/api .

CMD ["./api"]
