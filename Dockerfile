# ---- Build Stage ----
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# ---- Run Stage ----
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/server .
COPY templates ./templates
COPY static ./static

EXPOSE 8080

CMD ["./server"]
