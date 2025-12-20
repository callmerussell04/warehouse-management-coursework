FROM golang:alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o warehouse_management_systsem ./cmd/main.go

FROM alpine:latest

WORKDIR /app

RUN addgroup -S app && adduser -S -G app app

COPY --from=builder --chown=app:app /app/warehouse_management_systsem .
COPY --from=builder --chown=app:app /app/api/openapi.yml ./api/openapi.yml
COPY --from=builder --chown=app:app /app/assets/arial.ttf ./assets/arial.ttf

EXPOSE 8080

USER app

CMD ["./warehouse_management_systsem"]
