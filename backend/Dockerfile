FROM golang:alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o warehouse_management_systsem ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/warehouse_management_systsem .
COPY --from=builder /app/api/openapi.yml ./api/openapi.yml

EXPOSE 8080

CMD ["./warehouse_management_systsem"]