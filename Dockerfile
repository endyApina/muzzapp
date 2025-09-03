FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o muzzapp ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/muzzapp .
COPY .env ./

EXPOSE 50051
CMD ["./muzzapp"]
