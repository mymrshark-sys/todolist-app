# ---- Build Stage ----
FROM golang:1.22 AS builder

WORKDIR /app

# copy mod files first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy all source
COPY . .

# build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go


# ---- Run Stage ----
FROM alpine:latest

WORKDIR /app

# copy binary
COPY --from=builder /app/server .

# copy templates + static
COPY templates ./templates
COPY static ./static

# set env port (Railway sets automatically)
ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
