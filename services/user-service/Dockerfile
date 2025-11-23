FROM golang:1.25-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN apk add --no-cache git && go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/src/app/bin/user-service services/user-service/cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /usr/src/app/bin/user-service .

CMD ["./user-service"]