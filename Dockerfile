############################
# STEP 1: build executable binary
############################
FROM golang:1.24.1-alpine3.20 AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git tzdata

# Настройка окружения Go
ENV GO111MODULE=on
WORKDIR /app
COPY . .

RUN go mod download
RUN go mod verify

# Build with -v so failures show which package failed
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -installsuffix cgo -o main .

############################
# STEP 2: minimal runtime image
############################
FROM alpine:3.20

WORKDIR /go

COPY --from=builder /app/main /go/main

ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

# HEALTHCHECK --timeout=1s --start-period=10s --retries=3 --interval=30s \
#     CMD wget -nv -t1 --spider 'http://localhost:8080/health/check' || exit 1

ENTRYPOINT ["/go/main"]
